package tests

import (
	"github.com/revel/revel"
	"auth/app/emails"
//	"github.com/revel/revel/cache"
	"fmt"
	"github.com/revel/revel/testing"
)

var _ = fmt.Printf

type AppTest struct {
	testing.TestSuite
}

func (t *AppTest) Before() {
	//TODO: some setup
}

func (t *AppTest) TestThatIndexPageWorks() {
	t.Get("/")
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")
}

func (t *AppTest) TestEmailTemplates(){
	emailInfo := emails.EmailInfo{
		ToEmail: "xuan_tuong@mailinator.com",
		ToName: "Nguyen Xuan Tuong",
		Subject: "Account Activation",
		FromEmail: "noreply@auth.com",
		FromName: "Auth Team",
	}
	
	emailPlaceHolder := emails.EmailPlaceHolder{
		URL: "http://localhost:8888/accountActivation/1234",
		UserName: "Nguyen Xuan Tuong",
	}
	
	//now sending email
	err := emails.Send(emails.AccountActivation, emailInfo, emailPlaceHolder)
	if err != nil {
		revel.ERROR.Printf("error happen when sending email %s", err)
	}
}

func (t *AppTest) After() {
	//TODO: some teardown
}
