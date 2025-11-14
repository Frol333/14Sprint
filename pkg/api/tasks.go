package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Frol333/14Sprint/pkg/db"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// writeJson сериализует data в JSON и устанавливает заголовок.
func writeJson(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_ = json.NewEncoder(w).Encode(data)
}

// tasksHandler обрабатывает GET /api/tasks
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	// принимаем только GET
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Максимум 50 задач (как указано в примерах)
	tasks, err := db.Tasks(50)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	// Гарантируем, что вернем пустой массив, если задач нет
	if tasks == nil {
		tasks = make([]*db.Task, 0)
	}

	writeJson(w, TasksResp{Tasks: tasks})

}

// saveTaskHandler обрабатывает POST/PUT /api/tasks
// Сохраняет новую задачу или обновляет существующую.
func taskHandler(w http.ResponseWriter, r *http.Request) {
	// Чтение задачи из тела запроса

	switch r.Method {
	case http.MethodGet:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJson(w, map[string]string{"error": "Не указан идентификатор"})
			return
		}
		intID, err := strconv.Atoi(id)
		if err != nil {
			writeJson(w, map[string]string{"error": "Не корректный идентификатор"})
			return
		}
		if task, err := db.GetTask(r.Context(), int64(intID)); err != nil {
			writeJson(w, map[string]string{"error": err.Error()})
			return
		} else {
			writeJson(w, task)
			return
		}
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJson(w, map[string]string{"error": "Не указан идентификатор"})
			return
		}
		intID, err := strconv.Atoi(id)
		if err != nil {
			writeJson(w, map[string]string{"error": "Не корректный идентификатор"})
			return
		}
		if err := db.DeleteTask(r.Context(), int64(intID)); err != nil {
			writeJson(w, map[string]string{"error": err.Error()})
			return
		}

		return
	case http.MethodPost:
		var t db.Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeJson(w, map[string]string{"error": "invalid json"})
			return
		}
		// Простейшая валидация: дата и заголовок обязателены
		// (можно скорректировать под вашу модель)
		if t.Date == "" || t.Title == "" {
			writeJson(w, map[string]string{"error": "missing date or title"})
			return
		}
		// Обновление/создание задачи (upsert-логика предполагается в UpdateTask)
		if id, err := db.AddTask(r.Context(), &t); err != nil {
			writeJson(w, map[string]string{"error": err.Error()})
			return
		} else {
			writeJson(w, map[string]int64{"id": id})
		}

	case http.MethodPut:
		var t db.Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeJson(w, map[string]string{"error": "invalid json"})
			return
		}
		// Простейшая валидация: дата и заголовок обязателены
		// (можно скорректировать под вашу модель)
		if t.Date == "" || t.Title == "" {
			writeJson(w, map[string]string{"error": "missing date or title"})
			return
		}
		// Обновление/создание задачи (upsert-логика предполагается в UpdateTask)
		if err := db.UpdateTask(r.Context(), &t); err != nil {
			writeJson(w, map[string]string{"error": err.Error()})
			return
		}
		writeJson(w, map[string]interface{}{})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// doneHandler обрабатывает POST /api/tasks/done?id=…
func doneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "Не указан идентификатор"})
		return
	}
	intID, err := strconv.Atoi(id)
	if err != nil {
		writeJson(w, map[string]string{"error": "Не корректный идентификатор"})
		return
	}
	t, err := db.GetTask(r.Context(), int64(intID))
	if err != nil {
		http.Error(w, "In terner error ", http.StatusInternalServerError)
		return
	}

	// Если задача одноразовая (repeat пустой), удаляем
	if t.Repeat == "" {
		if err := db.DeleteTask(r.Context(), int64(intID)); err != nil {
			writeJson(w, map[string]string{"error": err.Error()})
			return
		}
		writeJson(w, map[string]interface{}{})
		return
	}

	// Периодическая задача: вычисляем следующую дату и обновляем её
	// Пример сигнатуры: NextDate(currentDate, repeat)
	// Распарсить дату в time.Time
	var currDate time.Time
	var parseErr error
	layouts := []string{"2006-01-02", "2006-01-02 15:04:05", time.RFC3339}
	for _, lay := range layouts {
		currDate, parseErr = time.Parse(lay, t.Date)
		if parseErr == nil {
			break
		}
	}
	if parseErr != nil {
		writeJson(w, map[string]string{"error": "invalid date format"})
		return
	}

	// Вызов NextDate с тремя аргументами
	nextDate, err := NextDate(currDate, t.Repeat, "")
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	if err := db.UpdateDate(nextDate, id); err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	writeJson(w, map[string]interface{}{})

}
