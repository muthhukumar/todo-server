package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	data "todo-server/data/quotes"
	"todo-server/internal"
	"todo-server/models"

	"database/sql"

	"todo-server/utils"

	"github.com/go-chi/chi/v5"
)

type HandlerFn struct {
	DB *sql.DB
}

func healthCheckWithDB(w http.ResponseWriter, r *http.Request) {
	query := "select 1"

	db := internal.SetupDatabase()
	defer db.Close()

	if _, err := db.Query(query); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		fmt.Println("DB Error: ", err.Error())

		w.Write([]byte("NOT OK"))
		return
	}

	w.WriteHeader(http.StatusOK)

	w.Write([]byte("OK"))
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("OK"))
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World!"))
}

func (h *HandlerFn) getTask(w http.ResponseWriter, r *http.Request) {
	taskId := chi.URLParam(r, "id")

	id, err := strconv.Atoi(taskId)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Task ID is not valid"})
		return
	}

	query := "select * from tasks where ID=$1"

	row := h.DB.QueryRow(query, id)

	var task models.Task

	if err := row.Scan(&task.ID, &task.Name, &task.Completed, &task.CompletedOn, &task.CreatedAt, &task.MarkedToday, &task.IsImportant, &task.DueDate); err != nil {
		utils.JsonResponse(w, http.StatusNotFound, models.MsgResponse{Message: fmt.Sprintf("Task with ID '%v' not found", id)})
		return
	}

	utils.JsonResponse(w, http.StatusOK, models.Response{Data: task})
}

func (h *HandlerFn) tasks(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	searchTerm := r.URL.Query().Get("query")
	showCompleted := r.URL.Query().Get("showCompleted")

	var query string
	var args []interface{}
	var completedFilter string

	// Determine the completed filter condition based on the showCompleted parameter
	switch showCompleted {
	case "true":
		completedFilter = " "
	case "false":
		completedFilter = " completed = false "
	default:
		completedFilter = " " // No filter for completed status
	}

	switch filter {
	case "":
		query = "SELECT * FROM tasks"
	case "my-day":
		today := time.Now().Format("2006-01-02")
		query = "SELECT * FROM tasks WHERE ((marked_today != '' AND DATE(marked_today) = $1) OR (due_date != '' AND DATE(due_date) = $1)) "
		args = append(args, today)
	case "important":
		query = "SELECT * FROM tasks where is_important = true"
	default:
		query = "SELECT * FROM tasks"
	}

	// If show completed is true then we don't have to add the filter.
	if showCompleted != "" && showCompleted == "false" {
		if len(args) > 0 || filter == "important" || filter == "my-day" {
			query += " AND"
		} else {
			query += " WHERE"
		}

		query += completedFilter
	}

	if searchTerm != "" {
		if len(args) > 0 || filter == "important" || filter == "my-day" || (showCompleted != "" && showCompleted == "false") {
			query += " AND"
		} else {
			query += " WHERE"
		}
		if filter == "my-day" {
			query += " name ILIKE '%' || $2 || '%'"
		} else {
			query += " name ILIKE '%' || $1 || '%'"
		}

		args = append(args, searchTerm)
	}

	query += " ORDER BY created_at DESC"

	rows, err := h.DB.Query(query, args...)

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, err.Error())
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
	INSERT INTO tasks (name, completed, completed_on, marked_today, is_important, due_date)
	VALUES ($1, $2, $3, $4, $5, $6);
`

	rows := h.DB.QueryRow(query, newTask.Name, newTask.Completed, newTask.CompletedOn, newTask.MarkedToday, newTask.IsImportant, newTask.DueDate)

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

func (h *HandlerFn) getQuotes(w http.ResponseWriter, r *http.Request) {
	sizeStr := r.URL.Query().Get("size")

	var size int

	if sizeStr != "" {
		var err error

		size, err = strconv.Atoi(sizeStr)

		size = max(0, size)

		if err != nil {
			utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Invalid size parameter"})
			return
		}
	}

	utils.Assert(size >= 0, "Size should be greater than or equal to zero")

	quotes := data.GetQuotes()

	if size > 0 {
		result := quotes[0:min(size, len(quotes))]

		utils.JsonResponse(w, http.StatusOK, models.QuotesResponse{Quotes: result, Size: len(result)})

		return
	}

	utils.JsonResponse(w, http.StatusOK, models.QuotesResponse{Quotes: quotes, Size: len(quotes)})
}

func SetupRoutes(r *chi.Mux, db *sql.DB) {
	routeHandler := HandlerFn{db}

	r.Get("/health", healthCheck)
	r.Get("/healthz", healthCheckWithDB)

	r.Get("/api/v1/hello-world", helloWorld)
	r.Get("/api/v1/quotes", routeHandler.getQuotes)

	r.Group(func(r chi.Router) {
		r.Use(internal.AuthWithApiKey)

		r.Get("/api/v1/tasks", routeHandler.tasks)
		r.Post("/api/v1/task/create", routeHandler.createTask)
		r.Get("/api/v1/task/{id}", routeHandler.getTask)
		r.Post("/api/v1/task/{id}", routeHandler.updateTask)
		r.Delete("/api/v1/task/{id}", routeHandler.deleteTask)
		r.Post("/api/v1/task/{id}/add/due-date", routeHandler.addDueDate)

		r.Post("/api/v1/task/{id}/completed/toggle", routeHandler.toggleTask)
		r.Post("/api/v1/task/{id}/important/toggle", routeHandler.toggleImportant)
		r.Post("/api/v1/task/{id}/add-to-my-day/toggle", routeHandler.toggleAddToMyToday)

	})

}
