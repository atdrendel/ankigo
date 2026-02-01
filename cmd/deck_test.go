package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/atdrendel/ankigo/internal/ankiconnect"
)

// mockClient is a test double for ankiconnect.Client
type mockClient struct {
	decks          []string
	deckIDs        map[string]int64
	deckStats      map[int64]ankiconnect.DeckStats
	err            error
	statsErr       error
	createDeckID   int64
	createDeckErr  error
	createdDeck    string // captures name passed to CreateDeck
	deleteDecksErr error
	deletedDecks   []string // captures names passed to DeleteDecks

	// Card search fields
	cardIDs      []int64
	cardInfo     []ankiconnect.CardInfo
	findCardsErr error
	cardsInfoErr error
	searchQuery  string // captures query passed to FindCards

	// Note add fields
	addedNote       *ankiconnect.Note // captures note passed to AddNote
	addNoteID       int64
	addNoteErr      error
	modelNames      []string
	modelNamesErr   error
	modelFieldNames map[string][]string // model name -> field names
	modelFieldsErr  error

	// Note delete fields
	deletedNotes   []int64 // captures IDs passed to DeleteNotes
	deleteNotesErr error

	// Note list fields
	noteIDs       []int64 // return value for FindNotes
	findNotesErr  error
	noteInfos     []ankiconnect.NoteInfo // return value for NotesInfo
	notesInfoErr  error
	noteQuery     string // captures query passed to FindNotes
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

func (m *mockClient) CreateDeck(name string) (int64, error) {
	m.createdDeck = name
	return m.createDeckID, m.createDeckErr
}

func (m *mockClient) DeleteDecks(decks []string) error {
	m.deletedDecks = decks
	return m.deleteDecksErr
}

func (m *mockClient) FindCards(query string) ([]int64, error) {
	m.searchQuery = query
	return m.cardIDs, m.findCardsErr
}

func (m *mockClient) CardsInfo(cardIDs []int64) ([]ankiconnect.CardInfo, error) {
	return m.cardInfo, m.cardsInfoErr
}

func (m *mockClient) AddNote(note ankiconnect.Note) (int64, error) {
	m.addedNote = &note
	return m.addNoteID, m.addNoteErr
}

func (m *mockClient) ModelNames() ([]string, error) {
	return m.modelNames, m.modelNamesErr
}

func (m *mockClient) ModelFieldNames(modelName string) ([]string, error) {
	if m.modelFieldsErr != nil {
		return nil, m.modelFieldsErr
	}
	if m.modelFieldNames == nil {
		return nil, nil
	}
	fields, ok := m.modelFieldNames[modelName]
	if !ok {
		return nil, fmt.Errorf("model was not found: %s", modelName)
	}
	return fields, nil
}

func (m *mockClient) DeleteNotes(notes []int64) error {
	m.deletedNotes = notes
	return m.deleteNotesErr
}

func (m *mockClient) FindNotes(query string) ([]int64, error) {
	m.noteQuery = query
	return m.noteIDs, m.findNotesErr
}

func (m *mockClient) NotesInfo(noteIDs []int64) ([]ankiconnect.NoteInfo, error) {
	return m.noteInfos, m.notesInfoErr
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

func TestDeckCreate_Success(t *testing.T) {
	mock := &mockClient{
		createDeckID: 1234567890,
	}

	var buf bytes.Buffer
	err := runDeckCreate(mock, &buf, "Test Deck")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.createdDeck != "Test Deck" {
		t.Errorf("expected createdDeck 'Test Deck', got %q", mock.createdDeck)
	}
	expected := "1234567890\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestDeckCreate_ConnectionError(t *testing.T) {
	mock := &mockClient{
		createDeckErr: errors.New("connection refused"),
	}

	var buf bytes.Buffer
	err := runDeckCreate(mock, &buf, "Test Deck")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "failed to create deck: connection refused" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDeckCreate_APIError(t *testing.T) {
	mock := &mockClient{
		createDeckErr: errors.New("collection is not available"),
	}

	var buf bytes.Buffer
	err := runDeckCreate(mock, &buf, "Test Deck")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "failed to create deck: collection is not available" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDeckCreate_ExistingDeck(t *testing.T) {
	// Creating an existing deck returns its ID (not an error)
	mock := &mockClient{
		createDeckID: 1,
	}

	var buf bytes.Buffer
	err := runDeckCreate(mock, &buf, "Default")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "1\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestDeckCreate_HierarchicalDeck(t *testing.T) {
	mock := &mockClient{
		createDeckID: 9876543210,
	}

	var buf bytes.Buffer
	err := runDeckCreate(mock, &buf, "Japanese::JLPT N3::Vocabulary")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.createdDeck != "Japanese::JLPT N3::Vocabulary" {
		t.Errorf("expected createdDeck 'Japanese::JLPT N3::Vocabulary', got %q", mock.createdDeck)
	}
	expected := "9876543210\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestDeckCreate_EmptyName(t *testing.T) {
	mock := &mockClient{}

	var buf bytes.Buffer
	err := runDeckCreate(mock, &buf, "")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "deck name cannot be empty" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDeckCreate_WhitespaceOnlyName(t *testing.T) {
	mock := &mockClient{}

	var buf bytes.Buffer
	err := runDeckCreate(mock, &buf, "   ")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "deck name cannot be empty" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDeckDelete_Success_SingleDeck(t *testing.T) {
	mock := &mockClient{
		decks: []string{"Test Deck"},
	}

	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, nil, &stdout, &stderr, []string{"Test Deck"}, true, false, false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.deletedDecks) != 1 || mock.deletedDecks[0] != "Test Deck" {
		t.Errorf("expected deletedDecks ['Test Deck'], got %v", mock.deletedDecks)
	}
	if stdout.String() != "" {
		t.Errorf("expected no stdout, got %q", stdout.String())
	}
	if stderr.String() != "Deleted Test Deck\n" {
		t.Errorf("expected stderr 'Deleted Test Deck\\n', got %q", stderr.String())
	}
}

func TestDeckDelete_Success_MultipleDecks(t *testing.T) {
	mock := &mockClient{
		decks: []string{"Deck1", "Deck2", "Deck3"},
	}

	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, nil, &stdout, &stderr, []string{"Deck1", "Deck2", "Deck3"}, true, false, false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.deletedDecks) != 3 {
		t.Errorf("expected 3 deletedDecks, got %d", len(mock.deletedDecks))
	}
	if mock.deletedDecks[0] != "Deck1" || mock.deletedDecks[1] != "Deck2" || mock.deletedDecks[2] != "Deck3" {
		t.Errorf("expected deletedDecks ['Deck1', 'Deck2', 'Deck3'], got %v", mock.deletedDecks)
	}
	expectedStderr := "Deleted Deck1\nDeleted Deck2\nDeleted Deck3\n"
	if stderr.String() != expectedStderr {
		t.Errorf("expected stderr %q, got %q", expectedStderr, stderr.String())
	}
}

func TestDeckDelete_APIError(t *testing.T) {
	mock := &mockClient{
		decks:          []string{"Test Deck"},
		deleteDecksErr: errors.New("collection is not available"),
	}

	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, nil, &stdout, &stderr, []string{"Test Deck"}, true, false, false)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "failed to delete decks: collection is not available" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDeckDelete_NoDeckNames(t *testing.T) {
	mock := &mockClient{}

	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, nil, &stdout, &stderr, []string{}, true, false, false)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "at least one deck name is required" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDeckDelete_DryRun(t *testing.T) {
	mock := &mockClient{
		decks: []string{"Deck1", "Deck2"},
	}

	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, nil, &stdout, &stderr, []string{"Deck1", "Deck2"}, true, true, false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should NOT call the API
	if mock.deletedDecks != nil {
		t.Errorf("expected no API call, but deletedDecks was set to %v", mock.deletedDecks)
	}
	// Should output deck names to stdout
	if stdout.String() != "Deck1\nDeck2\n" {
		t.Errorf("expected stdout 'Deck1\\nDeck2\\n', got %q", stdout.String())
	}
	// Should show info message on stderr
	if stderr.String() != "Would delete the following decks (and all their cards):\n" {
		t.Errorf("unexpected stderr: %q", stderr.String())
	}
}

func TestDeckDelete_ConfirmationYes(t *testing.T) {
	mock := &mockClient{
		decks: []string{"Test Deck"},
	}

	stdin := bytes.NewBufferString("y\n")
	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, stdin, &stdout, &stderr, []string{"Test Deck"}, false, false, false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should call the API
	if len(mock.deletedDecks) != 1 || mock.deletedDecks[0] != "Test Deck" {
		t.Errorf("expected deletedDecks ['Test Deck'], got %v", mock.deletedDecks)
	}
	// Should show prompt on stderr
	if !bytes.Contains(stderr.Bytes(), []byte("will be deleted")) {
		t.Errorf("expected warning on stderr, got %q", stderr.String())
	}
	if !bytes.Contains(stderr.Bytes(), []byte("Continue?")) {
		t.Errorf("expected confirmation prompt on stderr, got %q", stderr.String())
	}
	// Should show confirmation of deletion
	if !bytes.Contains(stderr.Bytes(), []byte("Deleted Test Deck")) {
		t.Errorf("expected deletion confirmation on stderr, got %q", stderr.String())
	}
}

func TestDeckDelete_ConfirmationYesFull(t *testing.T) {
	mock := &mockClient{
		decks: []string{"Test Deck"},
	}

	stdin := bytes.NewBufferString("yes\n")
	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, stdin, &stdout, &stderr, []string{"Test Deck"}, false, false, false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should call the API
	if len(mock.deletedDecks) != 1 || mock.deletedDecks[0] != "Test Deck" {
		t.Errorf("expected deletedDecks ['Test Deck'], got %v", mock.deletedDecks)
	}
	// Should show confirmation of deletion
	if !bytes.Contains(stderr.Bytes(), []byte("Deleted Test Deck")) {
		t.Errorf("expected deletion confirmation on stderr, got %q", stderr.String())
	}
}

func TestDeckDelete_ConfirmationNo(t *testing.T) {
	mock := &mockClient{
		decks: []string{"Test Deck"},
	}

	stdin := bytes.NewBufferString("n\n")
	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, stdin, &stdout, &stderr, []string{"Test Deck"}, false, false, false)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrCancelled) {
		t.Errorf("expected ErrCancelled, got: %v", err)
	}
	// Should NOT call the API
	if mock.deletedDecks != nil {
		t.Errorf("expected no API call, but deletedDecks was set to %v", mock.deletedDecks)
	}
}

func TestDeckDelete_ConfirmationOther(t *testing.T) {
	mock := &mockClient{
		decks: []string{"Test Deck"},
	}

	stdin := bytes.NewBufferString("maybe\n")
	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, stdin, &stdout, &stderr, []string{"Test Deck"}, false, false, false)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrCancelled) {
		t.Errorf("expected ErrCancelled, got: %v", err)
	}
	// Should NOT call the API
	if mock.deletedDecks != nil {
		t.Errorf("expected no API call, but deletedDecks was set to %v", mock.deletedDecks)
	}
}

