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
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "метод не поддерживается"})
		return
	}

	search := r.URL.Query().Get("search")
	searchDate, parseErr := time.Parse("02.01.2006", search)
	isDateSearch := parseErr == nil && searchDate.Format("02.01.2006") == search
	var (
		tasks []*db.Task
		err   error
	)
	switch {
	case search == "":
		tasks, err = db.Tasks(tasksLimit)
	case isDateSearch:
		tasks, err = db.TasksByDate(tasksLimit, searchDate.Format(dateFormat))
	default:
		tasks, err = db.TasksByText(tasksLimit, search)
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("не удалось загрузить задачи: %v", err)})
		return
	}

	if tasks == nil {
		tasks = make([]*db.Task, 0)
	}
	writeJSON(w, http.StatusOK, tasksResponse{Tasks: tasks})
}
