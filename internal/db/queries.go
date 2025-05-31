package db

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/The-OpenPlatform/backend/internal/models"
)

func ExecuteQuery(w http.ResponseWriter, query string) {
	cleaned := strings.TrimSpace(query)

	if !isQuerySafe(cleaned) {
		WriteJSONError(w, "Query contains unsafe operations", http.StatusForbidden)
		return
	}

	if isSelectQuery(cleaned) {
		executeSelectQuery(w, cleaned)
	} else if isModifyQuery(cleaned) {
		executeModifyQuery(w, cleaned)
	} else {
		WriteJSONError(w, "Unsupported query type", http.StatusBadRequest)
	}
}

func isQuerySafe(query string) bool {
	q := strings.ToLower(query)
	dangerous := []string{"drop", "truncate", "delete from", "alter", "create", "grant", "revoke", "load_file", "into outfile", "exec", "execute", "sp_", "xp_", "pg_"}
	for _, d := range dangerous {
		if strings.Contains(q, d) && !(strings.Contains(q, "delete from") && strings.Contains(q, "where")) {
			return false
		}
	}

	allowed := []string{"select", "insert", "update", "with"}
	for _, a := range allowed {
		if strings.HasPrefix(q, a) {
			return true
		}
	}

	return false
}

func isSelectQuery(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	return strings.HasPrefix(q, "select") || strings.HasPrefix(q, "with")
}

func isModifyQuery(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	for _, op := range []string{"insert", "update", "delete"} {
		if strings.HasPrefix(q, op) {
			return true
		}
	}
	return false
}

func executeSelectQuery(w http.ResponseWriter, query string) {
	rows, err := DB.Queryx(query)
	if err != nil {
		WriteJSONError(w, fmt.Sprintf("Query execution failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.MapScan(row); err != nil {
			WriteJSONError(w, fmt.Sprintf("Row scanning failed: %v", err), http.StatusInternalServerError)
			return
		}

		for k, v := range row {
			if b, ok := v.([]byte); ok {
				row[k] = string(b)
			}
		}

		results = append(results, row)
	}

	resp := models.QueryResponse{
		Success: true,
		Data:    results,
		Rows:    len(results),
	}

	WriteJSONResponse(w, resp)
}

func executeModifyQuery(w http.ResponseWriter, query string) {
	result, err := DB.Exec(query)
	if err != nil {
		WriteJSONError(w, fmt.Sprintf("Query execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		rowsAffected = 0
	}

	resp := models.QueryResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "Query executed successfully"},
		Rows:    int(rowsAffected),
	}

	WriteJSONResponse(w, resp)
}

func WriteJSONResponse(w http.ResponseWriter, response models.QueryResponse) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(models.QueryResponse{Success: false, Error: message})
	if err != nil {
		return
	}
}
