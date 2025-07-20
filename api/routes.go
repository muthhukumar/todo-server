package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"text/template"
	"time"
	data "todo-server/data/quotes"
	"todo-server/db"
	"todo-server/internal"
	"todo-server/models"

	"database/sql"

	"todo-server/utils"

	query "todo-server/internal/query"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
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
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
			Message: "Task ID is not valid",
			Status:  http.StatusBadRequest,
			Code:    internal.ErrorCodeErrorMessage,
		})
		return
	}

	query := `
	SELECT 
    t.id, 
    t.name, 
    t.completed, 
    t.completed_on, 
    t.created_at, 
    t.is_important, 
    t.marked_today, 
    t.due_date, 
    t.metadata, 
    t.start_date, 
    t.recurrence_pattern, 
    t.recurrence_interval,
		t.list_id,
    st.id AS sub_task_id, 
    st.name AS sub_task_name, 
    st.completed AS sub_task_completed, 
    st.created_at AS sub_task_created_at
	FROM 
    tasks t
	LEFT JOIN 
    sub_tasks st 
	ON 
    t.id = st.task_id
	WHERE 
    t.id = $1
	ORDER BY 
    st.created_at ASC;
  `

	rows, err := h.DB.Query(query, id)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, models.ErrorResponseV2{
			Message: "Failed to fetch tasks",
			Status:  http.StatusInternalServerError,
			Code:    internal.ErrorCodeErrorMessage,
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	var task models.Task
	var subTasks []models.SubTask

	hasSubTasks := false

	for rows.Next() {
		var subTask models.SubTask
		var subTaskID sql.NullInt64
		var subTaskName sql.NullString
		var subTaskCompleted sql.NullBool
		var completedOn sql.NullString
		var recurrencePattern sql.NullString
		var recurrenceInterval sql.NullInt64
		var subTaskCreatedAt sql.NullTime // Use sql.NullTime for nullable time
		var listID sql.NullInt64

		if err := rows.Scan(
			&task.ID,
			&task.Name,
			&task.Completed,
			&completedOn,
			&task.CreatedAt,
			&task.IsImportant,
			&task.MarkedToday,
			&task.DueDate,
			&task.Metadata,
			&task.StartDate,
			&recurrencePattern,
			&recurrenceInterval,
			&listID,
			&subTaskID,
			&subTaskName,
			&subTaskCompleted,
			&subTaskCreatedAt,
		); err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, models.ErrorResponseV2{
				Message: "Failed to scan task",
				Status:  http.StatusInternalServerError,
				Code:    internal.ErrorCodeErrorMessage,
				Error:   err.Error(),
			})
			return
		}

		if completedOn.Valid {
			task.CompletedOn = completedOn.String
		} else {
			task.CompletedOn = ""
		}

		if recurrencePattern.Valid {
			task.RecurrencePattern = recurrencePattern.String
		} else {
			task.RecurrencePattern = ""
		}

		if recurrenceInterval.Valid {
			task.RecurrenceInterval = int(recurrenceInterval.Int64)
		} else {
			task.RecurrenceInterval = 0
		}

		if subTaskID.Valid {
			subTask.ID = int(subTaskID.Int64)
			subTask.TaskID = task.ID
			if subTaskName.Valid {
				subTask.Name = subTaskName.String
			} else {
				subTask.Name = ""
			}

			if subTaskCompleted.Valid {
				subTask.Completed = subTaskCompleted.Bool
			} else {
				subTask.Completed = false
			}

			if subTaskCreatedAt.Valid {
				subTask.CreatedAt = subTaskCreatedAt.Time
			} else {
				subTask.CreatedAt = time.Time{}
			}

			subTasks = append(subTasks, subTask)
			hasSubTasks = true
		}

		if listID.Valid {
			task.ListID = int(listID.Int64)
		} else {
			task.ListID = 0
		}

	}

	if !hasSubTasks {
		task.SubTasks = nil
	} else {
		task.SubTasks = subTasks
	}

	utils.JsonResponse(w, http.StatusOK, models.Response{Data: task})
}

