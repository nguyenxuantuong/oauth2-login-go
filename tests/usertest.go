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
)

var _ = fmt.Printf

type UserTest struct {
	revel.TestSuite
}

func (t *UserTest) Before() {
	controllers.Gdb.Exec("TRUNCATE TABLE user;")
}

//just a gorm add user function
func (t *UserTest) TestAddUser() {
	newUser := models.User{
		FullName: "Nguyen Xuan Tuong",
		Email: "nguy0066@e.ntu.edu.sg",
		UserName: "nguy0066",
		Password: "111111",
	}
	
	//check if existing users exist or not
	var existingUsers []models.User
	controllers.Gdb.Where("email= ?", newUser.Email).Or("user_name= ?", newUser.UserName).Find(&existingUsers);

	t.AssertEqual(len(existingUsers), 0 )
	
	//now insert new users
	controllers.Gdb.Create(&newUser)
	
	//now verify that there is one user has been inserted
	controllers.Gdb.Where("email=?", newUser.Email).Or("user_name=?", newUser.UserName).Find(&existingUsers);
	t.AssertEqual(len(existingUsers), 1)
}


//API register testing
func (t *UserTest) TestUserRegister() {
	newUser := models.User{
		FullName: "Nguyen Xuan Tuong",
		Email: "nguy0066@e.ntu.edu.sg",
		UserName: "nguy0066",
		Password: utils.GetMD5Hash("111111"),
	}

	endpoint, _ := revel.Config.String("http.endpoint")

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

	//make sure both password are equal []byte(users[0].Password)
	t.AssertEqual(bcrypt.CompareHashAndPassword(users[0].HashedPassword, []byte(newUser.Password)), nil)
}

//test user login (register + login)
func (t *UserTest) TestUserLogin(){
	newUser := models.User{
		FullName: "Nguyen Xuan Tuong",
		Email: "nguy0066@e.ntu.edu.sg",
		UserName: "nguy0066",
		Password: utils.GetMD5Hash("111111"),
	}

	type UserResponse struct{
		Status string `json:"status"`
		Data   *models.User `json:"data"`
	}
	
	//register user
	endpoint, _ := revel.Config.String("http.endpoint")

	_, body, _ := gorequest.New().Post(endpoint + "/api/user/register").Send(newUser).End()

	jsonResponse := UserResponse{}
	json.Unmarshal([]byte(body), &jsonResponse)
	t.AssertEqual(jsonResponse.Data.FullName, newUser.FullName)
	
	//then login using the same credential
	_, body, _ = gorequest.New().Post(endpoint + "/api/user/login").
	Send(`{"Email":"nguy0066@e.ntu.edu.sg"}`).
	Send(`{"Password":"` + newUser.Password +`"}`).
	End()

	//unmarshal the response
	json.Unmarshal([]byte(body), &jsonResponse)
	t.AssertEqual(jsonResponse.Data.FullName, newUser.FullName)
}

func (t *UserTest) After() {
	controllers.Gdb.Exec("TRUNCATE TABLE user;")
}
