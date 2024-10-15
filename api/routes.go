package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"text/template"
	"time"
	"todo-server/chrome"
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
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Task ID is not valid", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})
		return
	}

	query := `
	SELECT 
    tasks.*,  -- Select all columns from tasks without renaming
    sub_tasks.id AS subtask_id,  -- Include subtask-specific fields with different names
    sub_tasks.name AS subtask_name,
    sub_tasks.completed AS subtask_completed,
    sub_tasks.created_at AS subtask_created_at
FROM 
    tasks
LEFT JOIN 
    sub_tasks ON sub_tasks.task_id = tasks.id
WHERE 
    tasks.id = $1
ORDER BY 
    sub_tasks.created_at ASC;`

	rows, err := h.DB.Query(query, id)

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, models.ErrorResponseV2{Message: "Failed to fetch tasks", Status: http.StatusInternalServerError, Code: internal.ErrorCodeErrorMessage})

		return
	}
	defer rows.Close()

	var task models.Task
	var subTasks []models.SubTask

	for rows.Next() {
		var subTaskID sql.NullInt64
		var subTaskName sql.NullString
		var subTaskCompleted sql.NullBool
		var subTaskCreatedAt sql.NullTime

		if err := rows.Scan(
			&task.ID, &task.Name, &task.Completed, &task.CompletedOn, &task.CreatedAt,
			&task.MarkedToday, &task.IsImportant, &task.DueDate, &task.Metadata,
			&subTaskID, &subTaskName, &subTaskCompleted, &subTaskCreatedAt); err != nil {
		}

		if subTaskID.Valid {
			subTask := models.SubTask{
				ID:        int(subTaskID.Int64),
				Name:      subTaskName.String,
				Completed: subTaskCompleted.Bool,
				CreatedAt: subTaskCreatedAt.Time,
			}
			subTasks = append(subTasks, subTask)
		}
	}

	task.SubTasks = subTasks

	if err := rows.Err(); err != nil {
		utils.JsonResponse(w, http.StatusNotFound, models.ErrorResponseV2{Message: fmt.Sprintf("Task with ID '%v' not found", id), Status: http.StatusNotFound, Code: internal.ErrorCodeErrorMessage})
		return
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

	query, args := query.GetTasksQuery(filter, searchTerm, showCompleted, size)

	rows, err := h.DB.Query(query, args...)

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, models.MsgResponse{Message: err.Error()})
		return
	}

	var tasks []models.Task

	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.Name, &task.Completed, &task.CompletedOn, &task.CreatedAt, &task.MarkedToday, &task.IsImportant, &task.DueDate, &task.Metadata, &task.InCompleteSubTaskCount, &task.SubTaskCount); err != nil {
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

	query := `
	INSERT INTO tasks (name, completed, completed_on, marked_today, is_important, due_date, metadata)
	VALUES ($1, $2, $3, $4, $5, $6, $7) 
	RETURNING id;
`

	var taskID int

	err = h.DB.QueryRow(query, newTask.Name, newTask.Completed, newTask.CompletedOn, newTask.MarkedToday, newTask.IsImportant, newTask.DueDate, newTask.Metadata).Scan(&taskID)

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, "Creating Task failed.")

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
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: "Toggling task failed", Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

		return
	}

	if rf, _ := result.RowsAffected(); rf != 1 {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{Message: fmt.Sprintf("Task with ID {%v} does not exist.", id), Status: http.StatusBadRequest, Code: internal.ErrorCodeErrorMessage})

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

	pageTitle, err := chrome.GetTitleFromURLUsingChrome(url)

	if err != nil {
		db.SaveOrUpdateURLTitle(h.DB, pageTitle, url, false)

		utils.JsonResponse(w, http.StatusInternalServerError, models.ErrorResponseV2{
			Status:  http.StatusBadRequest,
			Message: "Fetching Title using headless browser failed",
			Error:   err.Error(),
		})

		return
	}

	if pageTitle == "" {
		db.SaveOrUpdateURLTitle(h.DB, pageTitle, url, false)

		utils.JsonResponse(w, http.StatusUnprocessableEntity, models.MsgResponse{Message: "Page title not found."})
		return
	}

	db.SaveOrUpdateURLTitle(h.DB, pageTitle, url, true)

	utils.JsonResponse(w, http.StatusOK, models.Response{Data: pageTitle})
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
		r.Delete("/api/v1/task/{id}", routeHandler.deleteTask)
		r.Delete("/api/v1/sub-task/{id}", routeHandler.deleteSubTask)
		r.Post("/api/v1/task/{id}/add/due-date", routeHandler.addDueDate)

		r.Post("/api/v1/task/{id}/completed/toggle", routeHandler.toggleTask)
		r.Post("/api/v1/sub-task/{id}/completed/toggle", routeHandler.toggleSubTask)

		r.Post("/api/v1/task/{id}/important/toggle", routeHandler.toggleImportant)
		r.Post("/api/v1/task/{id}/add-to-my-day/toggle", routeHandler.toggleAddToMyToday)

		r.Get("/api/v1/fetch-title", routeHandler.fetchWebPageTitle)
		r.Get("/api/v1/title/sync", routeHandler.syncTitle)

		r.Get("/api/v1/log", routeHandler.logs)
		r.Post("/api/v1/log", routeHandler.createLog)
	})

}
