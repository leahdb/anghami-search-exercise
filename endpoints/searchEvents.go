package endpoints

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type ReportSearchRequest struct {
	SearchID    string    `json:"search_id"`
	SearchQuery string    `json:"search_query"`
	Timestamp   time.Time `json:"timestamp"`
}

func ReportSearchHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Log the request body for debugging
        body, err := io.ReadAll(r.Body)
        if err != nil {
            log.Printf("Error reading request body: %v", err)
        }

        // Parse request body and extract search event data
        var request ReportSearchRequest
        err = json.Unmarshal(body, &request)
        if err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

		// Insert search event data into the database
		err = insertSearchEvent(db, request)
		if err != nil {
			http.Error(w, "Error inserting search event", http.StatusInternalServerError)
			return
		}

		// Send success response
		w.WriteHeader(http.StatusOK)
	}
}

// insertSearchEvent inserts search event data into the search_events table in the database
func insertSearchEvent(db *sql.DB, request ReportSearchRequest) error {
	// Check if the search_events table exists, if not, create it
	if !isTableExists(db, "search_events") {
		err := createSearchEventsTable(db)
		if err != nil {
			return err
		}
	}

	// Get the current time
	timestamp := time.Now()

	// Prepare SQL query to insert data into search_events table
	query := `
		INSERT INTO search_events (search_id, search_query, timestamp)
		VALUES (?, ?, ?)
	`
	// Execute SQL query
	_, err := db.Exec(query, request.SearchID, request.SearchQuery, timestamp.Format("2006-01-02 15:04:05"))
	if err != nil {
		return err
	}

	return nil
}

// isTableExists checks if the given table exists in the database
func isTableExists(db *sql.DB, tableName string) bool {
	query := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = '%s'", tableName)
	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// createSearchEventsTable creates the search_events table in the database
func createSearchEventsTable(db *sql.DB) error {
	query := `
		CREATE TABLE search_events (
			id INT AUTO_INCREMENT PRIMARY KEY,
			search_id VARCHAR(255) NOT NULL,
			search_query VARCHAR(255) NOT NULL,
			timestamp TIMESTAMP NOT NULL
		)
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}
