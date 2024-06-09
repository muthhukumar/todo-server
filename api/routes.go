package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"todo-server/internal"
	"todo-server/models"

	"database/sql"

	"todo-server/utils"

	"github.com/go-chi/chi/v5"
)

type HandlerFn struct {
	DB *sql.DB
}

func (h *HandlerFn) healthCheck(w http.ResponseWriter, r *http.Request) {
	query := "select 1"

	if _, err := h.DB.Query(query); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		fmt.Println("DB Error: ", err.Error())

		w.Write([]byte("NOT OK"))
		return
	}

	w.WriteHeader(http.StatusOK)

	w.Write([]byte("OK"))
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World!"))
}

func (h *HandlerFn) tasks(w http.ResponseWriter, r *http.Request) {

	filter := r.URL.Query().Get("filter")

	var query string
	var args []interface{}

	switch filter {
	case "":
		query = "SELECT * FROM tasks ORDER BY created_at DESC"
	case "my-day":
		today := time.Now().Format("2006-01-02")
		query = "SELECT * FROM tasks WHERE marked_today != '' AND DATE(marked_today) = $1 ORDER BY created_at DESC"
		args = []interface{}{today}
	case "important":
		query = "SELECT * FROM tasks where is_important = true"
	default:
		query = "SELECT * FROM tasks ORDER BY created_at DESC"
	}

	rows, err := h.DB.Query(query, args...)

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	var tasks []models.Task

	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.Name, &task.Completed, &task.CompletedOn, &task.CreatedAt, &task.MarkedToday, &task.IsImportant, &task.DueDate); err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, models.MsgResponse{Message: err.Error()})

			return
		}
		tasks = append(tasks, task)
	}
	defer rows.Close()

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

func (h *HandlerFn) updateTask(w http.ResponseWriter, r *http.Request) {
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

func (h *HandlerFn) deleteTask(w http.ResponseWriter, r *http.Request) {
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

func (h *HandlerFn) toggleTask(w http.ResponseWriter, r *http.Request) {
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

func (h *HandlerFn) toggleImportant(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Task ID is not valid."})
		return
	}

	query := `
update
	tasks
set
	is_important = not is_important
where
	id = $1;
	`

	result, err := h.DB.Exec(query, id)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Toggling task importance failed"})

		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: fmt.Sprintf("Task with ID {%v} does not exist.", id)})

		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: "Toggled Task's important"})
}

func (h *HandlerFn) toggleAddToMyToday(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, id_err := strconv.Atoi(idStr)

	if id_err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Invalid task ID"})

		return
	}

	query := `
	UPDATE tasks
	SET marked_today = CASE
													WHEN marked_today = '' THEN CURRENT_TIMESTAMP::TEXT
													WHEN marked_today::DATE != CURRENT_DATE THEN CURRENT_TIMESTAMP::TEXT
													ELSE ''
											END
	WHERE id = $1;
	`

	result, err := h.DB.Exec(query, id)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Toggling task Add to my day failed"})

		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: fmt.Sprintf("Task with ID {%v} does not exist.", id)})

		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: "Toggled Task's Add to my day"})

}

func (h *HandlerFn) addDueDate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Invalid Task ID"})
		return
	}

	var task models.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Invalid request body"})

		return
	}

	_, err = time.Parse("2006-01-02", task.DueDate)

	if err != nil && task.DueDate != "" {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Invalid Due Date"})

		return
	}

	query := "update tasks set due_date=$1 where id = $2"

	result, err := h.DB.Exec(query, task.DueDate, id)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Updating Task Due date failed"})

		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		utils.JsonResponse(w, http.StatusNotFound, models.MsgResponse{Message: fmt.Sprintf("Task with %v ID does not exist", id)})
		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: "Due date updated successfully"})
}

func SetupRoutes(r *chi.Mux, db *sql.DB) {
	routeHandler := HandlerFn{db}

	r.Get("/health", routeHandler.healthCheck)

	r.Get("/api/v1/hello-world", helloWorld)

	r.Get("/api/v1/tasks", routeHandler.tasks)
	r.Post("/api/v1/task/create", routeHandler.createTask)
	r.Post("/api/v1/task/{id}", routeHandler.updateTask)
	r.Delete("/api/v1/task/{id}", routeHandler.deleteTask)
	r.Post("/api/v1/task/{id}/add/due-date", routeHandler.addDueDate)

	r.Post("/api/v1/task/{id}/completed/toggle", routeHandler.toggleTask)
	r.Post("/api/v1/task/{id}/important/toggle", routeHandler.toggleImportant)
	r.Post("/api/v1/task/{id}/add-to-my-day/toggle", routeHandler.toggleAddToMyToday)
}
