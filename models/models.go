package models

import (
	"fmt"
)

type Task struct {
	ID          int    `json:"id"`
	Name        string `json:"name" validate:"required,min=3,max=1000"`
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

type QuotesResponse struct {
	Quotes []string `json:"quotes"`
	Size   int      `json:"size"`
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
	ToEmail   string
}

type EmailTemplate struct {
	To      []string
	Body    string
	Subject string
}

type ErrorResponseV2 struct {
	Status        int            `json:"status"`
	Code          string         `json:"code"`
	Message       string         `json:"message"`
	RequestId     string         `json:"request_id"`
	InvalidFields []InvalidField `json:"invalid_fields"`
}

type InvalidField struct {
	ErrorMessage string `json:"error_message"`
	Field        string `json:"field"`
	IsValid      bool   `json:"is_invalid"`
}

func (e *EmailTemplate) GetMessage() (msg []byte) {
	to := fmt.Sprintf("To: %v\r\n", e.To[0])
	subject := fmt.Sprintf("Subject: %v\r\n", e.Subject)
	body := fmt.Sprintf("%v\r\n", e.Body)

	msg = []byte(to + subject + "\r\n" + body)

	return
}
