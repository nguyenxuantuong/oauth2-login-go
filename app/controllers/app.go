package controllers

import (
	"github.com/revel/revel"
	"golang.org/x/oauth2"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"auth/app/models"
	"auth/app/utils"
	"github.com/mrjones/oauth"
	"github.com/jinzhu/gorm"
	"time"
)

var _ = fmt.Printf

type App struct {
	BaseController
}

//index path
func (c App) Index() revel.Result {
	return c.RenderTemplate("App/Index.html")
}

func (c App) Home() revel.Result {
	//only allow authenticated user to login
	//TODO: move it into a seperated middleware
	var sessionKey string
	sessionKey = "s:user_" + c.Session.Id()

	var sessionUser models.User
	RCache.Get(sessionKey, &sessionUser)

	//if user doesn't login -- redirect him into the login page
	if sessionUser.Id == 0 {
		return c.Redirect(App.Login)
	}

	return c.RenderTemplate("Home/Home.html")
}

//render application error
func (c App) LoginError(title string, description string) revel.Result {
	c.RenderArgs["Title"] = title;
	c.RenderArgs["Description"] = description;
	//when there is login error; we clear the session inside redis
	var sessionKey string
	sessionKey = "s:user_"+c.Session.Id()

	RCache.Delete(sessionKey)

	return c.RenderTemplate("errors/application-error.html")
}

//normal login flow
func (c App) Login(client_id string, redirect_url string, response_type string, register_redirect string) revel.Result {
	c.SetSSOClientInformation(client_id, redirect_url, response_type, register_redirect)

	return c.RenderReactTemplate("render/login", "Login/Login.html")
}

//normal register
func (c App) Register() revel.Result {
	return c.RenderReactTemplate("render/register", "Register/Register.html")
}

//account activation
func (c App) Activation(client_id string, redirect_url string, response_type string, register_redirect string) revel.Result {
	var activationKey = c.Params.Get("activationKey");

	//check if activation key exist
	var accountActivation models.AccountActivation
	if err := Gdb.Where("activation_key=?", activationKey).First(&accountActivation).Error; err != nil {
		if err == gorm.RecordNotFound {
			return c.RenderBadRequest("Invalid activation key");
		} else {
			return c.RenderBadRequest("Internal Database error");
		}
	}

	//key is already expired
	if accountActivation.ExpiryDate.Before(time.Now()) {
		return c.RenderBadRequest("Activation key has already been expired")
	}

	//if there is activation key; then activate the user
	var userId = accountActivation.UserId
	var existingUser = models.User{}

	if err := Gdb.Where("id=?", userId).First(&existingUser).Error; err != nil {
		return c.RenderBadRequest("User does not exist")
	}

	existingUser.Status = models.USER_ACTIVE;

	if err := Gdb.Save(&existingUser).Error; err != nil {
		return c.RenderInternalServerError();
	}

	var sessionKey string
	sessionKey = "s:user_"+c.Session.Id()

	//everything ok, redirect him into home page or redirect URL
	//set the cache of the users
	RCache.Set(sessionKey, existingUser.Sanitize(), SessionExpire)

	//Now; delete the activation key
	if err := Gdb.Delete(&accountActivation).Error; err != nil {
		return c.RenderInternalServerError()
	}

	c.SetSSOClientInformation(client_id, redirect_url, response_type, register_redirect)

	return c.RedirectAfterLoginSuccess(App.Home, register_redirect)
}

//type new password using the password reset link
func (c App) ResetPassword() revel.Result {
	return c.RenderReactTemplate("render/resetPassword", "ResetPassword/ResetPassword.html")
}

//ask password to be sent
func (c App) ForgotPassword() revel.Result {
	return c.RenderReactTemplate("render/forgotPassword", "ForgotPassword/ForgotPassword.html")
}

