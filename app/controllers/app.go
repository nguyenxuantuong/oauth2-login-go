package controllers

import (
	"github.com/revel/revel"
)

type App struct {
	GormController
}

//interface for API response
type Response struct {
	Status string `json:"status"`
	Data   interface{} `json:"data"`
	Errors	interface {} `json:"errors"`
}

//index path
func (c App) Index() revel.Result {
	return c.RenderTemplate("App/Index.html")
}

func (c App) Login() revel.Result {
	return c.RenderTemplate("Login/Login.html")
}

