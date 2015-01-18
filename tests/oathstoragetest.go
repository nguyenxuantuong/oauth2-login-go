//NOTE: all of the test method has to started with Test prefix; otherwise, it will not being run
package tests

import (
	"github.com/revel/revel"
	"fmt"
	"bytes"
	"strings"
	"gopkg.in/mgo.v2/bson"
	"auth/app/controllers"
	"auth/app/utils"
	"github.com/RangelReale/osin"
//	"net/http"
//	"net/url"
	"encoding/json"
)

var _ = fmt.Printf
var _  = bytes.Index
var _ = strings.Index

type OAuthStorageTest struct {
	revel.TestSuite
}

type Client  struct {
	Id      string
	Secret   string
	RedirectUri   string
}

var (
	newAuthorizeData osin.AuthorizeData
	newAccessData osin.AccessData
)

func (t *OAuthStorageTest) Before() {
	MongoTestDB = revel.Config.StringDefault("mgo.dbname", "oauth_go_test")
	endpoint, _ = revel.Config.String("http.endpoint")

	//drop the collections
	controllers.Session.DB(MongoTestDB).C(utils.CLIENT_COL).DropCollection()
	controllers.Session.DB(MongoTestDB).C(utils.AUTHORIZE_COL).DropCollection()
	controllers.Session.DB(MongoTestDB).C(utils.ACCESS_COL).DropCollection()
	
	newClient := osin.DefaultClient{
		Id: "1234",
		Secret: "aabbccdd",
		RedirectUri: "http://localhost:9000",
	}
	
	//other initialization
	newAuthorizeData = osin.AuthorizeData{
		Client: &newClient,
		Code: "OTI2NWUyYWEtOGJhYy00ZjhiLTk2MjItNGViOTVmNWEwZWEw",
		ExpiresIn: 250,
		Scope: "",
		RedirectUri: "http://127.0.0.1:8889",
		State: "everything",
	}
	
	newAccessData = osin.AccessData{
		Client: &newClient,
		AuthorizeData: &newAuthorizeData,
		AccessToken: "M2Q4MzRhMGUtZmFhNC00OTA5LTkzODUtN2YzYjk1YjFiYzhl",
		RefreshToken: "MmNiOGE5NTgtYjJkYy00NWFhLTliYWItYTI1NGMzYmM3OTMw",
		ExpiresIn: 3600,
		Scope: "",
		RedirectUri: "http://localhost:9000",
	}
}

func (t *OAuthStorageTest) After() {
	//drop the collections
	controllers.Session.DB(MongoTestDB).C(utils.CLIENT_COL).DropCollection()
	controllers.Session.DB(MongoTestDB).C(utils.AUTHORIZE_COL).DropCollection()
	controllers.Session.DB(MongoTestDB).C(utils.ACCESS_COL).DropCollection()
}

//test get set access tokens
func (t *OAuthStorageTest) TestGetSetAccessStorage(){
	oauthStorage = utils.NewOAuthStorage(controllers.Session, MongoTestDB)

	if err := oauthStorage.SaveAccess(&newAccessData); err != nil {
		t.AssertEqual(err, nil)
	}

	existingAccessData, err := oauthStorage.LoadAccess(newAccessData.AccessToken)
	t.AssertEqual(err, nil)

	t.AssertEqual(existingAccessData.AccessToken, newAccessData.AccessToken)
	t.AssertEqual(existingAccessData.Client.GetId(), newAccessData.Client.GetId())
	t.AssertEqual(existingAccessData.Client.GetSecret(), newAccessData.Client.GetSecret())

	//also check if loading refresh token method work as expected
	existingAccessData, err = oauthStorage.LoadAccess(newAccessData.AccessToken)
	t.AssertEqual(err, nil)

	t.AssertEqual(existingAccessData.AccessToken, newAccessData.AccessToken)
	t.AssertEqual(existingAccessData.Client.GetId(), newAccessData.Client.GetId())
	t.AssertEqual(existingAccessData.Client.GetSecret(), newAccessData.Client.GetSecret())

	oldRefreshToken := existingAccessData.RefreshToken
	
	//remove refresh token
	if err = oauthStorage.RemoveRefresh(existingAccessData.RefreshToken); err != nil {
		t.AssertEqual(err, nil)
	}
	
	existingAccessData, err = oauthStorage.LoadAccess(newAccessData.AccessToken)
	t.AssertEqual(err, nil)
	
	t.AssertNotEqual(existingAccessData.RefreshToken, oldRefreshToken)
	
	//now remove the existing authorize tokens
	if err := oauthStorage.RemoveAccess(newAccessData.AccessToken); err != nil {
		t.AssertEqual(err, nil)
	}

	//it will get error -- not found
	existingAccessData, err = oauthStorage.LoadAccess(newAccessData.AccessToken)
	t.AssertNotEqual(err, nil)
}

func (t *OAuthStorageTest) TestGetSetClientStorage(){
	//init the oauth storage and insert a sample client
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

	existingClient, err := oauthStorage.GetClient(newClient.Id)

	if err != nil {
		t.AssertEqual(err, nil)
	}

	//assert if both are equal
	t.AssertEqual(existingClient.GetSecret(), newClient.GetSecret())
	t.AssertEqual(existingClient.GetId(), newClient.GetId())
	t.AssertEqual(existingClient.GetRedirectUri(), newClient.GetRedirectUri())
}

