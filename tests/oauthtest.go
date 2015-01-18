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
//	"encoding/json"
)

var _ = fmt.Printf
var _  = bytes.Index
var _ = strings.Index

type OAuthTest struct {
	revel.TestSuite
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
}

func (t *OAuthTest) After() {
	//drop the collections
	controllers.Session.DB(MongoTestDB).C(utils.CLIENT_COL).DropCollection()
	controllers.Session.DB(MongoTestDB).C(utils.AUTHORIZE_COL).DropCollection()
}


func (t *OAuthTest) TestAuthorizeCode(){
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
	
	//now try to get the authorize code
	authorizationData, err := oauthStorage.LoadAuthorize(code)
	if err != nil {
		t.AssertEqual(err, nil)		
	}		
	
	//assert the data
	t.AssertEqual(authorizationData.Client.GetId(), newClient.GetId())
	t.AssertEqual(authorizationData.Client.GetSecret(), newClient.GetSecret())
	t.AssertEqual(authorizationData.Code, code)
}



