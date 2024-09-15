package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"
	data "todo-server/data/quotes"
	"todo-server/internal"
	"todo-server/models"

	"database/sql"

	"todo-server/utils"

	query "todo-server/internal/query"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/chromedp/chromedp"
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

	query := "select * from tasks where ID=$1"

	row := h.DB.QueryRow(query, id)

	var task models.Task

	if err := row.Scan(&task.ID, &task.Name, &task.Completed, &task.CompletedOn, &task.CreatedAt, &task.MarkedToday, &task.IsImportant, &task.DueDate, &task.Metadata); err != nil {
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
	random := r.URL.Query().Get("random")
	size := internal.ParseSize(r.URL.Query().Get("size"))

	query, args := query.GetTasksQuery(filter, searchTerm, showCompleted, random, size)

	rows, err := h.DB.Query(query, args...)

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	var tasks []models.Task

	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.Name, &task.Completed, &task.CompletedOn, &task.CreatedAt, &task.MarkedToday, &task.IsImportant, &task.DueDate, &task.Metadata); err != nil {
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

func fetchWebPageTitle(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Query().Get("url")

	if url == "" {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
			Status:  http.StatusBadRequest,
			Message: "URL is not provided",
		})

		return
	}

	var chromePath string

	if chromePath = os.Getenv("CHROME_PATH"); chromePath == "" {
		chromePath = "/opt/render/project/.render/chrome/opt/google/chrome/"
	}

	// Create an ExecAllocator with the Chrome path
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
	)

	// Create a new context with the custom Chrome path
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var pageTitle string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),                      // Wait until the body is visible
		chromedp.Sleep(2*time.Second),                   // Wait a bit for JS execution
		chromedp.Evaluate(`document.title`, &pageTitle), // Evaluate JavaScript to get the updated title
	)

	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, models.ErrorResponseV2{
			Status:  http.StatusBadRequest,
			Message: "Fetching Title using headless browser failed",
			Error:   err.Error(),
		})

		return
	}

	if pageTitle != "" {
		w.Header().Set("Cache-Control", "max-age=10, must-revalidate")

		utils.JsonResponse(w, http.StatusOK, models.Response{Data: pageTitle})

		return
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, models.ErrorResponseV2{
			Status:  http.StatusInternalServerError,
			Message: "Creating request failed.",
		})

		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.JsonResponse(w, http.StatusUnprocessableEntity, models.ErrorResponseV2{
			Status:  resp.StatusCode,
			Message: "Unable to fetch title of the link",
		})
		return
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		utils.JsonResponse(w, http.StatusUnprocessableEntity, models.ErrorResponseV2{
			Status:  http.StatusUnprocessableEntity,
			Message: err.Error(),
		})
		return
	}

	title, err := internal.ExtractTitle(string(body))

	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, models.ErrorResponseV2{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Cache-Control", "max-age=10, must-revalidate")

	utils.JsonResponse(w, http.StatusOK, models.Response{Data: title})

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
		r.Get("/api/v1/task/{id}", routeHandler.getTask)
		r.Post("/api/v1/task/{id}", routeHandler.updateTask)
		r.Post("/api/v1/task/{id}/metadata", routeHandler.updateTaskMetadata)
		r.Delete("/api/v1/task/{id}", routeHandler.deleteTask)
		r.Post("/api/v1/task/{id}/add/due-date", routeHandler.addDueDate)

		r.Post("/api/v1/task/{id}/completed/toggle", routeHandler.toggleTask)
		r.Post("/api/v1/task/{id}/important/toggle", routeHandler.toggleImportant)
		r.Post("/api/v1/task/{id}/add-to-my-day/toggle", routeHandler.toggleAddToMyToday)

		r.Get("/api/v1/fetch-title", fetchWebPageTitle)
	})

}
