package ankiconnect

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_DeckNames_Success(t *testing.T) {
	// Setup: mock server returns valid deck list
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": ["Default", "Japanese::JLPT N3"], "error": null}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	decks, err := client.DeckNames()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decks) != 2 {
		t.Fatalf("expected 2 decks, got %d", len(decks))
	}
	if decks[0] != "Default" {
		t.Errorf("expected first deck to be 'Default', got %q", decks[0])
	}
	if decks[1] != "Japanese::JLPT N3" {
		t.Errorf("expected second deck to be 'Japanese::JLPT N3', got %q", decks[1])
	}
}

func TestClient_DeckNames_APIError(t *testing.T) {
	// Setup: mock server returns API error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": null, "error": "collection is not available"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	decks, err := client.DeckNames()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if decks != nil {
		t.Errorf("expected nil decks, got %v", decks)
	}
	if err.Error() != "collection is not available" {
		t.Errorf("expected error message 'collection is not available', got %q", err.Error())
	}
}

func TestClient_DeckNames_ConnectionError(t *testing.T) {
	// Setup: client with invalid URL (no server running)
	client := NewClient("http://localhost:59999")
	decks, err := client.DeckNames()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if decks != nil {
		t.Errorf("expected nil decks, got %v", decks)
	}
}

func TestClient_DeckNames_InvalidJSON(t *testing.T) {
	// Setup: mock server returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	decks, err := client.DeckNames()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if decks != nil {
		t.Errorf("expected nil decks, got %v", decks)
	}
}

func TestClient_DeckNames_EmptyList(t *testing.T) {
	// Setup: mock server returns empty deck list
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": [], "error": null}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	decks, err := client.DeckNames()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decks) != 0 {
		t.Errorf("expected 0 decks, got %d", len(decks))
	}
}

func TestClient_DeckNamesAndIds_Success(t *testing.T) {
	// Setup: mock server returns valid deck name/id map
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": {"Default": 1, "Japanese::JLPT N3": 1234567890}, "error": null}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	decks, err := client.DeckNamesAndIds()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decks) != 2 {
		t.Fatalf("expected 2 decks, got %d", len(decks))
	}
	if decks["Default"] != 1 {
		t.Errorf("expected Default id to be 1, got %d", decks["Default"])
	}
	if decks["Japanese::JLPT N3"] != 1234567890 {
		t.Errorf("expected Japanese::JLPT N3 id to be 1234567890, got %d", decks["Japanese::JLPT N3"])
	}
}

func TestClient_DeckNamesAndIds_APIError(t *testing.T) {
	// Setup: mock server returns API error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": null, "error": "collection is not available"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	decks, err := client.DeckNamesAndIds()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if decks != nil {
		t.Errorf("expected nil decks, got %v", decks)
	}
	if err.Error() != "collection is not available" {
		t.Errorf("expected error message 'collection is not available', got %q", err.Error())
	}
}

func TestClient_DeckNamesAndIds_Empty(t *testing.T) {
	// Setup: mock server returns empty map
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": {}, "error": null}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	decks, err := client.DeckNamesAndIds()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decks) != 0 {
		t.Errorf("expected 0 decks, got %d", len(decks))
	}
}

