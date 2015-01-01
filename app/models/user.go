package models

import (
	"time"
	"github.com/revel/revel"
	"regexp"
)

const (
	USER_INACTIVE = 0
	USER_ACTIVE = 1
	USER_SUSPEND = 2
)

type User struct {
	Id           	int64	`json:"id"`
	Status			int8	`json:"status"`
	UserName 	 	string  `sql:"size:50" json:"user_name"`
	Email 		 	string  `sql:"size:50" json:"email"`
	Password 		string  `sql:"size:50" json:"password"`
	HashedPassword  []byte	`json:"hashed_password"`
	FullName        string  `sql:"size:255" json:"full_name"`
	LastLogin    	time.Time `json:"last_login"`
	CreatedDate    	time.Time `json:"created_date"`
	UpdatedDate    	time.Time `json:"updated_date"`
	DeletedDate    	time.Time `json:"deleted_date"`
}

var userRegex = regexp.MustCompile("^\\w*$")

//note that sanitize will alter original object
func (user *User) Sanitize() *User{
	var sanitizeUser = &User {
		Id: user.Id,
		UserName: user.UserName,
		FullName: user.FullName,
		Email: user.Email,
	}
	
	return sanitizeUser
}

func (user *User) Validate(v *revel.Validation) {
	v.Check(user.UserName,
		revel.Required{},
		revel.MaxSize{15},
		revel.Match{userRegex},
	).Message("Username is a required field and must not exceed 15 characters");

	v.Check(user.Email,
		revel.Required{},
	).Message("Email is a required field");
	
	v.Check(user.FullName,
		revel.Required{},
		revel.MaxSize{100},
	).Message("Full name must not exceed 100 characters");

	//TODO: check password is MD5 encrypted; we don't allow user to send plain password
	//to prevent man-in-middle; non https
	v.Check(user.Password,
		revel.Required{},
	)
}


