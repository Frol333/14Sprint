package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Frol333/14Sprint/pkg/api"
	"github.com/Frol333/14Sprint/pkg/db"
)

// Port задаётся по умолчанию, можно переопределить через переменную окружения в тестах.
var port = 7540

// Путь к директории с фронтендом (файлы из ./web будут выдавать сервер)
var webDir = "./web"

func main() {
	// Инициализация БД
	dbFile := "scheduler.db"
	if env := os.Getenv("TODO_DBFILE"); env != "" {
		dbFile = env
	}
	if err := db.Init(dbFile); err != nil {
		log.Fatalf("DB init failed: %v", err)
	}

	api.Init()

	if v := os.Getenv("TODO_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}

	// http.FileServer будет отдавать файлы из webDir
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	addr := fmt.Sprintf("127.0.0.1:%v", port)
	log.Print("Starting server")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
