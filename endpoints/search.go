package endpoints

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "time"	
	"math/rand"
)

type SearchRequest struct {
    SearchQuery string `json:"search_query"`
}

type SearchResult struct {
    Title     string    `json:"title"`
    Type      string    `json:"type"`
    Rating    float64   `json:"rating"`
    Timestamp time.Time `json:"timestamp"`
}

type SearchResponse struct {
    Results  []SearchResult `json:"results"`
    Cached   bool           `json:"is_cached"`
    SearchID string         `json:"search_id"`
}

var SearchCache map[string][]SearchResult
const cacheExpiration = 30 * time.Second

// SearchHandler function to handle /search endpoint
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

		// Check if the search query exists in the cache
		cachedResults, found := SearchCache[request.SearchQuery]
		if found && !isExpired(cachedResults[0].Timestamp) {
			// If search query is found in the cache and it's not expired, return cached results
			response := SearchResponse{
				Results:  cachedResults,
				Cached:   true,
				SearchID: generateSearchID(),
			}
			sendResponse(w, response)
			return
		}

		// Perform search in the database
		results, err := performSearch(db, request.SearchQuery)
		if err != nil {
			http.Error(w, "Error performing search", http.StatusInternalServerError)
			return
		}

		// Cache the search results
		SearchCache[request.SearchQuery] = results

		// Construct response with search results
		response := SearchResponse{
			Results:  results,
			Cached:   false,
			SearchID: generateSearchID(),
		}

		// Send response back to client
		sendResponse(w, response)
	}
}

// sendResponse sends the response back to the client
func sendResponse(w http.ResponseWriter, response SearchResponse) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}


// performSearch executes a SQL query to perform the search operation in the database
func performSearch(db *sql.DB, searchQuery string) ([]SearchResult, error) {
	// Prepare SQL query to search for books and movies based on the search query
	query := `
		SELECT title, ratings_count, "book" AS type
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

func generateSearchID() string {
    // Set the seed for the random number generator based on current time
    rand.Seed(time.Now().UnixNano())

    // Define the characters that can be used in the random string
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

    // Define the length of the random string
    const length = 12

    // Create a byte slice to store the random string
    var result []byte
    for i := 0; i < length; i++ {
        // Generate a random index within the length of the charset
        index := rand.Intn(len(charset))
        // Append a randomly selected character from the charset to the result slice
        result = append(result, charset[index])
    }

    // Convert the byte slice to a string and return it as the search ID
    return string(result)
}

// isExpired checks if a cached entry is expired based on its timestamp
func isExpired(timestamp time.Time) bool {
	return time.Since(timestamp) > cacheExpiration
}