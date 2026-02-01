package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/atdrendel/ankigo/internal/ankiconnect"
	"github.com/spf13/cobra"
)

// cardSearchFields are the available fields for card search output.
var cardSearchFields = []string{
	"id", "note", "deck", "model", "ord",
	"question", "answer", "fields",
	"type", "queue", "due", "interval", "factor",
	"reps", "lapses", "left", "mod",
	"fieldOrder", "css",
}

// cardInfoFields are fields that require fetching card info.
var cardInfoFields = []string{
	"note", "deck", "model", "ord",
	"question", "answer", "fields",
	"type", "queue", "due", "interval", "factor",
	"reps", "lapses", "left", "mod",
	"fieldOrder", "css",
}

// cardSearchOptions holds the options for the card search command.
type cardSearchOptions struct {
	json   bool
	fields []string
}

var cardCmd = &cobra.Command{
	Use:   "card",
	Short: "Manage Anki cards",
	Long:  `Commands for searching Anki cards.`,
}

var cardSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for cards",
	Long:  `Search for cards in your Anki collection using a query string.`,
	Example: `  ankigo card search "deck:Default"
  ankigo card search "tag:japanese"
  ankigo card search "is:new"
  ankigo card search "is:due"
  ankigo card search "front:hello"
  ankigo card search "deck:\"My Spanish Deck\""`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := ankiconnect.DefaultClient()
		query := args[0]
		jsonFlag, _ := cmd.Flags().GetBool("json")
		fieldsStr, _ := cmd.Flags().GetString("fields")

		fields, err := parseFields(fieldsStr, cardSearchFields)
		if err != nil {
			return err
		}

		opts := cardSearchOptions{json: jsonFlag, fields: fields}
		return runCardSearch(client, cmd.OutOrStdout(), query, opts)
	},
}

// needsCardInfo returns true if any of the fields require fetching card info.
func needsCardInfo(fields []string) bool {
	for _, f := range fields {
		if contains(cardInfoFields, f) {
			return true
		}
	}
	return false
}

// cardEntry holds card data for output.
type cardEntry struct {
	id   int64
	info *ankiconnect.CardInfo
}

// runCardSearch is the testable implementation of card search.
func runCardSearch(client Client, out io.Writer, query string, opts cardSearchOptions) error {
	// Validate fields first
	if opts.fields != nil {
		for _, f := range opts.fields {
			if !contains(cardSearchFields, f) {
				return fmt.Errorf("unknown field: %s", f)
			}
		}
	}

	if opts.json {
		return runCardSearchJSON(client, out, query, opts.fields)
	}
	return runCardSearchText(client, out, query, opts.fields)
}

func runCardSearchText(client Client, out io.Writer, query string, fields []string) error {
	// If no fields specified, default to ["id"]
	if fields == nil {
		fields = []string{"id"}
	}

	// Find cards
	cardIDs, err := client.FindCards(query)
	if err != nil {
		return fmt.Errorf("failed to find cards: %w", err)
	}

	if len(cardIDs) == 0 {
		fmt.Fprintln(out, "No cards found")
		return nil
	}

	// Build entries
	entries := make([]cardEntry, len(cardIDs))
	for i, id := range cardIDs {
		entries[i] = cardEntry{id: id}
	}

	// Fetch card info if needed
	if needsCardInfo(fields) {
		cardInfos, err := client.CardsInfo(cardIDs)
		if err != nil {
			return fmt.Errorf("failed to get card info: %w", err)
		}

		// Build a map for lookup
		infoMap := make(map[int64]*ankiconnect.CardInfo)
		for i := range cardInfos {
			infoMap[cardInfos[i].CardID] = &cardInfos[i]
		}

		// Merge info into entries
		for i := range entries {
			if info, ok := infoMap[entries[i].id]; ok {
				entries[i].info = info
			}
		}
	}

	// Output
	for _, e := range entries {
		var vals []string
		for _, f := range fields {
			vals = append(vals, getCardFieldText(e, f))
		}
		fmt.Fprintln(out, strings.Join(vals, "\t"))
	}
	return nil
}

