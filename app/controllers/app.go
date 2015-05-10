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

//normal login flow
func (c App) Login() revel.Result {
	return c.RenderTemplate("Login/Login.html")
}

//normal register
func (c App) Register() revel.Result {
	return c.RenderTemplate("Register/Register.html")
}

//account activation
func (c App) Activation() revel.Result {
	return c.RenderTemplate("AccountActivation/AccountActivation.html")
}

//type new password using the password reset link
func (c App) ResetPassword() revel.Result {
	return c.RenderTemplate("ResetPassword/ResetPassword.html")
}

//ask password to be sent
func (c App) ForgotPassword() revel.Result {
	return c.RenderTemplate("ForgotPassword/ForgotPassword.html")
}


