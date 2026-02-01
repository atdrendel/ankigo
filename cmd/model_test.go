package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestModelList_Default(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{"Basic", "Cloze", "Basic (and reversed card)"},
	}

	var buf bytes.Buffer
	opts := modelListOptions{json: false, fields: nil}
	err := runModelList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should output names only, sorted alphabetically
	expected := "Basic\nBasic (and reversed card)\nCloze\n"
	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

func TestModelList_WithFields_ID(t *testing.T) {
	mock := &mockClient{
		modelNamesAndIDs: map[string]int64{
			"Basic": 1234567890,
			"Cloze": 9876543210,
		},
	}

	var buf bytes.Buffer
	opts := modelListOptions{json: false, fields: []string{"name", "id"}}
	err := runModelList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should output name and id, tab-separated, sorted by name
	if !strings.Contains(output, "Basic\t1234567890") {
		t.Errorf("expected Basic with ID, got %q", output)
	}
	if !strings.Contains(output, "Cloze\t9876543210") {
		t.Errorf("expected Cloze with ID, got %q", output)
	}
}

func TestModelList_WithFields_Fields(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{"Basic", "Cloze"},
		modelFieldNames: map[string][]string{
			"Basic": {"Front", "Back"},
			"Cloze": {"Text", "Extra"},
		},
	}

	var buf bytes.Buffer
	opts := modelListOptions{json: false, fields: []string{"name", "fields"}}
	err := runModelList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should output name and fields, tab-separated
	if !strings.Contains(output, "Basic\tFront,Back") {
		t.Errorf("expected Basic with fields, got %q", output)
	}
	if !strings.Contains(output, "Cloze\tText,Extra") {
		t.Errorf("expected Cloze with fields, got %q", output)
	}
}

func TestModelList_JSON_Default(t *testing.T) {
	mock := &mockClient{
		modelNamesAndIDs: map[string]int64{
			"Basic": 1234567890,
			"Cloze": 9876543210,
		},
		modelFieldNames: map[string][]string{
			"Basic": {"Front", "Back"},
			"Cloze": {"Text", "Extra"},
		},
	}

	var buf bytes.Buffer
	opts := modelListOptions{json: true, fields: nil}
	err := runModelList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse JSON output
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Should have all fields by default
	if len(result) != 2 {
		t.Errorf("expected 2 models, got %d", len(result))
	}

	// Find Basic model
	var basic map[string]interface{}
	for _, m := range result {
		if m["name"] == "Basic" {
			basic = m
			break
		}
	}

	if basic == nil {
		t.Fatal("Basic model not found in output")
	}

	if basic["id"].(float64) != 1234567890 {
		t.Errorf("expected id 1234567890, got %v", basic["id"])
	}

	fields := basic["fields"].([]interface{})
	if len(fields) != 2 || fields[0] != "Front" || fields[1] != "Back" {
		t.Errorf("expected fields [Front, Back], got %v", fields)
	}
}

func TestModelList_JSON_SelectedFields(t *testing.T) {
	mock := &mockClient{
		modelNamesAndIDs: map[string]int64{
			"Basic": 1234567890,
		},
	}

	var buf bytes.Buffer
	opts := modelListOptions{json: true, fields: []string{"name", "id"}}
	err := runModelList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 model, got %d", len(result))
	}

	// Should only have name and id, not fields
	if _, ok := result[0]["fields"]; ok {
		t.Error("should not include fields when not requested")
	}
	if result[0]["name"] != "Basic" {
		t.Errorf("expected name Basic, got %v", result[0]["name"])
	}
	if result[0]["id"].(float64) != 1234567890 {
		t.Errorf("expected id 1234567890, got %v", result[0]["id"])
	}
}

func TestModelList_EmptyResult(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{},
	}

	var buf bytes.Buffer
	opts := modelListOptions{json: false, fields: nil}
	err := runModelList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "No models found\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestModelList_EmptyResult_JSON(t *testing.T) {
	mock := &mockClient{
		modelNamesAndIDs: map[string]int64{},
	}

	var buf bytes.Buffer
	opts := modelListOptions{json: true, fields: nil}
	err := runModelList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty array, got %v", result)
	}
}

func TestModelList_InvalidField(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{"Basic"},
	}

	var buf bytes.Buffer
	opts := modelListOptions{json: false, fields: []string{"invalid_field"}}
	err := runModelList(mock, &buf, opts)

	if err == nil {
		t.Fatal("expected error for invalid field")
	}

	if !strings.Contains(err.Error(), "unknown field") {
		t.Errorf("expected 'unknown field' error, got %v", err)
	}
}

