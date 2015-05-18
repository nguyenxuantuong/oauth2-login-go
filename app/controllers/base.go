package controllers

import (
	"github.com/revel/revel"
	"fmt"
	"time"
	"github.com/mrjones/oauth"
	"github.com/parnurzeal/gorequest"
	"encoding/json"
	"html/template"
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

//bad request unknown error page
func (c BaseController) RenderBadRequest(error interface {}) revel.Result {
	revel.ERROR.Println(error)
	c.RenderArgs["StackTrace"] = error;
	return c.RenderTemplate("errors/500.html")
}

func (c BaseController) RenderInternalServerError() revel.Result {
	return c.RenderTemplate("errors/500.html")
}

//render react templates
func (c BaseController) RenderReactTemplate(nodejsPath string, revelPath string) revel.Result {
	//request to get content from nodejs server
	_, body, err := nodeHttpAgent.Get(nodeHttpServerUrl + "/" + nodejsPath).End()

	//fall-back to using go render directly
	if err != nil {
		return c.RenderTemplate(revelPath)
	}

	var profile map[string]interface{}
	if err := json.Unmarshal([]byte(body), &profile); err != nil {
		return c.RenderBadRequest(err)
	}

	c.RenderArgs["Markup"] = template.HTML(profile["data"].(string));
	return c.RenderTemplate(revelPath);
}

var (
	SessionExpire time.Duration //default configuration -- keep time when session expired
	WebURL string
	nodeHttpServerUrl string
	nodeHttpAgent *gorequest.SuperAgent
)

//TODO: remove this -- using redis instead -- don't have time to do it now
var tokens map[string]*oauth.RequestToken

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

	var ok bool

	if nodeHttpServerUrl, ok = revel.Config.String("http.nodeserver"); !ok {
		revel.ERROR.Println("Missing NodeJS http server endpoint")
	}

	nodeHttpAgent = gorequest.New()

	//init db connections
	InitDB()
	InitRedis()
	InitMgo()
	InitOAuthServer()

	//TODO: save into redis
	tokens = make(map[string]*oauth.RequestToken)
}

