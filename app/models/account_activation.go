package models

import (
	"time"
)

type AccountActivation struct {
	Id           	int64
	ActivationKey   string  `sql:"size:255"`
	ExpiryDate		time.Time
	CreatedDate    	time.Time
	UpdatedDate    	time.Time
	DeletedDate    	time.Time
}
