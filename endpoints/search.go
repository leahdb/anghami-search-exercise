package endpoints

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"
)

type SearchRequest struct {
    SearchQuery string `json:"search_query"`
}

type SearchResult struct {
	ID        int       `json:"id"`
    Title     string    `json:"title"`
    Type      string    `json:"type"`
    Rating    float64   `json:"rating"`
    Timestamp time.Time `json:"timestamp"`
	RelevanceScore int
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
			fmt.Println(err)
			http.Error(w, "Error performing search", http.StatusInternalServerError)
			return
		}

		// Cache the search results
		SearchCache[request.SearchQuery] = results

		 // Calculate relevance score for each search result based on Levenshtein distance
        for i := range results {
            results[i].RelevanceScore = LevenshteinDistance(request.SearchQuery, results[i].Title)
        }

        // Sort the search results by relevance score
        sort.Slice(results, func(i, j int) bool {
            return results[i].RelevanceScore < results[j].RelevanceScore
        })

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
		SELECT bookID, title, average_rating, "book" AS type
		FROM books
		WHERE title LIKE ?
		UNION ALL
		SELECT movieID, title, rating, "movie" AS type
		FROM movies
		WHERE title LIKE ?
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
		var id int
		var title string
		var rating float64
		var itemType string
		err := rows.Scan(&id, &title, &rating, &itemType)
		if err != nil {
            return nil, err
        }

		result := SearchResult{
			ID:     id,
			Title:  title,
			Rating: rating,
			Type:   itemType,
		}
		results = append(results, result)
	}

	// Sort the results by relevance
    sortResults(results, searchQuery)

    // Check for any errors during iteration
    err = rows.Err()
    if err != nil {
        return nil, err
    }

	return results, nil
}

// sortResults sorts the search results by relevance
func sortResults(results []SearchResult, searchQuery string) {
    // Define a relevance score for each search result
    for i, result := range results {
        // Calculate the relevance score based on factors like exact match and partial match
        score := calculateRelevanceScore(result.Title, searchQuery)
        // Assign the relevance score to the search result
        results[i].RelevanceScore = score
    }

    // Sort the results by relevance score (in descending order)
    sort.Slice(results, func(i, j int) bool {
        // If relevance scores are equal, prioritize shorter titles
        if results[i].RelevanceScore == results[j].RelevanceScore {
            return len(results[i].Title) < len(results[j].Title)
        }
        return results[i].RelevanceScore > results[j].RelevanceScore
    })
}

// calculateRelevanceScore calculates the relevance score for a search result
func calculateRelevanceScore(title, searchQuery string) int {
    // Initialize the relevance score
    score := 0

    // Check for exact match
    if strings.EqualFold(title, searchQuery) {
        score += 1000 // Add a high score for exact match
    }

    // Check for partial match
    if strings.Contains(strings.ToLower(title), strings.ToLower(searchQuery)) {
        score += 500 // Add a score for partial match
    }

    // You can add more relevancy metrics here

    return score
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

// LevenshteinDistance calculates the Levenshtein distance between two strings
func LevenshteinDistance(s1, s2 string) int {
    m, n := len(s1), len(s2)
    matrix := make([][]int, m+1)
    for i := range matrix {
        matrix[i] = make([]int, n+1)
        matrix[i][0] = i
    }
    for j := 0; j <= n; j++ {
        matrix[0][j] = j
    }
    for i := 1; i <= m; i++ {
        for j := 1; j <= n; j++ {
            cost := 0
            if s1[i-1] != s2[j-1] {
                cost = 1
            }
            matrix[i][j] = min(matrix[i-1][j]+1, matrix[i][j-1]+1, matrix[i-1][j-1]+cost)
        }
    }
    return matrix[m][n]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
    if a <= b && a <= c {
        return a
    } else if b <= a && b <= c {
        return b
    }
    return c
}
