package controllers

import (
	"github.com/revel/revel"
	"golang.org/x/oauth2"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"auth/app/models"
	"auth/app/utils"
)

var _ = fmt.Printf

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

//login using facebook
func (c App) LoginByGoogle() revel.Result {
	var googleClientId, googleClientSecret, googleRedirectUrl, googleUserInfoUrl string;
	var ok bool;

	if googleClientId, ok = revel.Config.String("google.clientId"); !ok {
		return c.RenderInternalServerError();
	}

	if googleClientSecret, ok = revel.Config.String("google.clientSecret"); !ok {
		return c.RenderInternalServerError();
	}

	if googleRedirectUrl, ok = revel.Config.String("google.redirectUrl"); !ok {
		return c.RenderInternalServerError();
	}

	if googleUserInfoUrl, ok = revel.Config.String("google.userinfoUrl"); !ok {
		return c.RenderInternalServerError();
	}

	conf := &oauth2.Config{
		ClientID:     googleClientId,
		ClientSecret: googleClientSecret,
		RedirectURL: googleRedirectUrl,
		Scopes:       []string{
			"https://www.googleapis.com/auth/plus.login",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
	}

	code := c.Params.Get("code")

	if code == "" {
		// Redirect user to consent page to ask for permission
		// for the scopes specified above.
		url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
		revel.INFO.Printf("Visit the URL for the auth dialog: %v", url)

		return c.Redirect(url)
	}

	//Use the authorization code that is pushed to the redirect URL.
	//NewTransportWithCode will do the handshake to retrieve
	//an access token and initiate a Transport that is
	//authorized and authenticated by the retrieved token.
	tok, err := conf.Exchange(oauth2.NoContext, code)

	if err != nil {
		revel.ERROR.Println(err)
		return c.RenderBadRequest(err)
	}

	client := conf.Client(oauth2.NoContext, tok)

	resp, err := client.Get(googleUserInfoUrl)

	if err != nil {
		return c.RenderBadRequest(err)
	}

	// Reading the body
	raw, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return c.RenderBadRequest(err)
	}

	var profile map[string]interface{}
	if err := json.Unmarshal(raw, &profile); err != nil {
		return c.RenderBadRequest(err)
	}

	//get google id of the user
	googleId := profile["id"].(string)

	var users []models.User
	if err := Gdb.Where("google_id= ?", googleId).Find(&users).Error; err != nil {
		return c.RenderBadRequest(err)
	}

	var sessionKey string
	sessionKey = "s:user_"+c.Session.Id()

	//if user has already register in the system before
	if len(users) == 1 {
		if users[0].Status == models.USER_ACTIVE {
			//logged user in directly
			RCache.Set(sessionKey, users[0], SessionExpire)

			//TODO: check redirect URL; if not redirect user into home page of Oath CMS
			return c.RenderTemplate("Home/Home.html")
		} else {
			//if user haven't activated account before, we show the bad request instead
			return c.RenderBadRequest(err)
		}
	}

	//for now, if user doesn't exist in the system; we register him directly
	//TODO: in the future, we might want to explicitly ask if user want to register using the current Google/Facebook acc
	//TODO: make sure that the verified_email of the google account is set
	newUser := models.User{
		FullName: utils.ChooseFirstNonEmpty(profile["name"].(string), profile["given_name"].(string), profile["family_name"].(string)),
		Email: profile["email"].(string),
		UserName: "",
		Password: "",
		Status: models.USER_ACTIVE,
		GoogleId: googleId,
	}

	//save user; if there is error; then show that error
	if err := Gdb.Save(&newUser).Error; err != nil {
		return c.RenderBadRequest(err)
	}

	//set the cache of the users
	RCache.Set(sessionKey, newUser, SessionExpire)


	//TODO: if there is redirect url; redirect him back to the main app; otherwise, log him in the Oauth CMS
	return c.RenderTemplate("Home/Home.html")
}


