package controllers

import (
	"github.com/revel/revel"
//	"github.com/revel/revel/cache"
	"auth/app/models"
	"auth/app/utils"
	"auth/app/emails"
	"encoding/json"
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/jinzhu/gorm"
	"fmt"
	"time"
)

var _ = fmt.Printf

//this controller handle login/logout/session, etc...
type Auth struct {
	BaseController
	GormController
}

//auth api
func (c Auth) Register() revel.Result {
	//unmarshal the request
	newUser := models.User{}

	if err := json.NewDecoder(c.Request.Body).Decode(&newUser); err != nil {
		c.RenderJsonError("Invalid post data. It is not in JSON format.")
	}

	//validate the post data
	newUser.Validate(c.Validation)

	// Handle errors
	if c.Validation.HasErrors() {
		return c.RenderJson(c.Validation.Errors)
	}

	//check if there is user with same username or email
	var existingUsers []models.User;

	if err := c.Txn.Where("email= ?", newUser.Email).Or("user_name= ?", newUser.UserName).Find(&existingUsers).Error; err != nil {
		revel.INFO.Printf("error %s", err)
		return c.RenderJsonError("Database Error")
	}

	if len(existingUsers) == 0 {
		//hash the password
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newUser.Password), 10)

		//assign hash, encrypted password for the new users
		newUser.HashedPassword = hashedPassword
		newUser.Password = ""

		//user has inactive status first
		newUser.Status = models.USER_INACTIVE
		
		//create new user
		c.Txn.NewRecord(newUser)
		c.Txn.Create(&newUser)

		//now trying to send email of activation link to him
		accountActivation := models.AccountActivation{}
		c.Txn.NewRecord(accountActivation)
		accountActivation.ActivationKey = utils.RandSeq(20)
		accountActivation.UserId = newUser.Id;
		accountActivation.ExpiryDate = time.Now().AddDate(0,0,3);
		
		if err := c.Txn.Create(&accountActivation).Error; err != nil {
			return c.RenderJsonError("Database Error")
		}

		emailInfo := emails.EmailInfo{
			ToEmail: newUser.Email,
			ToName: newUser.UserName,
			Subject: "Account Activation",
			FromEmail: "noreply@auth.com",
			FromName: "Auth Team",
		}

		accountActivationLink :=  WebURL + "/accountActivation/" + accountActivation.ActivationKey;
		
		//activation key
		emailPlaceHolder := emails.EmailPlaceHolder{
			URL: accountActivationLink,
			UserName: newUser.UserName,
		}

		//now sending email TODO: put email into queue to retry later
		//put it into different go routine to prevent block
		go func(){
			err := emails.Send(emails.AccountActivation, emailInfo, emailPlaceHolder)
			if err != nil {
				revel.ERROR.Printf("error happen when sending email %s", err)
			}
		}()
		
		//sanitize the new user
		return c.RenderJsonSuccess(newUser.Sanitize())

	} else {
		return c.RenderJsonError("Username or Email has been taken.")
	}
}

func (c Auth) Login() revel.Result {
	var sessionKey string
	sessionKey = "s:user_"+c.Session.Id()

	//if session is found; then return immediately
	var sessionUser models.User
	RCache.Get(sessionKey, &sessionUser)

	//TODO: how to check if result is found
	if sessionUser.Email != "" {
		return c.RenderJsonSuccess(sessionUser)
	}

	//prefer to use model + json decode because it's enable mapping case insensitive
	var requestUser = models.User{}

	if err := json.NewDecoder(c.Request.Body).Decode(&requestUser); err != nil {
		return c.RenderJsonError("Invalid post data. It is not in proper JSON format")
	}

	//verify password and email exist in post data
	c.Validation.Required(requestUser.Email)
	c.Validation.Required(requestUser.Password)

	if c.Validation.HasErrors() {
		return c.RenderJsonError("Missing required parameters email or password");
	}

	//otherwise check again database to find the user
	user := models.User{}
	if Gdb.Where("email = ?", requestUser.Email).First(&user).RecordNotFound() {
		return c.RenderJsonError("User with email " + user.Email + " does not exist");
	}

	//make sure that user is active
	if user.Status != models.USER_ACTIVE {
		return c.RenderJsonError("User was not activated. If you are new user, please check your email for activation link");
	}
	
	//compare password to validate the user
	if err := bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(requestUser.Password)); err != nil {
		return c.RenderJsonError("Unable to login. Passwords are miss-match");
	}

	//otherwise, set session in redis
	RCache.Set(sessionKey, user, SessionExpire)
	
	return c.RenderJsonSuccess(user)
}

//logout
func (c Auth) Logout() revel.Result {
	var sessionKey string
	sessionKey = "s:user_"+c.Session.Id()
	
	RCache.Delete(sessionKey)

	var response struct{}
	return c.RenderJsonSuccess(response)
}