func (t *OAuthStorageTest) TestGetSetAuthorizeStorage(){
	oauthStorage = utils.NewOAuthStorage(controllers.Session, MongoTestDB)
	
	if err := oauthStorage.SaveAuthorize(&newAuthorizeData); err != nil {
		t.AssertEqual(err, nil)
	}
	
	existingAuthorizeData, err := oauthStorage.LoadAuthorize(newAuthorizeData.Code)
	t.AssertEqual(err, nil)
	
	t.AssertEqual(existingAuthorizeData.Code, newAuthorizeData.Code)
	t.AssertEqual(existingAuthorizeData.Client.GetId(), newAuthorizeData.Client.GetId())
	t.AssertEqual(existingAuthorizeData.Client.GetSecret(), newAuthorizeData.Client.GetSecret())
	t.AssertEqual(existingAuthorizeData.RedirectUri, newAuthorizeData.RedirectUri)
	
	//now remove the existing authorize tokens
	if err := oauthStorage.RemoveAuthorize(newAuthorizeData.Code); err != nil {
		t.AssertEqual(err, nil)
	}

	//it will get error -- not found
	existingAuthorizeData, err = oauthStorage.LoadAuthorize(newAuthorizeData.Code)
	t.AssertNotEqual(err, nil)
}

//some raw functions
func (t *OAuthStorageTest) TestInsertAccessCode(){
	accessCol := controllers.Session.DB(MongoTestDB).C(utils.ACCESS_COL)
	
	if _, err := accessCol.UpsertId(newAccessData.AccessToken, &newAccessData); err != nil {
		t.AssertEqual(err, nil)
	}

	genericAccessData := make(map[string]interface{})

	if err := accessCol.FindId(newAccessData.AccessToken).One(&genericAccessData); err != nil {
		t.AssertEqual(err, nil)
	}

	jsonData, _ := json.Marshal(&genericAccessData)
	
	newClient := osin.DefaultClient{}
	authorizeData := osin.AuthorizeData{
		Client: &newClient,
	}
	
	existingAccessData := osin.AccessData{
		Client: &newClient,
		AuthorizeData: &authorizeData,
	}
	
	//now seriablize data
	if err := json.Unmarshal(jsonData, &existingAccessData); err != nil {
		t.AssertEqual(err, nil)
	}
	
	t.AssertEqual(existingAccessData.AuthorizeData.Client.GetId(), newClient.GetId())
	t.AssertEqual(existingAccessData.Client.GetId(), newClient.GetId())
	t.AssertEqual(existingAccessData.RefreshToken, newAccessData.RefreshToken)
}

func (t *OAuthStorageTest) TestInsertAuthorizeCode(){
	authorizations := controllers.Session.DB(MongoTestDB).C(utils.AUTHORIZE_COL)

	if _, err := authorizations.UpsertId(newAuthorizeData.Code, &newAuthorizeData); err != nil {
		t.AssertEqual(err, nil)
	}

	//we have to load it into a generic struct because the AuthorizeData contain the pointer to struct interface
	genericAuthorizeData := make(map[string]interface{})

	if err := authorizations.FindId(newAuthorizeData.Code).One(&genericAuthorizeData); err != nil {
		t.AssertEqual(err, nil)
	}

	jsonData, _ := json.Marshal(&genericAuthorizeData)

	//now need to seriablize data back to struct
	existingAuthorizeData := osin.AuthorizeData{
		Client: &osin.DefaultClient{},
	}

	if err := json.Unmarshal(jsonData, &existingAuthorizeData); err != nil {
		t.AssertEqual(err, nil)
	}

	//assert
	t.AssertEqual(existingAuthorizeData.Client.GetId(), newAuthorizeData.Client.GetId())
	t.AssertEqual(existingAuthorizeData.Client.GetSecret(), newAuthorizeData.Client.GetSecret())
	t.AssertEqual(existingAuthorizeData.Code, newAuthorizeData.Code)
	t.AssertEqual(existingAuthorizeData.RedirectUri, newAuthorizeData.RedirectUri)

	//try to unmarshall directly from mongo bson
	if err := authorizations.Find(bson.M{"code": newAuthorizeData.Code}).One(&existingAuthorizeData); err != nil {
		//there will be error when trying to unmarshal directly
		t.AssertNotEqual(err, nil)
	}
}

//insert client using mgo
func (t *OAuthStorageTest) TestInsertClient() {
	clients := controllers.Session.DB(MongoTestDB).C(utils.CLIENT_COL)
	newClient := Client{
		Id: "1234",
		Secret: "aabbccdd",
		RedirectUri: "http://localhost:9000",
	}

	//	if _, err := clients.UpsertId(newClient.Id, &newClient); err != nil {
	//		t.AssertEqual(err, nil)
	//	}

	if _, err := clients.Upsert(bson.M{"id": newClient.Id}, &newClient); err != nil {
		t.AssertEqual(err, nil)
	}

	existingClient := Client{}

	//finding using bson declaration
	clients.Find(bson.M{"id": newClient.Id}).One(&existingClient)

	//assert that new record has been inserted
	t.AssertEqual(existingClient.Id, newClient.Id)
	t.AssertEqual(existingClient.Secret, newClient.Secret)
	t.AssertEqual(existingClient.RedirectUri, newClient.RedirectUri)

	//find by ID
	clients.FindId(newClient.Id).One(&existingClient)

	t.AssertEqual(existingClient.Id, newClient.Id)
	t.AssertEqual(existingClient.Secret, newClient.Secret)
	t.AssertEqual(existingClient.RedirectUri, newClient.RedirectUri)
}
