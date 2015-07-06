//NOTE: all of the test method has to started with Test prefix; otherwise, it will not being run
package tests

import (
	"github.com/revel/revel"
	"fmt"
	"bytes"
	"strings"
//	"gopkg.in/mgo.v2"
//	"gopkg.in/mgo.v2/bson"
	"auth/app/controllers"
	"auth/app/utils"
	"github.com/RangelReale/osin"
	"net/http"
	"net/url"
	"encoding/base64"
	"github.com/parnurzeal/gorequest"
	"encoding/json"
	"auth/app/models"
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/revel/revel/testing"
)

var _ = fmt.Printf
var _  = bytes.Index
var _ = strings.Index

type OAuthTest struct {
	testing.TestSuite
}

var (
	oauthStorage *utils.OAuthStorage
	MongoTestDB string
	newClient osin.DefaultClient
)

func (t *OAuthTest) Before() {
	MongoTestDB = revel.Config.StringDefault("mgo.dbname", "oauth_go_test")
	endpoint, _ = revel.Config.String("http.endpoint")
	
	//drop the collections
	controllers.Session.DB(MongoTestDB).C(utils.CLIENT_COL).DropCollection()
	controllers.Session.DB(MongoTestDB).C(utils.AUTHORIZE_COL).DropCollection()
	controllers.Session.DB(MongoTestDB).C(utils.ACCESS_COL).DropCollection()
	controllers.Gdb.Exec("TRUNCATE TABLE user;")
}

func (t *OAuthTest) After() {
	//drop the collections
	controllers.Session.DB(MongoTestDB).C(utils.CLIENT_COL).DropCollection()
	controllers.Session.DB(MongoTestDB).C(utils.AUTHORIZE_COL).DropCollection()
	controllers.Session.DB(MongoTestDB).C(utils.ACCESS_COL).DropCollection()
	controllers.Gdb.Exec("TRUNCATE TABLE user;")
}


func (t *OAuthTest) RequestAuthorizationCode(oauthStorage *utils.OAuthStorage) (*osin.AuthorizeData, error) {
	newClient = osin.DefaultClient{
		Id: "1234",
		Secret: "aabbccdd",
		RedirectUri: "http://localhost:9000",
	}

	//function setclient has interface (NOTE: with interface, you can pass pointer directly)
	if err := oauthStorage.SetClient(newClient.Id, &newClient); err != nil {
		t.AssertEqual(err, nil)
	}

	req, err := http.NewRequest("GET", endpoint + "/api/oauth/authorize", nil)
	if err != nil {
		t.AssertEqual(err, nil)
	}

	//attach some params in the url parameter
	req.Form = make(url.Values)
	req.Form.Set("response_type", string(osin.CODE))
	req.Form.Set("client_id", newClient.GetId())
	req.Form.Set("state", "everything")

	//initialize server and try the handle authorize request, etc
	resp := controllers.OAuthServer.NewResponse()
	if ar := controllers.OAuthServer.HandleAuthorizeRequest(resp, req); ar != nil {
		ar.Authorized = true
		controllers.OAuthServer.FinishAuthorizeRequest(resp, req, ar)
	}

	code, found := resp.Output["code"].(string)

	t.AssertEqual(found, true)

	authorizationData, err := oauthStorage.LoadAuthorize(code)
	if err != nil {
		t.AssertEqual(err, nil)
	}
	
	return authorizationData, err
}

//test get authorize code
func (t *OAuthTest) TestAuthorizeCode(){
	//init the oauth storage and insert a sample client
	oauthStorage = utils.NewOAuthStorage(controllers.Session, MongoTestDB)

	//now try to get the authorize code
	authorizationData, err := t.RequestAuthorizationCode(oauthStorage)
	
	if err != nil {
		t.AssertEqual(err, nil)		
	}		
	
	//assert the data
	t.AssertEqual(authorizationData.Client.GetId(), newClient.GetId())
	t.AssertEqual(authorizationData.Client.GetSecret(), newClient.GetSecret())
}

