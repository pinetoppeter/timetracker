package commands

import (
	"fmt"

	"github.com/pinetoppeter/timetracker/internal/session"
	"github.com/spf13/cobra"
)

func NewPauseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pause",
		Short: "Pause the current session",
		Long:  "Pause the current time tracking session and its running record.",
		Run: func(cmd *cobra.Command, args []string) {
			err := session.PauseSession()
			if err != nil {
				fmt.Printf("Error pausing session: %v\n", err)
				return
			}
			fmt.Println("Session paused")
		},
	}
}
