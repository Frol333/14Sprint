package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"

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

// Вот что я добавил и плюс прописал в файле api.go
type Claims struct {
	PwdHash string `json:"pwdHash"`
	jwt.RegisteredClaims
}

func hashPassword(p string) string {
	h := sha256.Sum256([]byte(p))
	return hex.EncodeToString(h[:])
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	envPass := os.Getenv("TODO_PASSWORD")
	// Если пароль не задан в окружении — аутентификация не требуется
	if len(envPass) > 0 {
		if req.Password != envPass {
			// неверный пароль
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Неверный пароль"})
			return
		}
	}

	// формируем токен: подпись HS256, срок жизни 8 часов
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_secret" // можно заменить по месту на более строгий секрет
	}
	hash := hashPassword(envPass) // хэш текущего пароля
	claims := &Claims{
		PwdHash: hash,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "token generation error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})

}

// auth middleware: проверяет куку token, валидирует JWT и сравнивает pwdHash
func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		// если пароль не указан — фильтрация не требуется
		if len(pass) > 0 {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
			tokenStr := cookie.Value
			secret := os.Getenv("JWT_SECRET")
			if secret == "" {
				secret = "default_secret"
			}
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				// простая проверка подписи
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			// валидируем соответствие пароля текущему значению TODO_PASSWORD
			currentHash := hashPassword(pass)
			if claims.PwdHash != currentHash {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		// либо пароль пустой (аутентификация не требуется), либо прошли проверки
		next(w, r)
	}

}
