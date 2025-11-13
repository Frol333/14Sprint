package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Frol333/14Sprint/pkg/api"
	"github.com/Frol333/14Sprint/pkg/db"
)

func main() {

	// Инициализация БД
	// Используйте data/scheduler.db для надёжного хранения
	if err := db.Init("scheduler.db"); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Регистрация маршрутов API (POST /api/task и GET /api/tasks)
	api.Init()

	// Статические файлы
	webDir := "web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// Порт сервера
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}

}
