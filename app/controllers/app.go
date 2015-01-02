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

