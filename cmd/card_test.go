package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/atdrendel/ankigo/internal/ankiconnect"
)

func TestCardSearch_PlainText_Default(t *testing.T) {
	mock := &mockClient{
		cardIDs: []int64{1498938915662, 1502098034048},
	}

	var buf bytes.Buffer
	opts := cardSearchOptions{json: false, fields: nil}
	err := runCardSearch(mock, &buf, "deck:Default", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.searchQuery != "deck:Default" {
		t.Errorf("expected searchQuery 'deck:Default', got %q", mock.searchQuery)
	}

	expected := "1498938915662\n1502098034048\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestCardSearch_PlainText_Empty(t *testing.T) {
	mock := &mockClient{
		cardIDs: []int64{},
	}

	var buf bytes.Buffer
	opts := cardSearchOptions{json: false, fields: nil}
	err := runCardSearch(mock, &buf, "deck:NonExistent", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "No cards found\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestCardSearch_Error_FindCards(t *testing.T) {
	mock := &mockClient{
		findCardsErr: errors.New("connection refused"),
	}

	var buf bytes.Buffer
	opts := cardSearchOptions{json: false, fields: nil}
	err := runCardSearch(mock, &buf, "deck:Default", opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "failed to find cards: connection refused" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCardSearch_PlainText_Fields(t *testing.T) {
	mock := &mockClient{
		cardIDs: []int64{1498938915662, 1502098034048},
		cardInfo: []ankiconnect.CardInfo{
			{CardID: 1498938915662, DeckName: "Japanese", Question: "What is hello?"},
			{CardID: 1502098034048, DeckName: "Japanese", Question: "What is goodbye?"},
		},
	}

	var buf bytes.Buffer
	opts := cardSearchOptions{json: false, fields: []string{"id", "deck"}}
	err := runCardSearch(mock, &buf, "deck:Japanese", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "1498938915662\tJapanese\n1502098034048\tJapanese\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestCardSearch_PlainText_AllFields(t *testing.T) {
	mock := &mockClient{
		cardIDs: []int64{1498938915662},
		cardInfo: []ankiconnect.CardInfo{
			{
				CardID:     1498938915662,
				Note:       1502298033753,
				DeckName:   "Japanese",
				ModelName:  "Basic",
				Ord:        1,
				Question:   "What is hello?",
				Answer:     "konnichiwa",
				Fields:     map[string]ankiconnect.CardField{"Front": {Value: "hello", Order: 0}, "Back": {Value: "konnichiwa", Order: 1}},
				Type:       2,
				Queue:      2,
				Due:        100,
				Interval:   30,
				Factor:     2500,
				Reps:       5,
				Lapses:     1,
				Left:       0,
				Mod:        1629454092,
				FieldOrder: 0,
				CSS:        "p {font-family:Arial;}",
			},
		},
	}

	var buf bytes.Buffer
	// Test a subset of fields for readable output
	opts := cardSearchOptions{json: false, fields: []string{"id", "note", "deck", "model", "ord", "type", "queue", "due", "interval", "factor", "reps", "lapses", "left", "mod"}}
	err := runCardSearch(mock, &buf, "deck:Japanese", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "1498938915662\t1502298033753\tJapanese\tBasic\t1\t2\t2\t100\t30\t2500\t5\t1\t0\t1629454092\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestCardSearch_PlainText_FieldsMap(t *testing.T) {
	mock := &mockClient{
		cardIDs: []int64{1498938915662},
		cardInfo: []ankiconnect.CardInfo{
			{
				CardID: 1498938915662,
				Fields: map[string]ankiconnect.CardField{
					"Front": {Value: "hello", Order: 0},
					"Back":  {Value: "konnichiwa", Order: 1},
				},
			},
		},
	}

	var buf bytes.Buffer
	opts := cardSearchOptions{json: false, fields: []string{"id", "fields"}}
	err := runCardSearch(mock, &buf, "deck:Japanese", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Fields should be serialized as JSON
	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("Front")) {
		t.Errorf("expected output to contain 'Front', got %q", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("hello")) {
		t.Errorf("expected output to contain 'hello', got %q", output)
	}
}

func TestCardSearch_InvalidField(t *testing.T) {
	mock := &mockClient{
		cardIDs: []int64{1},
	}

	var buf bytes.Buffer
	opts := cardSearchOptions{json: false, fields: []string{"invalid"}}
	err := runCardSearch(mock, &buf, "deck:Default", opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "unknown field: invalid" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCardSearch_JSON_Default(t *testing.T) {
	mock := &mockClient{
		cardIDs: []int64{1498938915662, 1502098034048},
		cardInfo: []ankiconnect.CardInfo{
			{
				CardID:     1498938915662,
				Note:       1502298033753,
				DeckName:   "Japanese",
				ModelName:  "Basic",
				Ord:        1,
				Question:   "What is hello?",
				Answer:     "konnichiwa",
				Fields:     map[string]ankiconnect.CardField{"Front": {Value: "hello", Order: 0}},
				Type:       2,
				Queue:      2,
				Due:        100,
				Interval:   30,
				Factor:     2500,
				Reps:       5,
				Lapses:     1,
				Left:       0,
				Mod:        1629454092,
				FieldOrder: 0,
				CSS:        "p {font-family:Arial;}",
			},
			{
				CardID:    1502098034048,
				DeckName:  "Japanese",
				Question:  "What is goodbye?",
				Answer:    "sayonara",
				ModelName: "Basic",
				Due:       50,
				Interval:  15,
				Reps:      3,
				Lapses:    0,
			},
		},
	}

	var buf bytes.Buffer
	opts := cardSearchOptions{json: true, fields: nil}
	err := runCardSearch(mock, &buf, "deck:Japanese", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}

	// Check first entry has all fields
	first := result[0]
	if first["id"].(float64) != 1498938915662 {
		t.Errorf("expected id 1498938915662, got %v", first["id"])
	}
	if first["note"].(float64) != 1502298033753 {
		t.Errorf("expected note 1502298033753, got %v", first["note"])
	}
	if first["deck"] != "Japanese" {
		t.Errorf("expected deck 'Japanese', got %v", first["deck"])
	}
	if first["model"] != "Basic" {
		t.Errorf("expected model 'Basic', got %v", first["model"])
	}
	if first["ord"].(float64) != 1 {
		t.Errorf("expected ord 1, got %v", first["ord"])
	}
	if first["question"] != "What is hello?" {
		t.Errorf("expected question 'What is hello?', got %v", first["question"])
	}
	if first["answer"] != "konnichiwa" {
		t.Errorf("expected answer 'konnichiwa', got %v", first["answer"])
	}
	if first["type"].(float64) != 2 {
		t.Errorf("expected type 2, got %v", first["type"])
	}
	if first["queue"].(float64) != 2 {
		t.Errorf("expected queue 2, got %v", first["queue"])
	}
	if first["due"].(float64) != 100 {
		t.Errorf("expected due 100, got %v", first["due"])
	}
	if first["interval"].(float64) != 30 {
		t.Errorf("expected interval 30, got %v", first["interval"])
	}
	if first["factor"].(float64) != 2500 {
		t.Errorf("expected factor 2500, got %v", first["factor"])
	}
	if first["reps"].(float64) != 5 {
		t.Errorf("expected reps 5, got %v", first["reps"])
	}
	if first["lapses"].(float64) != 1 {
		t.Errorf("expected lapses 1, got %v", first["lapses"])
	}
	if first["left"].(float64) != 0 {
		t.Errorf("expected left 0, got %v", first["left"])
	}
	if first["mod"].(float64) != 1629454092 {
		t.Errorf("expected mod 1629454092, got %v", first["mod"])
	}
	if first["fieldOrder"].(float64) != 0 {
		t.Errorf("expected fieldOrder 0, got %v", first["fieldOrder"])
	}
	if first["css"] != "p {font-family:Arial;}" {
		t.Errorf("expected css 'p {font-family:Arial;}', got %v", first["css"])
	}
	// Check fields map is present
	if _, hasFields := first["fields"]; !hasFields {
		t.Error("expected fields to be present")
	}
}

func TestCardSearch_JSON_Empty(t *testing.T) {
	mock := &mockClient{
		cardIDs: []int64{},
	}

	var buf bytes.Buffer
	opts := cardSearchOptions{json: true, fields: nil}
	err := runCardSearch(mock, &buf, "deck:NonExistent", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "[]\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestCardSearch_JSON_Fields(t *testing.T) {
	mock := &mockClient{
		cardIDs: []int64{1498938915662},
		cardInfo: []ankiconnect.CardInfo{
			{
				CardID:    1498938915662,
				DeckName:  "Japanese",
				Question:  "What is hello?",
				Answer:    "konnichiwa",
				ModelName: "Basic",
			},
		},
	}

	var buf bytes.Buffer
	opts := cardSearchOptions{json: true, fields: []string{"id", "deck"}}
	err := runCardSearch(mock, &buf, "deck:Japanese", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}

	// Should only have id and deck fields
	entry := result[0]
	if entry["id"].(float64) != 1498938915662 {
		t.Errorf("expected id 1498938915662, got %v", entry["id"])
	}
	if entry["deck"] != "Japanese" {
		t.Errorf("expected deck 'Japanese', got %v", entry["deck"])
	}
	if _, hasQuestion := entry["question"]; hasQuestion {
		t.Error("expected question field to be absent")
	}
}

func TestCardSearch_CardsInfoMissing(t *testing.T) {
	// Tests when cardsInfo returns empty for some IDs
	mock := &mockClient{
		cardIDs:  []int64{1498938915662, 1502098034048},
		cardInfo: []ankiconnect.CardInfo{}, // No card info returned
	}

	var buf bytes.Buffer
	opts := cardSearchOptions{json: false, fields: []string{"id", "deck"}}
	err := runCardSearch(mock, &buf, "deck:Japanese", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should handle gracefully - output IDs with empty deck
	expected := "1498938915662\t\n1502098034048\t\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestCardSearch_CardsInfoPartialMissing(t *testing.T) {
	// Tests when cardsInfo returns data for only some IDs
	mock := &mockClient{
		cardIDs: []int64{1498938915662, 1502098034048},
		cardInfo: []ankiconnect.CardInfo{
			{CardID: 1498938915662, DeckName: "Japanese"},
			// Missing entry for 1502098034048
		},
	}

	var buf bytes.Buffer
	opts := cardSearchOptions{json: false, fields: []string{"id", "deck"}}
	err := runCardSearch(mock, &buf, "deck:Japanese", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should handle gracefully - output first card with deck, second with empty
	expected := "1498938915662\tJapanese\n1502098034048\t\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestCardSearch_CardsInfoError(t *testing.T) {
	mock := &mockClient{
		cardIDs:      []int64{1498938915662},
		cardsInfoErr: errors.New("collection is not available"),
	}

	var buf bytes.Buffer
	opts := cardSearchOptions{json: false, fields: []string{"id", "deck"}}
	err := runCardSearch(mock, &buf, "deck:Japanese", opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "failed to get card info: collection is not available" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCardSearch_IDOnlyNoCardsInfo(t *testing.T) {
	// When only "id" field is requested, should not call CardsInfo
	mock := &mockClient{
		cardIDs:      []int64{1498938915662, 1502098034048},
		cardsInfoErr: errors.New("should not be called"),
	}

	var buf bytes.Buffer
	opts := cardSearchOptions{json: false, fields: []string{"id"}}
	err := runCardSearch(mock, &buf, "deck:Japanese", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "1498938915662\n1502098034048\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestCardSearch_DefaultFieldsNoCardsInfo(t *testing.T) {
	// When no fields specified (default), should not call CardsInfo (only output IDs)
	mock := &mockClient{
		cardIDs:      []int64{1498938915662},
		cardsInfoErr: errors.New("should not be called"),
	}

	var buf bytes.Buffer
	opts := cardSearchOptions{json: false, fields: nil}
	err := runCardSearch(mock, &buf, "deck:Japanese", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "1498938915662\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestCardSearch_JSON_CardsInfoMissing(t *testing.T) {
	// JSON output with missing card info should use zero values
	mock := &mockClient{
		cardIDs:  []int64{1498938915662},
		cardInfo: []ankiconnect.CardInfo{}, // No card info
	}

	var buf bytes.Buffer
	opts := cardSearchOptions{json: true, fields: []string{"id", "deck", "reps"}}
	err := runCardSearch(mock, &buf, "deck:Japanese", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}

	entry := result[0]
	if entry["id"].(float64) != 1498938915662 {
		t.Errorf("expected id 1498938915662, got %v", entry["id"])
	}
	if entry["deck"] != "" {
		t.Errorf("expected empty deck, got %v", entry["deck"])
	}
	if entry["reps"].(float64) != 0 {
		t.Errorf("expected reps 0, got %v", entry["reps"])
	}
}

// Unused import check - strings is used in other test files via the shared mockClient
var _ = strings.Contains