func TestDeckDelete_HierarchicalDeck(t *testing.T) {
	mock := &mockClient{
		decks: []string{"Japanese::JLPT N3::Vocabulary"},
	}

	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, nil, &stdout, &stderr, []string{"Japanese::JLPT N3::Vocabulary"}, true, false, false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.deletedDecks) != 1 || mock.deletedDecks[0] != "Japanese::JLPT N3::Vocabulary" {
		t.Errorf("expected deletedDecks ['Japanese::JLPT N3::Vocabulary'], got %v", mock.deletedDecks)
	}
	if stderr.String() != "Deleted Japanese::JLPT N3::Vocabulary\n" {
		t.Errorf("expected stderr 'Deleted Japanese::JLPT N3::Vocabulary\\n', got %q", stderr.String())
	}
}

func TestDeckDelete_ByID_Success(t *testing.T) {
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default":   1,
			"Test Deck": 1234567890,
		},
	}

	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, nil, &stdout, &stderr, []string{"1234567890"}, true, false, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.deletedDecks) != 1 || mock.deletedDecks[0] != "Test Deck" {
		t.Errorf("expected deletedDecks ['Test Deck'], got %v", mock.deletedDecks)
	}
	if stdout.String() != "" {
		t.Errorf("expected no stdout, got %q", stdout.String())
	}
	if stderr.String() != "Deleted Test Deck\n" {
		t.Errorf("expected stderr 'Deleted Test Deck\\n', got %q", stderr.String())
	}
}

