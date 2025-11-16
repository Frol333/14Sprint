package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Frol333/14Sprint/pkg/db"
	"github.com/Frol333/14Sprint/pkg/model"
	"github.com/rs/zerolog/log"
)

// writeJson сериализует data в JSON и устанавливает заголовок.
func writeJson(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_ = json.NewEncoder(w).Encode(data)
}

// tasksHandler обрабатывает GET /api/tasks
func TasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := db.Tasks(r.Context(), 50)
	if err != nil {
		log.Debug().Err(err).Msgf("Failed to get tasks")
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}

	writeJson(w, model.TasksResp{Tasks: tasks})
}

// saveTaskHandler обрабатывает POST/PUT /api/tasks
func GetTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id should not be empty", http.StatusBadRequest)
		return
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Debug().Err(err).Msgf("Failed to convert id: %s", id)
		http.Error(w, "id should not be int", http.StatusBadRequest)
		return
	}
	task, err := db.GetTask(r.Context(), int64(intID))
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get task")
		http.Error(w, "Failed to get task", http.StatusInternalServerError)
		return
	}
	writeJson(w, task)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id should not be empty", http.StatusBadRequest)
		return
	}
	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Debug().Err(err).Msgf("Failed to convert id: %s", id)
		http.Error(w, "id should not be int", http.StatusBadRequest)
		return
	}
	if err := db.DeleteTask(r.Context(), int64(intID)); err != nil {
		log.Debug().Err(err).Msg("Failed to delete task")
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	writeJson(w, map[string]interface{}{})
}

func CrateTask(w http.ResponseWriter, r *http.Request) {
	var t model.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		log.Debug().Err(err).Msgf("Failed to decode task")
		http.Error(w, "json is invalid", http.StatusBadRequest)
		return
	}

	if t.Date == "" || t.Title == "" {
		http.Error(w, "missing date or title", http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(r.Context(), &t)
	if err != nil {
		log.Debug().Err(err).Interface("task", t).Msgf("Failed to add task")
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	writeJson(w, map[string]int64{"id": id})
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	var t model.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		log.Debug().Err(err).Msgf("Failed to decode task")
		http.Error(w, "json is invalid", http.StatusBadRequest)
		return
	}
	if t.Date == "" || t.Title == "" {
		http.Error(w, "missing date or title", http.StatusBadRequest)
		return
	}
	if err := db.UpdateTask(r.Context(), &t); err != nil {
		log.Debug().Err(err).Interface("task", t).Msgf("Failed to update task")
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}
	writeJson(w, map[string]interface{}{})
}

// doneHandler обрабатывает POST /api/tasks/done?id=…
func DoneHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id should not be empty", http.StatusBadRequest)
		return
	}
	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Debug().Err(err).Msgf("Failed to convert id: %s", id)
		http.Error(w, "id should not be int", http.StatusBadRequest)
		return
	}
	t, err := db.GetTask(r.Context(), int64(intID))
	if err != nil {
		log.Debug().Err(err).Int("", intID).Msgf("Failed to get task")
		http.Error(w, "Failed to get task", http.StatusInternalServerError)
		return
	}

	if t.Repeat == "" {
		if err := db.DeleteTask(r.Context(), int64(intID)); err != nil {
			log.Debug().Err(err).Int("", intID).Msgf("Failed to delete task")
			http.Error(w, "Failed to delete task", http.StatusInternalServerError)
			return
		}
		writeJson(w, map[string]interface{}{})
		return
	}

	var currDate time.Time
	var parseErr error
	layouts := []string{"2006-01-02", "2006-01-02 15:04:05", "20060102", time.RFC3339}
	for _, lay := range layouts {
		currDate, parseErr = time.Parse(lay, t.Date)
		if parseErr == nil {
			break
		}
	}

	nt, err := NextDate(currDate, t.Repeat)
	if err != nil {
		log.Debug().Err(err).
			Time("currDate", currDate).
			Str("repeat", t.Repeat).Msgf("Failed to get next data")
		http.Error(w, fmt.Sprintf("Failed to get next data with error: %v", err), http.StatusBadRequest)
		return
	}

	nextDate := nt.Format(dateLayout)

	if err := db.UpdateDate(nextDate, id); err != nil {
		log.Debug().Err(err).
			Str("nextDate", nextDate).
			Str("id", id).Msgf("Failed to update data")
		http.Error(w, fmt.Sprintf("Failed to update data with error: %v", err), http.StatusInternalServerError)
		return
	}

	writeJson(w, map[string]interface{}{})
}
