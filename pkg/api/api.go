package api

import "net/http"

func Init() {
	http.HandleFunc("/api/tasks", tasksHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/task/done", doneHandler)

}