//test get access code
func (t *OAuthTest) TestAccessCode(){
	//init the oauth storage and insert a sample client
	oauthStorage = utils.NewOAuthStorage(controllers.Session, MongoTestDB)

	//now try to get the authorize code
	authorizationData, err := t.RequestAuthorizationCode(oauthStorage)

	if err != nil {
		t.AssertEqual(err, nil)
	}

	req, err := http.NewRequest("POST", endpoint + "/api/oauth/token", nil)
	if err != nil {
		t.AssertEqual(err, nil)
	}
	
	//set clientId, clientSecret in authorization request
	req.SetBasicAuth(authorizationData.Client.GetId(), authorizationData.Client.GetSecret())

	req.Form = make(url.Values)
	req.Form.Set("grant_type", string(osin.AUTHORIZATION_CODE))
	req.Form.Set("code", authorizationData.Code)
	req.Form.Set("state", "everything")
	req.PostForm = make(url.Values)

	//initialize request
	resp := controllers.OAuthServer.NewResponse()
	if ar := controllers.OAuthServer.HandleAccessRequest(resp, req); ar != nil {
		ar.Authorized = true
		controllers.OAuthServer.FinishAccessRequest(resp, req, ar)
	}

	//verify that access_token and refresh_token has been created
	accessToken, found := resp.Output["access_token"].(string)
	t.AssertEqual(found, true)
	
	refreshToken, found := resp.Output["refresh_token"].(string)
	t.AssertEqual(found, true)

	existingAccessData, err := oauthStorage.LoadAccess(accessToken)
	t.AssertEqual(err, nil)
	
	existingAccessDataCloned, err := oauthStorage.LoadRefresh(refreshToken)
	t.AssertEqual(err, nil)
	
	t.AssertEqual(existingAccessData.AccessToken, existingAccessDataCloned.AccessToken)
}


