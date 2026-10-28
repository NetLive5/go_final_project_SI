package api

import (
	"net/http"
	"time"

	"go_final_project/pkg/db"
)

func doneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "метод не поддерживается"})
		return
	}

	id := r.URL.Query().Get("id")
	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	} else {
		next, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err := db.UpdateDate(next, id); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{})
}
