package tests

import (
	"github.com/revel/revel"
	"auth/app/models"
	"auth/app/utils"
	"auth/app/controllers"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"encoding/json"
	"code.google.com/p/go.crypto/bcrypt"
	"time"
	"math"
	"net/http"
)

var _ = fmt.Printf
var _ = http.Response{}

var endpoint string

type UserTest struct {
	revel.TestSuite
}

//some sugar function for testing easily
func createUser(t *UserTest) *models.User{
	newUser := models.User{
		FullName: "Nguyen Xuan Tuong",
		Email: "nguy0066@e.ntu.edu.sg",
		UserName: "nguy0066",
		Password: utils.GetMD5Hash("111111"),
		Status: models.USER_ACTIVE,
	}
	
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newUser.Password), 10)
	newUser.HashedPassword = hashedPassword;
	
	if err := controllers.Gdb.Save(&newUser).Error; err != nil {
		t.AssertEqual(err, nil)
	}
	
	return &newUser
}

//quick function to call register user API
func registerUser(t *UserTest) *models.User{
	newUser := models.User{
		FullName: "Nguyen Xuan Tuong",
		Email: "nguy0066@e.ntu.edu.sg",
		UserName: "nguy0066",
		Password: utils.GetMD5Hash("111111"),
	}

	_, body, _ := gorequest.New().Post(endpoint + "/api/user/register").Send(newUser).End()

	//decode the body -- tag is optional because json-decoder able to deal with lowercase
	type UserResponse struct{
		Status string `json:"status"`
		Data   *models.User `json:"data"`
	}

	jsonResponse := UserResponse{}
	json.Unmarshal([]byte(body), &jsonResponse)

	//assert the response body -- info of the newly created user
	t.AssertEqual(jsonResponse.Status, "success")
	t.AssertEqual(jsonResponse.Data.FullName, newUser.FullName)
	t.AssertEqual(jsonResponse.Data.Email, newUser.Email)
	t.AssertEqual(jsonResponse.Data.UserName, newUser.UserName)

	//check if new user has been created
	var users []models.User
	controllers.Gdb.Where("email= ?", newUser.Email).Or("user_name= ?", newUser.UserName).Find(&users)

	//there is only one found user
	t.AssertEqual(len(users), 1)
	t.AssertEqual(users[0].Email, newUser.Email)
	t.AssertEqual(users[0].Status, models.USER_INACTIVE)

	//make sure both password are equal []byte(users[0].Password)
	t.AssertEqual(bcrypt.CompareHashAndPassword(users[0].HashedPassword, []byte(newUser.Password)), nil)
	
	return &users[0]
}

//active account by calling API
func activateAccount(t *UserTest, user *models.User) bool {
	var accountActivations []models.AccountActivation
	controllers.Gdb.Where("user_id=?", user.Id).Find(&accountActivations)

	t.AssertEqual(len(accountActivations), 1)
	
	//make sure expiry date is valid
	var expiryDate = accountActivations[0].ExpiryDate;
	var expectedExpiryDate = time.Now().AddDate(0, 0, 3)
	t.AssertEqual(int(math.Floor(math.Abs(expectedExpiryDate.Sub(expiryDate).Hours()))), 0)

	//then activate account
	gorequest.New().Post(endpoint + "/api/user/activateAccount/"+accountActivations[0].ActivationKey).End()
	
	return true
}


//ACTUAL TEST
func (t *UserTest) Before() {
	endpoint, _ = revel.Config.String("http.endpoint")
	
	controllers.Gdb.Exec("TRUNCATE TABLE user;")
	controllers.Gdb.Exec("TRUNCATE TABLE account_activation;")
	controllers.Gdb.Exec("TRUNCATE TABLE password_reset;")
}

//API register testing
func (t *UserTest) TestUserRegister() {
	registerUser(t);
}