func TestDeckDelete_ByID_MultipleIDs(t *testing.T) {
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Deck A": 111,
			"Deck B": 222,
			"Deck C": 333,
		},
	}

	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, nil, &stdout, &stderr, []string{"111", "333"}, true, false, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.deletedDecks) != 2 {
		t.Errorf("expected 2 deletedDecks, got %d", len(mock.deletedDecks))
	}
	// Order should match input order
	if mock.deletedDecks[0] != "Deck A" || mock.deletedDecks[1] != "Deck C" {
		t.Errorf("expected deletedDecks ['Deck A', 'Deck C'], got %v", mock.deletedDecks)
	}
	expectedStderr := "Deleted Deck A\nDeleted Deck C\n"
	if stderr.String() != expectedStderr {
		t.Errorf("expected stderr %q, got %q", expectedStderr, stderr.String())
	}
}

func TestDeckDelete_ByID_NotFound(t *testing.T) {
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default": 1,
		},
	}

	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, nil, &stdout, &stderr, []string{"999999999"}, true, false, true)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "deck with ID 999999999 not found" {
		t.Errorf("unexpected error message: %v", err)
	}
	// Should NOT call the API
	if mock.deletedDecks != nil {
		t.Errorf("expected no API call, but deletedDecks was set to %v", mock.deletedDecks)
	}
}

