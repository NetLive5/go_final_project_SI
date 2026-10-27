package api

import "net/http"

func Init(mux *http.ServeMux) {
	mux.HandleFunc("/api/nextdate", nextDateHandler)
	mux.HandleFunc("/api/task", taskHandler)
	mux.HandleFunc("/api/task/done", doneHandler)
	mux.HandleFunc("/api/tasks", tasksHandler)
}
