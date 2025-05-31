package api

import (
	"encoding/json"
	"github.com/The-OpenPlatform/backend/internal/models"
	"net/http"
	"os"

	"github.com/The-OpenPlatform/backend/internal/db"
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

func executeQuery(w http.ResponseWriter, r *http.Request) {
	var req models.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		db.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		db.WriteJSONError(w, "Query cannot be empty", http.StatusBadRequest)
		return
	}

	db.ExecuteQuery(w, req.Query)
}
