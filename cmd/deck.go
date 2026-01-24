package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/atdrendel/ankigo/internal/ankiconnect"
	"github.com/spf13/cobra"
)

// Client interface for dependency injection in tests
type Client interface {
	DeckNames() ([]string, error)
	DeckNamesAndIds() (map[string]int64, error)
	GetDeckStats(decks []string) (map[int64]ankiconnect.DeckStats, error)
}

// deckListFields are the available fields for deck list output.
var deckListFields = []string{"id", "name", "new", "learn", "review", "total"}

// statsFields are fields that require fetching deck stats.
var statsFields = []string{"new", "learn", "review", "total"}

// deckListOptions holds the options for the deck list command.
type deckListOptions struct {
	json   bool
	fields []string
}

var jsonOutput bool
var fieldsFlag string

var deckCmd = &cobra.Command{
	Use:   "deck",
	Short: "Manage Anki decks",
	Long:  `Commands for listing, creating, and managing Anki decks.`,
}

var deckListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all decks",
	Long:  `List all decks in your Anki collection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := ankiconnect.DefaultClient()
		json, _ := cmd.Flags().GetBool("json")
		fieldsStr, _ := cmd.Flags().GetString("fields")

		fields, err := parseFields(fieldsStr, deckListFields)
		if err != nil {
			return err
		}

		opts := deckListOptions{json: json, fields: fields}
		return runDeckList(client, cmd.OutOrStdout(), opts)
	},
}

// parseFields parses a comma-separated field string and validates against available fields.
func parseFields(fieldsStr string, available []string) ([]string, error) {
	if fieldsStr == "" {
		return nil, nil
	}
	fields := strings.Split(fieldsStr, ",")
	for _, f := range fields {
		if !contains(available, f) {
			return nil, fmt.Errorf("unknown field: %s", f)
		}
	}
	return fields, nil
}

// contains checks if a string slice contains a value.
func contains(slice []string, value string) bool {
	for _, s := range slice {
		if s == value {
			return true
		}
	}
	return false
}

// needsStats returns true if any of the fields require fetching deck stats.
func needsStats(fields []string) bool {
	for _, f := range fields {
		if contains(statsFields, f) {
			return true
		}
	}
	return false
}

// deckEntry holds deck data for output.
type deckEntry struct {
	id    int64
	name  string
	stats *ankiconnect.DeckStats
}

// statValue returns a stat value, or 0 if stats are nil.
func (e *deckEntry) statValue(getter func(*ankiconnect.DeckStats) int) int {
	if e.stats == nil {
		return 0
	}
	return getter(e.stats)
}

// runDeckList is the testable implementation of deck list.
func runDeckList(client Client, out io.Writer, opts deckListOptions) error {
	// Validate fields first
	if opts.fields != nil {
		for _, f := range opts.fields {
			if !contains(deckListFields, f) {
				return fmt.Errorf("unknown field: %s", f)
			}
		}
	}

	if opts.json {
		return runDeckListJSON(client, out, opts.fields)
	}
	return runDeckListText(client, out, opts.fields)
}

func runDeckListText(client Client, out io.Writer, fields []string) error {
	// If no fields specified, default to ["name"] for backwards compatibility
	if fields == nil {
		fields = []string{"name"}
	}

	wantsID := contains(fields, "id")
	wantsStats := needsStats(fields)

	// Build entries based on what data we need
	var entries []deckEntry

	if wantsID || wantsStats {
		// Need IDs for stats lookup or direct output
		deckMap, err := client.DeckNamesAndIds()
		if err != nil {
			return fmt.Errorf("failed to get decks: %w", err)
		}

		if len(deckMap) == 0 {
			fmt.Fprintln(out, "No decks found")
			return nil
		}

		for name, id := range deckMap {
			entries = append(entries, deckEntry{id: id, name: name})
		}
	} else {
		// Only need names
		decks, err := client.DeckNames()
		if err != nil {
			return fmt.Errorf("failed to get deck names: %w", err)
		}

		if len(decks) == 0 {
			fmt.Fprintln(out, "No decks found")
			return nil
		}

		for _, name := range decks {
			entries = append(entries, deckEntry{name: name})
		}
	}

	// Sort by name for consistent output
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	// Fetch stats if needed
	if wantsStats {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.name
		}

		stats, err := client.GetDeckStats(names)
		if err != nil {
			return fmt.Errorf("failed to get deck stats: %w", err)
		}

		// Merge stats into entries
		for i := range entries {
			if s, ok := stats[entries[i].id]; ok {
				entries[i].stats = &s
			}
		}
	}

	// Output
	for _, e := range entries {
		var vals []string
		for _, f := range fields {
			switch f {
			case "id":
				vals = append(vals, fmt.Sprintf("%d", e.id))
			case "name":
				vals = append(vals, e.name)
			case "new":
				vals = append(vals, fmt.Sprintf("%d", e.statValue(func(s *ankiconnect.DeckStats) int { return s.NewCount })))
			case "learn":
				vals = append(vals, fmt.Sprintf("%d", e.statValue(func(s *ankiconnect.DeckStats) int { return s.LearnCount })))
			case "review":
				vals = append(vals, fmt.Sprintf("%d", e.statValue(func(s *ankiconnect.DeckStats) int { return s.ReviewCount })))
			case "total":
				vals = append(vals, fmt.Sprintf("%d", e.statValue(func(s *ankiconnect.DeckStats) int { return s.TotalInDeck })))
			}
		}
		fmt.Fprintln(out, strings.Join(vals, "\t"))
	}
	return nil
}

func runDeckListJSON(client Client, out io.Writer, fields []string) error {
	// If no fields specified, default to all fields
	if fields == nil {
		fields = deckListFields
	}

	// Always need IDs for JSON (for stats lookup and id field)
	deckMap, err := client.DeckNamesAndIds()
	if err != nil {
		return fmt.Errorf("failed to get decks: %w", err)
	}

	// Build entries
	var entries []deckEntry
	for name, id := range deckMap {
		entries = append(entries, deckEntry{id: id, name: name})
	}

	// Sort by name for consistent output
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	// Fetch stats if needed
	if needsStats(fields) {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.name
		}

		stats, err := client.GetDeckStats(names)
		if err != nil {
			return fmt.Errorf("failed to get deck stats: %w", err)
		}

		// Merge stats into entries
		for i := range entries {
			if s, ok := stats[entries[i].id]; ok {
				entries[i].stats = &s
			}
		}
	}

	// Build output based on requested fields
	var result []map[string]interface{}
	for _, e := range entries {
		obj := make(map[string]interface{})
		for _, f := range fields {
			switch f {
			case "id":
				obj["id"] = e.id
			case "name":
				obj["name"] = e.name
			case "new":
				obj["new"] = e.statValue(func(s *ankiconnect.DeckStats) int { return s.NewCount })
			case "learn":
				obj["learn"] = e.statValue(func(s *ankiconnect.DeckStats) int { return s.LearnCount })
			case "review":
				obj["review"] = e.statValue(func(s *ankiconnect.DeckStats) int { return s.ReviewCount })
			case "total":
				obj["total"] = e.statValue(func(s *ankiconnect.DeckStats) int { return s.TotalInDeck })
			}
		}
		result = append(result, obj)
	}

	// Ensure empty array instead of null
	if result == nil {
		result = []map[string]interface{}{}
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

var deckCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new deck",
	Long:  `Create a new deck with the specified name.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		fmt.Fprintf(cmd.OutOrStdout(), "deck create %q: not yet implemented\n", name)
		return nil
	},
}

func init() {
	deckListCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	deckListCmd.Flags().StringVarP(&fieldsFlag, "fields", "f", "", "Comma-separated list of fields (available: id, name, new, learn, review, total)")
	deckCmd.AddCommand(deckListCmd)
	deckCmd.AddCommand(deckCreateCmd)
	rootCmd.AddCommand(deckCmd)
}
