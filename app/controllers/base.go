package controllers

import (
	"github.com/revel/revel"
	"fmt"
	"time"
)

//For now, base controller is just an extension of revel controller
type BaseController struct {
	*revel.Controller
}

//interface for API response
type Response struct {
	Status string `json:"status"`
	Data   interface{} `json:"data"`
	Errors	interface {} `json:"errors"`
}

//sugar function to be shared among other controllers
func (c BaseController) RenderJsonError(errors interface {}) revel.Result{
	return c.RenderJson(Response{Status: "error", Errors: errors})
}

//sugar function to return success data
func (c BaseController) RenderJsonSuccess(data interface {}) revel.Result {
	return c.RenderJson(Response{Status: "success", Data: data})
}

var (
	SessionExpire time.Duration //default configuration -- keep time when session expired
	WebURL string
)

func Init(){
	//init session expire variable
	var expireAfterDuration time.Duration

	//keep the expire duration key
	var err error
	if expiresString, ok := revel.Config.String("session.expires"); !ok {
		expireAfterDuration = 30 * 24 * time.Hour
	} else if expiresString == "session" {
		expireAfterDuration = 0
	} else if expireAfterDuration, err = time.ParseDuration(expiresString); err != nil {
		panic(fmt.Errorf("session.expires invalid: %s", err))
	}

	SessionExpire = expireAfterDuration
	WebURL, _ = revel.Config.String("http.weburl")
	
	//init db connections
	InitDB()
	InitRedis()
	InitMgo()
}

