package commands

import (
	"fmt"
	"strings"

	"github.com/pinetoppeter/timetracker/internal/session"
	"github.com/pinetoppeter/timetracker/internal/storage"
	"github.com/spf13/cobra"
)

func NewStartCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "start [record-name]",
		Short: "Start a new time tracking session",
		Long:  "Start a new time tracking session for the day. If a session already exists, it will be resumed. Optionally name the first record.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				name = args[0]
			}

			// Check if there's already an active session
			currentSession, err := session.GetCurrentSession()
			if err == nil && currentSession != nil {
				// If the session was ended (stopped), treat it as no session and start a new one
				if currentSession.EndTime != nil {
					// Session was stopped, so start a new session instead of resuming
					// (fall through to new session creation below)
				} else {
					// Session exists and is active, switch to the new record
					err := session.SwitchAndNameRecord(name)
					if err != nil {
						fmt.Printf("Error switching record: %v\n", err)
						return
					}
					fmt.Printf("🔄 Resumed session %s\n", currentSession.ID)
					fmt.Printf("   Switched to record: %s\n", name)
					return
				}
			}

			// No existing session, start a new one
			result, err := session.StartSessionWithName(name)
			if err != nil {
				fmt.Printf("Error starting session: %v\n", err)
				return
			}

			if result.IsNewSession {
				fmt.Printf("🚀 Started new session %s\n", result.Session.ID)
				fmt.Printf("   Current record: %s\n", result.RecordName)
			} else {
				fmt.Printf("🔄 Resumed session %s\n", result.Session.ID)
				if result.RecordName != "" {
					fmt.Printf("   Current record: %s\n", result.RecordName)
				} else {
					fmt.Printf("   Current record: <unnamed>\n")
				}
			}
		},
	}

	// Add completion function for record names (argument completion)
	cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveDefault
		}

		store, err := storage.NewStorage()
		if err != nil {
			return nil, cobra.ShellCompDirectiveDefault
		}

		// Get all record names
		names, err := store.GetRecordNames()
		if err != nil {
			return nil, cobra.ShellCompDirectiveDefault
		}

		// Filter names based on what the user has typed so far
		var matches []string
		for _, name := range names {
			if strings.HasPrefix(name, toComplete) {
				matches = append(matches, name)
			}
		}

		return matches, cobra.ShellCompDirectiveDefault
	}

	return cmd
}
