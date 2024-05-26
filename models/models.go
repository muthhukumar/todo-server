package models

type Task struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Completed bool   `json:"completed"`

	CompletedOn string `json:"completed_on"`
	CreatedAt   string `json:"created_at"`
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
