package commands

import (
	"fmt"
	"time"

	"github.com/pinetoppeter/timetracker/internal/record"
	"github.com/pinetoppeter/timetracker/internal/session"
	"github.com/spf13/cobra"
)

func NewInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show session information",
		Long:  "Display information about the current session including the current record and all records with their durations.",
		Run: func(cmd *cobra.Command, args []string) {
			sess, err := session.GetCurrentSession()
			if err != nil {
				if err.Error() == "no current session found" {
					fmt.Println("🔍 No active session found. Start a new session with 'ttr start'")
				} else {
					fmt.Printf("Error getting session: %v\n", err)
				}
				return
			}

			if sess == nil {
				fmt.Println("🔍 No active session found. Start a new session with 'ttr start'")
				return
			}

			fmt.Printf("📋 Session %s (started: %s)\n", sess.ID, sess.StartTime.Format("2006-01-02 15:04:05"))
			if sess.EndTime != nil {
				fmt.Printf("   Ended: %s\n", sess.EndTime.Format("2006-01-02 15:04:05"))
			}

			// Find current record
			var currentRecord *record.Record
			for _, rec := range sess.Records {
				if rec.ID == sess.CurrentRecordID {
					currentRecord = rec
					break
				}
			}

			// Show current record
			if currentRecord != nil {
				fmt.Printf("\n🎯 Current Record:\n")
				fmt.Printf("   Name: %s\n", getRecordName(currentRecord))
				fmt.Printf("   Started: %s\n", currentRecord.StartTime.Format("15:04:05"))
				if currentRecord.EndTime != nil {
					fmt.Printf("   Ended: %s\n", currentRecord.EndTime.Format("15:04:05"))
					duration := currentRecord.EndTime.Sub(currentRecord.StartTime)
					fmt.Printf("   Duration: %s\n", formatDuration(duration))
				} else {
					duration := time.Since(currentRecord.StartTime)
					fmt.Printf("   Duration: %s (running)\n", formatDuration(duration))
				}
			}

			// Show all records
			fmt.Printf("\n📝 All Records (%d total):\n", len(sess.Records))
			for i, rec := range sess.Records {
				fmt.Printf("   %d. %s\n", i+1, getRecordName(rec))
				fmt.Printf("      Started: %s", rec.StartTime.Format("15:04:05"))
				if rec.EndTime != nil {
					fmt.Printf(", Ended: %s", rec.EndTime.Format("15:04:05"))
					duration := rec.EndTime.Sub(rec.StartTime)
					fmt.Printf(", Duration: %s\n", formatDuration(duration))
				} else {
					duration := time.Since(rec.StartTime)
					fmt.Printf(", Duration: %s (running)\n", formatDuration(duration))
				}
			}

			// Show session summary
			fmt.Printf("\n📊 Session Summary:\n")
			totalDuration := calculateTotalDuration(sess.Records)
			fmt.Printf("   Total duration: %s\n", formatDuration(totalDuration))

			namedCount := 0
			for _, rec := range sess.Records {
				if rec.Name != "" {
					namedCount++
				}
			}
			fmt.Printf("   Named records: %d/%d\n", namedCount, len(sess.Records))
		},
	}
}

func getRecordName(rec *record.Record) string {
	if rec.Name != "" {
		return rec.Name
	}
	return "<unnamed>"
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func calculateTotalDuration(records []*record.Record) time.Duration {
	var total time.Duration
	for _, rec := range records {
		endTime := time.Now()
		if rec.EndTime != nil {
			endTime = *rec.EndTime
		}
		total += endTime.Sub(rec.StartTime)
	}
	return total
}
