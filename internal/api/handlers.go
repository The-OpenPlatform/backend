package api

import (
	"net/http"
	"os"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to the root endpoint!"))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	podName, err := os.Hostname()
	if err != nil {
		podName = "unknown"
	}
	w.Write([]byte(podName))
}
