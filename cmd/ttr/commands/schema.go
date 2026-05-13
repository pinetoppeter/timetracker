package commands

import (
	"fmt"

	"github.com/pinetoppeter/timetracker/internal/storage"
	"github.com/spf13/cobra"
)

func NewSchemaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Manage export schema",
		Long:  "Manage the TimeTracker export schema for CSV exports. The schema defines which metadata fields (columns) are available for export.",
	}

	// Add list subcommand to show all schema fields
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all meta fields in the export schema",
		Long:  "Display all metadata fields (columns) defined in the export schema, including their names, display names, types, descriptions, and default values.",
		Run: func(cmd *cobra.Command, args []string) {
			// Create storage instance
			store, err := storage.NewStorage()
			if err != nil {
				fmt.Printf("Error creating storage: %v\n", err)
				return
			}

			// Load export schema
			schema, err := store.LoadExportSchema()
			if err != nil {
				fmt.Printf("Error loading export schema: %v\n", err)
				return
			}

			// Check if schema has columns
			if len(schema.Columns) == 0 {
				fmt.Println("No meta fields defined in the export schema.")
				fmt.Println("Add fields by using 'ttr meta <record> <key> <value>' or edit the schema file manually.")
				fmt.Printf("Schema file: %s\n", store.GetExportSchemaPath())
				return
			}

			fmt.Printf("Meta fields in export schema (%d total):\n\n", len(schema.Columns))

			// Display each column with its properties
			for i, col := range schema.Columns {
				fmt.Printf("%d. %s\n", i+1, col.Name)
				if col.DisplayName != "" && col.DisplayName != col.Name {
					fmt.Printf("   Display Name: %s\n", col.DisplayName)
				}
				if col.Type != "" {
					fmt.Printf("   Type: %s\n", col.Type)
				}
				if col.Description != "" {
					fmt.Printf("   Description: %s\n", col.Description)
				}
				if col.Default != "" {
					fmt.Printf("   Default: %s\n", col.Default)
				}
				fmt.Println()
			}

			fmt.Printf("Schema file: %s\n", store.GetExportSchemaPath())
		},
	}

	cmd.AddCommand(listCmd)

	return cmd
}
