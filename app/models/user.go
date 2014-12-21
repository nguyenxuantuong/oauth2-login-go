package models

import (
	"time"
	"github.com/revel/revel"
	"regexp"
)

type User struct {
	Id           	int64
	UserName 	 	string  `sql:"size:50"`
	Email 		 	string  `sql:"size:50"`
	Password 		string  `sql:"size:50"`
	HashedPassword  []byte
	FullName        string  `sql:"size:255"`
	LastLogin    	time.Time
	CreatedDate    	time.Time
	UpdatedDate    	time.Time
	DeletedDate    	time.Time
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


