//sending email using mandrill
package emails

import (
//	"errors"
	"github.com/mostafah/mandrill"
	"github.com/revel/revel"
	"strings"
)

//Mandrill Sender implement Mandrill interface
type MandrillSender struct {
	ApiKey string
}

//return instance of mandrill sender -- do nothing
func NewMandrillSender() MandrillSender {
	apiKey, ok := revel.Config.String("mandrill.apikey");
	
	if !ok {
		revel.ERROR.Println("Mandrill api key is missing from config file")
	}
	
	//now assign api key to mandrill
	mandrill.Key = apiKey
	err := mandrill.Ping()
	
	if err != nil {
		revel.ERROR.Println("unable to ping mandrill")
	}
	
	return MandrillSender{apiKey}	
}

//send function
func (sender MandrillSender) Send(emailType EmailType, emailInfo EmailInfo, placeHolder EmailPlaceHolder) error {
	var emailContent string
	
	//TODO: replace placeholder
	switch (emailType){
	case AccountActivation:
		emailContent = EmailTemplates.AccountActivation
	case PasswordReset:
		emailContent = EmailTemplates.PasswordReset
	case WelcomeUser:
		emailContent = EmailTemplates.WelcomeUser
	}
	
	//now trying to replace the email content
	if placeHolder.UserName != "" {
		emailContent = strings.Replace(emailContent, "[USERNAME]", placeHolder.UserName, -1)
	}
	
	if placeHolder.URL != ""{
		emailContent = strings.Replace(emailContent, "[URL]", placeHolder.URL, -1)
	}
	
	msg := mandrill.NewMessageTo(emailInfo.ToEmail, emailInfo.ToName)
	
	//init message details
	msg.HTML = emailContent
	msg.Text = emailContent
	msg.Subject = emailInfo.Subject
	msg.FromEmail = emailInfo.FromEmail
	msg.FromName = emailInfo.ToEmail
	
	res, err := msg.Send(false)
	
	//sending results
	if len(res) >= 1 {
		revel.INFO.Printf("Sending result %+v", res[0])
	}

	return err
}



