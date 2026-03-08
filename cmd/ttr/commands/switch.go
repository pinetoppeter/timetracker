package commands

import (
	"fmt"
	"strings"

	"github.com/pinetoppeter/timetracker/internal/autocomplete"
	"github.com/pinetoppeter/timetracker/internal/session"
	"github.com/pinetoppeter/timetracker/internal/storage"
	"github.com/spf13/cobra"
)

func NewSwitchCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "switch [record-name]",
		Short: "Switch to a new record",
		Long:  "Switch to a new record, ending the current running record. Optionally name the new record with autocomplete suggestions.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				name = args[0]
			} else {
				// Interactive mode with autocomplete
				name = promptForSwitchRecordName()
				if name == "" {
					return
				}
			}

			err := session.SwitchAndNameRecord(name)
			if err != nil {
				fmt.Printf("Error switching record: %v\n", err)
				return
			}
			fmt.Printf("Switched to new record: %s\n", name)
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

func promptForSwitchRecordName() string {
	store, err := storage.NewStorage()
	if err != nil {
		fmt.Printf("Warning: Could not load storage for autocomplete: %v\n", err)
		return simpleSwitchPrompt("")
	}

	ac := autocomplete.NewAutocomplete(store)

	// Get all suggestions first
	suggestions, err := ac.GetSuggestions("")
	if err != nil {
		fmt.Printf("Warning: Could not load suggestions: %v\n", err)
		return simpleSwitchPrompt("")
	}

	if len(suggestions) == 0 {
		return simpleSwitchPrompt("")
	}

	// Show suggestions and prompt
	fmt.Println("\nAvailable record names (type for suggestions):")
	for i, suggestion := range suggestions {
		fmt.Printf("  %d. %s\n", i+1, suggestion)
	}
	fmt.Println()

	return simpleSwitchPrompt("")
}

func simpleSwitchPrompt(defaultValue string) string {
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
