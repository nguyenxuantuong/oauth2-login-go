package models

import (
	"time"
)

type AccountActivation struct {
	Id           	int64 	`json:"id"`
	ActivationKey   string  `sql:"size:255" json: "activation_key"`
	ExpiryDate		time.Time `json:"expiry_date"`
	CreatedDate    	time.Time `json:"created_date"`
	UpdatedDate    	time.Time `json:"updated_date"`
	DeletedDate    	time.Time `json:"deleted_date"`
}
