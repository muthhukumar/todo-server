package models

import (
	"fmt"
	"time"
)

type Task struct {
	ID                     int       `json:"id"`
	Name                   string    `json:"name" validate:"required,min=3,max=1000"`
	Completed              bool      `json:"completed"`
	CompletedOn            string    `json:"completed_on"`
	CreatedAt              string    `json:"created_at"`
	IsImportant            bool      `json:"is_important"`
	MarkedToday            string    `json:"marked_today"`
	DueDate                string    `json:"due_date"`
	Metadata               string    `json:"metadata"`
	SubTasks               []SubTask `json:"sub_tasks"`
	InCompleteSubTaskCount int       `json:"incomplete_subtask_count"`
	SubTaskCount           int       `json:"subtask_count"`
}

type SubTask struct {
	ID        int       `json:"id"`
	Name      string    `json:"name" validate:"required,min=3,max=1000"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	TaskID    int       `json:"task_id"` // TODO: this is required
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

type CreateTaskResponse struct {
	Message string `json:"message"`
	ID      int    `json:"id"`
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
	Error         string         `json:"error"`
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

type URLTitle struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	IsValid   bool   `json:"is_valid"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Log struct {
	ID        int    `json:"id"`
	Log       string `json:"log" validate:"required,min=3,max=1000"`
	Level     string `json:"level" validate:"required"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type LogPayload struct {
	Data []Log `json:"data"`
}
