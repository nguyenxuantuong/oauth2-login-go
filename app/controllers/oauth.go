//contain oauth functionality
package controllers

import (
	"github.com/revel/revel"
	"github.com/RangelReale/osin"
	"fmt"
	"auth/app/utils"
	"auth/app/models"
	"auth/app/routes"
	"code.google.com/p/go.crypto/bcrypt"
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

const (
	AuthorizationCode = 1
	AccessToken = 2
	RefreshToken = 3
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

//handler to authorize request; it will redirect the users to obtain either authorize_token or access_token
//NOTE: this is mainly for web flow
func (c OAuth) Authorize(client_id string, redirect_url string, response_type string, register_redirect string) revel.Result {
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
			
			//TODO: if you want to save userdata -- do it here; put into ar object which will be saved into storage
			ar.UserData = sessionUser.Sanitize();

			//TODO: every redirect url need to be compared with the registered redirect url to make sure it have the same host
			//in case, we need to redirect to register url instead
			if redirect_url != "" {
				ar.RedirectUri = redirect_url
			}

			//we will alway force redirect using redirect url
			OAuthServer.FinishAuthorizeRequest(resp, req, ar)
		} else {
			return c.Redirect(routes.App.Login(client_id, redirect_url, response_type, register_redirect))
		}
	}

	//name might misleading; but it will just redirect with authorized code
	osin.OutputJSON(resp, c.Response.Out, req)
	return nil
}

//verify if authorization code / access code is valid
//this is for the other server to verify if the accessToken it received is credential or not
func (c OAuth) VerifyToken(code string, codeType int) revel.Result {
	if codeType == AuthorizationCode {
		//load authorization from the storage
		authorizeData, error := OAuthServer.Storage.LoadAuthorize(code)
		if error != nil || authorizeData.IsExpired() {
			return c.RenderJsonError("authorization code does not exist or has been expired")
		}

		//otherwise, return the user data
		return c.RenderJsonSuccess(authorizeData.UserData)
	} else if codeType == AccessToken {
		accessData, error := OAuthServer.Storage.LoadAccess(code)
		if error != nil || accessData.IsExpired() {
			return c.RenderJsonError("authorization code does not exist or has been expired")
		}
	} else if codeType == RefreshToken {
		refreshData, error := OAuthServer.Storage.LoadRefresh(code)
		if error != nil || refreshData.IsExpired() {
			return c.RenderJsonError("authorization code does not exist or has been expired")
		}
	} else {
		return c.RenderJsonError("unknown token type")
	}

	//if everything ok; just return success
	var response struct{}
	return c.RenderJsonSuccess(response)
}

//exchange for tokens -- it should be called in server side; unless client-side is ssl
//NOTE: this is mainly for mobile flow
func (c OAuth) AccessToken() revel.Result {
	resp := OAuthServer.NewResponse()
	defer resp.Close()

	req := c.Request.Request;

	var sessionKey string
	sessionKey = "s:user_"+c.Session.Id()

	//if session is found; then return immediately
	var sessionUser models.User
	RCache.Get(sessionKey, &sessionUser)
	
	if ar := OAuthServer.HandleAccessRequest(resp, req); ar != nil {
		//TODO: if we allow request access token; user-password grant type; then putting logic to check for user's credential overhere
		//also can add same logic in case user has already logged in; and want to request for access_token; no user-password check logic in the handle-request method
		grantType := c.Params.Get("grant_type")

		//TODO: if you want to save userdata -- do it here; put into ar object which will be saved into storage
		
		if grantType == osin.PASSWORD {
			if sessionUser.Id != 0 {
				ar.Authorized = true
				OAuthServer.FinishAccessRequest(resp, req, ar)
			} else {
				//get username and password from the request
				username := c.Params.Get("username")
				password := c.Params.Get("password")

				//check if such user exist in database
				user := models.User{}
				if Gdb.Where("email = ?", username).First(&user).RecordNotFound() {
					return c.RenderJsonError("Email does not exist");
				}
				
				if user.Status != models.USER_ACTIVE {
					return c.RenderJsonError("User was not activated.");
				}

				//compare password to validate the user
				if err := bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password)); err != nil {
					return c.RenderJsonError("Passwords are miss-match");
				}

				//password; username is OK; we pass through
				ar.Authorized = true
				OAuthServer.FinishAccessRequest(resp, req, ar)
			}
		} else {
			//just pass through
			ar.Authorized = true
			OAuthServer.FinishAccessRequest(resp, req, ar)
		}
	}

	if resp.IsError && resp.InternalError != nil {
		revel.ERROR.Printf("ERROR: %s\n", resp.InternalError)
	}
	
	osin.OutputJSON(resp, c.Response.Out, req)
	return nil
}