func TestDeckDelete_ByID_InvalidID(t *testing.T) {
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default": 1,
		},
	}

	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, nil, &stdout, &stderr, []string{"not-a-number"}, true, false, true)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != `invalid deck ID "not-a-number": must be a number` {
		t.Errorf("unexpected error message: %v", err)
	}
	// Should NOT call the API
	if mock.deletedDecks != nil {
		t.Errorf("expected no API call, but deletedDecks was set to %v", mock.deletedDecks)
	}
}

func TestDeckDelete_ByID_DryRun(t *testing.T) {
	mock := &mockClient{
		deckIDs: map[string]int64{
			"Default":   1,
			"Test Deck": 1234567890,
		},
	}

	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, nil, &stdout, &stderr, []string{"1234567890", "1"}, true, true, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should NOT call the API
	if mock.deletedDecks != nil {
		t.Errorf("expected no API call, but deletedDecks was set to %v", mock.deletedDecks)
	}
	// Should output resolved deck names to stdout (in input order)
	if stdout.String() != "Test Deck\nDefault\n" {
		t.Errorf("expected stdout 'Test Deck\\nDefault\\n', got %q", stdout.String())
	}
	// Should show info message on stderr
	if stderr.String() != "Would delete the following decks (and all their cards):\n" {
		t.Errorf("unexpected stderr: %q", stderr.String())
	}
}

func TestDeckDelete_NonExistentDeck(t *testing.T) {
	mock := &mockClient{
		decks: []string{"Default", "Existing"},
	}

	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, nil, &stdout, &stderr, []string{"Missing"}, true, false, false)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should return ErrSilent (error already reported via stderr)
	if !errors.Is(err, ErrSilent) {
		t.Errorf("expected ErrSilent, got: %v", err)
	}
	// Should print "Could not find" message to stderr
	if !strings.Contains(stderr.String(), "Could not find Missing") {
		t.Errorf("expected 'Could not find Missing' on stderr, got: %q", stderr.String())
	}
	// Should NOT call DeleteDecks
	if mock.deletedDecks != nil {
		t.Errorf("expected no DeleteDecks call, but got: %v", mock.deletedDecks)
	}
}

func TestDeckDelete_MixedExistentAndNonExistent(t *testing.T) {
	mock := &mockClient{
		decks: []string{"Default", "Existing"},
	}

	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, nil, &stdout, &stderr, []string{"Existing", "Missing"}, true, false, false)

	if err == nil {
		t.Fatal("expected error for partial failure, got nil")
	}
	// Should return ErrSilent (error already reported via stderr)
	if !errors.Is(err, ErrSilent) {
		t.Errorf("expected ErrSilent, got: %v", err)
	}
	// Should delete existing deck
	if len(mock.deletedDecks) != 1 || mock.deletedDecks[0] != "Existing" {
		t.Errorf("expected deletedDecks ['Existing'], got %v", mock.deletedDecks)
	}
	// Should print "Deleted Existing" to stderr
	if !strings.Contains(stderr.String(), "Deleted Existing") {
		t.Errorf("expected 'Deleted Existing' on stderr, got: %q", stderr.String())
	}
	// Should print "Could not find Missing" to stderr
	if !strings.Contains(stderr.String(), "Could not find Missing") {
		t.Errorf("expected 'Could not find Missing' on stderr, got: %q", stderr.String())
	}
}

func TestDeckDelete_AllNonExistent(t *testing.T) {
	mock := &mockClient{
		decks: []string{"Default"},
	}

	var stdout, stderr bytes.Buffer
	err := runDeckDelete(mock, nil, &stdout, &stderr, []string{"A", "B"}, true, false, false)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should return ErrSilent (error already reported via stderr)
	if !errors.Is(err, ErrSilent) {
		t.Errorf("expected ErrSilent, got: %v", err)
	}
	// Should print "Could not find" for both decks
	if !strings.Contains(stderr.String(), "Could not find A") {
		t.Errorf("expected 'Could not find A' on stderr, got: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "Could not find B") {
		t.Errorf("expected 'Could not find B' on stderr, got: %q", stderr.String())
	}
	// Should NOT call DeleteDecks
	if mock.deletedDecks != nil {
		t.Errorf("expected no DeleteDecks call, but got: %v", mock.deletedDecks)
	}
}