//login using twitter
func (c App) LoginByTwitter() revel.Result {
	var twitterClientId, twitterClientSecret, twitterRedirectUrl, twitterUserInfoUrl string;
	var ok bool;

	if twitterClientId, ok = revel.Config.String("twitter.clientId"); !ok {
		return c.RenderInternalServerError();
	}

	if twitterClientSecret, ok = revel.Config.String("twitter.clientSecret"); !ok {
		return c.RenderInternalServerError();
	}

	if twitterRedirectUrl, ok = revel.Config.String("twitter.redirectUrl"); !ok {
		return c.RenderInternalServerError();
	}

	if twitterUserInfoUrl, ok = revel.Config.String("twitter.userinfoUrl"); !ok {
		return c.RenderInternalServerError();
	}

	var twitter = oauth.NewConsumer(
		twitterClientId,
		twitterClientSecret,
		oauth.ServiceProvider{
			AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
			RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
			AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
		},
	)

	requestToken, url, err := twitter.GetRequestTokenAndUrl(twitterRedirectUrl)
	if err != nil {
		revel.ERROR.Fatal(err)
		return c.RenderInternalServerError()
	}

	revel.INFO.Println("(1) Go to: " + url)

	tokens[requestToken.Token] = requestToken

	_ , _, _ = requestToken, twitterRedirectUrl, twitterUserInfoUrl

	verificationCode := c.Params.Get("oauth_verifier")
	tokenKey := c.Params.Get("oauth_token")

	if(verificationCode == "" || tokenKey == ""){
		return c.Redirect(url)
	}

	//exchange authorization code for access token
	accessToken, err := twitter.AuthorizeToken(tokens[tokenKey], verificationCode)
	if err != nil {
		revel.ERROR.Fatal(err)
		return c.RenderInternalServerError()
	}

	//now: trying to get the user information
	response, err := twitter.Get(
		twitterUserInfoUrl,
		map[string]string{"count": "1"},
		accessToken)

	if err != nil {
		revel.ERROR.Fatal(err)
	}
	defer response.Body.Close()

	//now convert it into generic interface
	raw, err := ioutil.ReadAll(response.Body)

	var profile map[string]interface{}
	if err := json.Unmarshal(raw, &profile); err != nil {
		return c.RenderBadRequest(err)
	}

	//get google id of the user
	twitterId := profile["id_str"].(string)

	var users []models.User
	if err := Gdb.Where("twitter_id= ?", twitterId).Find(&users).Error; err != nil {
		return c.RenderBadRequest(err)
	}

	var sessionKey string
	sessionKey = "s:user_"+c.Session.Id()

	//if user has already register in the system before
	if len(users) == 1 {
		if users[0].Status == models.USER_ACTIVE {
			//logged user in directly
			RCache.Set(sessionKey, users[0], SessionExpire)

			return c.RedirectAfterLoginSuccess(App.Home)
		} else {
			//if user haven't activated account before, we show the bad request instead
			return c.RenderBadRequest(err)
		}
	}

	//for now, if user doesn't exist in the system; we register him directly
	//TODO: in the future, we might want to explicitly ask if user want to register using the current Google/Facebook/Twitter acc
	//TODO: make sure that the verified_email of the twitter account is set
	newUser := models.User{
		FullName: utils.ChooseFirstNonEmpty(profile["name"].(string), profile["screen_name"].(string)),
		Email: "",
		UserName: "",
		Password: "",
		Status: models.USER_ACTIVE,
		TwitterId: twitterId,
	}

	//save user; if there is error; then show that error
	if err := Gdb.Save(&newUser).Error; err != nil {
		return c.RenderBadRequest(err)
	}

	//set the cache of the users
	RCache.Set(sessionKey, newUser, SessionExpire)
	return c.RedirectAfterLoginSuccess(App.Home)
}

//login using facebook
func (c App) LoginByFacebook() revel.Result {
	var facebookClientId, facebookClientSecret, facebookRedirectUrl, facebookUserInfoUrl string;
	var ok bool;

	if facebookClientId, ok = revel.Config.String("facebook.clientId"); !ok {
		return c.RenderInternalServerError();
	}

	if facebookClientSecret, ok = revel.Config.String("facebook.clientSecret"); !ok {
		return c.RenderInternalServerError();
	}

	if facebookRedirectUrl, ok = revel.Config.String("facebook.redirectUrl"); !ok {
		return c.RenderInternalServerError();
	}

	if facebookUserInfoUrl, ok = revel.Config.String("facebook.userinfoUrl"); !ok {
		return c.RenderInternalServerError();
	}

	conf := &oauth2.Config{
		ClientID:     facebookClientId,
		ClientSecret: facebookClientSecret,
		RedirectURL: facebookRedirectUrl,
		Scopes:       []string{
			"email",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://graph.facebook.com/oauth/authorize",
			TokenURL: "https://graph.facebook.com/oauth/access_token",
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

	tok, err := conf.Exchange(oauth2.NoContext, code)

	if err != nil {
		revel.ERROR.Println(err)
		return c.RenderBadRequest(err)
	}

	//now; create a client to request for user info from facebook endpoint
	client := conf.Client(oauth2.NoContext, tok)

	resp, err := client.Get(facebookUserInfoUrl)

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
	facebookId := profile["id"].(string)

	var users []models.User
	if err := Gdb.Where("fb_id= ?", facebookId).Find(&users).Error; err != nil {
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
			return c.RedirectAfterLoginSuccess(App.Home)
		} else {
			//if user haven't activated account before, we show the bad request instead
			return c.RenderBadRequest(err)
		}
	}

	//for now, if user doesn't exist in the system; we register him directly
	//TODO: in the future, we might want to explicitly ask if user want to register using the current Google/Facebook acc
	//TODO: make sure that the verified_email of the google account is set
	newUser := models.User{
		FullName: utils.ChooseFirstNonEmpty(profile["name"].(string), profile["last_name"].(string), profile["first_name"].(string)),
		Email: profile["email"].(string),
		UserName: "",
		Password: "",
		Status: models.USER_ACTIVE,
		FbId: facebookId,
	}

	//save user; if there is error; then show that error
	if err := Gdb.Save(&newUser).Error; err != nil {
		return c.RenderBadRequest(err)
	}

	//set the cache of the users
	RCache.Set(sessionKey, newUser, SessionExpire)

	//TODO: if there is redirect url; redirect him back to the main app; otherwise, log him in the Oauth CMS
	return c.RedirectAfterLoginSuccess(App.Home)
}

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
			return c.RedirectAfterLoginSuccess(App.Home)
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

	return c.RedirectAfterLoginSuccess(App.Home)
}