func TestModelList_ConnectionError(t *testing.T) {
	mock := &mockClient{
		modelNamesErr: errors.New("connection refused"),
	}

	var buf bytes.Buffer
	opts := modelListOptions{json: false, fields: nil}
	err := runModelList(mock, &buf, opts)

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("expected connection error, got %v", err)
	}
}

func TestModelList_FieldsError(t *testing.T) {
	mock := &mockClient{
		modelNames:     []string{"Basic"},
		modelFieldsErr: errors.New("failed to fetch fields"),
	}

	var buf bytes.Buffer
	opts := modelListOptions{json: false, fields: []string{"name", "fields"}}
	err := runModelList(mock, &buf, opts)

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "failed to fetch fields") {
		t.Errorf("expected fields error, got %v", err)
	}
}

func TestModelList_WithAllFields(t *testing.T) {
	mock := &mockClient{
		modelNamesAndIDs: map[string]int64{
			"Basic": 1234567890,
		},
		modelFieldNames: map[string][]string{
			"Basic": {"Front", "Back"},
		},
	}

	var buf bytes.Buffer
	opts := modelListOptions{json: false, fields: []string{"name", "id", "fields"}}
	err := runModelList(mock, &buf, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should output all fields, tab-separated
	if !strings.Contains(output, "Basic\t1234567890\tFront,Back") {
		t.Errorf("expected all fields, got %q", output)
	}
}

// =============================================================================
// Model Create Tests
// =============================================================================

func TestModelCreate_Basic(t *testing.T) {
	mock := &mockClient{
		createModelResult: map[string]interface{}{"id": int64(1234567890)},
	}

	var stdout, stderr bytes.Buffer
	opts := modelCreateOptions{
		fields:    []string{"Front", "Back"},
		templates: []string{"Card 1,{{Front}},{{Back}}"},
	}
	err := runModelCreate(mock, &stdout, &stderr, "My Model", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify params passed to client
	if mock.createdModelParams == nil {
		t.Fatal("expected CreateModel to be called")
	}
	if mock.createdModelParams.ModelName != "My Model" {
		t.Errorf("expected model name 'My Model', got %q", mock.createdModelParams.ModelName)
	}
	if len(mock.createdModelParams.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(mock.createdModelParams.Fields))
	}
	if len(mock.createdModelParams.CardTemplates) != 1 {
		t.Errorf("expected 1 template, got %d", len(mock.createdModelParams.CardTemplates))
	}
}

func TestModelCreate_MultipleFields(t *testing.T) {
	mock := &mockClient{
		createModelResult: map[string]interface{}{"id": int64(1234567890)},
	}

	var stdout, stderr bytes.Buffer
	opts := modelCreateOptions{
		fields:    []string{"Question", "Answer", "Notes", "Source"},
		templates: []string{"Card 1,{{Question}},{{Answer}}"},
	}
	err := runModelCreate(mock, &stdout, &stderr, "Multi Field", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.createdModelParams.Fields) != 4 {
		t.Errorf("expected 4 fields, got %d", len(mock.createdModelParams.Fields))
	}
	expectedFields := []string{"Question", "Answer", "Notes", "Source"}
	for i, f := range expectedFields {
		if mock.createdModelParams.Fields[i] != f {
			t.Errorf("expected field %d to be %q, got %q", i, f, mock.createdModelParams.Fields[i])
		}
	}
}

func TestModelCreate_MultipleTemplates(t *testing.T) {
	mock := &mockClient{
		createModelResult: map[string]interface{}{"id": int64(1234567890)},
	}

	var stdout, stderr bytes.Buffer
	opts := modelCreateOptions{
		fields:    []string{"Front", "Back"},
		templates: []string{"Forward,{{Front}},{{Back}}", "Reverse,{{Back}},{{Front}}"},
	}
	err := runModelCreate(mock, &stdout, &stderr, "Bidirectional", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.createdModelParams.CardTemplates) != 2 {
		t.Errorf("expected 2 templates, got %d", len(mock.createdModelParams.CardTemplates))
	}
	if mock.createdModelParams.CardTemplates[0].Name != "Forward" {
		t.Errorf("expected first template name 'Forward', got %q", mock.createdModelParams.CardTemplates[0].Name)
	}
	if mock.createdModelParams.CardTemplates[1].Name != "Reverse" {
		t.Errorf("expected second template name 'Reverse', got %q", mock.createdModelParams.CardTemplates[1].Name)
	}
}

func TestModelCreate_WithCSS(t *testing.T) {
	mock := &mockClient{
		createModelResult: map[string]interface{}{"id": int64(1234567890)},
	}

	var stdout, stderr bytes.Buffer
	opts := modelCreateOptions{
		fields:    []string{"Q", "A"},
		templates: []string{"Card 1,{{Q}},{{A}}"},
		css:       ".card { font-size: 24px; }",
	}
	err := runModelCreate(mock, &stdout, &stderr, "Styled", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.createdModelParams.CSS != ".card { font-size: 24px; }" {
		t.Errorf("expected CSS to be set, got %q", mock.createdModelParams.CSS)
	}
}

func TestModelCreate_Cloze(t *testing.T) {
	mock := &mockClient{
		createModelResult: map[string]interface{}{"id": int64(1234567890)},
	}

	var stdout, stderr bytes.Buffer
	opts := modelCreateOptions{
		fields:    []string{"Text", "Extra"},
		templates: []string{"Cloze,{{cloze:Text}},{{Text}}"},
		isCloze:   true,
	}
	err := runModelCreate(mock, &stdout, &stderr, "My Cloze", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mock.createdModelParams.IsCloze {
		t.Error("expected IsCloze to be true")
	}
}

func TestModelCreate_MissingFields(t *testing.T) {
	mock := &mockClient{}

	var stdout, stderr bytes.Buffer
	opts := modelCreateOptions{
		fields:    []string{},
		templates: []string{"Card 1,X,Y"},
	}
	err := runModelCreate(mock, &stdout, &stderr, "No Fields", opts)

	if err == nil {
		t.Fatal("expected error for missing fields")
	}
	if !strings.Contains(err.Error(), "field") {
		t.Errorf("expected error about fields, got %v", err)
	}
}

func TestModelCreate_MissingTemplates(t *testing.T) {
	mock := &mockClient{}

	var stdout, stderr bytes.Buffer
	opts := modelCreateOptions{
		fields:    []string{"Front", "Back"},
		templates: []string{},
	}
	err := runModelCreate(mock, &stdout, &stderr, "No Templates", opts)

	if err == nil {
		t.Fatal("expected error for missing templates")
	}
	if !strings.Contains(err.Error(), "template") {
		t.Errorf("expected error about templates, got %v", err)
	}
}

func TestModelCreate_InvalidTemplateFormat(t *testing.T) {
	mock := &mockClient{}

	var stdout, stderr bytes.Buffer
	opts := modelCreateOptions{
		fields:    []string{"Front", "Back"},
		templates: []string{"InvalidFormat"},
	}
	err := runModelCreate(mock, &stdout, &stderr, "Bad Template", opts)

	if err == nil {
		t.Fatal("expected error for invalid template format")
	}
	if !strings.Contains(err.Error(), "template") {
		t.Errorf("expected error about template format, got %v", err)
	}
}

func TestModelCreate_DuplicateName(t *testing.T) {
	mock := &mockClient{
		createModelErr: errors.New("Model name already exists"),
	}

	var stdout, stderr bytes.Buffer
	opts := modelCreateOptions{
		fields:    []string{"Front", "Back"},
		templates: []string{"Card 1,{{Front}},{{Back}}"},
	}
	err := runModelCreate(mock, &stdout, &stderr, "Existing Model", opts)

	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got %v", err)
	}
}

func TestModelCreate_EmptyName(t *testing.T) {
	mock := &mockClient{}

	var stdout, stderr bytes.Buffer
	opts := modelCreateOptions{
		fields:    []string{"Front", "Back"},
		templates: []string{"Card 1,{{Front}},{{Back}}"},
	}
	err := runModelCreate(mock, &stdout, &stderr, "", opts)

	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error about name, got %v", err)
	}
}

func TestModelCreate_WhitespaceOnlyName(t *testing.T) {
	mock := &mockClient{}

	var stdout, stderr bytes.Buffer
	opts := modelCreateOptions{
		fields:    []string{"Front", "Back"},
		templates: []string{"Card 1,{{Front}},{{Back}}"},
	}
	err := runModelCreate(mock, &stdout, &stderr, "   ", opts)

	if err == nil {
		t.Fatal("expected error for whitespace-only name")
	}
}

// =============================================================================
// Model Prune Tests
// =============================================================================

func TestModelPrune_All_Force(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{"Empty1", "Empty2"},
		noteIDs:    []int64{}, // empty - no notes
	}

	var stdin bytes.Buffer
	var stdout, stderr bytes.Buffer
	opts := modelPruneOptions{force: true, dryRun: false}
	err := runModelPrune(mock, &stdin, &stdout, &stderr, []string{}, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have called RemoveEmptyNotes
	if !mock.removeEmptyNotesCalled {
		t.Error("expected RemoveEmptyNotes to be called")
	}
}

func TestModelPrune_Specific_Force(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{"Empty1", "Empty2"},
		noteIDs:    []int64{}, // no notes
	}

	var stdin bytes.Buffer
	var stdout, stderr bytes.Buffer
	opts := modelPruneOptions{force: true, dryRun: false}
	err := runModelPrune(mock, &stdin, &stdout, &stderr, []string{"Empty1"}, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that removeEmptyNotes was called
	if !mock.removeEmptyNotesCalled {
		t.Error("expected RemoveEmptyNotes to be called")
	}
}

func TestModelPrune_SkipsNonEmpty(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{"HasNotes"},
		noteIDs:    []int64{1, 2, 3}, // has notes
	}

	var stdin bytes.Buffer
	var stdout, stderr bytes.Buffer
	opts := modelPruneOptions{force: true, dryRun: false}
	err := runModelPrune(mock, &stdin, &stdout, &stderr, []string{"HasNotes"}, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should report that model was skipped
	if !strings.Contains(stderr.String(), "Skipped") {
		t.Errorf("expected 'Skipped' message, got %q", stderr.String())
	}
}

func TestModelPrune_NotFound(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{"Existing"},
	}

	var stdin bytes.Buffer
	var stdout, stderr bytes.Buffer
	opts := modelPruneOptions{force: true, dryRun: false}
	err := runModelPrune(mock, &stdin, &stdout, &stderr, []string{"NonExistent"}, opts)

	if err == nil {
		t.Fatal("expected error for non-existent model")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got %v", err)
	}
}

