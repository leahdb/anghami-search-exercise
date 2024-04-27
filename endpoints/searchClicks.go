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

type ClickData struct {
    SearchID       string `json:"search_id"`
    ResultType     string `json:"result_type"`
    ResultID       int    `json:"result_id"`
    ResultPosition int    `json:"result_position"`
	Timestamp   time.Time `json:"timestamp"`
}

// ReportClickHandler handles the /report-click endpoint
func ReportClickHandler(db *sql.DB) http.HandlerFunc {
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

		var request ClickData
		err = json.Unmarshal(body, &request)
        if err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

		// Insert click event data into the database
		err = insertClickEvent(db, request)
		if err != nil {
			log.Printf("Error inserting click event: %v", err)
			http.Error(w, "Error inserting click event", http.StatusInternalServerError)
			return
		}

		// Send success response
		w.WriteHeader(http.StatusOK)
	}
}

func insertClickEvent(db *sql.DB, request ClickData) error {
		// Check if the search_clicks table exists, if not, create it
	if !isTableExists(db, "search_clicks") {
		err := CreateSearchClicksTable(db)
		if err != nil {
			return err
		}
	}

	// Get the current time
	timestamp := time.Now()

	// Prepare SQL query to insert data into search_events table
	query := `
		INSERT INTO search_clicks (search_id, result_type, result_id, result_position, timestamp) VALUES (?, ?, ?, ?)
	`
	_, err := db.Exec(query, request.SearchID, request.ResultType, request.ResultID, request.ResultPosition, timestamp.Format("2006-01-02 15:04:05"))

	if err != nil {
		return err
	}

	return nil
}

// CreateSearchClicksTable creates the search_clicks table if it doesn't exist
func CreateSearchClicksTable(db *sql.DB) error {
    query := `
        CREATE TABLE IF NOT EXISTS search_clicks (
            id INT AUTO_INCREMENT PRIMARY KEY,
            search_id VARCHAR(255) NOT NULL,
            result_type ENUM('book', 'movie') NOT NULL,
            result_id INT NOT NULL,
            result_position INT NOT NULL,
            timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `

    _, err := db.Exec(query)
    if err != nil {
        return fmt.Errorf("error creating search_clicks table: %v", err)
    }

    return nil
}

