package commands

import (
	"fmt"
	"strings"

	"github.com/pinetoppeter/timetracker/internal/session"
	"github.com/pinetoppeter/timetracker/internal/storage"
	"github.com/spf13/cobra"
)

func NewEndCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "end [record-names...]",
		Aliases: []string{"stop"},
		Short:   "End the current session",
		Long:    "End the current time tracking session. Optionally provide names for unnamed records as arguments.",
		Run: func(cmd *cobra.Command, args []string) {
			var result *session.SessionEndResult
			var err error

			if len(args) > 0 {
				// Use provided record names
				result, err = session.EndSessionWithRecordNames(args)
			} else {
				// Use interactive prompting
				result, err = session.EndSession()
			}

			if err != nil {
				fmt.Printf("Error ending session: %v\n", err)
				return
			}

			fmt.Printf("🏁 Session %s ended\n", result.Session.ID)
			if result.RecordName != "" {
				fmt.Printf("   Final record: %s\n", result.RecordName)
			} else {
				fmt.Printf("   Final record: <unnamed>\n")
			}
			if result.RecordsNamed > 0 {
				fmt.Printf("   Named %d record(s) during session end\n", result.RecordsNamed)
			}
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// Provide tab completion for record names
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
		},
	}
}
