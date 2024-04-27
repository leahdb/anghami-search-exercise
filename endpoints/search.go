package endpoints

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/meilisearch/meilisearch-go"
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

// CacheEntry represents an entry in the cache
type CacheEntry struct {
    Results   []SearchResult
    Timestamp time.Time
}

var searchCache sync.Map
const cacheExpiration = 30 * time.Second

// Initialize MeiliSearch client
var client *meilisearch.Client

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	meiliHost := os.Getenv("MEILISEARCH_HOST")
	meiliKey := os.Getenv("MEILISEARCH_KEY")

	client = meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   meiliHost,
		APIKey: meiliKey,
	})

	client.CreateIndex(&meilisearch.IndexConfig{
		Uid: "books",
		PrimaryKey: "bookID",
	})
	
	client.CreateIndex(&meilisearch.IndexConfig{
		Uid: "movies",
		PrimaryKey: "Title",
	})
}

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
		cachedResults, found := searchCache.Load(request.SearchQuery)
		if found && !isExpired(cachedResults.(CacheEntry).Timestamp) {
			// If search query is found in the cache and it's not expired, return cached results
			response := SearchResponse{
				Results:  cachedResults.(CacheEntry).Results,
				Cached:   true,
				SearchID: generateSearchID(),
			}
			sendResponse(w, response)
			return
		}

		// Perform search in the database
		results, err := performSearch(request.SearchQuery)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Error performing search", http.StatusInternalServerError)
			return
		}

		// Cache the search results
		searchCache.Store(request.SearchQuery, CacheEntry{Results: results, Timestamp: time.Now()})

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


// performSearch performs a search using MeiliSearch
func performSearch(searchQuery string) ([]SearchResult, error) {
    // Perform search on the "books" index
    searchRes, err := client.Index("books").Search(searchQuery, &meilisearch.SearchRequest{
        Limit: 10,
    })
    if err != nil {
        return nil, err
    }

    // Process search response and extract search results
    var results []SearchResult
    for _, hit := range searchRes.Hits {
        // Convert hit to map[string]interface{}
        hitMap := hit.(map[string]interface{})

        // Retrieve fields from the hit map
        title, okTitle := hitMap["title"].(string)
        itemType, okType := hitMap["type"].(string)
        rating, okRating := hitMap["rating"].(float64)

        // Check if all required fields exist
        if !okTitle || !okType || !okRating {
            return nil, fmt.Errorf("missing or invalid fields in search result")
        }

        // Create SearchResult object
        result := SearchResult{
            Title:  title,
            Type:   itemType,
            Rating: rating,
        }
        results = append(results, result)
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