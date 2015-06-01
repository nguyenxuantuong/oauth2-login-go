package controllers

import (
	"github.com/revel/revel"
	"auth/app/models"
//	"auth/app/utils"
//	"encoding/json"
//	"code.google.com/p/go.crypto/bcrypt"
//	"github.com/jinzhu/gorm"
//	"github.com/parnurzeal/gorequest"
//	"time"
	"fmt"
//	"strconv"
)

var _ = fmt.Printf

//this controller handle login/logout/session, etc...
type UserController struct {
	BaseController
	GormController
}

func (c UserController) GetUserList() revel.Result {
	paginatedParams := PaginatedParams{}
	var err error

	if paginatedParams, err = c.GetPaginationParams(); err != nil {
		return c.RenderJsonError(err.Error())
	}

	limit := paginatedParams.Limit
	offset := paginatedParams.Offset

	var userList []models.User
	var totalUser int

	if err := Gdb.Limit(limit).Offset(offset).Select("id, status, user_name, email, full_name, fb_id, google_id, twitter_id").
	Find(&userList).Count(&totalUser).Error; err != nil {
		return c.RenderJsonError(err.Error())
	}

	return c.RenderPaginatedJsonSuccess(userList, totalUser)
}

func (c UserController) UpdateUserInfo() revel.Result {
	return c.RenderJsonSuccess(1)
}

func (c UserController) GetUserInfo() revel.Result {
	return c.RenderJsonSuccess(2)
}