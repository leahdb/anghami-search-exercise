package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// SearchRequest represents the structure of the request sent to the /search endpoint
type SearchRequest struct {
	SearchQuery string `json:"search_query"`
}

// SearchResult represents the structure of a search result item
type SearchResult struct {
	Title   string `json:"title"`
	Type    string `json:"type"`
	Rating    float64 `json:"rating"`
	// Add other fields as needed
}

// SearchResponse represents the structure of the response sent by the /search endpoint
type SearchResponse struct {
	Results       []SearchResult `json:"results"`
	Cached        bool           `json:"cached"`
	SearchID      string         `json:"search_id"`
}

// Define SearchHandler function to handle /search endpoint
func SearchHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body and extract search query
		var request SearchRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Perform search in the database
		results, err := performSearch(db, request.SearchQuery)
		if err != nil {
			http.Error(w, "Error performing search", http.StatusInternalServerError)
			return
		}

		// Construct response with search results
		response := SearchResponse{
			Results:  results,
			Cached:   false, // Placeholder, you will implement caching later
			SearchID: generateSearchID(),
		}

		// Send response back to client
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}
}


// performSearch executes a SQL query to perform the search operation in the database
func performSearch(db *sql.DB, searchQuery string) ([]SearchResult, error) {
	// Prepare SQL query to search for books and movies based on the search query
	query := `
		SELECT title, rating, "book" AS type
		FROM books
		WHERE title LIKE ?
		UNION ALL
		SELECT title, rating, "movie" AS type
		FROM movies
		WHERE title LIKE ?
		ORDER BY title
	`

	// Execute the SQL query with placeholders for search query
	rows, err := db.Query(query, "%"+searchQuery+"%", "%"+searchQuery+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the query results and construct search results
	var results []SearchResult
	for rows.Next() {
		var title string
		var rating float64
		var itemType string
		err := rows.Scan(&title, &rating, &itemType)
		if err != nil {
			return nil, err
		}

		result := SearchResult{
			Title:  title,
			Rating: rating,
			Type:   itemType,
		}
		results = append(results, result)
	}

	// Check for any errors during iteration
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return results, nil
}


// generateSearchID generates a unique search ID
func generateSearchID() string {
	// Implement function to generate a unique search ID (e.g., using UUID)
	// For simplicity, you can use a random string or timestamp-based ID
	return "unique_search_id"
}

// Other handlers for additional endpoints (e.g., search reporting, analytics) can be added here

