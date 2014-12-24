package controllers

import (
	"github.com/revel/revel"
	"github.com/revel/revel/cache"
	"auth/app/models"
	"encoding/json"
	"code.google.com/p/go.crypto/bcrypt"
	"fmt"
//	"time"
)

var _ = fmt.Printf

type Auth struct {
	GormController
}

//auth api
func (c Auth) Register() revel.Result {
	//unmarshal the request
	newUser := models.User{}
	err := json.NewDecoder(c.Request.Body).Decode(&newUser)
	revel.INFO.Printf("new decoded user %+v", newUser);
	if err != nil {
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

	c.Txn.Where("email= ?", newUser.Email).Or("user_name= ?", newUser.UserName).Find(&existingUsers);

	if len(existingUsers) == 0 {
		//hash the password
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newUser.Password), 10)

		//assign hash, encrypted password for the new users
		newUser.HashedPassword = hashedPassword
		newUser.Password = ""

		newUser.Status = models.USER_INACTIVE
		
		//create new user
		c.Txn.NewRecord(newUser)
		c.Txn.Create(&newUser)

		//sanitize the new user
		return c.RenderJsonSuccess(newUser.Sanitize())

	} else {
		return c.RenderJsonError("Username or Email has been taken.")
	}
}

func (c Auth) Login() revel.Result {
	var sessionKey string
	sessionKey = "s:user_" + c.Session.Id()

	//if session is found; then return immediately
	var sessionUser models.User
	cache.Get(sessionKey, &sessionUser)

	//TODO: how to check if result is found
	if sessionUser.Email != "" {
		return c.RenderJsonSuccess(sessionUser)
	}
	
	var f interface{}

	if err := json.NewDecoder(c.Request.Body).Decode(&f); err != nil {
		return c.RenderJsonError("Invalid post data. It is not in proper JSON format")
	}

	m := f.(map[string]interface{})

	//verify password and email exist in post data
	c.Validation.Required(m["Email"])
	c.Validation.Required(m["Password"])
	
	if c.Validation.HasErrors() {
		return c.RenderJsonError("Missing required parameters email or password");
	}
	
	//otherwise check again database to find the user
	user := models.User{}
	if Gdb.Where("email = ?", m["Email"]).First(&user).RecordNotFound() {
		return c.RenderJsonError("User with email " + user.Email + " does not exist" );
	}

	//compare password to validate the user
	if err := bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(m["Password"].(string))); err != nil {
		return c.RenderJsonError("Unable to login. Passwords are miss-match");
	}

	//otherwise, set session in redis
	go cache.Set(sessionKey, user, c.getSessionExpire())
	
	return c.RenderJsonSuccess(user)
}

func (c Auth) ForgotPassword() revel.Result {
	return nil
}

func (c Auth) ResetPassword() revel.Result {
	return nil
}

func (c Auth) ActivateAccount() revel.Result {
	return nil
}
