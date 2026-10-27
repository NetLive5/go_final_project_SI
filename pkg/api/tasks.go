package api

import (
	"fmt"
	"net/http"
	"time"

	"go_final_project/pkg/db"
)

const tasksLimit = 50

type tasksResponse struct {
	Tasks []*db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	search := r.URL.Query().Get("search")
	var (
		tasks []*db.Task
		err   error
	)
	if search == "" {
		tasks, err = db.Tasks(tasksLimit)
	} else if date, parseErr := time.Parse("02.01.2006", search); parseErr == nil && date.Format("02.01.2006") == search {
		tasks, err = db.TasksByDate(tasksLimit, date.Format(dateFormat))
	} else {
		tasks, err = db.TasksByText(tasksLimit, search)
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("load tasks: %v", err)})
		return
	}

	if tasks == nil {
		tasks = make([]*db.Task, 0)
	}
	writeJSON(w, http.StatusOK, tasksResponse{Tasks: tasks})
}
