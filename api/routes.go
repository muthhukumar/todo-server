package api

import (
	"encoding/json"
	"net/http"
	"todo-server/internal"
	"todo-server/models"

	"database/sql"

	"todo-server/utils"

	"github.com/go-chi/chi/v5"
)

type HandlerFn struct {
	DB *sql.DB
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Worlld!"))
}

func (h *HandlerFn) tasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "applications/json")

	rows, err := h.DB.Query("SELECT * FROM tasks")

	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server error.")
		return
	}

	var tasks []models.Task

	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.Name, &task.Completed, &task.CompletedOn); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server error.")

			return
		}
		tasks = append(tasks, task)
	}

	if len(tasks) == 0 {
		json.NewEncoder(w).Encode(models.Response{Data: []models.Task{}})

		return
	}

	json.NewEncoder(w).Encode(models.Response{Data: tasks})
}

func (h *HandlerFn) createTask(w http.ResponseWriter, r *http.Request) {
	var newTask models.Task

	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")

		return
	}

	if isValid, invalidFields := internal.ValidateTodo(newTask); !isValid {
		w.Header().Set("Content-Type", "applications/json")
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(models.Response{Data: invalidFields})
		// TODO - use the response with error function here too
		// utils.RespondWithError(w, http.StatusBadRequest, )

		return
	}

	query := `
	INSERT INTO tasks (name, completed, completed_on)
	VALUES ($1, $2, $3);
`

	rows := h.DB.QueryRow(query, newTask.Name, newTask.Completed, newTask.CompletedOn)

	if err := rows.Err(); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Creating Task failed.")

		return
	}

	w.Header().Set("Content-Type", "applications/json")

	json.NewEncoder(w).Encode(models.MsgResponse{Message: "Task created successfully"})

}

func SetupRoutes(r *chi.Mux, db *sql.DB) {
	r.Get("/api/v1/hello-world", helloWorld)

	routeHandler := HandlerFn{db}

	r.Get("/api/v1/tasks", routeHandler.tasks)
	r.Post("/api/v1/task/create", routeHandler.createTask)
}
