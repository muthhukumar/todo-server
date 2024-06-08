package models

import "fmt"

type Task struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Completed   bool   `json:"completed"`
	CompletedOn string `json:"completed_on"`
	CreatedAt   string `json:"created_at"`
	IsImportant bool   `json:"is_important"`
	MarkedToday string `json:"marked_today"`
	DueDate     string `json:"due_date"`
}

type Response struct {
	Data any `json:"data"`
}

type MsgResponse struct {
	Message string `json:"message"`
}

type FieldValidation struct {
	IsValid      bool   `json:"is_valid"`
	Field        string `json:"field"`
	ErrorMessage string `json:"error_message"`
}

type EmailAuth struct {
	FromEmail string
	Password  string
}

type EmailTemplate struct {
	To      []string
	Body    string
	Subject string
}

func (e *EmailTemplate) GetMessage() (msg []byte) {
	to := fmt.Sprintf("To: %v\r\n", e.To[0])
	subject := fmt.Sprintf("Subject: %v\r\n", e.Subject)
	body := fmt.Sprintf("%v\r\n", e.Body)

	msg = []byte(to + subject + "\r\n" + body)

	return
}
