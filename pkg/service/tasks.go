package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Frol333/14Sprint/pkg/db"
	"github.com/Frol333/14Sprint/pkg/model"
	"github.com/rs/zerolog/log"
)

type taskPayload struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// writeJson сериализует data в JSON и устанавливает заголовок.
func writeJson(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// tasksHandler обрабатывает GET /api/tasks
func TasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := db.Tasks(r.Context(), 0)
	if err != nil {
		log.Debug().Err(err).Msgf("Failed to get tasks")
		writeError(w, http.StatusInternalServerError, "Failed to get tasks")
		return
	}

	resp := make([]map[string]string, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, taskToResponse(t))
	}

	writeJson(w, map[string][]map[string]string{"tasks": resp})
}

// saveTaskHandler обрабатывает POST/PUT /api/tasks
func GetTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id should not be empty")
		return
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Debug().Err(err).Msgf("Failed to convert id: %s", id)
		writeError(w, http.StatusBadRequest, "id should not be int")
		return
	}
	task, err := db.GetTask(r.Context(), int64(intID))
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get task")
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJson(w, taskToResponse(task))
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id should not be empty")
		return
	}
	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Debug().Err(err).Msgf("Failed to convert id: %s", id)
		writeError(w, http.StatusBadRequest, "id should not be int")
		return
	}
	if err := db.DeleteTask(r.Context(), int64(intID)); err != nil {
		log.Debug().Err(err).Msg("Failed to delete task")
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJson(w, map[string]interface{}{})
}

func CrateTask(w http.ResponseWriter, r *http.Request) {
	var payload taskPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Debug().Err(err).Msgf("Failed to decode task")
		writeError(w, http.StatusBadRequest, "json is invalid")
		return
	}

	task, err := payload.normalize(false)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := db.AddTask(r.Context(), task)
	if err != nil {
		log.Debug().Err(err).Interface("task", task).Msgf("Failed to add task")
		writeError(w, http.StatusInternalServerError, "Failed to create task")
		return
	}

	writeJson(w, map[string]int64{"id": id})
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	var payload taskPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Debug().Err(err).Msgf("Failed to decode task")
		writeError(w, http.StatusBadRequest, "json is invalid")
		return
	}
	task, err := payload.normalize(true)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := db.UpdateTask(r.Context(), task); err != nil {
		log.Debug().Err(err).Interface("task", task).Msgf("Failed to update task")
		writeError(w, http.StatusInternalServerError, "Failed to create task")
		return
	}
	writeJson(w, map[string]interface{}{})
}

// doneHandler обрабатывает POST /api/tasks/done?id=…
func DoneHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id should not be empty")
		return
	}
	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Debug().Err(err).Msgf("Failed to convert id: %s", id)
		writeError(w, http.StatusBadRequest, "id should not be int")
		return
	}
	t, err := db.GetTask(r.Context(), int64(intID))
	if err != nil {
		log.Debug().Err(err).Int("", intID).Msgf("Failed to get task")
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if t.Repeat == "" {
		if err := db.DeleteTask(r.Context(), int64(intID)); err != nil {
			log.Debug().Err(err).Int("", intID).Msgf("Failed to delete task")
			writeError(w, http.StatusInternalServerError, "Failed to delete task")
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
	if parseErr != nil {
		writeError(w, http.StatusBadRequest, "Failed to parse task date")
		return
	}

	nt, err := NextDate(currDate, t.Repeat)
	if err != nil {
		log.Debug().Err(err).
			Time("currDate", currDate).
			Str("repeat", t.Repeat).Msgf("Failed to get next data")
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to get next data with error: %v", err))
		return
	}

	nextDate := nt.Format(dateLayout)

	if err := db.UpdateDate(nextDate, id); err != nil {
		log.Debug().Err(err).
			Str("nextDate", nextDate).
			Str("id", id).Msgf("Failed to update data")
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update data with error: %v", err))
		return
	}

	writeJson(w, map[string]interface{}{})
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowParam := r.URL.Query().Get("now")
	var now time.Time
	var err error
	if nowParam == "" {
		now = time.Now()
	} else {
		now, err = time.Parse(dateLayout, nowParam)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse now with error: %v", err), http.StatusBadRequest)
			return
		}
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "date should not be empty", http.StatusBadRequest)
		return
	}

	repeat := r.URL.Query().Get("repeat")
	if repeat == "" {
		http.Error(w, "repeat should not be empty", http.StatusBadRequest)
		return
	}

	currDate, err := time.Parse(dateLayout, date)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse date with error: %v", err), http.StatusBadRequest)
		return
	}

	nextDate := currDate
	for {
		nextDate, err = NextDate(nextDate, repeat)
		if err != nil {
			log.Debug().Err(err).
				Time("currDate", nextDate).
				Str("repeat", repeat).Msgf("Failed to get next date")
			http.Error(w, fmt.Sprintf("Failed to get next date with error: %v", err), http.StatusBadRequest)
			return
		}
		if !nextDate.Before(now) {
			break
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err = w.Write([]byte(nextDate.Format(dateLayout))); err != nil {
		log.Debug().Err(err).Msg("Failed to write next date response")
	}
}

func (p *taskPayload) normalize(requireID bool) (*model.Task, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, fmt.Errorf("title should not be empty")
	}
	comment := strings.TrimSpace(p.Comment)

	dateStr := strings.TrimSpace(p.Date)
	if dateStr == "" {
		dateStr = today.Format(dateLayout)
	}
	date, err := time.Parse(dateLayout, dateStr)
	if err != nil {
		return nil, fmt.Errorf("date should be in format %s", dateLayout)
	}
	if date.Before(today) {
		date = today
	}

	repeat := strings.TrimSpace(p.Repeat)
	if repeat != "" {
		if _, err := NextDate(date, repeat); err != nil {
			return nil, fmt.Errorf("repeat is invalid")
		}
	}

	var id int64
	if strings.TrimSpace(p.ID) != "" {
		id, err = strconv.ParseInt(strings.TrimSpace(p.ID), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("id should be int")
		}
	}
	if requireID && id == 0 {
		return nil, fmt.Errorf("id should not be empty")
	}

	return &model.Task{
		ID:      id,
		Date:    date.Format(dateLayout),
		Title:   title,
		Comment: comment,
		Repeat:  repeat,
	}, nil
}

func taskToResponse(t *model.Task) map[string]string {
	return map[string]string{
		"id":      strconv.FormatInt(t.ID, 10),
		"date":    t.Date,
		"title":   t.Title,
		"comment": t.Comment,
		"repeat":  t.Repeat,
	}
}