func root(w http.ResponseWriter, r *http.Request) {
	funnyMessages := []string{"Oops, nothing to see here! Just a wild goose chase. 🦢",
		"You’ve reached the end of the internet. Congratulations!",
		"404: Fun not found. Try again later!",
		"Welcome to the void! It’s pretty empty here, huh?",
		"Under construction: Please wear your hard hat at all times. 🚧",
		"You’re lost, aren’t you? Let’s find our way back together!",
		"You're suppose to not see this page.",
	}

	idx := rand.Intn(len(funnyMessages))

	tmpl, err := template.New("root").Parse(`
<!DOCTYPE html>
<html
  lang="en"
  style="
		height: 100%;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    background-color: white;
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue',
      Arial, sans-serif;
  "
>
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>MK Todo Server</title>
  </head>
		<style>
		* {
			box-sizing: border-box;
		}
		</style>
  <body style="
		height: 100%;
		">
    <div
      style="
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
      "
    >
		<h1 style="padding: 0; margin: 0;text-align: center;">
		{{.}}
      </h1>
    </div>
  </body>
</html>
		`)

	if err != nil {
		w.Write([]byte("Oops!! Something bad happened."))

		return
	}

	tmpl.Execute(w, funnyMessages[idx])
}

func (h *HandlerFn) tasks(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	searchTerm := r.URL.Query().Get("query")
	showCompleted := r.URL.Query().Get("showCompleted")
	size := internal.ParseSize(r.URL.Query().Get("size"))
	listID := internal.ParseSize(r.URL.Query().Get("list_id"))
	showAllTasks := r.URL.Query().Get("show_all_tasks")

	query, args := query.GetTasksQuery(filter, searchTerm, showCompleted, size, listID, showAllTasks)

	rows, err := h.DB.Query(query, args...)

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, models.MsgResponse{Message: err.Error()})
		return
	}

	var tasks []models.Task

	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.Name, &task.Completed, &task.CompletedOn, &task.CreatedAt, &task.MarkedToday, &task.IsImportant, &task.DueDate, &task.Metadata, &task.ListID, &task.RecurrencePattern, &task.InCompleteSubTaskCount, &task.SubTaskCount); err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, models.ErrorResponseV2{Message: err.Error(), Status: http.StatusInternalServerError, Code: internal.ErrorCodeErrorMessage})

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
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid request body", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	validate := validator.New()

	err := validate.Struct(newTask)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
			Status:        http.StatusBadRequest,
			Code:          internal.ErrorCodeValidationFailed,
			Message:       "One or more fields are invalid",
			InvalidFields: internal.ConstructInvalidFieldData(err)})
		// TODO: add request id here

		return
	}

	utils.Assert(len(newTask.Name) > 0, "Task name length should be greater than 0")

	var taskID int

	if newTask.ListID == 0 {
		query := `
		INSERT INTO tasks (name, completed, completed_on, marked_today, is_important, due_date, metadata, list_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id;
		`

		err = h.DB.QueryRow(query, newTask.Name, newTask.Completed, newTask.CompletedOn, newTask.MarkedToday, newTask.IsImportant, newTask.DueDate, newTask.Metadata, nil).Scan(&taskID)
	} else {

		query := `
		INSERT INTO tasks (name, completed, completed_on, marked_today, is_important, due_date, metadata, list_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id;
		`
		err = h.DB.QueryRow(query, newTask.Name, newTask.Completed, newTask.CompletedOn, newTask.MarkedToday, newTask.IsImportant, newTask.DueDate, newTask.Metadata, newTask.ListID).Scan(&taskID)

	}

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, models.MsgResponse{Message: err.Error()})

		return
	}

	utils.JsonResponse(w, http.StatusCreated, models.CreateTaskResponse{Message: "Task created successfully", ID: taskID})
}