//authorization + refresh_token grant type
func (t *OAuthTest) TestRequestAccessTokenAuthorizationGrant(){
	oauthStorage = utils.NewOAuthStorage(controllers.Session, MongoTestDB)

	//now try to get the authorize code
	authorizationData, err := t.RequestAuthorizationCode(oauthStorage)

	if err != nil {
		t.AssertEqual(err, nil)
	}
	
	request := gorequest.New()
	
	//request directly from token endpoint
	_,body,_ := request.Post(endpoint + "/api/oauth/token").
	Type("form").
	Set("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(authorizationData.Client.GetId() + ":" + authorizationData.Client.GetSecret()))).
	Send(`{"grant_type":"` + string(osin.AUTHORIZATION_CODE) +`"}`).
	Send(`{"code":"` + authorizationData.Code +`"}`).
	Send(`{"state":"everything"}`).
	End()

	//unmarshal json
	var m map[string]interface{}
	err = json.Unmarshal([]byte(body), &m)
	
	t.AssertEqual(err, nil)
	
	//attract token from body
	accessToken, found := m["access_token"].(string)
	t.AssertEqual(found, true)
	refreshToken, found := m["refresh_token"].(string)
	t.AssertEqual(found, true)

	//verify that access token exist in database
	existingAccessData, err := oauthStorage.LoadAccess(accessToken)
	t.AssertEqual(err, nil)
	
	t.AssertEqual(existingAccessData.AuthorizeData.Code, authorizationData.Code)
	t.AssertEqual(existingAccessData.RefreshToken, refreshToken)
	
	//verify that authorize token has been deleted from database
	_, err = oauthStorage.LoadAuthorize(authorizationData.Code)
	t.AssertNotEqual(err, nil)
	
	//now trying to send the same request -- but with refresh token grant type
	_,body,_ = request.Post(endpoint + "/api/oauth/token").
	Type("form").
	Set("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(authorizationData.Client.GetId() + ":" + authorizationData.Client.GetSecret()))).
	Send(`{"grant_type":"` + string(osin.REFRESH_TOKEN) +`"}`).
	Send(`{"refresh_token":"` + refreshToken +`"}`).
	Send(`{"state":"everything"}`).
	End()

	err = json.Unmarshal([]byte(body), &m)

	t.AssertEqual(err, nil)

	//new access_token should be returned
	accessTokenNew, found := m["access_token"].(string)
	t.AssertEqual(found, true)
	refreshTokenNew, found := m["refresh_token"].(string)
	t.AssertEqual(found, true)

	//both of this will exist
	existingAccessData2, err := oauthStorage.LoadAccess(accessTokenNew)
	t.AssertEqual(err, nil)
	
	existingAccessData3, err := oauthStorage.LoadRefresh(refreshTokenNew)
	t.AssertEqual(err, nil)
	
	//both of them are actually the same
	t.AssertEqual(existingAccessData2.AccessToken, existingAccessData3.AccessToken)
}


//client credential grant_type
func (t *OAuthTest) TestRequestAccessTokenClientGrant(){
	oauthStorage = utils.NewOAuthStorage(controllers.Session, MongoTestDB)
	
	newClient = osin.DefaultClient{
		Id: "1234",
		Secret: "aabbccdd",
		RedirectUri: "http://localhost:9000",
	}

	//function setclient has interface (NOTE: with interface, you can pass pointer directly)
	if err := oauthStorage.SetClient(newClient.Id, &newClient); err != nil {
		t.AssertEqual(err, nil)
	}

	request := gorequest.New()

	//request directly from token endpoint
	_,body,_ := request.Post(endpoint + "/api/oauth/token").
	Type("form").
	Set("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(newClient.GetId() + ":" + newClient.GetSecret()))).
	Send(`{"grant_type":"` + string(osin.CLIENT_CREDENTIALS) +`"}`).
	Send(`{"state":"everything"}`).
	End()
	
	//unmarshal json
	var m map[string]interface{}
	err := json.Unmarshal([]byte(body), &m)

	t.AssertEqual(err, nil)

	//attract token from body
	accessToken, found := m["access_token"].(string)
	t.AssertEqual(found, true)
	
	//NO Refresh_Token is granted
	_, found = m["refresh_token"].(string)
	t.AssertEqual(found, false)

	//verify that access token exist in database
	existingAccessData, err := oauthStorage.LoadAccess(accessToken)
	t.AssertEqual(err, nil)
	t.AssertEqual(existingAccessData.AccessToken, accessToken)
}

func (t *OAuthTest) TestRequestAccessTokenPasswordGrant(){
	oauthStorage = utils.NewOAuthStorage(controllers.Session, MongoTestDB)

	newClient = osin.DefaultClient{
		Id: "1234",
		Secret: "aabbccdd",
		RedirectUri: "http://localhost:9000",
	}

	//function setclient has interface (NOTE: with interface, you can pass pointer directly)
	if err := oauthStorage.SetClient(newClient.Id, &newClient); err != nil {
		t.AssertEqual(err, nil)
	}

	request := gorequest.New()

	newUser := models.User{
		FullName: "Nguyen Xuan Tuong",
		Email: "nguy0066@e.ntu.edu.sg",
		UserName: "nguy0066",
		Password: utils.GetMD5Hash("111111"),
		Status: models.USER_ACTIVE,
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newUser.Password), 10)
	newUser.HashedPassword = hashedPassword;

	if err := controllers.Gdb.Save(&newUser).Error; err != nil {
		t.AssertEqual(err, nil)
	}

	//then login using the same credential
	_, body, _ := request.Post(endpoint + "/api/user/login").
	Send(`{"Email":"nguy0066@e.ntu.edu.sg"}`).
	Send(`{"Password":"` + newUser.Password +`"}`).
	End()
	
	//now request token using PassWord grant type
	_,body,_ = request.Post(endpoint + "/api/oauth/token").
	Type("form").
	Set("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(newClient.GetId() + ":" + newClient.GetSecret()))).
	Send(`{"grant_type":"` + string(osin.PASSWORD) +`"}`).
	Send(`{"username":"` + string(newUser.Email) +`"}`).
	Send(`{"password":"` + string(newUser.Password) +`"}`).
	Send(`{"state":"everything"}`).
	End()

	//unmarshal json
	var m map[string]interface{}
	err := json.Unmarshal([]byte(body), &m)

	t.AssertEqual(err, nil)

	//attract token from body
	accessToken, found := m["access_token"].(string)
	t.AssertEqual(found, true)

	//NO Refresh_Token is granted
	_, found = m["refresh_token"].(string)
	t.AssertEqual(found, true)

	//verify that access token exist in database
	existingAccessData, err := oauthStorage.LoadAccess(accessToken)
	t.AssertEqual(err, nil)
	t.AssertEqual(existingAccessData.AccessToken, accessToken)
}
