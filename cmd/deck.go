package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/atdrendel/ankigo/internal/ankiconnect"
	"github.com/spf13/cobra"
)

// Client interface for dependency injection in tests
type Client interface {
	DeckNames() ([]string, error)
	DeckNamesAndIds() (map[string]int64, error)
	GetDeckStats(decks []string) (map[int64]ankiconnect.DeckStats, error)
	CreateDeck(name string) (int64, error)
	DeleteDecks(decks []string) error
	FindCards(query string) ([]int64, error)
	CardsInfo(cardIDs []int64) ([]ankiconnect.CardInfo, error)
	AddNote(note ankiconnect.Note) (int64, error)
	DeleteNotes(notes []int64) error
	ModelNames() ([]string, error)
	ModelFieldNames(modelName string) ([]string, error)
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
		client := ankiconnect.DefaultClient()
		return runDeckCreate(client, cmd.OutOrStdout(), args[0])
	},
}

// runDeckCreate is the testable implementation of deck create.
func runDeckCreate(client Client, out io.Writer, name string) error {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return fmt.Errorf("deck name cannot be empty")
	}

	deckID, err := client.CreateDeck(name)
	if err != nil {
		return fmt.Errorf("failed to create deck: %w", err)
	}

	fmt.Fprintln(out, deckID)
	return nil
}

var deckDeleteCmd = &cobra.Command{
	Use:   "delete [deck-names...]",
	Short: "Delete one or more decks",
	Long:  `Delete one or more decks and all their cards. Requires --force flag or confirmation.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Silence usage for errors that happen during execution (not arg validation)
		cmd.SilenceUsage = true

		client := ankiconnect.DefaultClient()
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		byID, _ := cmd.Flags().GetBool("id")
		return runDeckDelete(client, os.Stdin, cmd.OutOrStdout(), cmd.ErrOrStderr(), args, force, dryRun, byID)
	},
}

// runDeckDelete is the testable implementation of deck delete.
func runDeckDelete(client Client, stdin io.Reader, stdout, stderr io.Writer, args []string, force, dryRun, byID bool) error {
	if len(args) == 0 {
		return fmt.Errorf("at least one deck name is required")
	}

	var decks []string
	var existingDecks map[string]struct{}

	if byID {
		// Parse IDs and resolve to names
		deckMap, err := client.DeckNamesAndIds()
		if err != nil {
			return fmt.Errorf("failed to get deck names: %w", err)
		}
		// Invert map: id -> name
		idToName := make(map[int64]string)
		for name, id := range deckMap {
			idToName[id] = name
		}
		for _, arg := range args {
			id, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid deck ID %q: must be a number", arg)
			}
			name, ok := idToName[id]
			if !ok {
				return fmt.Errorf("deck with ID %d not found", id)
			}
			decks = append(decks, name)
		}
		// All decks exist since we resolved them from IDs
		existingDecks = make(map[string]struct{})
		for _, name := range decks {
			existingDecks[name] = struct{}{}
		}
	} else {
		decks = args
		// Fetch existing deck names to validate
		existingNames, err := client.DeckNames()
		if err != nil {
			return fmt.Errorf("failed to get deck names: %w", err)
		}
		existingDecks = make(map[string]struct{})
		for _, name := range existingNames {
			existingDecks[name] = struct{}{}
		}
	}

	// Partition decks into found and not found
	var found, notFound []string
	for _, deck := range decks {
		if _, ok := existingDecks[deck]; ok {
			found = append(found, deck)
		} else {
			notFound = append(notFound, deck)
		}
	}

	if dryRun {
		fmt.Fprintln(stderr, "Would delete the following decks (and all their cards):")
		for _, deck := range decks {
			fmt.Fprintln(stdout, deck)
		}
		return nil
	}

	if !force {
		fmt.Fprintln(stderr, "The following decks will be deleted (including all cards):")
		for _, deck := range decks {
			fmt.Fprintf(stderr, "  - %s\n", deck)
		}
		fmt.Fprint(stderr, "Continue? [y/N] ")

		reader := bufio.NewReader(stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return ErrCancelled
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return ErrCancelled
		}
	}

	// Delete only the decks that exist
	if len(found) > 0 {
		if err := client.DeleteDecks(found); err != nil {
			return fmt.Errorf("failed to delete decks: %w", err)
		}
		for _, deck := range found {
			fmt.Fprintf(stderr, "Deleted %s\n", deck)
		}
	}

	// Report decks that were not found
	for _, deck := range notFound {
		fmt.Fprintf(stderr, "Could not find %s\n", deck)
	}

	if len(notFound) > 0 {
		return ErrSilent
	}
	return nil
}

func init() {
	deckListCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	deckListCmd.Flags().StringVarP(&fieldsFlag, "fields", "f", "", "Comma-separated list of fields (available: id, name, new, learn, review, total)")
	deckDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	deckDeleteCmd.Flags().Bool("dry-run", false, "Show what would be deleted without executing")
	deckDeleteCmd.Flags().Bool("id", false, "Treat arguments as deck IDs instead of names")
	deckCmd.AddCommand(deckListCmd)
	deckCmd.AddCommand(deckCreateCmd)
	deckCmd.AddCommand(deckDeleteCmd)
	rootCmd.AddCommand(deckCmd)
}
