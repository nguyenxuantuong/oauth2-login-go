package controllers

import (
	"github.com/revel/revel"
)

type App struct {
	BaseController
}

//index path
func (c App) Index() revel.Result {
	return c.RenderTemplate("App/Index.html")
}

func (c App) Login() revel.Result {
	return c.RenderTemplate("Login/Login.html")
}

func (c App) Register() revel.Result {
	return c.RenderTemplate("Register/Register.html")
}

func (c App) Activation() revel.Result {
	return c.RenderTemplate("AccountActivation/AccountActivation.html")
}

func (c App) PasswordReset() revel.Result {
	return c.RenderTemplate("ForgotPassword/ForgotPassword.html")
}


