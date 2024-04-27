package analytics

import (
	"anghami-exercise/endpoints"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// Insights represents the daily insights
type Insights struct {
	Top10Clicked       []string  `json:"top_10_clicked"`
	AverageClickPos    float64   `json:"average_click_position"`
	TotalSearches      int       `json:"total_searches"`
	TotalClicks        int       `json:"total_clicks"`
	ClickThroughRate   float64   `json:"click_through_rate"`
	Date               time.Time `json:"date"`
}

func FetchClickData(db *sql.DB) ([]endpoints.ClickData, error) {
	// Query to fetch click data from the last 24 hours
	query := `
		SELECT search_id, result_type, result_id, result_position, timestamp
		FROM search_clicks
		WHERE timestamp >= NOW() - INTERVAL 1 DAY
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fmt.Println(rows)

	var clickData []endpoints.ClickData
	for rows.Next() {
		var click endpoints.ClickData
		var timestampStr string
		err := rows.Scan(&click.SearchID, &click.ResultType, &click.ResultID, &click.ResultPosition, &timestampStr)
		if err != nil {
			return nil, err
		}
		click.Timestamp, err = time.Parse("2006-01-02 15:04:05", timestampStr)
		if err != nil {
			return nil, err
		}
		clickData = append(clickData, click)
	}

	return clickData, nil
}

func GenerateInsights(clickData []endpoints.ClickData) Insights {
	// Initialize variables to store insights
	var (
		top10ClickedMap  = make(map[string]int)
		totalClicks      int
		totalSearches    = len(clickData)
		totalClickPosSum int
	)

	// Calculate total clicks and click position sum
	for _, click := range clickData {
		totalClicks++
		totalClickPosSum += click.ResultPosition

		// Count clicks for each result type
		top10ClickedMap[click.ResultType]++
	}

	// Calculate average click position
	var averageClickPos float64
	if totalClicks > 0 {
		averageClickPos = float64(totalClickPosSum) / float64(totalClicks)
	}

	// Sort the top 10 clicked items
	top10Clicked := make([]string, 0, 10)
	for resultType, clickCount := range top10ClickedMap {
		top10Clicked = append(top10Clicked, fmt.Sprintf("%s: %d clicks", resultType, clickCount))
	}
	sort.Slice(top10Clicked, func(i, j int) bool {
		return top10ClickedMap[top10Clicked[i]] > top10ClickedMap[top10Clicked[j]]
	})
	if len(top10Clicked) > 10 {
		top10Clicked = top10Clicked[:10]
	}

	// Calculate click-through rate
	var clickThroughRate float64
	if totalSearches > 0 {
		clickThroughRate = (float64(totalClicks) / float64(totalSearches)) * 100
	}

	// Get current date
	date := time.Now()

	return Insights{
		Top10Clicked:     top10Clicked,
		AverageClickPos:  averageClickPos,
		TotalSearches:    totalSearches,
		TotalClicks:      totalClicks,
		ClickThroughRate: clickThroughRate,
		Date:             date,
	}
}


func SaveInsightsToFile(insights Insights) error {
	// Serialize insights to JSON
	insightsJSON, err := json.MarshalIndent(insights, "", "  ")
	if err != nil {
		return err
	}

	// Create folder if it doesn't exist
	err = os.MkdirAll("Insights", 0755)
	if err != nil {
		return err
	}

	// Create filename based on current date (YYYY-MM-DD format) inside the Insights folder
	filename := "Insights/" + time.Now().Format("2006-01-02") + ".json"

	// Write JSON data to file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(insightsJSON)
	if err != nil {
		return err
	}

	fmt.Println("Insights saved to", filename)
	return nil
}
