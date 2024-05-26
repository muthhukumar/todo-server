package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	w.Write([]byte("Hello World!"))
}

func (h *HandlerFn) tasks(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT * FROM tasks")

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, models.MsgResponse{Message: "Internal Server error."})
		return
	}

	var tasks []models.Task

	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.Name, &task.Completed, &task.CompletedOn, &task.CreatedAt); err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, err)

			return
		}
		tasks = append(tasks, task)
	}

	if len(tasks) == 0 {
		utils.JsonResponse(w, http.StatusOK, models.Response{Data: []models.Task{}})

		return
	}

	utils.JsonResponse(w, http.StatusOK, models.Response{Data: tasks})
}

func (h *HandlerFn) createTask(w http.ResponseWriter, r *http.Request) {
	var newTask models.Task

	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Invalid request body"})

		return
	}

	if isValid, invalidFields := internal.ValidateTodo(newTask); !isValid {
		utils.JsonResponse(w, http.StatusBadRequest, models.Response{Data: invalidFields})

		return
	}

	query := `
	INSERT INTO tasks (name, completed, completed_on)
	VALUES ($1, $2, $3);
`

	rows := h.DB.QueryRow(query, newTask.Name, newTask.Completed, newTask.CompletedOn)

	if err := rows.Err(); err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, "Creating Task failed.")

		return
	}

	utils.JsonResponse(w, http.StatusCreated, models.MsgResponse{Message: "Task created successfully"})
}

func (h HandlerFn) updateTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, id_err := strconv.Atoi(idStr)

	if id_err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Invalid task ID"})
	}

	var task models.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Invalid request body"})

		return
	}

	query := "UPDATE tasks SET name=$1 WHERE id=$2"

	result, err := h.DB.Exec(query, task.Name, id)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: fmt.Sprintf("Updating task with ID {%v} failed.", id)})
		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		// TODO - check whether not found here is okay
		utils.JsonResponse(w, http.StatusNotFound, models.MsgResponse{Message: fmt.Sprintf("Updating task with ID {%v} failed. Task may not be available.", id)})
		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: "Updated Task successfully."})
}

func (h HandlerFn) deleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, id_err := strconv.Atoi(idStr)

	if id_err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Invalid task ID"})

		return
	}

	result, error := h.DB.Exec("DELETE FROM tasks where id = $1", id)

	if error != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Deleting task failed"})

		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: fmt.Sprintf("Task either already deleted or task with ID {%v} does not exist.", id)})

		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: fmt.Sprintf("Deleted task with ID {%v} successfully.", id)})
}

func (h HandlerFn) toggleTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, id_err := strconv.Atoi(idStr)

	if id_err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Invalid task ID"})

		return
	}

	query := `
	UPDATE tasks 
	SET completed = NOT completed, 
			completed_on = CASE 
					WHEN NOT completed THEN TO_CHAR(CURRENT_TIMESTAMP, 'YYYY-MM-DD HH24:MI:SS')
					ELSE '' 
			END 
	WHERE id = $1
	`

	result, error := h.DB.Exec(query, id)

	if error != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Toggling task failed"})

		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: fmt.Sprintf("Task with ID {%v} does not exist.", id)})

		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: fmt.Sprintf("Toggled task with ID {%v} successfully.", id)})
}

func SetupRoutes(r *chi.Mux, db *sql.DB) {
	r.Get("/api/v1/hello-world", helloWorld)

	routeHandler := HandlerFn{db}

	r.Get("/api/v1/tasks", routeHandler.tasks)
	r.Post("/api/v1/task/create", routeHandler.createTask)
	r.Post("/api/v1/task/{id}", routeHandler.updateTask)
	r.Delete("/api/v1/task/{id}", routeHandler.deleteTask)
	r.Post("/api/v1/task/{id}/toggle", routeHandler.toggleTask)
}