func TestModelPrune_DryRun(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{"Empty1", "Empty2"},
		noteIDs:    []int64{}, // no notes
	}

	var stdin bytes.Buffer
	var stdout, stderr bytes.Buffer
	opts := modelPruneOptions{dryRun: true}
	err := runModelPrune(mock, &stdin, &stdout, &stderr, []string{}, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should NOT call removeEmptyNotes
	if mock.removeEmptyNotesCalled {
		t.Error("expected RemoveEmptyNotes to NOT be called in dry-run")
	}

	// Should show preview
	if !strings.Contains(stderr.String(), "Would") {
		t.Errorf("expected dry-run preview message, got %q", stderr.String())
	}
}

func TestModelPrune_NoEmptyModels(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{"HasNotes1", "HasNotes2"},
		noteIDs:    []int64{1}, // all have notes
	}

	var stdin bytes.Buffer
	var stdout, stderr bytes.Buffer
	opts := modelPruneOptions{force: true, dryRun: false}
	err := runModelPrune(mock, &stdin, &stdout, &stderr, []string{}, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should report nothing to prune
	output := stderr.String()
	if !strings.Contains(output, "No empty models") && !strings.Contains(output, "nothing") {
		t.Errorf("expected message about no empty models, got %q", output)
	}
}

func TestModelPrune_ConfirmYes(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{"Empty1"},
		noteIDs:    []int64{}, // no notes
	}

	stdin := bytes.NewBufferString("y\n")
	var stdout, stderr bytes.Buffer
	opts := modelPruneOptions{force: false, dryRun: false}
	err := runModelPrune(mock, stdin, &stdout, &stderr, []string{}, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have called RemoveEmptyNotes
	if !mock.removeEmptyNotesCalled {
		t.Error("expected RemoveEmptyNotes to be called after confirmation")
	}

	// Should show confirmation prompt
	if !strings.Contains(stderr.String(), "Continue?") {
		t.Errorf("expected confirmation prompt, got %q", stderr.String())
	}
}

