//contain oauth functionality
package controllers

import (
	"github.com/revel/revel"
	"github.com/RangelReale/osin"
	"fmt"
	"auth/app/utils"
	"auth/app/models"
	"auth/app/routes"
//	"strings"
)

var _ = fmt.Printf

type OAuth struct {
	BaseController
	GormController
}

var (
	OAuthServer *osin.Server
)

func InitOAuthServer(){
	sconfig := osin.NewServerConfig()
	
	//allow both authorize code (for token exchange) and token itself
	sconfig.AllowedAuthorizeTypes = osin.AllowedAuthorizeType{osin.CODE, osin.TOKEN}
	sconfig.AllowedAccessTypes = osin.AllowedAccessType{osin.AUTHORIZATION_CODE, osin.REFRESH_TOKEN, osin.PASSWORD, osin.CLIENT_CREDENTIALS, osin.ASSERTION}
	sconfig.AllowGetAccessRequest = true
	
	//initialize server which take the config params
	mgoDbName := revel.Config.StringDefault("mgo.dbname", "oauth_go")
	OAuthServer = osin.NewServer(sconfig, utils.NewOAuthStorage(Session, mgoDbName))
}

//handler to authorize request
func (c OAuth) Authorize() revel.Result {
	//create resp object and defer close
	resp := OAuthServer.NewResponse()
	defer resp.Close()

	//TODO: check if user session is there;
	var sessionKey string
	sessionKey = "s:user_"+c.Session.Id()

	//if session is found; then return immediately
	var sessionUser models.User
	RCache.Get(sessionKey, &sessionUser)
	
	//get raw request object
	req := c.Request.Request;

	//using osin server to handle the request -- which return authorizeRequest object
	if ar := OAuthServer.HandleAuthorizeRequest(resp, req); ar != nil {
		if sessionUser.Id != 0 {
			//if user has logged in; allow him to get authorized code immediately
			ar.Authorized = true
			
			//we will alway force redirect using redirect url
			OAuthServer.FinishAuthorizeRequest(resp, req, ar)
		} else {
			return c.Redirect(routes.App.Login() + "?" + c.Params.Query.Encode())
		}
	}

	//name might misleading; but it will just redirect with authorized code
	osin.OutputJSON(resp, c.Response.Out, req)
	return nil
}

//exchange for tokens
func (c OAuth) AccessToken() revel.Result {
	resp := OAuthServer.NewResponse()
	defer resp.Close()

	req := c.Request.Request;
	
	if ar := OAuthServer.HandleAccessRequest(resp, req); ar != nil {
		ar.Authorized = true
		OAuthServer.FinishAccessRequest(resp, req, ar)
	}

	if resp.IsError && resp.InternalError != nil {
		revel.ERROR.Printf("ERROR: %s\n", resp.InternalError)
	}
	
	osin.OutputJSON(resp, c.Response.Out, req)
	return nil
}

