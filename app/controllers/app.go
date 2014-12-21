package controllers

import (
	"github.com/revel/revel"
//	"auth/app/models"
//	"encoding/json"
//	"code.google.com/p/go.crypto/bcrypt"
)

type App struct {
	GormController
}

//interface for API response
type Response struct {
	Status string
	Data   interface{}
	Errors	interface {}
}

//index path
func (c App) Index() revel.Result {
	return c.RenderTemplate("App/Index.html")
}

func (c App) Login() revel.Result {
	return c.RenderTemplate("Login/Login.html")
}

