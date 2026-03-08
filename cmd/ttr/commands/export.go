package commands

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pinetoppeter/timetracker/internal/storage"
	"github.com/spf13/cobra"
)

func NewExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export records to CSV",
		Long:  "Export all records from all sessions to a CSV file, accumulating durations by record name and including metadata properties.",
		Run: func(cmd *cobra.Command, args []string) {
			// Get month and year from user input
			month, year := getMonthAndYear()
			if month == 0 || year == 0 {
				fmt.Println("Invalid month or year")
				return
			}

			// Create storage instance
			store, err := storage.NewStorage()
			if err != nil {
				fmt.Printf("Error creating storage: %v\n", err)
				return
			}

			// Get all sessions for the specified month/year
			sessions, err := store.GetSessionsByMonthYear(month, year)
			if err != nil {
				fmt.Printf("Error getting sessions: %v\n", err)
				return
			}

			if len(sessions) == 0 {
				fmt.Printf("No sessions found for %d-%02d\n", year, month)
				return
			}

			// Process records and accumulate durations
			recordDurations := make(map[string]time.Duration)

			for _, session := range sessions {
				for _, record := range session.Records {
					if record.Name == "" {
						continue // Skip unnamed records
					}

					duration := calculateRecordDuration(record)
					if duration > 0 {
						recordDurations[record.Name] += duration
					}
				}
			}

			if len(recordDurations) == 0 {
				fmt.Println("No named records found for export")
				return
			}

			// Load export schema to get predefined columns
			schema, err := store.LoadExportSchema()
			if err != nil {
				fmt.Printf("Warning: Could not load export schema: %v\n", err)
				schema = storage.NewDefaultExportSchema()
			}

			// Parse schema-defined columns
			columns, err := store.ParseExportColumns(schema)
			if err != nil {
				fmt.Printf("Warning: Could not parse export schema columns: %v\n", err)
				columns = []storage.ExportColumn{}
			}

			// Load metadata for all unique record names from record JSON files
			recordMetadata := make(map[string]*storage.RecordMetadata)

			for recordName := range recordDurations {
				metadata, err := store.LoadRecordMetadata(recordName)
				if err != nil {
					fmt.Printf("Warning: Could not load metadata for %s: %v\n", recordName, err)
					continue
				}
				recordMetadata[recordName] = metadata
			}

			// Create CSV file in .timetracker directory
			filename := fmt.Sprintf("timetracker-export-%d-%02d.csv", year, month)
			filePath := filepath.Join(store.GetBaseDir(), filename)
			file, err := os.Create(filePath)
			if err != nil {
				fmt.Printf("Error creating CSV file: %v\n", err)
				return
			}
			defer file.Close()

			writer := csv.NewWriter(file)
			defer writer.Flush()

			// Write header: Record Name, Total Duration (hours), [schema-defined columns]
			header := []string{"Record Name", "Total Duration (hours)"}
			columnNames := storage.GetSchemaColumnNames(columns)
			for _, colName := range columnNames {
				header = append(header, colName)
			}

			if err := writer.Write(header); err != nil {
				fmt.Printf("Error writing CSV header: %v\n", err)
				return
			}

			// Write data
			for name, duration := range recordDurations {
				hours := duration.Hours()
				row := []string{name, fmt.Sprintf("%.2f", hours)}

				// Add schema-defined metadata properties
				metadata := recordMetadata[name]
				if metadata != nil {
					for _, col := range columns {
						if val, exists := metadata.Metadata[col.Name]; exists {
							// Convert metadata value to string
							var valStr string
							switch v := val.(type) {
							case string:
								valStr = v
							case int, int32, int64:
								valStr = fmt.Sprintf("%d", v)
							case float32, float64:
								valStr = fmt.Sprintf("%f", v)
							case bool:
								valStr = fmt.Sprintf("%t", v)
							default:
								valStr = fmt.Sprintf("%v", v)
							}
							row = append(row, valStr)
						} else {
							// Use default value if available, otherwise empty
							if col.Default != "" {
								row = append(row, col.Default)
							} else {
								row = append(row, "")
							}
						}
					}
				} else {
					// Add empty columns for all schema properties
					for range columns {
						row = append(row, "")
					}
				}

				if err := writer.Write(row); err != nil {
					fmt.Printf("Error writing CSV data: %v\n", err)
					return
				}
			}

			fmt.Printf("Export completed successfully: %s\n", filePath)
			fmt.Printf("Exported %d unique record names\n", len(recordDurations))
			fmt.Printf("Included %d schema-defined columns\n", len(columns))
		},
	}
}

func getMonthAndYear() (int, int) {
	now := time.Now()

	// Default to last month of current year
	month := int(now.Month())
	year := now.Year()

	if month == 1 {
		// If current month is January, default to December of previous year
		month = 12
		year--
	} else {
		month--
	}

	fmt.Printf("Export records for month (1-12) [%d]: ", month)
	var monthInput string
	fmt.Scanln(&monthInput)

	if monthInput != "" {
		parsedMonth, err := strconv.Atoi(monthInput)
		if err == nil && parsedMonth >= 1 && parsedMonth <= 12 {
			month = parsedMonth
		}
	}

	fmt.Printf("Export records for year [%d]: ", year)
	var yearInput string
	fmt.Scanln(&yearInput)

	if yearInput != "" {
		parsedYear, err := strconv.Atoi(yearInput)
		if err == nil && parsedYear > 0 {
			year = parsedYear
		}
	}

	return month, year
}

func calculateRecordDuration(record storage.Record) time.Duration {
	if record.StartTime.IsZero() {
		return 0
	}

	endTime := time.Now()
	if record.EndTime != nil {
		endTime = *record.EndTime
	}

	return endTime.Sub(record.StartTime)
}