//test user login (register + login)
func (t *UserTest) TestUserAccountActivation(){
	newUser := models.User{
		FullName: "Nguyen Xuan Tuong",
		Email: "nguy0066@e.ntu.edu.sg",
		UserName: "nguy0066",
		Password: utils.GetMD5Hash("111111"),
	}

	//register user
	_, body, _ := gorequest.New().Post(endpoint + "/api/user/register").Send(newUser).End()

	
	//then login using the same credential
	_, body, _ = gorequest.New().Post(endpoint + "/api/user/login").
	Send(`{"Email":"nguy0066@e.ntu.edu.sg"}`).
	Send(`{"Password":"` + newUser.Password +`"}`).
	End()

	type UserResponse struct{
		Status string `json:"status"`
		Data   *models.User `json:"data"`
		Errors	interface {} `json:"errors"`
	}

	//unable to login
	//error due to inactive user
	jsonResponse := UserResponse{}
	json.Unmarshal([]byte(body), &jsonResponse)
	t.AssertEqual(jsonResponse.Status, "error")

	var users []models.User
	controllers.Gdb.Where("email= ?", newUser.Email).Or("user_name= ?", newUser.UserName).Find(&users)
	t.AssertEqual(len(users), 1)
	
	activateAccount(t, &users[0])
	
	//then try to login again
	_, body, _ = gorequest.New().Post(endpoint + "/api/user/login").
	Send(`{"Email":"nguy0066@e.ntu.edu.sg"}`).
	Send(`{"Password":"` + newUser.Password +`"}`).
	End()

	//unmarshal the response
	json.Unmarshal([]byte(body), &jsonResponse)
	t.AssertEqual(jsonResponse.Data.FullName, newUser.FullName)

	//reset user to be empty array
	users = []models.User{}
	controllers.Gdb.Where("email= ?", newUser.Email).Or("user_name= ?", newUser.UserName).Find(&users)
	t.AssertEqual(users[0].Status, models.USER_ACTIVE)
	//TODO: also make sure that redis session has been set
}

//test password reset
func (t * UserTest) TestRequestPasswordReset(){
	newUser := registerUser(t)

	gorequest.New().Post(endpoint + "/api/user/requestPasswordReset").
	Send("email="+newUser.Email).End()
	
	var passwordReset []models.PasswordReset
	controllers.Gdb.Where("user_id=?", newUser.Id).Find(&passwordReset)
	t.AssertEqual(len(passwordReset), 1)
	
	//now send actual password reset
	gorequest.New().Post(endpoint + "/api/user/resetPassword/" + passwordReset[0].PasswordResetKey).
	Send("newPassword=" + utils.GetMD5Hash("111112")).End()

	var updatedUser = models.User{}
	controllers.Gdb.First(&updatedUser)
	
	//verify password has been updated
	t.AssertEqual(bcrypt.CompareHashAndPassword(updatedUser.HashedPassword, []byte(utils.GetMD5Hash("111112"))), nil)
}

//test change user password
func (t *UserTest) TestChangePassword() {
	createUser(t)

	//now change password
	request := gorequest.New()
	
	//login first to establish the session
	resp, _, _ := request.Post(endpoint + "/api/user/login").
	Send(`{"Email":"nguy0066@e.ntu.edu.sg"}`).
	Send(`{"Password":"` + utils.GetMD5Hash("111111") +`"}`).
	End()
	
	//get cookies to set for the next request
	var httpResp *http.Response
	httpResp = (*http.Response)(resp)

	//actually don't need to set cookies; but for reference only
	request.Post(endpoint + "/api/user/changePassword").
	AddCookie(httpResp.Cookies()[0]).
	Send("oldPassword=" + utils.GetMD5Hash("111111")).Send("newPassword=" + utils.GetMD5Hash("111112")).End()

	//now check password of the new user
	var user = models.User{}
	controllers.Gdb.First(&user)
	
	t.AssertEqual(bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(utils.GetMD5Hash("111112"))), nil)
}

func (t *UserTest) After() {
	controllers.Gdb.Exec("TRUNCATE TABLE user;")
	controllers.Gdb.Exec("TRUNCATE TABLE account_activation;")
	controllers.Gdb.Exec("TRUNCATE TABLE password_reset;")
}
