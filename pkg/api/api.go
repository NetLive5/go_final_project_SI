package api

import "net/http"

func Init(mux *http.ServeMux, password string) {
	mux.HandleFunc("/api/nextdate", nextDateHandler)
	mux.Handle("/api/signin", signinHandler(password))
	mux.Handle("/api/task", auth(password, http.HandlerFunc(taskHandler)))
	mux.Handle("/api/task/done", auth(password, http.HandlerFunc(doneHandler)))
	mux.Handle("/api/tasks", auth(password, http.HandlerFunc(tasksHandler)))
}
