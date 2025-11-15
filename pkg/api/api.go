package api

import "net/http"

var routesRegistered bool

func Init() {
	if routesRegistered {
		return
	}
	// http.HandleFunc("/api/tasks", tasksHandler)
	// http.HandleFunc("/api/task", taskHandler)
	// //http.HandleFunc("/api/task/done", doneHandler)
	// http.HandleFunc("/api/signin", signinHandler)
	http.HandleFunc("/api/signin", signinHandler)
	http.HandleFunc("/api/tasks", auth(tasksHandler))
	http.HandleFunc("/api/task", auth(taskHandler))
	http.HandleFunc("/api/task/done", auth(doneHandler))
	routesRegistered = true
}
