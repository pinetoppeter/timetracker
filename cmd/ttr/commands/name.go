package commands

import (
	"fmt"
	"strings"

	"github.com/pinetoppeter/timetracker/internal/autocomplete"
	"github.com/pinetoppeter/timetracker/internal/session"
	"github.com/pinetoppeter/timetracker/internal/storage"
	"github.com/spf13/cobra"
)

func NewNameCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "name [record-name]",
		Short: "Name the current record",
		Long:  "Set the name of the currently running record. Provides autocomplete suggestions from previous record names.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				name = args[0]
			} else {
				// Interactive mode with autocomplete
				name = promptForRecordName()
				if name == "" {
					return
				}
			}

			err := session.NameCurrentRecord(name)
			if err != nil {
				fmt.Printf("Error naming record: %v\n", err)
				return
			}
			fmt.Printf("Record named: %s\n", name)
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

func promptForRecordName() string {
	store, err := storage.NewStorage()
	if err != nil {
		fmt.Printf("Warning: Could not load storage for autocomplete: %v\n", err)
		return simplePrompt("")
	}

	ac := autocomplete.NewAutocomplete(store)

	// Get all suggestions first
	suggestions, err := ac.GetSuggestions("")
	if err != nil {
		fmt.Printf("Warning: Could not load suggestions: %v\n", err)
		return simplePrompt("")
	}

	if len(suggestions) == 0 {
		return simplePrompt("")
	}

	// Show suggestions and prompt
	fmt.Println("\nAvailable record names (type for suggestions):")
	for i, suggestion := range suggestions {
		fmt.Printf("  %d. %s\n", i+1, suggestion)
	}
	fmt.Println()

	return simplePrompt("")
}

func simplePrompt(defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("Enter record name [%s]: ", defaultValue)
	} else {
		fmt.Print("Enter record name: ")
	}

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return ""
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}

	return input
}
