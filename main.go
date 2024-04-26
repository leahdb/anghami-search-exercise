package main

import (
	"anghami-exercise/endpoints"
	"anghami-exercise/importCSV"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)


func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize the searchCache map
    endpoints.SearchCache = make(map[string][]endpoints.SearchResult)
	
	// Database connection setup
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	// Connect to MySQL database
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName))
	if err != nil {
		log.Fatalf("Error connecting to MySQL: %v", err)
	}
	defer db.Close()

	if err := importCSV.CreateTables(db); err != nil {
		log.Fatalf("Error creating tables: %v", err)
	}

	if err := importCSV.ImportDataFromCSV(db, "books.csv", "books"); err != nil {
		log.Fatalf("Error importing data from books CSV: %v", err)
	}

	if err := importCSV.ImportDataFromCSV(db, "movies.csv", "movies"); err != nil {
		log.Fatalf("Error importing data from movies CSV: %v", err)
	}

	fmt.Println("Data import successful!")

	// Define HTTP routes
	http.HandleFunc("/search", endpoints.SearchHandler(db))
	http.HandleFunc("/report-search", endpoints.ReportSearchHandler(db))
	http.HandleFunc("/report-click", endpoints.ReportClickHandler(db))

	// Start HTTP server
	http.ListenAndServe(":8080", nil)
}