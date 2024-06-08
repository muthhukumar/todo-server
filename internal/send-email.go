package internal

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"todo-server/models"
)

func LoadEmailCredentials() models.EmailAuth {
	fromEmail := os.Getenv("FROM_EMAIL")
	password := os.Getenv("EMAIL_PASSWORD")
	toEmail := os.Getenv("TO_EMAIL")

	if fromEmail == "" {
		log.Fatal("FROM_EMAIL environment variable not set")
	}

	if password == "" {
		log.Fatal("EMAIL_PASSWORD environment variable not set")
	}

	if toEmail == "" {
		log.Fatal("EMAIL_PASSWORD environment variable not set")
	}

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