func (h *HandlerFn) updateTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, id_err := strconv.Atoi(idStr)

	if id_err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid task ID", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
	}

	var task models.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid request body", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	validate := validator.New()

	err := validate.Struct(task)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
			Status:        http.StatusBadRequest,
			Code:          internal.ErrorCodeValidationFailed,
			Message:       "One or more fields are invalid",
			InvalidFields: internal.ConstructInvalidFieldData(err)})

		return
	}

	query := "UPDATE tasks SET name=$1 WHERE id=$2"

	result, err := h.DB.Exec(query, task.Name, id)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: fmt.Sprintf("Updating task with ID {%v} failed.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		// TODO - check whether not found here is okay
		utils.JsonResponse(w, http.StatusNotFound, models.ErrorResponseV2{Message: fmt.Sprintf("Updating task with ID {%v} failed. Task may not be available.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: "Updated Task successfully."})
}

func (h *HandlerFn) updateSubTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, id_err := strconv.Atoi(idStr)

	if id_err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid sub task ID", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
	}

	var subTask models.SubTask

	if err := json.NewDecoder(r.Body).Decode(&subTask); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid request body", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	validate := validator.New()

	err := validate.Struct(subTask)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
			Status:        http.StatusBadRequest,
			Code:          internal.ErrorCodeValidationFailed,
			Message:       "One or more fields are invalid",
			InvalidFields: internal.ConstructInvalidFieldData(err)})

		return
	}

	query := "UPDATE sub_tasks SET name=$1 WHERE id=$2"

	result, err := h.DB.Exec(query, subTask.Name, id)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: fmt.Sprintf("Updating sub task with ID {%v} failed.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		// TODO - check whether not found here is okay
		utils.JsonResponse(w, http.StatusNotFound, models.ErrorResponseV2{Message: fmt.Sprintf("Updating sub task with ID {%v} failed. Task may not be available.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: "Updated sub Task successfully."})
}

func (h *HandlerFn) updateTaskMetadata(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, id_err := strconv.Atoi(idStr)

	if id_err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid task ID", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
	}

	var task models.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid request body", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	// validate := validator.New()
	//
	// err := validate.Struct(task)
	//
	// if err != nil {
	// 	utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
	// 		Status:        http.StatusBadRequest,
	// 		Code:          internal.ErrorCodeValidationFailed,
	// 		Message:       "One or more fields are invalid",
	// 		InvalidFields: internal.ConstructInvalidFieldData(err)})
	//
	// 	return
	// }

	query := "UPDATE tasks SET metadata=$1 WHERE id=$2"

	result, err := h.DB.Exec(query, task.Metadata, id)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: fmt.Sprintf("Updating task with ID {%v} failed.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		utils.JsonResponse(w, http.StatusNotFound, models.ErrorResponseV2{Message: fmt.Sprintf("Updating task with ID {%v} failed. Task may not be available.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: "Updated Task metadata successfully."})
}

func (h *HandlerFn) deleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, id_err := strconv.Atoi(idStr)

	if id_err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid task ID", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	result, error := h.DB.Exec("DELETE FROM tasks where id = $1", id)

	if error != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Deleting task failed", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: fmt.Sprintf("Task either already deleted or task with ID {%v} does not exist.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: fmt.Sprintf("Deleted task with ID {%v} successfully.", id)})
}

func (h *HandlerFn) deleteSubTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, id_err := strconv.Atoi(idStr)

	if id_err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid sub task ID", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	result, error := h.DB.Exec("DELETE FROM sub_tasks where id = $1", id)

	if error != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Deleting sub task failed", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: fmt.Sprintf("Sub Task either already deleted or task with ID {%v} does not exist.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: fmt.Sprintf("Deleted sub task with ID {%v} successfully.", id)})
}

func (h *HandlerFn) toggleTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, id_err := strconv.Atoi(idStr)

	if id_err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid task ID", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	error := db.ToggleTaskAndHandleRecurrence(h.DB, id)

	if error != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Toggling task failed", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage, Error: error.Error()})

		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: fmt.Sprintf("Toggled task with ID {%v} successfully.", id)})
}

