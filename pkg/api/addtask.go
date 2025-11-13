package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Frol333/14Sprint/pkg/db"
)

// taskHandler обрабатывает методы для /api/task
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
		// future methods: GET, DELETE и т.д. будут добавлены позже
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// addTaskHandler обрабатывает POST /api/task — создание задачи
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	defer r.Body.Close()

	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, map[string]string{"error": "Invalid JSON"})
		return
	}

	// Проверка обязательного поля title
	if strings.TrimSpace(task.Title) == "" {
		writeJson(w, map[string]string{"error": "Не указан заголовок задачи"})
		return
	}

	// Логика даты и повторения
	now := time.Now()
	todayMid := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Если date не указана — устанавливаем сегодняшнюю дату
	if strings.TrimSpace(task.Date) == "" {
		task.Date = now.Format("20060102")
	}

	// Парсим дату
	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		writeJson(w, map[string]string{"error": "Дата представлена в формате, отличном от 20060102"})
		return
	}

	// Если дата прошла (раньше сегодня), корректируем в зависимости от Repeat
	if parsedDate.Before(todayMid) {
		trimmedRepeat := strings.TrimSpace(task.Repeat)
		if trimmedRepeat == "" {
			// Без повторения — ставим текущую дату
			task.Date = todayMid.Format("20060102")
		} else {
			// С повторением — вычисляем следующую дату
			next, err := NextDate(now, task.Date, trimmedRepeat)
			if err != nil {
				writeJson(w, map[string]string{"error": err.Error()})
				return
			}
			task.Date = next
		}
	}
	// Иначе (дата ≥ сегодня) — оставляем как есть

	// Добавляем задачу в БД
	id, err := db.AddTask(&task)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	// Возвращаем id как строку
	writeJson(w, map[string]string{
		"id": strconv.FormatInt(id, 10),
	})

}

// writeJson отправляет JSON-ответ (без pretty-print)
func writeJson(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