//activate account when receive activation link
func (c Auth) ActivateAccount() revel.Result {
	var activationKey = c.Params.Get("activationKey");
	
	//check if activation key exist
	var accountActivation models.AccountActivation
	if err := Gdb.Where("activation_key=?", activationKey).First(&accountActivation).Error; err != nil {
		if err == gorm.RecordNotFound {
			return c.RenderJsonError("Invalid activation key");
		} else {
			return c.RenderJsonError("Internal Database error");
		}
	}
	
	//key is already expired
	if accountActivation.ExpiryDate.Before(time.Now()) {
		return c.RenderJsonError("Activation key has already been expired")
	}
	
	//if there is activation key; then activate the user
    var userId = accountActivation.UserId
    var existingUser = models.User{}

	if err := Gdb.Where("id=?", userId).First(&existingUser).Error; err != nil {
        return c.RenderJsonError("User does not exist")
	}
	
    existingUser.Status = models.USER_ACTIVE;
	
    if err := Gdb.Save(&existingUser).Error; err != nil {
		return c.RenderJsonError("Internal Database error");
	}
	
	var response struct{}
    return c.RenderJsonSuccess(response)
}

//request password reset being sent
func (c Auth) RequestPasswordReset() revel.Result {
	var email = c.Params.Get("email")
	
	var user = models.User{}
	//check user withe email exist
	if Gdb.Where("email = ?", email).First(&user).RecordNotFound() {
		return c.RenderJsonError("User with email " + email + " does not exist");
	}

	//now trying to send email of activation link to him
	passwordReset := models.PasswordReset{}
	c.Txn.NewRecord(passwordReset)
	passwordReset.PasswordResetKey = utils.RandSeq(20)
	passwordReset.UserId = user.Id;
	passwordReset.ExpiryDate = time.Now().AddDate(0,0,3);

	if err := c.Txn.Create(&passwordReset).Error; err != nil {
		return c.RenderJsonError("Database Error")
	}

	emailInfo := emails.EmailInfo{
		ToEmail: user.Email,
		ToName: user.UserName,
		Subject: "Password Reset",
		FromEmail: "noreply@auth.com",
		FromName: "Auth Team",
	}

	passwordResetLink :=  WebURL + "/login.html#passwordReset/" + passwordReset.PasswordResetKey;

	//activation key
	emailPlaceHolder := emails.EmailPlaceHolder{
		URL: passwordResetLink,
		UserName: user.UserName,
	}

	//now sending email TODO: put email into queue to retry later
	//put it into different go routine to prevent block
	go func(){
		err := emails.Send(emails.PasswordReset, emailInfo, emailPlaceHolder)
		if err != nil {
			revel.ERROR.Printf("error happen when sending email %s", err)
		}
	}()
	
	var response struct{}
	return c.RenderJsonSuccess(response)
}

//actual reset password (update old password)
func (c Auth) ResetPassword() revel.Result {
	//password reset key
	var passwordResetKey = c.Params.Get("passwordResetKey")
	var newPassword	= c.Params.Get("newPassword")

	//check if activation key exist
	var passwordReset models.PasswordReset
	if err := Gdb.Where("password_reset_key=?", passwordResetKey).First(&passwordReset).Error; err != nil {
		if err == gorm.RecordNotFound {
			return c.RenderJsonError("Invalid password reset key");
		} else {
			return c.RenderJsonError("Internal Database error");
		}
	}

	//key is already expired
	if passwordReset.ExpiryDate.Before(time.Now()) {
		return c.RenderJsonError("Activation key has already been expired")
	}

	//if there is activation key; then activate the user
	var userId = passwordReset.UserId
	var existingUser = models.User{}

	if err := Gdb.Where("id=?", userId).First(&existingUser).Error; err != nil {
		return c.RenderJsonError("User does not exist")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newPassword), 10)

	//assign hash, encrypted password for the new users
	existingUser.HashedPassword = hashedPassword
	existingUser.Password = ""


	if err := Gdb.Save(&existingUser); err != nil {
		return c.RenderJsonError("Internal Database error");
	}

	var response struct{}
	return c.RenderJsonSuccess(response)
}

//update existing userpassword
func (c Auth) ChangePassword() revel.Result {
	var oldPassword = c.Params.Get("oldPassword")
	var newPassword	= c.Params.Get("newPassword")
	
	//get user from session
	var sessionKey string
	sessionKey = "s:user_"+c.Session.Id()

	//if session is found; then return immediately
	var sessionUser models.User
	RCache.Get(sessionKey, &sessionUser)
	
	if sessionUser.Id == 0 {
		c.RenderJsonError("User must be logged in first")
	}
	
	//get user from database
	var userId = sessionUser.Id
	var existingUser = models.User{}

	//find active user
	if err := Gdb.Where("id=?", userId).Where("status=?", models.USER_ACTIVE).First(&existingUser).Error; err != nil {
		return c.RenderJsonError("User does not exist")
	}

	if bcrypt.CompareHashAndPassword(existingUser.HashedPassword, []byte(oldPassword)) != nil {
		return c.RenderJsonError("Please fill in  your existing password correctly")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newPassword), 10)

	//assign hash, encrypted password for the new users
	existingUser.HashedPassword = hashedPassword
	existingUser.Password = ""


	if err := Gdb.Save(&existingUser).Error; err != nil {
		return c.RenderJsonError("Internal Database error");
	}

	var response struct{}
	return c.RenderJsonSuccess(response)
}

