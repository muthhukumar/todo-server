package internal

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"todo-server/models"
	"todo-server/utils"
)

func LoadEmailCredentials() models.EmailAuth {
	fromEmail := os.Getenv("FROM_EMAIL")
	password := os.Getenv("EMAIL_PASSWORD")
	toEmail := os.Getenv("TO_EMAIL")

	utils.Assert(fromEmail != "", "FROM_EMAIL env is set")
	utils.Assert(password != "", "EMAIL_PASSWORD env is set")
	utils.Assert(toEmail != "", "TO_EMAIL env is set")

	return models.EmailAuth{FromEmail: fromEmail, Password: password, ToEmail: toEmail}

}

func SendEmail(emailAuth models.EmailAuth, emailTemplate models.EmailTemplate) (success bool) {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	addr := fmt.Sprintf("%v:%v", smtpHost, smtpPort)
	auth := smtp.PlainAuth("", emailAuth.FromEmail, emailAuth.Password, smtpHost)

	err := smtp.SendMail(addr, auth, emailAuth.FromEmail, emailTemplate.To, emailTemplate.GetMessage())

	if err != nil {
		log.Println("Sending email failed", err)

		return false
	}

	return true
}

func SendHtmlEmail(emailAuth models.EmailAuth, to []string, emailTemplate []byte) (success bool) {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	addr := fmt.Sprintf("%v:%v", smtpHost, smtpPort)
	auth := smtp.PlainAuth("", emailAuth.FromEmail, emailAuth.Password, smtpHost)

	err := smtp.SendMail(addr, auth, emailAuth.FromEmail, to, emailTemplate)

	if err != nil {
		log.Println("Sending email failed", err)

		return false
	}

	return true
}
