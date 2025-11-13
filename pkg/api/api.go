package api

import "net/http"

func Init() {
	http.HandleFunc("/api/tasks", tasksHandler)

	http.HandleFunc("/api/tasks/save", saveTaskHandler)
	http.HandleFunc("/api/tasks/done", doneHandler)

	http.HandleFunc("/api/tasks/delete", delHandler)
	http.HandleFunc("/api/tasks", taskHandler)

}
