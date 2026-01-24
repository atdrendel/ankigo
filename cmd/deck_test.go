package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/atdrendel/ankigo/internal/ankiconnect"
)

// mockClient is a test double for ankiconnect.Client
type mockClient struct {
	decks     []string
	deckIDs   map[string]int64
	deckStats map[int64]ankiconnect.DeckStats
	err       error
	statsErr  error
}

func (m *mockClient) DeckNames() ([]string, error) {
	return m.decks, m.err
}

func (m *mockClient) DeckNamesAndIds() (map[string]int64, error) {
	return m.deckIDs, m.err
}

func (m *mockClient) GetDeckStats(decks []string) (map[int64]ankiconnect.DeckStats, error) {
	return m.deckStats, m.statsErr
}

func TestDeckList_PlainText_Default(t *testing.T) {
	mock := &mockClient{
		decks: []string{"Default", "Japanese::JLPT N3"},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: false, fields: nil}
	err := runDeckList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Default\nJapanese::JLPT N3\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestDeckList_PlainText_Fields(t *testing.T) {
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default":           1,
			"Japanese::JLPT N3": 1234567890,
		},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: false, fields: []string{"id", "name"}}
	err := runDeckList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "1\tDefault\n1234567890\tJapanese::JLPT N3\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestDeckList_PlainText_SingleField_ID(t *testing.T) {
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default":           1,
			"Japanese::JLPT N3": 1234567890,
		},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: false, fields: []string{"id"}}
	err := runDeckList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "1\n1234567890\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestDeckList_PlainText_SingleField_Name(t *testing.T) {
	mock := &mockClient{
		decks: []string{"Default", "Japanese::JLPT N3"},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: false, fields: []string{"name"}}
	err := runDeckList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Default\nJapanese::JLPT N3\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestDeckList_JSON_Default(t *testing.T) {
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default":           1,
			"Japanese::JLPT N3": 1234567890,
		},
		deckStats: map[int64]ankiconnect.DeckStats{
			1:          {DeckID: 1, Name: "Default", NewCount: 10, LearnCount: 5, ReviewCount: 20, TotalInDeck: 150},
			1234567890: {DeckID: 1234567890, Name: "Japanese::JLPT N3", NewCount: 25, LearnCount: 3, ReviewCount: 50, TotalInDeck: 500},
		},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: true, fields: nil}
	err := runDeckList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse the output as JSON to verify structure
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}

	// Results should be sorted by name
	if result[0]["name"] != "Default" {
		t.Errorf("expected first entry name to be 'Default', got %v", result[0]["name"])
	}
	if result[0]["id"].(float64) != 1 {
		t.Errorf("expected first entry id to be 1, got %v", result[0]["id"])
	}
	if result[1]["name"] != "Japanese::JLPT N3" {
		t.Errorf("expected second entry name to be 'Japanese::JLPT N3', got %v", result[1]["name"])
	}
	if result[1]["id"].(float64) != 1234567890 {
		t.Errorf("expected second entry id to be 1234567890, got %v", result[1]["id"])
	}
}

func TestDeckList_JSON_Fields(t *testing.T) {
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default":           1,
			"Japanese::JLPT N3": 1234567890,
		},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: true, fields: []string{"name"}}
	err := runDeckList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse the output as JSON to verify structure
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}

	// Should only have name field
	if result[0]["name"] != "Default" {
		t.Errorf("expected first entry name to be 'Default', got %v", result[0]["name"])
	}
	if _, hasID := result[0]["id"]; hasID {
		t.Errorf("expected id field to be absent, but it was present")
	}
}

func TestDeckList_JSON_Fields_IDOnly(t *testing.T) {
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default":           1,
			"Japanese::JLPT N3": 1234567890,
		},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: true, fields: []string{"id"}}
	err := runDeckList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse the output as JSON to verify structure
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}

	// Should only have id field
	if result[0]["id"].(float64) != 1 {
		t.Errorf("expected first entry id to be 1, got %v", result[0]["id"])
	}
	if _, hasName := result[0]["name"]; hasName {
		t.Errorf("expected name field to be absent, but it was present")
	}
}

