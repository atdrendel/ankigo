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

// CardField represents a single field in a card.
type CardField struct {
	Value string `json:"value"`
	Order int    `json:"order"`
}

// Note represents a note to be added to Anki.
type Note struct {
	DeckName  string            `json:"deckName"`
	ModelName string            `json:"modelName"`
	Fields    map[string]string `json:"fields"`
	Tags      []string          `json:"tags,omitempty"`
	Options   *NoteOptions      `json:"options,omitempty"`
	Audio     []MediaAttachment `json:"audio,omitempty"`
	Video     []MediaAttachment `json:"video,omitempty"`
	Picture   []MediaAttachment `json:"picture,omitempty"`
}

// NoteOptions contains options for duplicate handling.
type NoteOptions struct {
	AllowDuplicate        bool                   `json:"allowDuplicate,omitempty"`
	DuplicateScope        string                 `json:"duplicateScope,omitempty"`
	DuplicateScopeOptions *DuplicateScopeOptions `json:"duplicateScopeOptions,omitempty"`
}

// DuplicateScopeOptions contains advanced duplicate checking options.
type DuplicateScopeOptions struct {
	DeckName       string `json:"deckName,omitempty"`
	CheckChildren  bool   `json:"checkChildren,omitempty"`
	CheckAllModels bool   `json:"checkAllModels,omitempty"`
}

// MediaAttachment represents a media file to attach to a note.
type MediaAttachment struct {
	Filename       string   `json:"filename"`
	URL            string   `json:"url,omitempty"`
	Path           string   `json:"path,omitempty"`
	Data           string   `json:"data,omitempty"`
	SkipHash       string   `json:"skipHash,omitempty"`
	DeleteExisting *bool    `json:"deleteExisting,omitempty"`
	Fields         []string `json:"fields,omitempty"`
}

// CardInfo contains detailed information about a card.
type CardInfo struct {
	CardID     int64                `json:"cardId"`
	Fields     map[string]CardField `json:"fields"`
	FieldOrder int                  `json:"fieldOrder"`
	Question   string               `json:"question"`
	Answer     string               `json:"answer"`
	ModelName  string               `json:"modelName"`
	Ord        int                  `json:"ord"`
	DeckName   string               `json:"deckName"`
	CSS        string               `json:"css"`
	Factor     int                  `json:"factor"`
	Interval   int                  `json:"interval"`
	Note       int64                `json:"note"`
	Type       int                  `json:"type"`
	Queue      int                  `json:"queue"`
	Due        int                  `json:"due"`
	Reps       int                  `json:"reps"`
	Lapses     int                  `json:"lapses"`
	Left       int                  `json:"left"`
	Mod        int64                `json:"mod"`
}

// Client defines the interface for interacting with anki-connect.
type Client interface {
	DeckNames() ([]string, error)
	DeckNamesAndIds() (map[string]int64, error)
	GetDeckStats(decks []string) (map[int64]DeckStats, error)
	CreateDeck(name string) (int64, error)
	DeleteDecks(decks []string) error
	FindCards(query string) ([]int64, error)
	CardsInfo(cardIDs []int64) ([]CardInfo, error)
	AddNote(note Note) (int64, error)
	ModelNames() ([]string, error)
	ModelFieldNames(modelName string) ([]string, error)
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

// FindCards searches for cards matching the query and returns card IDs.
func (c *HTTPClient) FindCards(query string) ([]int64, error) {
	req := request{
		Action:  "findCards",
		Version: 6,
		Params:  map[string]interface{}{"query": query},
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

	var cardIDs []int64
	if err := json.Unmarshal(apiResp.Result, &cardIDs); err != nil {
		return nil, err
	}

	return cardIDs, nil
}

// CardsInfo returns detailed information for the given card IDs.
func (c *HTTPClient) CardsInfo(cardIDs []int64) ([]CardInfo, error) {
	req := request{
		Action:  "cardsInfo",
		Version: 6,
		Params:  map[string]interface{}{"cards": cardIDs},
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

	var cards []CardInfo
	if err := json.Unmarshal(apiResp.Result, &cards); err != nil {
		return nil, err
	}

	return cards, nil
}

// AddNote creates a note and returns the note ID.
func (c *HTTPClient) AddNote(note Note) (int64, error) {
	req := request{
		Action:  "addNote",
		Version: 6,
		Params:  map[string]interface{}{"note": note},
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

	// Result can be null on failure
	if string(apiResp.Result) == "null" {
		return 0, errors.New("failed to create note")
	}

	var noteID int64
	if err := json.Unmarshal(apiResp.Result, &noteID); err != nil {
		return 0, err
	}

	return noteID, nil
}

// ModelNames returns all model (note type) names.
func (c *HTTPClient) ModelNames() ([]string, error) {
	req := request{
		Action:  "modelNames",
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

	var models []string
	if err := json.Unmarshal(apiResp.Result, &models); err != nil {
		return nil, err
	}

	return models, nil
}

// ModelFieldNames returns the field names for a given model.
func (c *HTTPClient) ModelFieldNames(modelName string) ([]string, error) {
	req := request{
		Action:  "modelFieldNames",
		Version: 6,
		Params:  map[string]interface{}{"modelName": modelName},
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

	var fields []string
	if err := json.Unmarshal(apiResp.Result, &fields); err != nil {
		return nil, err
	}

	return fields, nil
}

// DeleteNotes deletes notes by their IDs.
func (c *HTTPClient) DeleteNotes(notes []int64) error {
	req := request{
		Action:  "deleteNotes",
		Version: 6,
		Params:  map[string]interface{}{"notes": notes},
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
