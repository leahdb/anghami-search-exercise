package importCSV

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type CustomCSVReader struct {
    *csv.Reader
}

// NewCustomCSVReader creates a new CustomCSVReader instance
func NewCustomCSVReader(r io.Reader) *CustomCSVReader {
    csvReader := csv.NewReader(r)
    csvReader.LazyQuotes = true // Allow quotes inside quoted fields
    return &CustomCSVReader{csvReader}
}

// Read reads a single CSV record from the underlying reader and replaces problematic quotes
func (r *CustomCSVReader) Read() ([]string, error) {
    record, err := r.Reader.Read()
    if err != nil {
        return nil, err
    }

    // Replace problematic quotes in each field
    for i := range record {
        record[i] = strings.ReplaceAll(record[i], `"`, `""`)
    }

    return record, nil
}


func ImportDataFromCSV(db *sql.DB, fileName, tableName string) error {
    fmt.Printf("Importing %s ...\n", tableName)
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("error opening %s CSV file: %v", fileName, err)
	}
	defer file.Close()

	reader := NewCustomCSVReader(file)
	if err != nil {
		return fmt.Errorf("error creating custom CSV reader: %v", err)
	}

	return importData(db, reader, tableName)
}


// createTables creates tables in the database if they don't already exist
func CreateTables(db *sql.DB) error {
	// Table creation statements
	tables := map[string]string{
		"movies": "",
	}

	// Open CSV files to get column names
	for tableName := range tables {
		csvFile, err := os.Open(tableName + ".csv")
		if err != nil {
			return fmt.Errorf("error opening %s.csv file: %v", tableName, err)
		}
		defer csvFile.Close()

		csvReader := csv.NewReader(csvFile)
		headers, err := csvReader.Read()
		if err != nil {
			return fmt.Errorf("error reading CSV header for %s.csv: %v", tableName, err)
		}

		// Sanitize column names and enclose them in backticks
        sanitizedHeaders := make([]string, len(headers))
        for i, header := range headers {
            sanitizedHeaders[i] = sanitizeColumnName(header)
        }

		// Construct CREATE TABLE statement
		tables[tableName] = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", tableName)
		for i, header := range sanitizedHeaders {
			tables[tableName] += fmt.Sprintf("%s VARCHAR(255)", header)
			if i < len(headers)-1 {
				tables[tableName] += ", "
			}
		}
		tables[tableName] += ") ENGINE=InnoDB;"
	}

	// Execute table creation statements
	for _, query := range tables {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("error creating table: %v", err)
		}
	}

	return nil
}

func sanitizeColumnName(name string) string {
	name = strings.ToLower(name)
    // Remove double quotes and spaces, replace with underscores
    name = strings.ReplaceAll(name, `"`, "")
    name = strings.ReplaceAll(name, " ", "_")
    // Remove other special characters
    name = regexp.MustCompile("[^a-zA-Z0-9_]").ReplaceAllString(name, "")
    // Trim leading and trailing underscores
    name = strings.Trim(name, "_")
    return name
}

// importData imports data from a CSV reader into the specified table in the database
func importData(db *sql.DB, reader *CustomCSVReader, tableName string) error {

	//clearing the tables
	_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", tableName))
    if err != nil {
        return fmt.Errorf("error truncating table %s: %v", tableName, err)
    }

	skippedRecords := make(map[int]string)
    // Open log file for recording errors and skipped records
    logFile, err := os.OpenFile("import_errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
        return fmt.Errorf("error opening log file: %v", err)
    }

    defer logFile.Close()

    // Read CSV header
    headers, err := reader.Read()

    if err != nil {
        return fmt.Errorf("error reading CSV header: %v", err)
    }

	sanitizedHeaders := make([]string, len(headers))
	for i, header := range headers {
		sanitizedHeaders[i] = sanitizeColumnName(header)
	}

    // Prepare SQL statement for inserting data
    placeholders := make([]string, len(headers))
    for i := range placeholders {
        placeholders[i] = "?"
    }
    query := fmt.Sprintf("INSERT IGNORE INTO %s (%s) VALUES (%s)", tableName, strings.Join(sanitizedHeaders, ", "), strings.Join(placeholders, ", "))
    stmt, err := db.Prepare(query)
    if err != nil {
        return fmt.Errorf("error preparing SQL statement: %v", err)
    }
    defer stmt.Close()

    // Read CSV records and insert into database
    for index := 1; index<=3886; index++ {
        record, err := reader.Read()
		//40146 44439

		// Check if the number of fields matches the number of headers
        if len(record) != len(headers) {
            skippedRecords[index] = fmt.Sprintf("Incorrect number of fields: %v", record)
            continue
        }

        if err != nil {
            if err == io.EOF {
                break // End of file reached
            }
            return fmt.Errorf("error reading CSV record: %v", err)
        }

		// Replace slashes with commas for authors data if needed
        record[2] = strings.ReplaceAll(record[2], "/", ", ")

		for i, value := range record {
            if len(value) > 255 {
                record[i] = value[:255]
            }
        }

        // Execute SQL statement to insert record
		recordInterfaces := make([]interface{}, len(record))
		for i, v := range record {
			recordInterfaces[i] = v
		}
        if _, err := stmt.Exec(recordInterfaces...); err != nil {
            skippedRecords[index] = fmt.Sprintf("error inserting record into %s table: %v", tableName, err)
        }
    }

    // Log skipped records to file
    logWriter := io.MultiWriter(os.Stdout, logFile) // Write to both console and log file
    log.SetOutput(logWriter)
	for recordNumber, errorMessage := range skippedRecords {
        log.Printf("Skipped record %d from %s CSV: %s", recordNumber, tableName, errorMessage)
    }

    return nil
}