func (h *HandlerFn) toggleSubTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, id_err := strconv.Atoi(idStr)

	if id_err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid sub task ID", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	query := `
	UPDATE sub_tasks 
	SET completed = NOT completed
	WHERE id = $1;
	`

	result, error := h.DB.Exec(query, id)

	if error != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Toggling sub task failed", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage, Error: error.Error()})

		return
	}

	if rf, error := result.RowsAffected(); rf != 1 {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: fmt.Sprintf("Sub Task with ID {%v} does not exist.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage, Error: error.Error()})

		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: fmt.Sprintf("Toggled sub task with ID {%v} successfully.", id)})
}

func (h *HandlerFn) toggleImportant(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Task ID is not valid.", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
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
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Toggling task importance failed", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: fmt.Sprintf("Task with ID {%v} does not exist.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: "Toggled Task's important"})
}

func (h *HandlerFn) toggleAddToMyToday(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, id_err := strconv.Atoi(idStr)

	if id_err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid task ID", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

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
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Toggling task Add to my day failed", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: fmt.Sprintf("Task with ID {%v} does not exist.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: "Toggled Task's Add to my day"})

}

func (h *HandlerFn) addDueDate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid Task ID", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
		return
	}

	var task models.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid request body", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	_, err = time.Parse("2006-01-02", task.DueDate)

	if err != nil && task.DueDate != "" {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid Due Date", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	query := "update tasks set due_date=$1 where id = $2"

	result, err := h.DB.Exec(query, task.DueDate, id)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Updating Task Due date failed", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		utils.JsonResponse(w, http.StatusNotFound, models.ErrorResponseV2{Message: fmt.Sprintf("Task with %v ID does not exist", id), Status: http.StatusNotFound, Code: internal.ErrorCodeErrorMessage})
		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: "Due date updated successfully"})
}

func (h *HandlerFn) getQuotes(w http.ResponseWriter, r *http.Request) {
	sizeStr := r.URL.Query().Get("size")
	random := r.URL.Query().Get("random")
	stream := r.URL.Query().Get("stream")

	var size int

	if sizeStr != "" {
		var err error

		size, err = strconv.Atoi(sizeStr)

		size = max(0, size)

		if err != nil {
			utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid size parameter", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
			return
		}
	}

	utils.Assert(size >= 0, "Size should be greater than or equal to zero")

	quotes := data.GetQuotes()

	var result []string = quotes

	if size > 0 {
		result = quotes[0:min(size, len(quotes))]
	}

	if random == "true" {
		if size <= 0 {
			size = len(quotes)
		}

		result = data.GetRandomQuotes(quotes, size)
	}

	utils.Assert(result != nil, "Result should never be nil")
	utils.Assert(len(result) >= 0, "Result should be greater than or equal to zero")

	if stream == "true" {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		for _, item := range result {
			select {
			case <-r.Context().Done():
				return
			default:
				fmt.Fprintf(w, "%s\n\n", item)
				w.(http.Flusher).Flush()
				time.Sleep(time.Millisecond * 100)
			}
		}
		return
	}

	utils.JsonResponse(w, http.StatusOK, models.QuotesResponse{Quotes: result, Size: len(quotes)})
}

func (h *HandlerFn) fetchWebPageTitle(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")

	if url == "" {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
			Status:  http.StatusBadRequest,
			Message: "URL is not provided",
		})

		return
	}

	w.Header().Set("Cache-Control", "public, max-age=604800") // 1week

	url_title, _ := db.GetURLTitle(h.DB, url)

	if url_title != nil {
		if url_title.IsValid {
			utils.JsonResponse(w, http.StatusOK, models.Response{Data: url_title.Title})
		} else {
			utils.JsonResponse(w, http.StatusUnprocessableEntity, models.MsgResponse{Message: "This URL is marked as Invalid."})
		}
		return
	}

	// pageTitle, err := chrome.GetTitleFromURLUsingChrome(url)

	// if err != nil {
	// 	db.SaveOrUpdateURLTitle(h.DB, pageTitle, url, false)

	// 	utils.JsonResponse(w, http.StatusInternalServerError, models.ErrorResponseV2{
	// 		Status:  http.StatusBadRequest,
	// 		Message: "Fetching Title using headless browser failed",
	// 		Error:   err.Error(),
	// 	})

	// 	return
	// }

	// if pageTitle == "" {
	// 	db.SaveOrUpdateURLTitle(h.DB, pageTitle, url, false)

	// 	utils.JsonResponse(w, http.StatusUnprocessableEntity, models.MsgResponse{Message: "Page title not found."})
	// 	return
	// }

	// db.SaveOrUpdateURLTitle(h.DB, pageTitle, url, true)

	// utils.JsonResponse(w, http.StatusOK, models.Response{Data: pageTitle})

	// TODO: for now this the response for this api. Once the server is migrated to better server will enabled it.
	utils.JsonResponse(w, http.StatusUnprocessableEntity, models.MsgResponse{Message: "This URL is marked as Invalid."})
}

func (h *HandlerFn) syncTitle(w http.ResponseWriter, r *http.Request) {
	internal.SyncURLTitle(h.DB)

	w.Write([]byte("OK"))
	return
}

func (h *HandlerFn) createLog(w http.ResponseWriter, r *http.Request) {
	var payload models.LogPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid request body", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	// validate := validator.New()
	//
	// err := validate.Struct(newLogs)
	//
	// if err != nil {
	// 	utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
	// 		Status:        http.StatusBadRequest,
	// 		Code:          internal.ErrorCodeValidationFailed,
	// 		Message:       "One or more fields are invalid",
	// 		InvalidFields: internal.ConstructInvalidFieldData(err)})
	//
	// 	return
	// }

	query := "INSERT INTO log (log, level, created_at) VALUES "
	values := []interface{}{}
	for i, logEntry := range payload.Data {
		query += fmt.Sprintf("($%d, $%d, $%d),", i*3+1, i*3+2, i*3+3)
		values = append(values, logEntry.Log, logEntry.Level, logEntry.CreatedAt)
	}

	var logStr string

	for _, item := range payload.Data {
		if item.Level == "error" || item.Level == "ERROR" {
			logStr += fmt.Sprintf("%s | %s | %s\n", time.Now().Format("Monday, January 2 2006"), item.Level, item.Log)
			logStr += "-----------------------------------------------\n"
		}
	}

	if logStr != "" {
		emailAuth := internal.LoadEmailCredentials()

		template := models.EmailTemplate{
			To:      []string{emailAuth.ToEmail},
			Subject: fmt.Sprintf("Critical Error Log: [MKTodo] - [%s]", time.Now().Format("Monday, January 2 2006")),
			Body:    logStr,
		}

		_ = internal.SendEmail(emailAuth, template)
	}

	query = query[:len(query)-1]

	_, err := h.DB.Exec(query, values...)

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, models.MsgResponse{Message: err.Error()})

		return
	}

	utils.JsonResponse(w, http.StatusCreated, models.MsgResponse{Message: "Logged successfully"})
}

func (h *HandlerFn) logs(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("select * from log ORDER BY created_at DESC")
	var logs []models.Log

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, models.MsgResponse{Message: err.Error()})
		return
	}

	for rows.Next() {
		var log models.Log
		if err := rows.Scan(&log.ID, &log.Log, &log.Level, &log.CreatedAt, &log.UpdatedAt); err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, models.ErrorResponseV2{Message: err.Error(), Status: http.StatusInternalServerError, Code: internal.ErrorCodeErrorMessage})

			return
		}
		logs = append(logs, log)
	}
	defer rows.Close()

	if len(logs) == 0 {
		utils.JsonResponse(w, http.StatusOK, models.Response{Data: []models.Log{}})

		return
	}

	utils.JsonResponse(w, http.StatusOK, models.Response{Data: logs})

}

func (h *HandlerFn) createSubTask(w http.ResponseWriter, r *http.Request) {
	var newSubTask models.SubTask

	if err := json.NewDecoder(r.Body).Decode(&newSubTask); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid request body", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	validate := validator.New()

	err := validate.Struct(newSubTask)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
			Status:        http.StatusBadRequest,
			Code:          internal.ErrorCodeValidationFailed,
			Message:       "One or more fields are invalid",
			InvalidFields: internal.ConstructInvalidFieldData(err)})
		// TODO: add request id here

		return
	}

	utils.Assert(len(newSubTask.Name) > 0, "Sub Task name length should be greater than 0")

	query := `
	INSERT INTO sub_tasks (name, task_id, completed)
	VALUES ($1, $2, $3)
	RETURNING id;