func TestClient_GetDeckStats_Success(t *testing.T) {
	// Setup: mock server returns stats for two decks
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"result": {
				"1": {"deck_id": 1, "name": "Default", "new_count": 10, "learn_count": 5, "review_count": 20, "total_in_deck": 150},
				"1234567890": {"deck_id": 1234567890, "name": "Japanese::JLPT N3", "new_count": 25, "learn_count": 3, "review_count": 50, "total_in_deck": 500}
			},
			"error": null
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	stats, err := client.GetDeckStats([]string{"Default", "Japanese::JLPT N3"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 deck stats, got %d", len(stats))
	}

	// Check Default deck stats
	defaultStats, ok := stats[1]
	if !ok {
		t.Fatal("expected stats for deck ID 1")
	}
	if defaultStats.Name != "Default" {
		t.Errorf("expected name 'Default', got %q", defaultStats.Name)
	}
	if defaultStats.NewCount != 10 {
		t.Errorf("expected new_count 10, got %d", defaultStats.NewCount)
	}
	if defaultStats.LearnCount != 5 {
		t.Errorf("expected learn_count 5, got %d", defaultStats.LearnCount)
	}
	if defaultStats.ReviewCount != 20 {
		t.Errorf("expected review_count 20, got %d", defaultStats.ReviewCount)
	}
	if defaultStats.TotalInDeck != 150 {
		t.Errorf("expected total_in_deck 150, got %d", defaultStats.TotalInDeck)
	}

	// Check Japanese deck stats
	japaneseStats, ok := stats[1234567890]
	if !ok {
		t.Fatal("expected stats for deck ID 1234567890")
	}
	if japaneseStats.TotalInDeck != 500 {
		t.Errorf("expected total_in_deck 500, got %d", japaneseStats.TotalInDeck)
	}
}

func TestClient_GetDeckStats_APIError(t *testing.T) {
	// Setup: mock server returns API error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": null, "error": "collection is not available"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	stats, err := client.GetDeckStats([]string{"Default"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if stats != nil {
		t.Errorf("expected nil stats, got %v", stats)
	}
	if err.Error() != "collection is not available" {
		t.Errorf("expected error message 'collection is not available', got %q", err.Error())
	}
}

func TestClient_GetDeckStats_Empty(t *testing.T) {
	// Setup: mock server returns empty result
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": {}, "error": null}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	stats, err := client.GetDeckStats([]string{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected 0 stats, got %d", len(stats))
	}
}

func TestClient_CreateDeck_Success(t *testing.T) {
	// Setup: mock server returns deck ID
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": 1234567890, "error": null}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	deckID, err := client.CreateDeck("Test Deck")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deckID != 1234567890 {
		t.Errorf("expected deck ID 1234567890, got %d", deckID)
	}
}

func TestClient_CreateDeck_APIError(t *testing.T) {
	// Setup: mock server returns API error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": null, "error": "collection is not available"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	deckID, err := client.CreateDeck("Test Deck")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if deckID != 0 {
		t.Errorf("expected deck ID 0, got %d", deckID)
	}
	if err.Error() != "collection is not available" {
		t.Errorf("expected error message 'collection is not available', got %q", err.Error())
	}
}

func TestClient_CreateDeck_ConnectionError(t *testing.T) {
	// Setup: client with invalid URL (no server running)
	client := NewClient("http://localhost:59999")
	deckID, err := client.CreateDeck("Test Deck")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if deckID != 0 {
		t.Errorf("expected deck ID 0, got %d", deckID)
	}
}

func TestClient_CreateDeck_InvalidJSON(t *testing.T) {
	// Setup: mock server returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	deckID, err := client.CreateDeck("Test Deck")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if deckID != 0 {
		t.Errorf("expected deck ID 0, got %d", deckID)
	}
}

func TestClient_CreateDeck_ExistingDeck(t *testing.T) {
	// Setup: mock server returns existing deck's ID (not an error)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// anki-connect returns the existing deck ID when creating an existing deck
		w.Write([]byte(`{"result": 1, "error": null}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	deckID, err := client.CreateDeck("Default")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deckID != 1 {
		t.Errorf("expected deck ID 1, got %d", deckID)
	}
}

func TestClient_DeleteDecks_Success(t *testing.T) {
	// Setup: mock server accepts delete request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": null, "error": null}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.DeleteDecks([]string{"Test Deck", "Japanese::JLPT N3"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DeleteDecks_APIError(t *testing.T) {
	// Setup: mock server returns API error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": null, "error": "collection is not available"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.DeleteDecks([]string{"Test Deck"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "collection is not available" {
		t.Errorf("expected error message 'collection is not available', got %q", err.Error())
	}
}

func TestClient_DeleteDecks_ConnectionError(t *testing.T) {
	// Setup: client with invalid URL (no server running)
	client := NewClient("http://localhost:59999")
	err := client.DeleteDecks([]string{"Test Deck"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_DeleteDecks_EmptyList(t *testing.T) {
	// Empty list is valid (no-op)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": null, "error": null}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.DeleteDecks([]string{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