func TestDeckList_InvalidField(t *testing.T) {
	mock := &mockClient{
		decks: []string{"Default"},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: false, fields: []string{"invalid"}}
	err := runDeckList(mock, &buf, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "unknown field: invalid" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDeckList_PlainText_Empty(t *testing.T) {
	mock := &mockClient{
		decks: []string{},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: false, fields: nil}
	err := runDeckList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "No decks found\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestDeckList_JSON_Empty(t *testing.T) {
	mock := &mockClient{
		deckIDs:   map[string]int64{},
		deckStats: map[int64]ankiconnect.DeckStats{},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: true, fields: nil}
	err := runDeckList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "[]\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestDeckList_Error(t *testing.T) {
	mock := &mockClient{
		err: errors.New("connection refused"),
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: false, fields: nil}
	err := runDeckList(mock, &buf, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "failed to get deck names: connection refused" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDeckList_JSON_Error(t *testing.T) {
	mock := &mockClient{
		err: errors.New("connection refused"),
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: true, fields: nil}
	err := runDeckList(mock, &buf, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "failed to get decks: connection refused" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDeckList_PlainText_StatsFields(t *testing.T) {
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default":           1,
			"Japanese::JLPT N3": 1234567890,
		},
		deckStats: map[int64]ankiconnect.DeckStats{
			1: {DeckID: 1, Name: "Default", NewCount: 10, LearnCount: 5, ReviewCount: 20, TotalInDeck: 150},
			1234567890: {DeckID: 1234567890, Name: "Japanese::JLPT N3", NewCount: 25, LearnCount: 3, ReviewCount: 50, TotalInDeck: 500},
		},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: false, fields: []string{"name", "total", "new"}}
	err := runDeckList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Default\t150\t10\nJapanese::JLPT N3\t500\t25\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestDeckList_PlainText_AllStatsFields(t *testing.T) {
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default": 1,
		},
		deckStats: map[int64]ankiconnect.DeckStats{
			1: {DeckID: 1, Name: "Default", NewCount: 10, LearnCount: 5, ReviewCount: 20, TotalInDeck: 150},
		},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: false, fields: []string{"id", "name", "new", "learn", "review", "total"}}
	err := runDeckList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "1\tDefault\t10\t5\t20\t150\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestDeckList_JSON_AllFields(t *testing.T) {
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default": 1,
		},
		deckStats: map[int64]ankiconnect.DeckStats{
			1: {DeckID: 1, Name: "Default", NewCount: 10, LearnCount: 5, ReviewCount: 20, TotalInDeck: 150},
		},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: true, fields: nil}
	err := runDeckList(mock, &buf, opts)

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

	// Check all fields are present
	entry := result[0]
	if entry["id"].(float64) != 1 {
		t.Errorf("expected id 1, got %v", entry["id"])
	}
	if entry["name"] != "Default" {
		t.Errorf("expected name 'Default', got %v", entry["name"])
	}
	if entry["new"].(float64) != 10 {
		t.Errorf("expected new 10, got %v", entry["new"])
	}
	if entry["learn"].(float64) != 5 {
		t.Errorf("expected learn 5, got %v", entry["learn"])
	}
	if entry["review"].(float64) != 20 {
		t.Errorf("expected review 20, got %v", entry["review"])
	}
	if entry["total"].(float64) != 150 {
		t.Errorf("expected total 150, got %v", entry["total"])
	}
}

func TestDeckList_JSON_StatsFieldsOnly(t *testing.T) {
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default": 1,
		},
		deckStats: map[int64]ankiconnect.DeckStats{
			1: {DeckID: 1, Name: "Default", NewCount: 10, LearnCount: 5, ReviewCount: 20, TotalInDeck: 150},
		},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: true, fields: []string{"new", "review"}}
	err := runDeckList(mock, &buf, opts)

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
	if entry["new"].(float64) != 10 {
		t.Errorf("expected new 10, got %v", entry["new"])
	}
	if entry["review"].(float64) != 20 {
		t.Errorf("expected review 20, got %v", entry["review"])
	}
	// Should not have other fields
	if _, hasID := entry["id"]; hasID {
		t.Error("expected id field to be absent")
	}
	if _, hasName := entry["name"]; hasName {
		t.Error("expected name field to be absent")
	}
}

func TestDeckList_StatsError(t *testing.T) {
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default": 1,
		},
		statsErr: errors.New("stats unavailable"),
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: false, fields: []string{"total"}}
	err := runDeckList(mock, &buf, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "failed to get deck stats: stats unavailable" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDeckList_PlainText_StatsMissing(t *testing.T) {
	// Reproduces crash: stats fields requested but stats not found for deck
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default": 1,
		},
		deckStats: map[int64]ankiconnect.DeckStats{
			// Stats missing for deck ID 1
		},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: false, fields: []string{"new", "name", "id"}}
	err := runDeckList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should output zeros for missing stats
	expected := "0\tDefault\t1\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestDeckList_JSON_StatsMissing(t *testing.T) {
	// Reproduces crash: stats fields requested but stats not found for deck
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default": 1,
		},
		deckStats: map[int64]ankiconnect.DeckStats{
			// Stats missing for deck ID 1
		},
	}

	var buf bytes.Buffer
	opts := deckListOptions{json: true, fields: []string{"name", "new"}}
	err := runDeckList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}

	// Should have zero for missing stats
	if result[0]["new"].(float64) != 0 {
		t.Errorf("expected new=0, got %v", result[0]["new"])
	}
}
