package api

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/The-OpenPlatform/backend/internal/models"

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

func getRandomMessage(w http.ResponseWriter, r *http.Request) {
	messages, err := db.FetchMessages()
	if err != nil {
		http.Error(w, "Failed to query messages", http.StatusInternalServerError)
		return
	}

	if len(messages) == 0 {
		http.Error(w, "No messages found", http.StatusNotFound)
		return
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := rand.Intn(len(messages))
	w.Write([]byte(messages[randomIndex]))
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
