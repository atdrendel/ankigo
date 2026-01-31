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

// === Card Create Tests ===

func TestCardCreate_Basic_Success(t *testing.T) {
	mock := &mockClient{
		addNoteID:  1234567890,
		modelNames: []string{"Basic", "Cloze"},
		modelFieldNames: map[string][]string{
			"Basic": {"Front", "Back"},
		},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		back:  "Answer",
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.addedNote == nil {
		t.Fatal("expected AddNote to be called")
	}
	if mock.addedNote.DeckName != "Default" {
		t.Errorf("expected deck 'Default', got %q", mock.addedNote.DeckName)
	}
	if mock.addedNote.ModelName != "Basic" {
		t.Errorf("expected model 'Basic', got %q", mock.addedNote.ModelName)
	}
	if mock.addedNote.Fields["Front"] != "Question?" {
		t.Errorf("expected Front 'Question?', got %q", mock.addedNote.Fields["Front"])
	}
	if mock.addedNote.Fields["Back"] != "Answer" {
		t.Errorf("expected Back 'Answer', got %q", mock.addedNote.Fields["Back"])
	}
	if stdout.String() != "1234567890\n" {
		t.Errorf("expected stdout '1234567890\\n', got %q", stdout.String())
	}
}

func TestCardCreate_Basic_MissingFront(t *testing.T) {
	mock := &mockClient{
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic",
		back:  "Answer",
		// front is missing
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "--front is required for Basic model" {
		t.Errorf("unexpected error: %v", err)
	}
	if mock.addedNote != nil {
		t.Error("expected AddNote NOT to be called")
	}
}

func TestCardCreate_Basic_MissingBack(t *testing.T) {
	mock := &mockClient{
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		// back is missing
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "--back is required for Basic model" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCardCreate_WithTags(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Q",
		back:  "A",
		tags:  []string{"tag1", "tag2", "tag3"},
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.addedNote.Tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(mock.addedNote.Tags))
	}
	if mock.addedNote.Tags[0] != "tag1" {
		t.Errorf("expected first tag 'tag1', got %q", mock.addedNote.Tags[0])
	}
}

func TestCardCreate_CustomDeck(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Japanese::JLPT N3",
		model: "Basic",
		front: "日本",
		back:  "Japan",
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.addedNote.DeckName != "Japanese::JLPT N3" {
		t.Errorf("expected deck 'Japanese::JLPT N3', got %q", mock.addedNote.DeckName)
	}
}

func TestCardCreate_ModelNotFound(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{"Basic", "Cloze"},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "NonExistent",
		front: "Q",
		back:  "A",
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != `model "NonExistent" not found` {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCardCreate_DuplicateError(t *testing.T) {
	mock := &mockClient{
		addNoteErr:      errors.New("cannot create note because it is a duplicate"),
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Q",
		back:  "A",
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "card already exists (use --allow-duplicate to add anyway)" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCardCreate_AllowDuplicate(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:           "Default",
		model:          "Basic",
		front:          "Q",
		back:           "A",
		allowDuplicate: true,
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.addedNote.Options == nil {
		t.Fatal("expected Options to be set")
	}
	if !mock.addedNote.Options.AllowDuplicate {
		t.Error("expected AllowDuplicate to be true")
	}
}

func TestCardCreate_DuplicateScopeDeck(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:           "Default",
		model:          "Basic",
		front:          "Q",
		back:           "A",
		duplicateScope: "deck",
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.addedNote.Options == nil {
		t.Fatal("expected Options to be set")
	}
	if mock.addedNote.Options.DuplicateScope != "deck" {
		t.Errorf("expected DuplicateScope 'deck', got %q", mock.addedNote.Options.DuplicateScope)
	}
}

func TestCardCreate_ClozeModel(t *testing.T) {
	mock := &mockClient{
		addNoteID:  1234567890,
		modelNames: []string{"Basic", "Cloze"},
		modelFieldNames: map[string][]string{
			"Cloze": {"Text", "Extra"},
		},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Cloze",
		fields: map[string]string{
			"Text":  "The capital of {{c1::France}} is {{c2::Paris}}",
			"Extra": "Geography",
		},
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.addedNote.ModelName != "Cloze" {
		t.Errorf("expected model 'Cloze', got %q", mock.addedNote.ModelName)
	}
	if mock.addedNote.Fields["Text"] != "The capital of {{c1::France}} is {{c2::Paris}}" {
		t.Errorf("unexpected Text field: %q", mock.addedNote.Fields["Text"])
	}
	if mock.addedNote.Fields["Extra"] != "Geography" {
		t.Errorf("unexpected Extra field: %q", mock.addedNote.Fields["Extra"])
	}
}

func TestCardCreate_MixedFrontBackAndField(t *testing.T) {
	mock := &mockClient{
		addNoteID:  1234567890,
		modelNames: []string{"Basic (and reversed card)"},
		modelFieldNames: map[string][]string{
			"Basic (and reversed card)": {"Front", "Back"},
		},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic (and reversed card)",
		front: "Q",
		back:  "A",
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.addedNote.Fields["Front"] != "Q" {
		t.Errorf("expected Front 'Q', got %q", mock.addedNote.Fields["Front"])
	}
	if mock.addedNote.Fields["Back"] != "A" {
		t.Errorf("expected Back 'A', got %q", mock.addedNote.Fields["Back"])
	}
}

func TestCardCreate_InvalidFieldWarning(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Q",
		back:  "A",
		fields: map[string]string{
			"InvalidField": "value",
		},
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	// Should succeed but warn
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr.String(), "warning") {
		t.Errorf("expected warning on stderr, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "InvalidField") {
		t.Errorf("expected warning to mention 'InvalidField', got %q", stderr.String())
	}
}

func TestCardCreate_NoFieldsError(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{"Custom"},
		modelFieldNames: map[string][]string{
			"Custom": {"Field1", "Field2"},
		},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Custom",
		// No front, back, or fields
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "at least one field must be provided") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCardCreate_ConnectionError(t *testing.T) {
	mock := &mockClient{
		modelNamesErr: errors.New("connection refused"),
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Q",
		back:  "A",
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get model names") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCardCreate_DeckNotFoundError(t *testing.T) {
	mock := &mockClient{
		addNoteErr:      errors.New("deck was not found: NonExistent"),
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "NonExistent",
		model: "Basic",
		front: "Q",
		back:  "A",
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != `deck "NonExistent" not found` {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCardCreate_EmptyContentError(t *testing.T) {
	mock := &mockClient{
		addNoteErr:      errors.New("cannot create note because it is empty"),
		modelNames:      []string{"Custom"},
		modelFieldNames: map[string][]string{"Custom": {"Field1"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Custom",
		fields: map[string]string{
			"Field1": "",
		},
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "card content cannot be empty" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCardCreate_ModelFieldsError_StillSucceeds(t *testing.T) {
	// If we can't fetch field names for validation, the command should still work
	mock := &mockClient{
		addNoteID:      1234567890,
		modelNames:     []string{"Basic"},
		modelFieldsErr: errors.New("some error"),
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Q",
		back:  "A",
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	// Should succeed - field validation is optional
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout.String() != "1234567890\n" {
		t.Errorf("expected note ID output, got %q", stdout.String())
	}
}

func TestCardCreate_FieldOverridesFrontBack(t *testing.T) {
	// When both --field and --front/--back are provided, --front/--back take precedence
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "from front flag",
		back:  "from back flag",
		fields: map[string]string{
			"Front": "from field flag",
		},
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// --front/--back are applied after --field, so they win
	if mock.addedNote.Fields["Front"] != "from front flag" {
		t.Errorf("expected Front 'from front flag', got %q", mock.addedNote.Fields["Front"])
	}
	if mock.addedNote.Fields["Back"] != "from back flag" {
		t.Errorf("expected Back 'from back flag', got %q", mock.addedNote.Fields["Back"])
	}
}

// === Media Spec Parsing Tests ===

func TestParseMediaSpec_LocalPath(t *testing.T) {
	spec := "filename=audio.mp3,path=/tmp/test.mp3,fields=Back"

	media, err := parseMediaSpec(spec)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if media.Filename != "audio.mp3" {
		t.Errorf("expected filename 'audio.mp3', got %q", media.Filename)
	}
	if media.Path != "/tmp/test.mp3" {
		t.Errorf("expected path '/tmp/test.mp3', got %q", media.Path)
	}
	if len(media.Fields) != 1 || media.Fields[0] != "Back" {
		t.Errorf("expected fields ['Back'], got %v", media.Fields)
	}
	// URL and Data should be empty
	if media.URL != "" {
		t.Errorf("expected empty URL, got %q", media.URL)
	}
	if media.Data != "" {
		t.Errorf("expected empty Data, got %q", media.Data)
	}
}

func TestParseMediaSpec_URL(t *testing.T) {
	spec := "filename=pronunciation.mp3,url=https://example.com/audio.mp3,fields=Front"

	media, err := parseMediaSpec(spec)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if media.Filename != "pronunciation.mp3" {
		t.Errorf("expected filename 'pronunciation.mp3', got %q", media.Filename)
	}
	if media.URL != "https://example.com/audio.mp3" {
		t.Errorf("expected URL 'https://example.com/audio.mp3', got %q", media.URL)
	}
	if len(media.Fields) != 1 || media.Fields[0] != "Front" {
		t.Errorf("expected fields ['Front'], got %v", media.Fields)
	}
	// Path should be empty
	if media.Path != "" {
		t.Errorf("expected empty Path, got %q", media.Path)
	}
}

func TestParseMediaSpec_Base64Data(t *testing.T) {
	spec := "filename=image.png,data=SGVsbG8gV29ybGQ=,fields=Back"

	media, err := parseMediaSpec(spec)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if media.Filename != "image.png" {
		t.Errorf("expected filename 'image.png', got %q", media.Filename)
	}
	if media.Data != "SGVsbG8gV29ybGQ=" {
		t.Errorf("expected Data 'SGVsbG8gV29ybGQ=', got %q", media.Data)
	}
	if len(media.Fields) != 1 || media.Fields[0] != "Back" {
		t.Errorf("expected fields ['Back'], got %v", media.Fields)
	}
}

func TestParseMediaSpec_MultipleFields(t *testing.T) {
	spec := "filename=image.jpg,path=/tmp/img.jpg,fields=Front;Back"

	media, err := parseMediaSpec(spec)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(media.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(media.Fields))
	}
	if media.Fields[0] != "Front" {
		t.Errorf("expected first field 'Front', got %q", media.Fields[0])
	}
	if media.Fields[1] != "Back" {
		t.Errorf("expected second field 'Back', got %q", media.Fields[1])
	}
}

func TestParseMediaSpec_MissingFilename(t *testing.T) {
	spec := "path=/tmp/test.mp3,fields=Back"

	_, err := parseMediaSpec(spec)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "filename") {
		t.Errorf("expected error about missing filename, got: %v", err)
	}
}

func TestParseMediaSpec_MissingSource(t *testing.T) {
	spec := "filename=test.mp3,fields=Back"

	_, err := parseMediaSpec(spec)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "path") || !strings.Contains(err.Error(), "url") || !strings.Contains(err.Error(), "data") {
		t.Errorf("expected error about missing source (path/url/data), got: %v", err)
	}
}

func TestParseMediaSpec_MissingFields(t *testing.T) {
	spec := "filename=test.mp3,path=/tmp/test.mp3"

	_, err := parseMediaSpec(spec)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "fields") {
		t.Errorf("expected error about missing fields, got: %v", err)
	}
}

func TestParseMediaSpec_InvalidFormat(t *testing.T) {
	spec := "invalid"

	_, err := parseMediaSpec(spec)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseMediaSpec_EmptySpec(t *testing.T) {
	spec := ""

	_, err := parseMediaSpec(spec)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseMediaSpec_RelativePathConvertedToAbsolute(t *testing.T) {
	spec := "filename=audio.mp3,path=relative/path/file.mp3,fields=Back"

	media, err := parseMediaSpec(spec)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Path should be converted to absolute
	if !strings.HasPrefix(media.Path, "/") {
		t.Errorf("expected absolute path starting with '/', got %q", media.Path)
	}
	if !strings.HasSuffix(media.Path, "relative/path/file.mp3") {
		t.Errorf("expected path to end with 'relative/path/file.mp3', got %q", media.Path)
	}
}

func TestParseMediaSpec_AbsolutePathUnchanged(t *testing.T) {
	spec := "filename=audio.mp3,path=/absolute/path/file.mp3,fields=Back"

	media, err := parseMediaSpec(spec)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if media.Path != "/absolute/path/file.mp3" {
		t.Errorf("expected path '/absolute/path/file.mp3', got %q", media.Path)
	}
}

// === Card Create with Media Tests ===

func TestCardCreate_WithAudio_NoteHasAudioField(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		back:  "Answer",
		audio: []string{"filename=test.mp3,path=/tmp/test.mp3,fields=Back"},
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.addedNote == nil {
		t.Fatal("expected AddNote to be called")
	}
	if len(mock.addedNote.Audio) != 1 {
		t.Fatalf("expected 1 audio attachment, got %d", len(mock.addedNote.Audio))
	}
	audio := mock.addedNote.Audio[0]
	if audio.Filename != "test.mp3" {
		t.Errorf("expected filename 'test.mp3', got %q", audio.Filename)
	}
	if audio.Path != "/tmp/test.mp3" {
		t.Errorf("expected path '/tmp/test.mp3', got %q", audio.Path)
	}
	if len(audio.Fields) != 1 || audio.Fields[0] != "Back" {
		t.Errorf("expected fields ['Back'], got %v", audio.Fields)
	}
}

func TestCardCreate_WithVideo_NoteHasVideoField(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		back:  "Answer",
		video: []string{"filename=clip.mp4,url=https://example.com/video.mp4,fields=Back"},
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.addedNote == nil {
		t.Fatal("expected AddNote to be called")
	}
	if len(mock.addedNote.Video) != 1 {
		t.Fatalf("expected 1 video attachment, got %d", len(mock.addedNote.Video))
	}
	video := mock.addedNote.Video[0]
	if video.Filename != "clip.mp4" {
		t.Errorf("expected filename 'clip.mp4', got %q", video.Filename)
	}
	if video.URL != "https://example.com/video.mp4" {
		t.Errorf("expected URL 'https://example.com/video.mp4', got %q", video.URL)
	}
}

func TestCardCreate_WithPicture_NoteHasPictureField(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:    "Default",
		model:   "Basic",
		front:   "Question?",
		back:    "Answer",
		picture: []string{"filename=image.jpg,path=/tmp/image.jpg,fields=Front"},
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.addedNote == nil {
		t.Fatal("expected AddNote to be called")
	}
	if len(mock.addedNote.Picture) != 1 {
		t.Fatalf("expected 1 picture attachment, got %d", len(mock.addedNote.Picture))
	}
	picture := mock.addedNote.Picture[0]
	if picture.Filename != "image.jpg" {
		t.Errorf("expected filename 'image.jpg', got %q", picture.Filename)
	}
	if picture.Path != "/tmp/image.jpg" {
		t.Errorf("expected path '/tmp/image.jpg', got %q", picture.Path)
	}
}

func TestCardCreate_MultipleAudio_NoteHasAllAudio(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		back:  "Answer",
		audio: []string{
			"filename=audio1.mp3,path=/tmp/a1.mp3,fields=Front",
			"filename=audio2.mp3,path=/tmp/a2.mp3,fields=Back",
		},
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.addedNote.Audio) != 2 {
		t.Fatalf("expected 2 audio attachments, got %d", len(mock.addedNote.Audio))
	}
	if mock.addedNote.Audio[0].Filename != "audio1.mp3" {
		t.Errorf("expected first audio filename 'audio1.mp3', got %q", mock.addedNote.Audio[0].Filename)
	}
	if mock.addedNote.Audio[1].Filename != "audio2.mp3" {
		t.Errorf("expected second audio filename 'audio2.mp3', got %q", mock.addedNote.Audio[1].Filename)
	}
}

func TestCardCreate_MixedMedia_NoteHasAllMediaTypes(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:    "Default",
		model:   "Basic",
		front:   "Question?",
		back:    "Answer",
		audio:   []string{"filename=audio.mp3,path=/tmp/a.mp3,fields=Back"},
		video:   []string{"filename=video.mp4,url=https://example.com/v.mp4,fields=Back"},
		picture: []string{"filename=image.png,path=/tmp/i.png,fields=Front"},
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.addedNote.Audio) != 1 {
		t.Errorf("expected 1 audio, got %d", len(mock.addedNote.Audio))
	}
	if len(mock.addedNote.Video) != 1 {
		t.Errorf("expected 1 video, got %d", len(mock.addedNote.Video))
	}
	if len(mock.addedNote.Picture) != 1 {
		t.Errorf("expected 1 picture, got %d", len(mock.addedNote.Picture))
	}
}

func TestCardCreate_MediaWithTags(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		back:  "Answer",
		tags:  []string{"tag1", "tag2"},
		audio: []string{"filename=audio.mp3,path=/tmp/a.mp3,fields=Back"},
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.addedNote.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(mock.addedNote.Tags))
	}
	if len(mock.addedNote.Audio) != 1 {
		t.Errorf("expected 1 audio, got %d", len(mock.addedNote.Audio))
	}
}

func TestCardCreate_MediaFieldNotInModel_Warning(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		back:  "Answer",
		audio: []string{"filename=audio.mp3,path=/tmp/a.mp3,fields=NonExistent"},
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	// Should succeed but warn
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr.String(), "warning") {
		t.Errorf("expected warning on stderr, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "NonExistent") {
		t.Errorf("expected warning to mention 'NonExistent', got %q", stderr.String())
	}
}

func TestCardCreate_InvalidMediaSpec(t *testing.T) {
	mock := &mockClient{
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := cardCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		back:  "Answer",
		audio: []string{"invalid-spec"},
	}

	err := runCardCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should not have called AddNote
	if mock.addedNote != nil {
		t.Error("expected AddNote NOT to be called")
	}
}
