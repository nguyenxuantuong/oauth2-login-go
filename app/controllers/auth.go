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
	"github.com/parnurzeal/gorequest"
//	"auth/app/routes"
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
		return c.RenderJsonError(c.Validation.Errors)
	}

	//check if there is user with same username or email
	var existingUsers []models.User;

	if err := c.Txn.Where("email= ?", newUser.Email).Or("user_name= ?", newUser.UserName).Find(&existingUsers).Error; err != nil {
		revel.INFO.Printf("error %s", err)
		return c.RenderJsonError("Internal Server Error")
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
			return c.RenderJsonError("Internal Server Error")
		}

		emailInfo := emails.EmailInfo{
			ToEmail: newUser.Email,
			ToName: newUser.UserName,
			Subject: "Account Activation",
			FromEmail: "noreply@auth.com",
			FromName: "Auth Team",
		}

		accountActivationLink :=  WebURL + "/activation/" + accountActivation.ActivationKey;
		
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
			} else {
				revel.INFO.Println("Sending activation email successfully")
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
//	var sessionUser models.User
//	RCache.Get(sessionKey, &sessionUser)
//
//	//TODO: how to check if result is found
//	if sessionUser.Email != "" {
//		return c.RenderJsonSuccess(sessionUser)
//	}

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
		return c.RenderJsonError("Unable to login. Passwords are miss-matched");
	}

	//otherwise, set session in redis
	RCache.Set(sessionKey, user, SessionExpire)
	
	//otherwise; just return response as usual
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

	if email == "" {
		return c.RenderJsonError("Please fill in the email field");
	}

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

	passwordResetLink :=  WebURL + "/resetPassword/" + passwordReset.PasswordResetKey;

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
		} else {
			revel.INFO.Printf("Password reset has been sent to email %s", email)
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

	if newPassword == "" {
		return c.RenderJsonError("Missing new password")
	}

	revel.INFO.Println("Req params", passwordResetKey, newPassword)

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

//	assign hash, encrypted password for the new users
	existingUser.HashedPassword = hashedPassword
	existingUser.Password = ""

	revel.INFO.Println("Existing User", existingUser)

	if err := Gdb.Save(&existingUser).Error; err != nil {
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
		return c.RenderJsonError("Please fill in your existing password correctly")
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

//get user info from session
func (c Auth) UserInfo() revel.Result {
	var sessionKey string
	sessionKey = "s:user_" + c.Session.Id()
	
	var sessionUser models.User
	RCache.Get(sessionKey, &sessionUser)

	if sessionUser.Id == 0 {
		c.RenderJsonError("User must be logged in first")
	}

	//get latest user info from database
	var userId = sessionUser.Id
	var existingUser = models.User{}

	//find active user
	if err := Gdb.Where("id=?", userId).Where("status=?", models.USER_ACTIVE).First(&existingUser).Error; err != nil {
		return c.RenderJsonError("User does not exist")
	}

	//TODO: might want to include other information here
	return c.RenderJsonSuccess(existingUser.Sanitize())
}

//function to validate the access tokens
func ValidateFacebookAccessToken(fbId string, accessToken string) bool {
	url := "https://graph.facebook.com/me?access_token=" + accessToken;
	_, body, _ := gorequest.New().Get(url).End()

	//unmarshal arbitrary data
	var f interface{}
	err := json.Unmarshal([]byte(body), &f)
	
	if err != nil {
		revel.ERROR.Println("unable to unmarshal json data accesstoken")
		return false
	}

	m := f.(map[string]interface{})
	
	if m["id"] == fbId {
		return true
	} else {
	 	return false
	}
}

//validate access tokens
func ValidateGoogleAccessToken(googleId string, accessToken string) bool {
	url := "https://www.googleapis.com/oauth2/v1/userinfo?access_token=" + accessToken;
	_, body, _ := gorequest.New().Get(url).End()

	var f interface{}
	err := json.Unmarshal([]byte(body), &f)

	if err != nil {
		revel.ERROR.Println("unable to unmarshal json data accesstoken")
		return false
	}

	m := f.(map[string]interface{})
	
	if m["id"] == googleId {
		return true
	} else {
		return false
	}
}

//register using facebook
func (c Auth) RegisterUsingFacebook() revel.Result {
	email := c.Params.Get("email")
	name := c.Params.Get("name")
	fbId := c.Params.Get("fb_id")
	accessTokens := c.Params.Get("access_token")

	//Note: For now, email id is required;
	//TODO: check to see if can ignore email; note that sometimes email not being returned from Oath2
	c.Validation.Check(email, revel.Required{})
	c.Validation.Check(name, revel.Required{})
	c.Validation.Check(fbId, revel.Required{})
	c.Validation.Check(accessTokens, revel.Required{})

	// Handle errors
	if c.Validation.HasErrors() {
		return c.RenderJson(c.Validation.Errors)
	}
	
	//validate facebook access token
	if(!ValidateFacebookAccessToken(fbId, accessTokens)){
		return c.RenderJsonError("Invalid facebook access tokens")
	}
	
	//now check if user with fbId exist
	var users []models.User
	if err := Gdb.Where("fb_id= ?", fbId).Find(&users).Error; err != nil {
		return c.RenderJsonError("Internal Server Error")
	}

	if len(users) >= 2 {
		return c.RenderJsonError("There are two account with the same fbId")		
	}
	
	var sessionKey string
	sessionKey = "s:user_"+c.Session.Id()
	
	if len(users) == 1 {
		if users[0].Status == models.USER_ACTIVE {
			//logged user in directly
			RCache.Set(sessionKey, users[0], SessionExpire)
			return c.RenderJsonSuccess(users[0])
		} else {
			return c.RenderJsonError("User was not activated")
		}
	}
	
	//if there is no such user with fbId; check email and register user
	users = []models.User{}
	if err := Gdb.Where("email= ?", email).Find(&users).Error; err != nil {
		return c.RenderJsonError("Internal Server Error")
	}
	
	if len(users) >= 2 {
		return c.RenderJsonError("There are two account with the same email id")
	}

	if len(users) == 1 {
		//update fbId field of that account
		existingUser := users[0]
		existingUser.FbId = fbId;
		
		Gdb.Save(&existingUser)
		
		RCache.Set(sessionKey, existingUser, SessionExpire)
		return c.RenderJsonSuccess(existingUser)
	}
	
	//otherwise just create new users
	newUser := models.User{
		FullName: name,
		Email: email,
		UserName: email,
		Password: "",
		Status: models.USER_ACTIVE,
		FbId: fbId,
	}

	Gdb.Save(&newUser)

	RCache.Set(sessionKey, newUser, SessionExpire)
	return c.RenderJsonSuccess(newUser)
}

//register using google
func (c Auth) RegisterUsingGoogle() revel.Result {
	email := c.Params.Get("email")
	name := c.Params.Get("name")
	googleId := c.Params.Get("google_id")
	accessTokens := c.Params.Get("access_token")

	//Note: For now, email id is required;
	//TODO: check to see if can ignore email; note that sometimes email not being returned from Oath2
	c.Validation.Check(email, revel.Required{})
	c.Validation.Check(name, revel.Required{})
	c.Validation.Check(googleId, revel.Required{})
	c.Validation.Check(accessTokens, revel.Required{})

	// Handle errors
	if c.Validation.HasErrors() {
		return c.RenderJson(c.Validation.Errors)
	}

	//validate facebook access token
	if(!ValidateGoogleAccessToken(googleId, accessTokens)){
		return c.RenderJsonError("Invalid facebook access tokens")
	}

	//now check if user with fbId exist
	var users []models.User
	if err := Gdb.Where("google_id= ?", googleId).Find(&users).Error; err != nil {
		return c.RenderJsonError("Internal Server Error")
	}

	if len(users) >= 2 {
		return c.RenderJsonError("There are two account with the same google id")
	}

	var sessionKey string
	sessionKey = "s:user_"+c.Session.Id()

	if len(users) == 1 {
		if users[0].Status == models.USER_ACTIVE {
			//logged user in directly
			RCache.Set(sessionKey, users[0], SessionExpire)
			return c.RenderJsonSuccess(users[0])
		} else {
			return c.RenderJsonError("User was not activated")
		}
	}

	//if there is no such user with fbId; check email and register user
	users = []models.User{}
	if err := Gdb.Where("email= ?", email).Find(&users).Error; err != nil {
		return c.RenderJsonError("Internal Server Error")
	}

	if len(users) >= 2 {
		return c.RenderJsonError("There are two account with the same email id")
	}

	if len(users) == 1 {
		//update fbId field of that account
		existingUser := users[0]
		existingUser.GoogleId = googleId;

		Gdb.Save(&existingUser)

		RCache.Set(sessionKey, existingUser, SessionExpire)
		return c.RenderJsonSuccess(existingUser)
	}

	//otherwise just create new users -- oauth login/register doesn't need to activate the account
	newUser := models.User{
		FullName: name,
		Email: email,
		UserName: email,
		Password: "",
		Status: models.USER_ACTIVE,
		GoogleId: googleId,
	}

	Gdb.Save(&newUser)

	RCache.Set(sessionKey, newUser, SessionExpire)
	return c.RenderJsonSuccess(newUser)
}

func (c Auth) LoginUsingFacebook() revel.Result {
	fbId := c.Params.Get("fb_id")
	accessTokens := c.Params.Get("access_token")

	c.Validation.Check(fbId, revel.Required{})
	c.Validation.Check(accessTokens, revel.Required{})

	// Handle errors
	if c.Validation.HasErrors() {
		return c.RenderJson(c.Validation.Errors)
	}

	//validate facebook access token
	if(!ValidateFacebookAccessToken(fbId, accessTokens)){
		return c.RenderJsonError("Invalid facebook access tokens")
	}

	//now check if user with fbId exist
	var users []models.User
	if err := Gdb.Where("fb_id= ?", fbId).Find(&users).Error; err != nil {
		return c.RenderJsonError("Internal Server Error")
	}

	if len(users) >= 2 {
		return c.RenderJsonError("There are two account with the same fbId")
	}

	var sessionKey string
	sessionKey = "s:user_"+c.Session.Id()

	if len(users) == 1 {
		if users[0].Status == models.USER_ACTIVE {
			//logged user in directly
			RCache.Set(sessionKey, users[0], SessionExpire)
			return c.RenderJsonSuccess(users[0])
		} else {
			return c.RenderJsonError("User was not activated")
		}
	} else {
		return c.RenderJsonError("User with facebook id does not exist")
	}
}

func (c Auth) LoginUsingGoogle() revel.Result {
	googleId := c.Params.Get("google_id")
	accessTokens := c.Params.Get("access_token")

	c.Validation.Check(googleId, revel.Required{})
	c.Validation.Check(accessTokens, revel.Required{})

	// Handle errors
	if c.Validation.HasErrors() {
		return c.RenderJson(c.Validation.Errors)
	}

	//validate facebook access token
	if(!ValidateGoogleAccessToken(googleId, accessTokens)){
		return c.RenderJsonError("Invalid google access tokens")
	}

	//now check if user with fbId exist
	var users []models.User
	if err := Gdb.Where("google_id= ?", googleId).Find(&users).Error; err != nil {
		return c.RenderJsonError("Internal Server Error")
	}

	if len(users) >= 2 {
		return c.RenderJsonError("There are two account with the same google id")
	}

	var sessionKey string
	sessionKey = "s:user_"+c.Session.Id()

	if len(users) == 1 {
		if users[0].Status == models.USER_ACTIVE {
			//logged user in directly
			RCache.Set(sessionKey, users[0], SessionExpire)
			return c.RenderJsonSuccess(users[0])
		} else {
			return c.RenderJsonError("User was not activated")
		}
	} else {
		return c.RenderJsonError("User with facebook id does not exist")
	}
}

//link and unlink facebook, google
func (c Auth) LinkAccountWithFacebook() revel.Result {
	var sessionKey string
	sessionKey = "s:user_" + c.Session.Id()

	var sessionUser models.User
	RCache.Get(sessionKey, &sessionUser)

	if sessionUser.Id == 0 {
		c.RenderJsonError("User must be logged in first")
	}

	fbId := c.Params.Get("fb_id")
	accessTokens := c.Params.Get("access_token")

	c.Validation.Check(fbId, revel.Required{})
	c.Validation.Check(accessTokens, revel.Required{})

	// Handle errors
	if c.Validation.HasErrors() {
		return c.RenderJson(c.Validation.Errors)
	}

	//validate facebook access token
	if(!ValidateFacebookAccessToken(fbId, accessTokens)){
		return c.RenderJsonError("Invalid facebook access tokens")
	}

	//get latest user info from database
	var userId = sessionUser.Id
	var existingUser = models.User{}

	//find active user
	if err := Gdb.Where("id=?", userId).Where("status=?", models.USER_ACTIVE).First(&existingUser).Error; err != nil {
		return c.RenderJsonError("User does not exist")
	}

	//update the existing user
	existingUser.FbId = fbId;
	Gdb.Save(&existingUser);
	
	//TODO: might want to include other information here
	return c.RenderJsonSuccess(existingUser.Sanitize())
}

func (c Auth) LinkAccountWithGoogle() revel.Result {
	var sessionKey string
	sessionKey = "s:user_" + c.Session.Id()

	var sessionUser models.User
	RCache.Get(sessionKey, &sessionUser)

	if sessionUser.Id == 0 {
		c.RenderJsonError("User must be logged in first")
	}

	googleId := c.Params.Get("google_id")
	accessTokens := c.Params.Get("access_token")

	c.Validation.Check(googleId, revel.Required{})
	c.Validation.Check(accessTokens, revel.Required{})

	// Handle errors
	if c.Validation.HasErrors() {
		return c.RenderJson(c.Validation.Errors)
	}

	//validate facebook access token
	if(!ValidateGoogleAccessToken(googleId, accessTokens)){
		return c.RenderJsonError("Invalid google access tokens")
	}

	//get latest user info from database
	var userId = sessionUser.Id
	var existingUser = models.User{}

	//find active user
	if err := Gdb.Where("id=?", userId).Where("status=?", models.USER_ACTIVE).First(&existingUser).Error; err != nil {
		return c.RenderJsonError("User does not exist")
	}

	//update the existing user
	existingUser.GoogleId = googleId;
	Gdb.Save(&existingUser);

	//TODO: might want to include other information here
	return c.RenderJsonSuccess(existingUser.Sanitize())
}


func (c Auth) UnlinkAccountWithFacebook() revel.Result {
	var sessionKey string
	sessionKey = "s:user_" + c.Session.Id()

	var sessionUser models.User
	RCache.Get(sessionKey, &sessionUser)

	if sessionUser.Id == 0 {
		c.RenderJsonError("User must be logged in first")
	}

	existingUser := models.User{}
	if err := Gdb.First(&existingUser).Update("facebook_id", "").Error; err != nil {
		return c.RenderJsonError("Unable to unlink account with facebook")
	}

	var response struct{}
	return c.RenderJsonSuccess(response)
}


func (c Auth) UnlinkAccountWithGoogle() revel.Result {
	var sessionKey string
	sessionKey = "s:user_" + c.Session.Id()

	var sessionUser models.User
	RCache.Get(sessionKey, &sessionUser)

	if sessionUser.Id == 0 {
		c.RenderJsonError("User must be logged in first")
	}
	
	existingUser := models.User{}
	if err := Gdb.First(&existingUser).Update("google_id", "").Error; err != nil {
		return c.RenderJsonError("Unable to unlink account with google")
	}

	var response struct{}
	return c.RenderJsonSuccess(response)
}



