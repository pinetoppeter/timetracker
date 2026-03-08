package commands

import (
	"fmt"

	"github.com/pinetoppeter/timetracker/internal/session"
	"github.com/spf13/cobra"
)

func NewResumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resume",
		Short: "Resume the last named record",
		Long:  "End the current record and start a new one with the same name as the last named record in the session.",
		Run: func(cmd *cobra.Command, args []string) {
			err := session.ResumeLastRecord()
			if err != nil {
				fmt.Printf("Error resuming record: %v\n", err)
				return
			}
			fmt.Println("⏯️  Resumed last record")
		},
	}
}