func TestModelPrune_ConfirmNo(t *testing.T) {
	// When confirmation is declined, the command returns ErrCancelled.
	// The Cobra command must have cmd.SilenceUsage = true to prevent
	// usage from being printed when this error is returned.
	// See modelPruneCmd in model.go.
	mock := &mockClient{
		modelNames: []string{"Empty1"},
		noteIDs:    []int64{}, // no notes
	}

	stdin := bytes.NewBufferString("n\n")
	var stdout, stderr bytes.Buffer
	opts := modelPruneOptions{force: false, dryRun: false}
	err := runModelPrune(mock, stdin, &stdout, &stderr, []string{}, opts)

	// Should return ErrCancelled
	if err != ErrCancelled {
		t.Errorf("expected ErrCancelled, got %v", err)
	}

	// Should NOT call RemoveEmptyNotes
	if mock.removeEmptyNotesCalled {
		t.Error("expected RemoveEmptyNotes to NOT be called when cancelled")
	}
}

func TestModelPrune_ConfirmEmpty(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{"Empty1"},
		noteIDs:    []int64{}, // no notes
	}

	stdin := bytes.NewBufferString("\n") // empty response = no
	var stdout, stderr bytes.Buffer
	opts := modelPruneOptions{force: false, dryRun: false}
	err := runModelPrune(mock, stdin, &stdout, &stderr, []string{}, opts)

	// Should return ErrCancelled
	if err != ErrCancelled {
		t.Errorf("expected ErrCancelled for empty response, got %v", err)
	}
}
