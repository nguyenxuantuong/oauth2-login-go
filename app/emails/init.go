package emails

import (
//	"strings"
	"io/ioutil"
	"github.com/revel/revel"
	"errors"
)

type EmailType int

//enum of email types
const (
	AccountActivation EmailType = 1 + iota
	PasswordReset	
	WelcomeUser
)

//struct contain email content of email templates
type EmailTemplateContents struct {
	AccountActivation string
	PasswordReset	string
	WelcomeUser		string
}

//struct contains necessary information for sending email
type EmailInfo struct {
	ToEmail string
	ToName string
	Subject	string
	FromEmail string
	FromName  string
}

//struct contains place-holder content for the templates
type EmailPlaceHolder struct {
	UserName	string
	URL			string
	UserEmail	string
}

//local variable contain content of email to be sent
var EmailTemplates EmailTemplateContents

//this interface contains sugar method for sending email -- can be Mandrill, normal sender, etc...
type EmailSender interface {
	Send(emailType EmailType, emailInfo EmailInfo, placeHolder EmailPlaceHolder) error
}

//Instance to keep the instance of EmailSender interface
var (
	Instance EmailSender

	ErrNotSent = errors.New("email. email was not sent successfully")
)

//sending email templates using placeholder
func Send(emailType EmailType, emailInfo EmailInfo, placeHolder EmailPlaceHolder) error  { return Instance.Send(emailType, emailInfo, placeHolder) }

//init function
func Init(){
	loadEmailContent()	
	
	//init the instance of EmailSender interface
	//by default using Mandrill sender over-here
	Instance = NewMandrillSender()
}

//load email content into template struct
func loadEmailContent() {
	//account activation
	emailActivation, err := ioutil.ReadFile(revel.BasePath + "/app/emails/tpls/email-activation.html")
	
	if err != nil {
		revel.ERROR.Println(err)
	}
	
	EmailTemplates.AccountActivation = string(emailActivation)

	//password reset
	passwordReset, err := ioutil.ReadFile(revel.BasePath + "/app/emails/tpls/email-password-reset.html")

	if err != nil {
		revel.ERROR.Println(err)
	}

	EmailTemplates.PasswordReset = string(passwordReset);
	
	//email welcome
	emailWelcome, err := ioutil.ReadFile(revel.BasePath + "/app/emails/tpls/email-welcome.html")

	if err != nil {
		revel.ERROR.Println(err)
	}

	EmailTemplates.WelcomeUser = string(emailWelcome);
}

