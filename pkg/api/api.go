package api

import "net/http"

func Init(mux *http.ServeMux) {
	mux.HandleFunc("/api/nextdate", nextDateHandler)
	mux.Handle("/api/signin", http.HandlerFunc(signinHandler))
	mux.Handle("/api/task", auth(http.HandlerFunc(taskHandler)))
	mux.Handle("/api/task/done", auth(http.HandlerFunc(doneHandler)))
	mux.Handle("/api/tasks", auth(http.HandlerFunc(tasksHandler)))
}