// getCardFieldText returns the text value for a card field.
func getCardFieldText(e cardEntry, field string) string {
	switch field {
	case "id":
		return fmt.Sprintf("%d", e.id)
	case "note":
		if e.info != nil {
			return fmt.Sprintf("%d", e.info.Note)
		}
		return "0"
	case "deck":
		if e.info != nil {
			return e.info.DeckName
		}
		return ""
	case "model":
		if e.info != nil {
			return e.info.ModelName
		}
		return ""
	case "ord":
		if e.info != nil {
			return fmt.Sprintf("%d", e.info.Ord)
		}
		return "0"
	case "question":
		if e.info != nil {
			return e.info.Question
		}
		return ""
	case "answer":
		if e.info != nil {
			return e.info.Answer
		}
		return ""
	case "fields":
		if e.info != nil && e.info.Fields != nil {
			// Serialize fields as JSON for text output
			b, _ := json.Marshal(e.info.Fields)
			return string(b)
		}
		return "{}"
	case "type":
		if e.info != nil {
			return fmt.Sprintf("%d", e.info.Type)
		}
		return "0"
	case "queue":
		if e.info != nil {
			return fmt.Sprintf("%d", e.info.Queue)
		}
		return "0"
	case "due":
		if e.info != nil {
			return fmt.Sprintf("%d", e.info.Due)
		}
		return "0"
	case "interval":
		if e.info != nil {
			return fmt.Sprintf("%d", e.info.Interval)
		}
		return "0"
	case "factor":
		if e.info != nil {
			return fmt.Sprintf("%d", e.info.Factor)
		}
		return "0"
	case "reps":
		if e.info != nil {
			return fmt.Sprintf("%d", e.info.Reps)
		}
		return "0"
	case "lapses":
		if e.info != nil {
			return fmt.Sprintf("%d", e.info.Lapses)
		}
		return "0"
	case "left":
		if e.info != nil {
			return fmt.Sprintf("%d", e.info.Left)
		}
		return "0"
	case "mod":
		if e.info != nil {
			return fmt.Sprintf("%d", e.info.Mod)
		}
		return "0"
	case "fieldOrder":
		if e.info != nil {
			return fmt.Sprintf("%d", e.info.FieldOrder)
		}
		return "0"
	case "css":
		if e.info != nil {
			return e.info.CSS
		}
		return ""
	}
	return ""
}

func runCardSearchJSON(client Client, out io.Writer, query string, fields []string) error {
	// If no fields specified, default to all fields
	if fields == nil {
		fields = cardSearchFields
	}

	// Find cards
	cardIDs, err := client.FindCards(query)
	if err != nil {
		return fmt.Errorf("failed to find cards: %w", err)
	}

	// Handle empty results
	if len(cardIDs) == 0 {
		fmt.Fprintln(out, "[]")
		return nil
	}

	// Build entries
	entries := make([]cardEntry, len(cardIDs))
	for i, id := range cardIDs {
		entries[i] = cardEntry{id: id}
	}

	// Fetch card info if needed
	if needsCardInfo(fields) {
		cardInfos, err := client.CardsInfo(cardIDs)
		if err != nil {
			return fmt.Errorf("failed to get card info: %w", err)
		}

		// Build a map for lookup
		infoMap := make(map[int64]*ankiconnect.CardInfo)
		for i := range cardInfos {
			infoMap[cardInfos[i].CardID] = &cardInfos[i]
		}

		// Merge info into entries
		for i := range entries {
			if info, ok := infoMap[entries[i].id]; ok {
				entries[i].info = info
			}
		}
	}

	// Build output based on requested fields
	var result []map[string]interface{}
	for _, e := range entries {
		obj := make(map[string]interface{})
		for _, f := range fields {
			obj[f] = getCardFieldJSON(e, f)
		}
		result = append(result, obj)
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// getCardFieldJSON returns the JSON value for a card field.
func getCardFieldJSON(e cardEntry, field string) interface{} {
	switch field {
	case "id":
		return e.id
	case "note":
		if e.info != nil {
			return e.info.Note
		}
		return int64(0)
	case "deck":
		if e.info != nil {
			return e.info.DeckName
		}
		return ""
	case "model":
		if e.info != nil {
			return e.info.ModelName
		}
		return ""
	case "ord":
		if e.info != nil {
			return e.info.Ord
		}
		return 0
	case "question":
		if e.info != nil {
			return e.info.Question
		}
		return ""
	case "answer":
		if e.info != nil {
			return e.info.Answer
		}
		return ""
	case "fields":
		if e.info != nil && e.info.Fields != nil {
			return e.info.Fields
		}
		return map[string]ankiconnect.CardField{}
	case "type":
		if e.info != nil {
			return e.info.Type
		}
		return 0
	case "queue":
		if e.info != nil {
			return e.info.Queue
		}
		return 0
	case "due":
		if e.info != nil {
			return e.info.Due
		}
		return 0
	case "interval":
		if e.info != nil {
			return e.info.Interval
		}
		return 0
	case "factor":
		if e.info != nil {
			return e.info.Factor
		}
		return 0
	case "reps":
		if e.info != nil {
			return e.info.Reps
		}
		return 0
	case "lapses":
		if e.info != nil {
			return e.info.Lapses
		}
		return 0
	case "left":
		if e.info != nil {
			return e.info.Left
		}
		return 0
	case "mod":
		if e.info != nil {
			return e.info.Mod
		}
		return int64(0)
	case "fieldOrder":
		if e.info != nil {
			return e.info.FieldOrder
		}
		return 0
	case "css":
		if e.info != nil {
			return e.info.CSS
		}
		return ""
	}
	return nil
}

func init() {
	cardSearchCmd.Flags().Bool("json", false, "Output in JSON format")
	cardSearchCmd.Flags().StringP("fields", "f", "", "Comma-separated list of fields (available: id, note, deck, model, ord, question, answer, fields, type, queue, due, interval, factor, reps, lapses, left, mod, fieldOrder, css)")

	cardCmd.AddCommand(cardSearchCmd)
	rootCmd.AddCommand(cardCmd)
}
