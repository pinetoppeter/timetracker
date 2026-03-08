package commands

import (
	"fmt"

	"github.com/pinetoppeter/timetracker/internal/session"
	"github.com/spf13/cobra"
)

func NewMetaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meta [record-name] [key] [value]",
		Short: "Add metadata to a record",
		Long:  "Add metadata to a record. If record-name is provided, adds metadata to that specific record. If no record-name is provided, adds metadata to the currently running record. Metadata can be used for categorization, project tracking, billing, etc.",
		Args:  cobra.RangeArgs(2, 3),
		Run: func(cmd *cobra.Command, args []string) {
			var recordName, key, value string
			
			if len(args) == 3 {
				// Specific record metadata
				recordName = args[0]
				key = args[1]
				value = args[2]
				
				err := session.AddMetadataToRecord(recordName, key, value)
				if err != nil {
					fmt.Printf("Error adding metadata to record '%s': %v\n", recordName, err)
					return
				}
				fmt.Printf("Added metadata to record '%s': %s=%s\n", recordName, key, value)
			} else {
				// Current record metadata (backward compatibility)
				key = args[0]
				value = args[1]
				
				err := session.AddMetadataToCurrentRecord(key, value)
				if err != nil {
					fmt.Printf("Error adding metadata: %v\n", err)
					return
				}
				fmt.Printf("Added metadata: %s=%s\n", key, value)
			}
		},
	}

	// Add a subcommand to show current metadata
	listCmd := &cobra.Command{
		Use:   "list [record-name]",
		Short: "List metadata for a record",
		Long:  "Show all metadata associated with a record. If record-name is provided, shows metadata for that specific record. If no record-name is provided, shows metadata for the currently running record.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var metadata map[string]interface{}
			var err error
			
			if len(args) == 1 {
				// Specific record metadata
				recordName := args[0]
				metadata, err = session.GetRecordMetadata(recordName)
				if err != nil {
					fmt.Printf("Error getting metadata for record '%s': %v\n", recordName, err)
					return
				}
				
				if len(metadata) == 0 {
					fmt.Printf("No metadata found for record '%s'.\n", recordName)
					return
				}
				
				fmt.Printf("Metadata for record '%s':\n", recordName)
			} else {
				// Current record metadata (backward compatibility)
				metadata, err = session.GetCurrentRecordMetadata()
				if err != nil {
					fmt.Printf("Error getting metadata: %v\n", err)
					return
				}

				if len(metadata) == 0 {
					fmt.Println("No metadata found for current record.")
					return
				}

				fmt.Println("Current record metadata:")
			}

			for key, value := range metadata {
				fmt.Printf("  %s: %v\n", key, value)
			}
		},
	}

	cmd.AddCommand(listCmd)

	return cmd
}
