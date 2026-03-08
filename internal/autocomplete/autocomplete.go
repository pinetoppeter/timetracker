package autocomplete

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pinetoppeter/timetracker/internal/storage"
)

type Autocomplete struct {
	storage *storage.Storage
}

func NewAutocomplete(store *storage.Storage) *Autocomplete {
	return &Autocomplete{storage: store}
}

func (a *Autocomplete) GetSuggestions(prefix string) ([]string, error) {
	// Get all existing record names
	names, err := a.storage.GetRecordNames()
	if err != nil {
		return nil, err
	}

	// Filter and sort suggestions
	var suggestions []string
	for _, name := range names {
		if prefix == "" || strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
			suggestions = append(suggestions, name)
		}
	}

	sort.Strings(suggestions)
	return suggestions, nil
}

func (a *Autocomplete) PrintSuggestions(prefix string) error {
	suggestions, err := a.GetSuggestions(prefix)
	if err != nil {
		return err
	}

	if len(suggestions) == 0 {
		return nil
	}

	fmt.Println("\nSuggestions:")
	for i, suggestion := range suggestions {
		fmt.Printf("%d. %s\n", i+1, suggestion)
	}
	fmt.Print("\nEnter record name: ")

	return nil
}
