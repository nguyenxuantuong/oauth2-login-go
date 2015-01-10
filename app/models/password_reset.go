package models

import (
	"time"
)

type PasswordReset struct {
	Id           	int64 	`json:"id"`
	UserId			int64	`json:"user_id"`
	PasswordResetKey   string  `sql:"size:255" json: "password_reset_key"`
	ExpiryDate		time.Time `json:"expiry_date"`
	CreatedDate    	time.Time `json:"created_date"`
	UpdatedDate    	time.Time `json:"updated_date"`
	DeletedDate    	time.Time `json:"deleted_date"`
}
