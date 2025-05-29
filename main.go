package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB

type QueryRequest struct {
	Query string `json:"query"`
}

type QueryResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Rows    int         `json:"rows,omitempty"`
}

func main() {
	var err error

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	db, err = sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/", rootHandler)
		r.Get("/hello", helloHandler)
		r.Get("/message", getRandomMessage)
		r.Get("/status", getStatus)
		r.Post("/query", executeQuery)
	})

	log.Println("Server is running on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}

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
	var messages []string
	err := db.Select(&messages, "SELECT message FROM messages")
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
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		writeJSONError(w, "Query cannot be empty", http.StatusBadRequest)
		return
	}

	// Validate query safety
	if !isQuerySafe(req.Query) {
		writeJSONError(w, "Query contains unsafe operations", http.StatusForbidden)
		return
	}

	// Execute query
	query := strings.TrimSpace(req.Query)

	if isSelectQuery(query) {
		executeSelectQuery(w, query)
	} else if isModifyQuery(query) {
		executeModifyQuery(w, query)
	} else {
		writeJSONError(w, "Unsupported query type", http.StatusBadRequest)
	}
}

func isQuerySafe(query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))

	// For now not allowed
	dangerous := []string{
		"drop", "truncate", "delete from", "alter", "create",
		"grant", "revoke", "load_file", "into outfile",
		"exec", "execute", "sp_", "xp_", "pg_",
	}

	for _, danger := range dangerous {
		if strings.Contains(query, danger) {
			// Allow DELETE with WHERE clause
			if strings.Contains(query, "delete from") && strings.Contains(query, "where") {
				continue
			}
			return false
		}
	}

	// Safe operations
	allowedStarts := []string{"select", "insert", "update", "with"}
	for _, allowed := range allowedStarts {
		if strings.HasPrefix(query, allowed) {
			return true
		}
	}

	return false
}

func isSelectQuery(query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	return strings.HasPrefix(query, "select") || strings.HasPrefix(query, "with")
}

func isModifyQuery(query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	modifyOps := []string{"insert", "update", "delete"}
	for _, op := range modifyOps {
		if strings.HasPrefix(query, op) {
			return true
		}
	}
	return false
}

func executeSelectQuery(w http.ResponseWriter, query string) {
	rows, err := db.Queryx(query)
	if err != nil {
		writeJSONError(w, fmt.Sprintf("Query execution failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.MapScan(row); err != nil {
			writeJSONError(w, fmt.Sprintf("Row scanning failed: %v", err), http.StatusInternalServerError)
			return
		}

		for key, value := range row {
			if bytes, ok := value.([]byte); ok {
				row[key] = string(bytes)
			}
		}

		results = append(results, row)
	}

	response := QueryResponse{
		Success: true,
		Data:    results,
		Rows:    len(results),
	}

	writeJSONResponse(w, response)
}

func executeModifyQuery(w http.ResponseWriter, query string) {
	result, err := db.Exec(query)
	if err != nil {
		writeJSONError(w, fmt.Sprintf("Query execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		rowsAffected = 0
	}

	response := QueryResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "Query executed successfully"},
		Rows:    int(rowsAffected),
	}

	writeJSONResponse(w, response)
}

func writeJSONResponse(w http.ResponseWriter, response QueryResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := QueryResponse{
		Success: false,
		Error:   message,
	}
	json.NewEncoder(w).Encode(response)
}
