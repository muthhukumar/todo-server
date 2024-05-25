package api

import (
	"encoding/json"
	"net/http"
	"todo-server/internal"
	"todo-server/models"

	"sync"

	"github.com/go-chi/chi/v5"
)

var (
	todos []models.Task
	mu    sync.Mutex
)

func addTask(todo models.Task) {
	mu.Lock()
	defer mu.Unlock()

	todos = append(todos, todo)
}

func getTasks() []models.Task {
	mu.Lock()
	defer mu.Unlock()

	return todos
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Worlld!"))
}

func tasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "applications/json")

	if len(getTasks()) == 0 {
		json.NewEncoder(w).Encode(models.Response{Data: []models.Task{}})

		return
	}

	json.NewEncoder(w).Encode(models.Response{Data: getTasks()})
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var newTask models.Task

	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		w.Header().Set("Content-Type", "applications/json")
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(models.MsgResponse{Message: "Invalid request body"})

		return
	}

	if isValid, invalidFields := internal.ValidateTodo(newTask); !isValid {
		w.Header().Set("Content-Type", "applications/json")
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(models.Response{Data: invalidFields})

		return
	}

	newTask.ID = len(getTasks()) + 1

	addTask(newTask)

	w.Header().Set("Content-Type", "applications/json")

	json.NewEncoder(w).Encode(models.MsgResponse{Message: "Task created successfully"})

}

func SetupRoutes(r *chi.Mux) {
	r.Get("/api/v1/hello-world", helloWorld)

	r.Get("/api/v1/tasks", tasks)
	r.Post("/api/v1/task/create", createTask)
}
