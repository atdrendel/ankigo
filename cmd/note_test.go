package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/atdrendel/ankigo/internal/ankiconnect"
)

// === Note Create Tests ===

func TestNoteCreate_Basic_Success(t *testing.T) {
	mock := &mockClient{
		addNoteID:  1234567890,
		modelNames: []string{"Basic", "Cloze"},
		modelFieldNames: map[string][]string{
			"Basic": {"Front", "Back"},
		},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		back:  "Answer",
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

func TestNoteCreate_Basic_MissingFront(t *testing.T) {
	mock := &mockClient{
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic",
		back:  "Answer",
		// front is missing
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

func TestNoteCreate_Basic_MissingBack(t *testing.T) {
	mock := &mockClient{
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		// back is missing
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "--back is required for Basic model" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoteCreate_WithTags(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Q",
		back:  "A",
		tags:  []string{"tag1", "tag2", "tag3"},
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

func TestNoteCreate_CustomDeck(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Japanese::JLPT N3",
		model: "Basic",
		front: "日本",
		back:  "Japan",
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.addedNote.DeckName != "Japanese::JLPT N3" {
		t.Errorf("expected deck 'Japanese::JLPT N3', got %q", mock.addedNote.DeckName)
	}
}

func TestNoteCreate_ModelNotFound(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{"Basic", "Cloze"},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "NonExistent",
		front: "Q",
		back:  "A",
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != `model "NonExistent" not found` {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoteCreate_DuplicateError(t *testing.T) {
	mock := &mockClient{
		addNoteErr:      errors.New("cannot create note because it is a duplicate"),
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Q",
		back:  "A",
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "note already exists (use --allow-duplicate to add anyway)" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoteCreate_AllowDuplicate(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:           "Default",
		model:          "Basic",
		front:          "Q",
		back:           "A",
		allowDuplicate: true,
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

func TestNoteCreate_DuplicateScopeDeck(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:           "Default",
		model:          "Basic",
		front:          "Q",
		back:           "A",
		duplicateScope: "deck",
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

func TestNoteCreate_ClozeModel(t *testing.T) {
	mock := &mockClient{
		addNoteID:  1234567890,
		modelNames: []string{"Basic", "Cloze"},
		modelFieldNames: map[string][]string{
			"Cloze": {"Text", "Extra"},
		},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Cloze",
		fields: map[string]string{
			"Text":  "The capital of {{c1::France}} is {{c2::Paris}}",
			"Extra": "Geography",
		},
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

func TestNoteCreate_MixedFrontBackAndField(t *testing.T) {
	mock := &mockClient{
		addNoteID:  1234567890,
		modelNames: []string{"Basic (and reversed card)"},
		modelFieldNames: map[string][]string{
			"Basic (and reversed card)": {"Front", "Back"},
		},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic (and reversed card)",
		front: "Q",
		back:  "A",
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

func TestNoteCreate_InvalidFieldWarning(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Q",
		back:  "A",
		fields: map[string]string{
			"InvalidField": "value",
		},
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

func TestNoteCreate_NoFieldsError(t *testing.T) {
	mock := &mockClient{
		modelNames: []string{"Custom"},
		modelFieldNames: map[string][]string{
			"Custom": {"Field1", "Field2"},
		},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Custom",
		// No front, back, or fields
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "at least one field must be provided") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoteCreate_ConnectionError(t *testing.T) {
	mock := &mockClient{
		modelNamesErr: errors.New("connection refused"),
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Q",
		back:  "A",
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get model names") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoteCreate_DeckNotFoundError(t *testing.T) {
	mock := &mockClient{
		addNoteErr:      errors.New("deck was not found: NonExistent"),
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "NonExistent",
		model: "Basic",
		front: "Q",
		back:  "A",
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != `deck "NonExistent" not found` {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoteCreate_EmptyContentError(t *testing.T) {
	mock := &mockClient{
		addNoteErr:      errors.New("cannot create note because it is empty"),
		modelNames:      []string{"Custom"},
		modelFieldNames: map[string][]string{"Custom": {"Field1"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Custom",
		fields: map[string]string{
			"Field1": "",
		},
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "note content cannot be empty" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoteCreate_ModelFieldsError_StillSucceeds(t *testing.T) {
	// If we can't fetch field names for validation, the command should still work
	mock := &mockClient{
		addNoteID:      1234567890,
		modelNames:     []string{"Basic"},
		modelFieldsErr: errors.New("some error"),
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Q",
		back:  "A",
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

	// Should succeed - field validation is optional
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout.String() != "1234567890\n" {
		t.Errorf("expected note ID output, got %q", stdout.String())
	}
}

func TestNoteCreate_FieldOverridesFrontBack(t *testing.T) {
	// When both --field and --front/--back are provided, --front/--back take precedence
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "from front flag",
		back:  "from back flag",
		fields: map[string]string{
			"Front": "from field flag",
		},
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

// === Note Create with Media Tests ===

func TestNoteCreate_WithAudio_NoteHasAudioField(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		back:  "Answer",
		audio: []string{"filename=test.mp3,path=/tmp/test.mp3,fields=Back"},
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

func TestNoteCreate_WithVideo_NoteHasVideoField(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		back:  "Answer",
		video: []string{"filename=clip.mp4,url=https://example.com/video.mp4,fields=Back"},
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

func TestNoteCreate_WithPicture_NoteHasPictureField(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:    "Default",
		model:   "Basic",
		front:   "Question?",
		back:    "Answer",
		picture: []string{"filename=image.jpg,path=/tmp/image.jpg,fields=Front"},
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

func TestNoteCreate_MultipleAudio_NoteHasAllAudio(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		back:  "Answer",
		audio: []string{
			"filename=audio1.mp3,path=/tmp/a1.mp3,fields=Front",
			"filename=audio2.mp3,path=/tmp/a2.mp3,fields=Back",
		},
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

func TestNoteCreate_MixedMedia_NoteHasAllMediaTypes(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:    "Default",
		model:   "Basic",
		front:   "Question?",
		back:    "Answer",
		audio:   []string{"filename=audio.mp3,path=/tmp/a.mp3,fields=Back"},
		video:   []string{"filename=video.mp4,url=https://example.com/v.mp4,fields=Back"},
		picture: []string{"filename=image.png,path=/tmp/i.png,fields=Front"},
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

func TestNoteCreate_MediaWithTags(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		back:  "Answer",
		tags:  []string{"tag1", "tag2"},
		audio: []string{"filename=audio.mp3,path=/tmp/a.mp3,fields=Back"},
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

func TestNoteCreate_MediaFieldNotInModel_Warning(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		back:  "Answer",
		audio: []string{"filename=audio.mp3,path=/tmp/a.mp3,fields=NonExistent"},
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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

func TestNoteCreate_InvalidMediaSpec(t *testing.T) {
	mock := &mockClient{
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		deck:  "Default",
		model: "Basic",
		front: "Question?",
		back:  "Answer",
		audio: []string{"invalid-spec"},
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should not have called AddNote
	if mock.addedNote != nil {
		t.Error("expected AddNote NOT to be called")
	}
}

// === Note Create JSON Input Tests ===

func TestNoteCreate_InputJSON_Basic(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		inputJSON: `{"deckName":"Default","modelName":"Basic","fields":{"Front":"Q","Back":"A"}}`,
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

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
	if mock.addedNote.Fields["Front"] != "Q" {
		t.Errorf("expected Front 'Q', got %q", mock.addedNote.Fields["Front"])
	}
	if mock.addedNote.Fields["Back"] != "A" {
		t.Errorf("expected Back 'A', got %q", mock.addedNote.Fields["Back"])
	}
	if stdout.String() != "1234567890\n" {
		t.Errorf("expected stdout '1234567890\\n', got %q", stdout.String())
	}
}

func TestNoteCreate_InputJSON_WithTags(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		inputJSON: `{"deckName":"Default","modelName":"Basic","fields":{"Front":"Q","Back":"A"},"tags":["tag1","tag2"]}`,
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.addedNote.Tags) != 2 || mock.addedNote.Tags[0] != "tag1" || mock.addedNote.Tags[1] != "tag2" {
		t.Errorf("expected tags [tag1, tag2], got %v", mock.addedNote.Tags)
	}
}

func TestNoteCreate_InputJSON_WithMedia(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		inputJSON: `{"deckName":"Default","modelName":"Basic","fields":{"Front":"Q","Back":"A"},"audio":[{"filename":"a.mp3","path":"/tmp/a.mp3","fields":["Back"]}]}`,
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.addedNote.Audio) != 1 {
		t.Fatalf("expected 1 audio attachment, got %d", len(mock.addedNote.Audio))
	}
	if mock.addedNote.Audio[0].Filename != "a.mp3" {
		t.Errorf("expected filename 'a.mp3', got %q", mock.addedNote.Audio[0].Filename)
	}
	if mock.addedNote.Audio[0].Path != "/tmp/a.mp3" {
		t.Errorf("expected path '/tmp/a.mp3', got %q", mock.addedNote.Audio[0].Path)
	}
	if len(mock.addedNote.Audio[0].Fields) != 1 || mock.addedNote.Audio[0].Fields[0] != "Back" {
		t.Errorf("expected fields [Back], got %v", mock.addedNote.Audio[0].Fields)
	}
}

func TestNoteCreate_InputJSON_WithOptions(t *testing.T) {
	mock := &mockClient{
		addNoteID:       1234567890,
		modelNames:      []string{"Basic"},
		modelFieldNames: map[string][]string{"Basic": {"Front", "Back"}},
	}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		inputJSON: `{"deckName":"Default","modelName":"Basic","fields":{"Front":"Q","Back":"A"},"allowDuplicate":true,"duplicateScope":"deck"}`,
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.addedNote.Options == nil {
		t.Fatal("expected Options to be set")
	}
	if !mock.addedNote.Options.AllowDuplicate {
		t.Error("expected AllowDuplicate to be true")
	}
	if mock.addedNote.Options.DuplicateScope != "deck" {
		t.Errorf("expected DuplicateScope 'deck', got %q", mock.addedNote.Options.DuplicateScope)
	}
}

func TestNoteCreate_InputJSON_InvalidJSON(t *testing.T) {
	mock := &mockClient{}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		inputJSON: `{invalid json`,
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("expected 'invalid JSON' error, got: %v", err)
	}
}

func TestNoteCreate_InputJSON_ConflictWithFlags(t *testing.T) {
	mock := &mockClient{}

	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{
		inputJSON: `{"deckName":"Default","modelName":"Basic","fields":{"Front":"Q","Back":"A"}}`,
		front:     "Q",
	}

	err := runNoteCreate(mock, &stdout, &stderr, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--input-json") {
		t.Errorf("expected error about --input-json conflict, got: %v", err)
	}
}

// === Note Create Schema Tests ===

func TestNoteCreate_Schema(t *testing.T) {
	var stdout, stderr bytes.Buffer
	opts := noteCreateOptions{schema: true}

	// Should work with nil client — no Anki connection needed
	err := runNoteCreate(nil, &stdout, &stderr, opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should output valid JSON
	var schema map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &schema); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	// Verify top-level structure
	if schema["type"] != "object" {
		t.Errorf("expected type 'object', got %v", schema["type"])
	}

	required, ok := schema["required"].([]interface{})
	if !ok {
		t.Fatal("expected 'required' to be an array")
	}
	requiredSet := make(map[string]bool)
	for _, r := range required {
		requiredSet[r.(string)] = true
	}
	for _, field := range []string{"deckName", "modelName", "fields"} {
		if !requiredSet[field] {
			t.Errorf("expected %q in required fields", field)
		}
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'properties' to be an object")
	}
	for _, field := range []string{"deckName", "modelName", "fields", "tags", "audio", "video", "picture"} {
		if _, ok := props[field]; !ok {
			t.Errorf("expected property %q in schema", field)
		}
	}

	// Should have $defs with mediaAttachment
	defs, ok := schema["$defs"].(map[string]interface{})
	if !ok {
		t.Fatal("expected '$defs' to be an object")
	}
	if _, ok := defs["mediaAttachment"]; !ok {
		t.Error("expected 'mediaAttachment' in $defs")
	}
}

// === Note Delete Tests ===

func TestNoteDelete_Success_SingleNote(t *testing.T) {
	mock := &mockClient{}

	var stdout, stderr bytes.Buffer
	err := runNoteDelete(mock, nil, &stdout, &stderr, []int64{1234567890}, true, false, alwaysInteractive)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.deletedNotes) != 1 || mock.deletedNotes[0] != 1234567890 {
		t.Errorf("expected deletedNotes [1234567890], got %v", mock.deletedNotes)
	}
	if stderr.String() != "Deleted 1234567890\n" {
		t.Errorf("expected stderr 'Deleted 1234567890\\n', got %q", stderr.String())
	}
}

func TestNoteDelete_Success_MultipleNotes(t *testing.T) {
	mock := &mockClient{}

	var stdout, stderr bytes.Buffer
	err := runNoteDelete(mock, nil, &stdout, &stderr, []int64{111, 222, 333}, true, false, alwaysInteractive)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.deletedNotes) != 3 {
		t.Errorf("expected 3 deletedNotes, got %d", len(mock.deletedNotes))
	}
	expectedStderr := "Deleted 111\nDeleted 222\nDeleted 333\n"
	if stderr.String() != expectedStderr {
		t.Errorf("expected stderr %q, got %q", expectedStderr, stderr.String())
	}
}

func TestNoteDelete_DryRun(t *testing.T) {
	mock := &mockClient{}

	var stdout, stderr bytes.Buffer
	err := runNoteDelete(mock, nil, &stdout, &stderr, []int64{1234567890}, true, true, alwaysInteractive)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should NOT call the API
	if mock.deletedNotes != nil {
		t.Errorf("expected no API call, but deletedNotes was set to %v", mock.deletedNotes)
	}
	// Should output note ID to stdout
	if stdout.String() != "1234567890\n" {
		t.Errorf("expected stdout '1234567890\\n', got %q", stdout.String())
	}
	// Should show info message on stderr
	if !strings.Contains(stderr.String(), "Would delete") {
		t.Errorf("expected 'Would delete' on stderr, got %q", stderr.String())
	}
}

func TestNoteDelete_ConfirmationYes(t *testing.T) {
	mock := &mockClient{}

	stdin := bytes.NewBufferString("y\n")
	var stdout, stderr bytes.Buffer
	err := runNoteDelete(mock, stdin, &stdout, &stderr, []int64{1234567890}, false, false, alwaysInteractive)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should call the API
	if len(mock.deletedNotes) != 1 || mock.deletedNotes[0] != 1234567890 {
		t.Errorf("expected deletedNotes [1234567890], got %v", mock.deletedNotes)
	}
	// Should show prompt on stderr
	if !strings.Contains(stderr.String(), "will be deleted") {
		t.Errorf("expected warning on stderr, got %q", stderr.String())
	}
}

func TestNoteDelete_ConfirmationNo(t *testing.T) {
	mock := &mockClient{}

	stdin := bytes.NewBufferString("n\n")
	var stdout, stderr bytes.Buffer
	err := runNoteDelete(mock, stdin, &stdout, &stderr, []int64{1234567890}, false, false, alwaysInteractive)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrCancelled) {
		t.Errorf("expected ErrCancelled, got: %v", err)
	}
	// Should NOT call the API
	if mock.deletedNotes != nil {
		t.Errorf("expected no API call, but deletedNotes was set to %v", mock.deletedNotes)
	}
}

func TestNoteDelete_NonInteractiveWithoutForce(t *testing.T) {
	mock := &mockClient{}

	var stdout, stderr bytes.Buffer
	err := runNoteDelete(mock, nil, &stdout, &stderr, []int64{1234567890}, false, false, neverInteractive)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("expected error mentioning --force, got: %v", err)
	}
	// Should NOT call the API
	if mock.deletedNotes != nil {
		t.Errorf("expected no API call, but deletedNotes was set to %v", mock.deletedNotes)
	}
}

func TestNoteDelete_ConnectionError(t *testing.T) {
	mock := &mockClient{
		deleteNotesErr: errors.New("connection refused"),
	}

	var stdout, stderr bytes.Buffer
	err := runNoteDelete(mock, nil, &stdout, &stderr, []int64{1234567890}, true, false, alwaysInteractive)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to delete notes") {
		t.Errorf("unexpected error: %v", err)
	}
}

// === Note List Tests ===

func TestNoteList_NoQuery_ListsAllNotes(t *testing.T) {
	mock := &mockClient{
		noteIDs: []int64{111, 222, 333},
	}

	var buf bytes.Buffer
	opts := noteListOptions{json: false, fields: nil}
	err := runNoteList(mock, &buf, "", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// When no query is provided, should use "deck:*" to match all notes
	if mock.noteQuery != "deck:*" {
		t.Errorf("expected query 'deck:*', got %q", mock.noteQuery)
	}
	// Default output is just note IDs
	expected := "111\n222\n333\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestNoteList_WithQuery_FiltersNotes(t *testing.T) {
	mock := &mockClient{
		noteIDs: []int64{123, 456},
	}

	var buf bytes.Buffer
	opts := noteListOptions{json: false, fields: nil}
	err := runNoteList(mock, &buf, "deck:Japanese", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.noteQuery != "deck:Japanese" {
		t.Errorf("expected query 'deck:Japanese', got %q", mock.noteQuery)
	}
	expected := "123\n456\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestNoteList_EmptyResult(t *testing.T) {
	mock := &mockClient{
		noteIDs: []int64{},
	}

	var buf bytes.Buffer
	opts := noteListOptions{json: false, fields: nil}
	err := runNoteList(mock, &buf, "deck:NonExistent", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "No notes found\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestNoteList_WithFields_Model(t *testing.T) {
	mock := &mockClient{
		noteIDs: []int64{111, 222},
		noteInfos: []ankiconnect.NoteInfo{
			{NoteID: 111, ModelName: "Basic"},
			{NoteID: 222, ModelName: "Cloze"},
		},
	}

	var buf bytes.Buffer
	opts := noteListOptions{json: false, fields: []string{"id", "model"}}
	err := runNoteList(mock, &buf, "deck:Test", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "111\tBasic\n222\tCloze\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestNoteList_WithFields_Tags(t *testing.T) {
	mock := &mockClient{
		noteIDs: []int64{111},
		noteInfos: []ankiconnect.NoteInfo{
			{NoteID: 111, Tags: []string{"vocab", "japanese"}},
		},
	}

	var buf bytes.Buffer
	opts := noteListOptions{json: false, fields: []string{"id", "tags"}}
	err := runNoteList(mock, &buf, "deck:Test", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Tags should be comma-separated in text output
	expected := "111\tvocab,japanese\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestNoteList_WithFields_Multiple(t *testing.T) {
	mock := &mockClient{
		noteIDs: []int64{111},
		noteInfos: []ankiconnect.NoteInfo{
			{
				NoteID:    111,
				ModelName: "Basic",
				Tags:      []string{"tag1"},
				Mod:       1609459200,
				Cards:     []int64{999, 888},
			},
		},
	}

	var buf bytes.Buffer
	opts := noteListOptions{json: false, fields: []string{"id", "model", "tags", "mod", "cards"}}
	err := runNoteList(mock, &buf, "deck:Test", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Cards should be comma-separated in text output
	expected := "111\tBasic\ttag1\t1609459200\t999,888\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestNoteList_JSON_AllFields(t *testing.T) {
	mock := &mockClient{
		noteIDs: []int64{111},
		noteInfos: []ankiconnect.NoteInfo{
			{
				NoteID:    111,
				ModelName: "Basic",
				Tags:      []string{"tag1", "tag2"},
				Fields: map[string]ankiconnect.NoteFieldValue{
					"Front": {Value: "Q", Order: 0},
					"Back":  {Value: "A", Order: 1},
				},
				Mod:   1609459200,
				Cards: []int64{999},
			},
		},
	}

	var buf bytes.Buffer
	opts := noteListOptions{json: true, fields: nil}
	err := runNoteList(mock, &buf, "deck:Test", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse JSON output
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}

	entry := result[0]
	if entry["id"].(float64) != 111 {
		t.Errorf("expected id 111, got %v", entry["id"])
	}
	if entry["model"] != "Basic" {
		t.Errorf("expected model 'Basic', got %v", entry["model"])
	}
	tags := entry["tags"].([]interface{})
	if len(tags) != 2 || tags[0] != "tag1" || tags[1] != "tag2" {
		t.Errorf("expected tags ['tag1', 'tag2'], got %v", tags)
	}
}

func TestNoteList_JSON_SelectedFields(t *testing.T) {
	mock := &mockClient{
		noteIDs: []int64{111},
		noteInfos: []ankiconnect.NoteInfo{
			{NoteID: 111, ModelName: "Basic", Tags: []string{"tag1"}},
		},
	}

	var buf bytes.Buffer
	opts := noteListOptions{json: true, fields: []string{"id", "model"}}
	err := runNoteList(mock, &buf, "deck:Test", opts)

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

	entry := result[0]
	if entry["id"].(float64) != 111 {
		t.Errorf("expected id 111, got %v", entry["id"])
	}
	if entry["model"] != "Basic" {
		t.Errorf("expected model 'Basic', got %v", entry["model"])
	}
	// Should not have tags since it wasn't requested
	if _, hasTags := entry["tags"]; hasTags {
		t.Error("expected tags field to be absent")
	}
}

func TestNoteList_JSON_EmptyResult(t *testing.T) {
	mock := &mockClient{
		noteIDs: []int64{},
	}

	var buf bytes.Buffer
	opts := noteListOptions{json: true, fields: nil}
	err := runNoteList(mock, &buf, "deck:Test", opts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "[]\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestNoteList_InvalidField(t *testing.T) {
	mock := &mockClient{
		noteIDs: []int64{111},
	}

	var buf bytes.Buffer
	opts := noteListOptions{json: false, fields: []string{"invalid_field"}}
	err := runNoteList(mock, &buf, "deck:Test", opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "unknown field: invalid_field" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoteList_ConnectionError(t *testing.T) {
	mock := &mockClient{
		findNotesErr: errors.New("connection refused"),
	}

	var buf bytes.Buffer
	opts := noteListOptions{json: false, fields: nil}
	err := runNoteList(mock, &buf, "deck:Test", opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to find notes") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoteList_MissingNoteInfo(t *testing.T) {
	// When notesInfo returns fewer entries than expected
	mock := &mockClient{
		noteIDs: []int64{111, 222},
		noteInfos: []ankiconnect.NoteInfo{
			{NoteID: 111, ModelName: "Basic"}, // Only one entry
		},
	}

	var buf bytes.Buffer
	opts := noteListOptions{json: false, fields: []string{"id", "model"}}
	err := runNoteList(mock, &buf, "deck:Test", opts)

	// Should still succeed, outputting empty model for missing note
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "111\tBasic\n222\t\n"
	if buf.String() != expected {
		t.Errorf("expected output %q, got %q", expected, buf.String())
	}
}

func TestNoteList_NotesInfoError(t *testing.T) {
	mock := &mockClient{
		noteIDs:      []int64{111},
		notesInfoErr: errors.New("notes info error"),
	}

	var buf bytes.Buffer
	opts := noteListOptions{json: false, fields: []string{"id", "model"}}
	err := runNoteList(mock, &buf, "deck:Test", opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get note info") {
		t.Errorf("unexpected error: %v", err)
	}
}
