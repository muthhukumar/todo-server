package internal

import (
	"fmt"
	"net/smtp"
	"os"
	"todo-server/models"
	"todo-server/utils"
)

func LoadEmailCredentials() models.EmailAuth {
	fromEmail := os.Getenv("FROM_EMAIL")
	password := os.Getenv("EMAIL_PASSWORD")
	toEmail := os.Getenv("TO_EMAIL")

	utils.Assert(fromEmail != "", "FROM_EMAIL env should not be empty")
	utils.Assert(password != "", "EMAIL_PASSWORD env should not be empty")
	utils.Assert(toEmail != "", "TO_EMAIL env should not be empty")

	return models.EmailAuth{FromEmail: fromEmail, Password: password, ToEmail: toEmail}

}

func SendEmail(emailAuth models.EmailAuth, emailTemplate models.EmailTemplate) (success bool) {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	addr := fmt.Sprintf("%v:%v", smtpHost, smtpPort)
	auth := smtp.PlainAuth("", emailAuth.FromEmail, emailAuth.Password, smtpHost)

	err := smtp.SendMail(addr, auth, emailAuth.FromEmail, emailTemplate.To, emailTemplate.GetMessage())

	if err != nil {
		fmt.Println("Sending email failed", err)

		return false
	}

	return true
}