`

	var subTaskID int

	err = h.DB.QueryRow(query, newSubTask.Name, newSubTask.TaskID, newSubTask.Completed).Scan(&subTaskID)

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, models.MsgResponse{Message: err.Error()})

		return
	}

	utils.JsonResponse(w, http.StatusCreated, models.CreateTaskResponse{Message: "Sub Task created successfully", ID: subTaskID})
}

func (h *HandlerFn) updateRecurrencePattern(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, id_err := strconv.Atoi(idStr)

	if id_err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid task ID", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	var task models.RecurringTask

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid request body", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	validate := validator.New()

	err := validate.Struct(task)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
			Status:        http.StatusBadRequest,
			Code:          internal.ErrorCodeValidationFailed,
			Message:       "One or more fields are invalid",
			InvalidFields: internal.ConstructInvalidFieldData(err)})

		return
	}

	query := "UPDATE tasks SET recurrence_pattern=$1, recurrence_interval=$2, start_date=$3, due_date=$4 WHERE id=$5"

	var result sql.Result

	if task.RecurrencePattern == "" {
		result, err = h.DB.Exec(query, nil, task.RecurrenceInterval, task.StartDate, task.StartDate, id)
	} else {
		result, err = h.DB.Exec(query, task.RecurrencePattern, task.RecurrenceInterval, task.StartDate, task.StartDate, id)
	}

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: fmt.Sprintf("Updating task recurrence pattern with ID {%v} failed.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage, Error: err.Error()})
		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		// TODO - check whether not found here is okay
		utils.JsonResponse(w, http.StatusNotFound, models.ErrorResponseV2{Message: fmt.Sprintf("Updating task with ID {%v} failed. Task may not be available.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: "Updated Task recurring details successfully."})
}

func (h *HandlerFn) createList(w http.ResponseWriter, r *http.Request) {
	var list models.List

	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid request body", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	validate := validator.New()

	err := validate.Struct(list)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
			Status:        http.StatusBadRequest,
			Code:          internal.ErrorCodeValidationFailed,
			Message:       "One or more fields are invalid",
			InvalidFields: internal.ConstructInvalidFieldData(err)})

		return
	}

	query := `
	INSERT INTO lists (name) VALUES ($1) RETURNING id;
	`

	var listID int

	err = h.DB.QueryRow(query, list.Name).Scan(&listID)

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, "Creating list failed.")

		return
	}

	utils.JsonResponse(w, http.StatusCreated, models.CreateTaskResponse{Message: "List created successfully", ID: listID})
}

func (h *HandlerFn) lists(w http.ResponseWriter, r *http.Request) {
	query := `
	SELECT 
    l.id, 
    l.name, 
    l.created_at, 
    COUNT(t.id) AS tasks_count
	FROM 
    lists l
	LEFT JOIN 
    tasks t ON l.id = t.list_id
	GROUP BY 
    l.id, l.name, l.created_at;
	`

	rows, err := h.DB.Query(query)

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, models.MsgResponse{Message: err.Error()})
		return
	}

	var lists []models.List

	for rows.Next() {
		var list models.List
		if err := rows.Scan(&list.ID, &list.Name, &list.CreatedAt, &list.TasksCount); err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, models.ErrorResponseV2{Message: err.Error(), Status: http.StatusInternalServerError, Code: internal.ErrorCodeErrorMessage})

			return
		}
		lists = append(lists, list)
	}
	defer rows.Close()

	if len(lists) == 0 {
		utils.JsonResponse(w, http.StatusOK, models.Response{Data: []models.List{}})

		return
	}

	utils.JsonResponse(w, http.StatusOK, models.Response{Data: lists})
}

func (h *HandlerFn) updateTaskListId(w http.ResponseWriter, r *http.Request) {
	taskIdStr := chi.URLParam(r, "taskId")
	taskId, err := strconv.Atoi(taskIdStr)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
			Message: "Task ID is not valid",
			Status:  http.StatusBadRequest,
			Code:    internal.ErrorCodeErrorMessage,
		})
		return
	}

	var listID models.GetListID

	if err := json.NewDecoder(r.Body).Decode(&listID); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid request body", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	query := `
	update tasks SET list_id=$1 WHERE id=$2;
	`

	var result sql.Result

	if listID.ListID == 0 {
		result, err = h.DB.Exec(query, nil, taskId)
	} else {
		result, err = h.DB.Exec(query, listID.ListID, taskId)
	}

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: fmt.Sprintf("Updating task's list id with ID {%v} failed.", taskId), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage, Error: err.Error()})
		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		utils.JsonResponse(w, http.StatusNotFound, models.ErrorResponseV2{Message: fmt.Sprintf("Updating task's list with ID {%v} failed. Task may not be available.", taskId), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: "Updated Task's list id successfully."})

}

func (h *HandlerFn) updateListName(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, id_err := strconv.Atoi(idStr)

	if id_err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid list ID", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
	}

	var list models.List

	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid request body", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	validate := validator.New()

	err := validate.Struct(list)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
			Status:        http.StatusBadRequest,
			Code:          internal.ErrorCodeValidationFailed,
			Message:       "One or more fields are invalid",
			InvalidFields: internal.ConstructInvalidFieldData(err)})

		return
	}

	query := "UPDATE lists SET name=$1 WHERE id=$2"

	result, err := h.DB.Exec(query, list.Name, id)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: fmt.Sprintf("Updating list with ID {%v} failed.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		// TODO - check whether not found here is okay
		utils.JsonResponse(w, http.StatusNotFound, models.ErrorResponseV2{Message: fmt.Sprintf("Updating list with ID {%v} failed. list may not be available.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: "Updated list successfully."})
}

func (h *HandlerFn) deleteList(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, id_err := strconv.Atoi(idStr)

	if id_err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Invalid list ID", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	result, error := h.DB.Exec("DELETE FROM lists where id = $1", id)

	if error != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Deleting list failed", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: fmt.Sprintf("List either already deleted or list with ID {%v} does not exist.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	utils.JsonResponse(w, http.StatusOK, models.MsgResponse{Message: fmt.Sprintf("Deleted list with ID {%v} successfully.", id)})
}

func (h *HandlerFn) getList(w http.ResponseWriter, r *http.Request) {
	listId := chi.URLParam(r, "id")

	id, err := strconv.Atoi(listId)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
			Message: "List ID is not valid",
			Status:  http.StatusBadRequest,
			Code:    internal.ErrorCodeErrorMessage,
		})
		return
	}

	query := `SELECT id, name, created_at from lists where id=$1;`

	var list models.List

	row := h.DB.QueryRow(query, id)

	err = row.Scan(&list.ID, &list.Name, &list.CreatedAt)

	if err != nil {
		var message string
		if err == sql.ErrNoRows {
			message = "No list found with the give ID"
		} else {
			message = "Failed to get list details"
		}

		utils.JsonResponse(w, http.StatusInternalServerError, models.ErrorResponseV2{
			Message: message,
			Status:  http.StatusInternalServerError,
			Code:    internal.ErrorCodeErrorMessage,
			Error:   err.Error(),
		})
		return
	}

	utils.JsonResponse(w, http.StatusOK, models.Response{Data: list})
}

func SetupRoutes(r *chi.Mux, db *sql.DB) {
	routeHandler := HandlerFn{db}

	r.Get("/", root)
	r.Get("/health", healthCheck)
	r.Get("/healthz", healthCheckWithDB)

	r.Get("/api/v1/hello-world", helloWorld)
	r.Get("/api/v1/quotes", routeHandler.getQuotes)

	r.Group(func(r chi.Router) {
		r.Use(internal.AuthWithApiKey)

		r.Get("/api/v1/tasks", routeHandler.tasks)
		r.Post("/api/v1/task/create", routeHandler.createTask)
		r.Post("/api/v1/task/sub-task/create", routeHandler.createSubTask)

		r.Get("/api/v1/task/{id}", routeHandler.getTask)
		r.Post("/api/v1/task/{id}", routeHandler.updateTask)
		r.Post("/api/v1/sub-task/{id}", routeHandler.updateSubTask)

		r.Post("/api/v1/task/{id}/metadata", routeHandler.updateTaskMetadata)
		r.Post("/api/v1/task/{id}/recurrence", routeHandler.updateRecurrencePattern)

		r.Delete("/api/v1/task/{id}", routeHandler.deleteTask)
		r.Delete("/api/v1/sub-task/{id}", routeHandler.deleteSubTask)
		r.Post("/api/v1/task/{id}/add/due-date", routeHandler.addDueDate)

		r.Post("/api/v1/task/{id}/completed/toggle", routeHandler.toggleTask)
		r.Post("/api/v1/sub-task/{id}/completed/toggle", routeHandler.toggleSubTask)

		r.Post("/api/v1/task/{id}/important/toggle", routeHandler.toggleImportant)
		r.Post("/api/v1/task/{id}/add-to-my-day/toggle", routeHandler.toggleAddToMyToday)

		r.Get("/api/v1/fetch-title", routeHandler.fetchWebPageTitle)
		r.Get("/api/v1/title/sync", routeHandler.syncTitle)

		r.Post("/api/v1/list/new", routeHandler.createList)
		r.Get("/api/v1/lists", routeHandler.lists)
		r.Post("/api/v1/list/{id}", routeHandler.updateListName)
		r.Delete("/api/v1/list/{id}", routeHandler.deleteList)
		r.Get("/api/v1/list/{id}", routeHandler.getList)

		r.Post("/api/v1/task/{taskId}/list/update", routeHandler.updateTaskListId)

		r.Get("/api/v1/log", routeHandler.logs)
		r.Post("/api/v1/log", routeHandler.createLog)
	})

}
