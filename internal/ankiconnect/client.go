package ankiconnect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// DeckStats contains statistics for a deck from the getDeckStats API.
type DeckStats struct {
	DeckID      int64  `json:"deck_id"`
	Name        string `json:"name"`
	NewCount    int    `json:"new_count"`
	LearnCount  int    `json:"learn_count"`
	ReviewCount int    `json:"review_count"`
	TotalInDeck int    `json:"total_in_deck"`
}

// Client defines the interface for interacting with anki-connect.
type Client interface {
	DeckNames() ([]string, error)
	DeckNamesAndIds() (map[string]int64, error)
	GetDeckStats(decks []string) (map[int64]DeckStats, error)
	CreateDeck(name string) (int64, error)
	DeleteDecks(decks []string) error
}

// HTTPClient is the real implementation that communicates with anki-connect.
type HTTPClient struct {
	BaseURL    string
	httpClient *http.Client
}

// NewClient creates a new anki-connect client with the given base URL.
func NewClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// DefaultClient creates a new client with the default anki-connect URL.
func DefaultClient() *HTTPClient {
	return NewClient("http://localhost:8765")
}

// request represents an anki-connect API request.
type request struct {
	Action  string      `json:"action"`
	Version int         `json:"version"`
	Params  interface{} `json:"params,omitempty"`
}

// response represents an anki-connect API response.
type response struct {
	Result json.RawMessage `json:"result"`
	Error  *string         `json:"error"`
}

// DeckNames returns the names of all decks in the collection.
func (c *HTTPClient) DeckNames() ([]string, error) {
	req := request{
		Action:  "deckNames",
		Version: 6,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.BaseURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if apiResp.Error != nil {
		return nil, errors.New(*apiResp.Error)
	}

	var decks []string
	if err := json.Unmarshal(apiResp.Result, &decks); err != nil {
		return nil, err
	}

	return decks, nil
}

// DeckNamesAndIds returns a map of deck names to their IDs.
func (c *HTTPClient) DeckNamesAndIds() (map[string]int64, error) {
	req := request{
		Action:  "deckNamesAndIds",
		Version: 6,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.BaseURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if apiResp.Error != nil {
		return nil, errors.New(*apiResp.Error)
	}

	var decks map[string]int64
	if err := json.Unmarshal(apiResp.Result, &decks); err != nil {
		return nil, err
	}

	return decks, nil
}

// GetDeckStats returns statistics for the specified decks.
func (c *HTTPClient) GetDeckStats(decks []string) (map[int64]DeckStats, error) {
	req := request{
		Action:  "getDeckStats",
		Version: 6,
		Params:  map[string]interface{}{"decks": decks},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.BaseURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if apiResp.Error != nil {
		return nil, errors.New(*apiResp.Error)
	}

	// API returns map with string keys (deck IDs as strings)
	var rawStats map[string]DeckStats
	if err := json.Unmarshal(apiResp.Result, &rawStats); err != nil {
		return nil, err
	}

	// Convert string keys to int64
	result := make(map[int64]DeckStats)
	for idStr, stats := range rawStats {
		var id int64
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
			return nil, fmt.Errorf("invalid deck ID %q: %w", idStr, err)
		}
		result[id] = stats
	}

	return result, nil
}

// CreateDeck creates a new deck with the given name and returns its ID.
// If the deck already exists, it returns the existing deck's ID.
func (c *HTTPClient) CreateDeck(name string) (int64, error) {
	req := request{
		Action:  "createDeck",
		Version: 6,
		Params:  map[string]interface{}{"deck": name},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}

	resp, err := c.httpClient.Post(c.BaseURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var apiResp response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return 0, err
	}

	if apiResp.Error != nil {
		return 0, errors.New(*apiResp.Error)
	}

	var deckID int64
	if err := json.Unmarshal(apiResp.Result, &deckID); err != nil {
		return 0, err
	}

	return deckID, nil
}

// DeleteDecks deletes the specified decks and all their cards.
func (c *HTTPClient) DeleteDecks(decks []string) error {
	req := request{
		Action:  "deleteDecks",
		Version: 6,
		Params:  map[string]interface{}{"decks": decks, "cardsToo": true},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(c.BaseURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var apiResp response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return err
	}

	if apiResp.Error != nil {
		return errors.New(*apiResp.Error)
	}

	return nil
}
