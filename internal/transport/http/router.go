package transporthttp

import (
	"net/http"

	"github.com/gorilla/mux"

	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
)

func NewRouter(
	taskHandler *httphandlers.TaskHandler,
	repeatTaskHandler *httphandlers.RepeatTaskHandler,
	docsHandler *swaggerdocs.Handler,
) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// Swagger документация
	router.HandleFunc("/swagger/openapi.json", docsHandler.ServeSpec).Methods(http.MethodGet)
	router.HandleFunc("/swagger/", docsHandler.ServeUI).Methods(http.MethodGet)
	router.HandleFunc("/swagger", docsHandler.RedirectToUI).Methods(http.MethodGet)

	api := router.PathPrefix("/api/v1").Subrouter()

	// Задачи (CRUD)
	api.HandleFunc("/tasks", taskHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/tasks", taskHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.GetByID).Methods(http.MethodGet)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.Delete).Methods(http.MethodDelete)

	// Периоды (CRUD)
	api.HandleFunc("/periods", repeatTaskHandler.ListPeriods).Methods(http.MethodGet)
	api.HandleFunc("/periods", repeatTaskHandler.CreatePeriod).Methods(http.MethodPost)
	api.HandleFunc("/periods/{id:[0-9]+}", repeatTaskHandler.GetPeriodByID).Methods(http.MethodGet)
	api.HandleFunc("/periods/{id:[0-9]+}", repeatTaskHandler.UpdatePeriod).Methods(http.MethodPut)
	api.HandleFunc("/periods/{id:[0-9]+}", repeatTaskHandler.DeletePeriod).Methods(http.MethodDelete)

	// Настройки повторения
	api.HandleFunc("/repeat-tasks", repeatTaskHandler.ListRepeatTasks).Methods(http.MethodGet)
	api.HandleFunc("/repeat-tasks/active", repeatTaskHandler.ListActiveRepeatTasks).Methods(http.MethodGet)
	api.HandleFunc("/repeat-tasks", repeatTaskHandler.CreateRepeatTask).Methods(http.MethodPost)
	api.HandleFunc("/repeat-tasks/{id:[0-9]+}", repeatTaskHandler.GetRepeatTaskByID).Methods(http.MethodGet)
	api.HandleFunc("/repeat-tasks/{id:[0-9]+}", repeatTaskHandler.UpdateRepeatTask).Methods(http.MethodPut)
	api.HandleFunc("/repeat-tasks/{id:[0-9]+}", repeatTaskHandler.DeleteRepeatTask).Methods(http.MethodDelete)
	api.HandleFunc("/repeat-tasks/{id:[0-9]+}/generate", repeatTaskHandler.GenerateTasks).Methods(http.MethodPost)
	// api.HandleFunc("/repeat-tasks/by-task/{taskId:[0-9]+}", repeatTaskHandler.GetRepeatTaskByTaskID).Methods(http.MethodGet)

	// Атомарное создание задачи + настройки повторения (удобный эндпоинт)
	api.HandleFunc("/tasks-with-repeat", repeatTaskHandler.CreateTaskWithRepeat).Methods(http.MethodPost)

	// Получение задач вместе с информацией о повторении (с пагинацией)
	api.HandleFunc("/tasks-with-repeat", repeatTaskHandler.GetTasksWithRepeatInfo).Methods(http.MethodGet)

	api.HandleFunc("/tasks-generate", repeatTaskHandler.GenerateAllTasks).Methods(http.MethodGet)


	return router
}