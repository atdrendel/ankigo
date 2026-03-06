package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/atdrendel/ankigo/internal/ankiconnect"
	"github.com/spf13/cobra"
)

// noteCreateOptions holds all options for the note create command.
type noteCreateOptions struct {
	deck           string
	model          string
	front          string
	back           string
	fields         map[string]string
	tags           []string
	allowDuplicate bool
	duplicateScope string
	audio          []string
	video          []string
	picture        []string
	inputJSON      string
	schema         bool
}

// noteCreateSchemaJSON is the JSON Schema for --input-json on note create.
const noteCreateSchemaJSON = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "description": "Input schema for ankigo note create --input-json",
  "type": "object",
  "required": ["deckName", "modelName", "fields"],
  "properties": {
    "deckName": { "type": "string", "description": "Target deck name" },
    "modelName": { "type": "string", "description": "Note type (model) name" },
    "fields": {
      "type": "object",
      "additionalProperties": { "type": "string" },
      "description": "Field name to value mapping"
    },
    "tags": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Tags for the note"
    },
    "allowDuplicate": { "type": "boolean", "description": "Allow adding duplicate notes" },
    "duplicateScope": {
      "type": "string",
      "enum": ["deck", ""],
      "description": "Scope for duplicate check: \"deck\" or empty for collection-wide"
    },
    "audio": { "type": "array", "items": { "$ref": "#/$defs/mediaAttachment" }, "description": "Audio attachments" },
    "video": { "type": "array", "items": { "$ref": "#/$defs/mediaAttachment" }, "description": "Video attachments" },
    "picture": { "type": "array", "items": { "$ref": "#/$defs/mediaAttachment" }, "description": "Picture attachments" }
  },
  "$defs": {
    "mediaAttachment": {
      "type": "object",
      "required": ["filename"],
      "properties": {
        "filename": { "type": "string", "description": "Filename for the media in Anki's collection" },
        "url": { "type": "string", "description": "URL to download the media from" },
        "path": { "type": "string", "description": "Local file path (relative paths resolved to absolute)" },
        "data": { "type": "string", "description": "Base64-encoded file data" },
        "skipHash": { "type": "string", "description": "MD5 hash — skip storage if file with this hash exists" },
        "deleteExisting": { "type": "boolean", "description": "Delete existing file with same name (default: true)" },
        "fields": { "type": "array", "items": { "type": "string" }, "description": "Fields to add the media reference to" }
      }
    }
  }
}
`

// noteCreateInput is the JSON input structure for note create.
type noteCreateInput struct {
	DeckName       string                        `json:"deckName"`
	ModelName      string                        `json:"modelName"`
	Fields         map[string]string              `json:"fields"`
	Tags           []string                       `json:"tags,omitempty"`
	AllowDuplicate bool                           `json:"allowDuplicate,omitempty"`
	DuplicateScope string                         `json:"duplicateScope,omitempty"`
	Audio          []ankiconnect.MediaAttachment  `json:"audio,omitempty"`
	Video          []ankiconnect.MediaAttachment  `json:"video,omitempty"`
	Picture        []ankiconnect.MediaAttachment  `json:"picture,omitempty"`
}

// parseMediaSpec parses a media specification string into a MediaAttachment.
// Format: filename=<name>,<source>,fields=<f1>;<f2>
// Source is one of: path=/path, url=https://..., data=base64...
func parseMediaSpec(spec string) (ankiconnect.MediaAttachment, error) {
	var media ankiconnect.MediaAttachment

	if spec == "" {
		return media, fmt.Errorf("media specification cannot be empty")
	}

	// Parse key=value pairs separated by commas
	// Handle the case where values might contain = (like URLs)
	parts := strings.Split(spec, ",")
	for _, part := range parts {
		idx := strings.Index(part, "=")
		if idx == -1 {
			return media, fmt.Errorf("invalid media specification: expected key=value pairs, got %q", part)
		}
		key := part[:idx]
		value := part[idx+1:]

		switch key {
		case "filename":
			media.Filename = value
		case "path":
			media.Path = value
		case "url":
			media.URL = value
		case "data":
			media.Data = value
		case "fields":
			media.Fields = strings.Split(value, ";")
		default:
			return media, fmt.Errorf("unknown media specification key: %q", key)
		}
	}

	// Validate required fields
	if media.Filename == "" {
		return media, fmt.Errorf("media specification missing required 'filename'")
	}
	if media.Path == "" && media.URL == "" && media.Data == "" {
		return media, fmt.Errorf("media specification missing source: must specify 'path', 'url', or 'data'")
	}
	if len(media.Fields) == 0 {
		return media, fmt.Errorf("media specification missing required 'fields'")
	}

	// Convert relative paths to absolute paths (anki-connect requires absolute paths)
	if media.Path != "" && !filepath.IsAbs(media.Path) {
		absPath, err := filepath.Abs(media.Path)
		if err != nil {
			return media, fmt.Errorf("failed to resolve path %q: %w", media.Path, err)
		}
		media.Path = absPath
	}

	return media, nil
}

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Manage Anki notes",
	Long:  `Commands for creating and deleting Anki notes.`,
}

var noteCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new note",
	Long: `Create a new note in a deck.

For the default "Basic" model, use --front and --back flags:
  ankigo note create --front "Question" --back "Answer"

For other models, use --field to set arbitrary fields:
  ankigo note create --model "Cloze" --field "Text={{c1::answer}}"

The --front and --back flags are shortcuts that set "Front" and "Back" fields.`,
	Example: `  # Basic note
  ankigo note create --front "What is Go?" --back "A programming language"

  # Note with tags
  ankigo note create --front "猫" --back "cat" --tags japanese,vocabulary

  # Note in a specific deck
  ankigo note create --deck "Japanese::JLPT N3" --front "日本" --back "Japan"

  # Cloze deletion note
  ankigo note create --model "Cloze" --field "Text=The capital of {{c1::France}} is {{c2::Paris}}"

  # Allow duplicate
  ankigo note create --front "repeat" --back "answer" --allow-duplicate

  # Note with audio from local file
  ankigo note create --front "猫" --back "cat" --audio "filename=neko.mp3,path=./neko.mp3,fields=Front"

  # Note with image from URL
  ankigo note create --front "Q" --back "A" --picture "filename=img.jpg,url=https://example.com/img.jpg,fields=Front"

  # Note with multiple media attachments
  ankigo note create --front "Q" --back "A" \
    --audio "filename=a.mp3,path=./a.mp3,fields=Back" \
    --picture "filename=i.png,path=./i.png,fields=Front"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		schema, _ := cmd.Flags().GetBool("schema")
		if schema {
			return runNoteCreate(nil, cmd.OutOrStdout(), cmd.ErrOrStderr(), noteCreateOptions{schema: true})
		}

		client := ankiconnect.DefaultClient()

		inputJSON, _ := cmd.Flags().GetString("input-json")

		deck, _ := cmd.Flags().GetString("deck")
		model, _ := cmd.Flags().GetString("model")
		front, _ := cmd.Flags().GetString("front")
		back, _ := cmd.Flags().GetString("back")
		fieldFlags, _ := cmd.Flags().GetStringArray("field")
		tagsFlag, _ := cmd.Flags().GetStringSlice("tags")
		allowDup, _ := cmd.Flags().GetBool("allow-duplicate")
		dupScope, _ := cmd.Flags().GetString("duplicate-scope")
		audioFlags, _ := cmd.Flags().GetStringArray("audio")
		videoFlags, _ := cmd.Flags().GetStringArray("video")
		pictureFlags, _ := cmd.Flags().GetStringArray("picture")

		// Parse --field flags into map
		fields := make(map[string]string)
		for _, f := range fieldFlags {
			parts := strings.SplitN(f, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid field format %q: expected Name=value", f)
			}
			fields[parts[0]] = parts[1]
		}

		opts := noteCreateOptions{
			deck:           deck,
			model:          model,
			front:          front,
			back:           back,
			fields:         fields,
			tags:           tagsFlag,
			allowDuplicate: allowDup,
			duplicateScope: dupScope,
			audio:          audioFlags,
			video:          videoFlags,
			picture:        pictureFlags,
			inputJSON:      inputJSON,
		}

		return runNoteCreate(client, cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
	},
}

// resolveMediaPaths converts relative paths to absolute paths in media attachments.
func resolveMediaPaths(media []ankiconnect.MediaAttachment) {
	for i := range media {
		if media[i].Path != "" && !filepath.IsAbs(media[i].Path) {
			if absPath, err := filepath.Abs(media[i].Path); err == nil {
				media[i].Path = absPath
			}
		}
	}
}

// runNoteCreate is the testable implementation of note create.
func runNoteCreate(client Client, stdout, stderr io.Writer, opts noteCreateOptions) error {
	// Handle --schema: output JSON Schema and return
	if opts.schema {
		fmt.Fprint(stdout, noteCreateSchemaJSON)
		return nil
	}

	// Handle --input-json: parse and delegate to the common creation path
	if opts.inputJSON != "" {
		// Check for conflicts with other flags
		if opts.front != "" || opts.back != "" || len(opts.fields) > 0 || len(opts.audio) > 0 || len(opts.video) > 0 || len(opts.picture) > 0 {
			return fmt.Errorf("--input-json cannot be combined with --front, --back, --field, --audio, --video, or --picture")
		}

		var input noteCreateInput
		if err := json.Unmarshal([]byte(opts.inputJSON), &input); err != nil {
			return fmt.Errorf("invalid JSON input: %w", err)
		}

		// Resolve relative paths
		resolveMediaPaths(input.Audio)
		resolveMediaPaths(input.Video)
		resolveMediaPaths(input.Picture)

		// Build note directly from JSON input
		note := ankiconnect.Note{
			DeckName:  input.DeckName,
			ModelName: input.ModelName,
			Fields:    input.Fields,
			Tags:      input.Tags,
			Audio:     input.Audio,
			Video:     input.Video,
			Picture:   input.Picture,
		}
		if input.AllowDuplicate || input.DuplicateScope != "" {
			note.Options = &ankiconnect.NoteOptions{
				AllowDuplicate: input.AllowDuplicate,
				DuplicateScope: input.DuplicateScope,
			}
		}

		return createNote(client, stdout, note)
	}

	// Build the fields map
	fields := make(map[string]string)

	// Copy explicit field flags first
	for k, v := range opts.fields {
		fields[k] = v
	}

	// Apply --front and --back as convenience shortcuts (they override --field)
	if opts.front != "" {
		fields["Front"] = opts.front
	}
	if opts.back != "" {
		fields["Back"] = opts.back
	}

	// Validate: for Basic model, require Front and Back
	if opts.model == "Basic" {
		if fields["Front"] == "" {
			return fmt.Errorf("--front is required for Basic model")
		}
		if fields["Back"] == "" {
			return fmt.Errorf("--back is required for Basic model")
		}
	}

	// Validate: at least one field must be set
	if len(fields) == 0 {
		return fmt.Errorf("at least one field must be provided (use --front/--back or --field)")
	}

	// Validate model exists
	modelNames, err := client.ModelNames()
	if err != nil {
		return fmt.Errorf("failed to get model names: %w", err)
	}

	modelExists := false
	for _, m := range modelNames {
		if m == opts.model {
			modelExists = true
			break
		}
	}
	if !modelExists {
		return fmt.Errorf("model %q not found", opts.model)
	}

	// Validate field names (warn, don't fail)
	modelFields, err := client.ModelFieldNames(opts.model)
	var modelFieldSet map[string]bool
	if err == nil {
		modelFieldSet = make(map[string]bool)
		for _, f := range modelFields {
			modelFieldSet[f] = true
		}
		for fieldName := range fields {
			if !modelFieldSet[fieldName] {
				fmt.Fprintf(stderr, "warning: field %q is not in model %q (available: %s)\n",
					fieldName, opts.model, strings.Join(modelFields, ", "))
			}
		}
	}

	// Parse media attachments
	var audioAttachments []ankiconnect.MediaAttachment
	for _, spec := range opts.audio {
		media, err := parseMediaSpec(spec)
		if err != nil {
			return fmt.Errorf("invalid audio specification: %w", err)
		}
		// Warn if media field not in model
		if modelFieldSet != nil {
			for _, f := range media.Fields {
				if !modelFieldSet[f] {
					fmt.Fprintf(stderr, "warning: audio field %q is not in model %q (available: %s)\n",
						f, opts.model, strings.Join(modelFields, ", "))
				}
			}
		}
		audioAttachments = append(audioAttachments, media)
	}

	var videoAttachments []ankiconnect.MediaAttachment
	for _, spec := range opts.video {
		media, err := parseMediaSpec(spec)
		if err != nil {
			return fmt.Errorf("invalid video specification: %w", err)
		}
		if modelFieldSet != nil {
			for _, f := range media.Fields {
				if !modelFieldSet[f] {
					fmt.Fprintf(stderr, "warning: video field %q is not in model %q (available: %s)\n",
						f, opts.model, strings.Join(modelFields, ", "))
				}
			}
		}
		videoAttachments = append(videoAttachments, media)
	}

	var pictureAttachments []ankiconnect.MediaAttachment
	for _, spec := range opts.picture {
		media, err := parseMediaSpec(spec)
		if err != nil {
			return fmt.Errorf("invalid picture specification: %w", err)
		}
		if modelFieldSet != nil {
			for _, f := range media.Fields {
				if !modelFieldSet[f] {
					fmt.Fprintf(stderr, "warning: picture field %q is not in model %q (available: %s)\n",
						f, opts.model, strings.Join(modelFields, ", "))
				}
			}
		}
		pictureAttachments = append(pictureAttachments, media)
	}

	// Build the note
	note := ankiconnect.Note{
		DeckName:  opts.deck,
		ModelName: opts.model,
		Fields:    fields,
		Tags:      opts.tags,
		Audio:     audioAttachments,
		Video:     videoAttachments,
		Picture:   pictureAttachments,
	}

	// Add duplicate options if specified
	if opts.allowDuplicate || opts.duplicateScope != "" {
		note.Options = &ankiconnect.NoteOptions{
			AllowDuplicate: opts.allowDuplicate,
			DuplicateScope: opts.duplicateScope,
		}
	}

	return createNote(client, stdout, note)
}

// createNote sends the note to anki-connect and handles common error messages.
func createNote(client Client, stdout io.Writer, note ankiconnect.Note) error {
	noteID, err := client.AddNote(note)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "cannot create note because it is a duplicate") {
			return fmt.Errorf("note already exists (use --allow-duplicate to add anyway)")
		}
		if strings.Contains(errMsg, "cannot create note because it is empty") {
			return fmt.Errorf("note content cannot be empty")
		}
		if strings.Contains(errMsg, "model was not found") {
			return fmt.Errorf("model %q not found", note.ModelName)
		}
		if strings.Contains(errMsg, "deck was not found") {
			return fmt.Errorf("deck %q not found", note.DeckName)
		}
		return fmt.Errorf("failed to add note: %w", err)
	}

	fmt.Fprintln(stdout, noteID)
	return nil
}

var noteDeleteCmd = &cobra.Command{
	Use:   "delete [note-ids...]",
	Short: "Delete one or more notes",
	Long:  `Delete notes by ID. This also deletes all cards generated from those notes.`,
	Args:  cobra.MinimumNArgs(1),
	Example: `  # Delete a single note
  ankigo note delete 1234567890

  # Delete multiple notes
  ankigo note delete 1234567890 9876543210

  # Delete with confirmation skipped (for scripting)
  ankigo note delete 1234567890 --force

  # Preview what would be deleted
  ankigo note delete 1234567890 --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Silence usage for errors that happen during execution (not arg validation)
		cmd.SilenceUsage = true

		client := ankiconnect.DefaultClient()
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Parse note IDs from args
		noteIDs := make([]int64, 0, len(args))
		for _, arg := range args {
			id, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid note ID %q: must be a number", arg)
			}
			noteIDs = append(noteIDs, id)
		}

		return runNoteDelete(client, os.Stdin, cmd.OutOrStdout(), cmd.ErrOrStderr(), noteIDs, force, dryRun, isStdinTerminal)
	},
}

// runNoteDelete is the testable implementation of note delete.
func runNoteDelete(client Client, stdin io.Reader, stdout, stderr io.Writer, noteIDs []int64, force, dryRun bool, isTerminal func() bool) error {
	// Dry run: show what would be deleted
	if dryRun {
		fmt.Fprintln(stderr, "Would delete the following notes (and all their cards):")
		for _, id := range noteIDs {
			fmt.Fprintln(stdout, id)
		}
		return nil
	}

	// Confirmation prompt
	if !force {
		fmt.Fprintln(stderr, "The following notes will be deleted (including all their cards):")
		for _, id := range noteIDs {
			fmt.Fprintf(stderr, "  - %d\n", id)
		}
		if err := requireConfirmation(stdin, stderr, isTerminal); err != nil {
			return err
		}
	}

	// Delete notes
	if err := client.DeleteNotes(noteIDs); err != nil {
		return fmt.Errorf("failed to delete notes: %w", err)
	}

	// Report success
	for _, id := range noteIDs {
		fmt.Fprintf(stderr, "Deleted %d\n", id)
	}

	return nil
}

// noteListFields are the available fields for note list output.
var noteListFields = []string{"id", "model", "tags", "fields", "mod", "cards"}

// noteInfoFields are fields that require fetching note info.
var noteInfoFields = []string{"model", "tags", "fields", "mod", "cards"}

// noteListOptions holds the options for the note list command.
type noteListOptions struct {
	json   bool
	fields []string
}

var noteListCmd = &cobra.Command{
	Use:   "list [query]",
	Short: "List notes",
	Long:  `List notes in your Anki collection. Optionally filter using a query string.`,
	Example: `  ankigo note list
  ankigo note list "deck:Default"
  ankigo note list "tag:japanese"
  ankigo note list "note:Cloze"
  ankigo note list "edited:1"
  ankigo note list "note:\"Basic (and reversed card)\""`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := ankiconnect.DefaultClient()

		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		jsonFlag, _ := cmd.Flags().GetBool("json")
		fieldsStr, _ := cmd.Flags().GetString("fields")

		fields, err := parseFields(fieldsStr, noteListFields)
		if err != nil {
			return err
		}

		opts := noteListOptions{json: jsonFlag, fields: fields}
		return runNoteList(client, cmd.OutOrStdout(), query, opts)
	},
}

// needsNoteInfo returns true if any of the fields require fetching note info.
func needsNoteInfo(fields []string) bool {
	for _, f := range fields {
		if contains(noteInfoFields, f) {
			return true
		}
	}
	return false
}

// noteEntry holds note data for output.
type noteEntry struct {
	id   int64
	info *ankiconnect.NoteInfo
}

// runNoteList is the testable implementation of note list.
func runNoteList(client Client, out io.Writer, query string, opts noteListOptions) error {
	// Validate fields first
	if opts.fields != nil {
		for _, f := range opts.fields {
			if !contains(noteListFields, f) {
				return fmt.Errorf("unknown field: %s", f)
			}
		}
	}

	if opts.json {
		return runNoteListJSON(client, out, query, opts.fields)
	}
	return runNoteListText(client, out, query, opts.fields)
}

func runNoteListText(client Client, out io.Writer, query string, fields []string) error {
	// If no fields specified, default to ["id"]
	if fields == nil {
		fields = []string{"id"}
	}

	// If no query, use "deck:*" to match all notes
	if query == "" {
		query = "deck:*"
	}

	// Find notes
	noteIDs, err := client.FindNotes(query)
	if err != nil {
		return fmt.Errorf("failed to find notes: %w", err)
	}

	if len(noteIDs) == 0 {
		fmt.Fprintln(out, "No notes found")
		return nil
	}

	// Build entries
	entries := make([]noteEntry, len(noteIDs))
	for i, id := range noteIDs {
		entries[i] = noteEntry{id: id}
	}

	// Fetch note info if needed
	if needsNoteInfo(fields) {
		noteInfos, err := client.NotesInfo(noteIDs)
		if err != nil {
			return fmt.Errorf("failed to get note info: %w", err)
		}

		// Build a map for lookup
		infoMap := make(map[int64]*ankiconnect.NoteInfo)
		for i := range noteInfos {
			infoMap[noteInfos[i].NoteID] = &noteInfos[i]
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
			vals = append(vals, getNoteFieldText(e, f))
		}
		fmt.Fprintln(out, strings.Join(vals, "\t"))
	}
	return nil
}

// getNoteFieldText returns the text value for a note field.
func getNoteFieldText(e noteEntry, field string) string {
	switch field {
	case "id":
		return fmt.Sprintf("%d", e.id)
	case "model":
		if e.info != nil {
			return e.info.ModelName
		}
		return ""
	case "tags":
		if e.info != nil && len(e.info.Tags) > 0 {
			return strings.Join(e.info.Tags, ",")
		}
		return ""
	case "fields":
		if e.info != nil && e.info.Fields != nil {
			// Serialize fields as JSON for text output
			b, _ := json.Marshal(e.info.Fields)
			return string(b)
		}
		return "{}"
	case "mod":
		if e.info != nil {
			return fmt.Sprintf("%d", e.info.Mod)
		}
		return "0"
	case "cards":
		if e.info != nil && len(e.info.Cards) > 0 {
			var cardStrs []string
			for _, c := range e.info.Cards {
				cardStrs = append(cardStrs, fmt.Sprintf("%d", c))
			}
			return strings.Join(cardStrs, ",")
		}
		return ""
	}
	return ""
}

func runNoteListJSON(client Client, out io.Writer, query string, fields []string) error {
	// If no fields specified, default to all fields
	if fields == nil {
		fields = noteListFields
	}

	// If no query, use "deck:*" to match all notes
	if query == "" {
		query = "deck:*"
	}

	// Find notes
	noteIDs, err := client.FindNotes(query)
	if err != nil {
		return fmt.Errorf("failed to find notes: %w", err)
	}

	// Handle empty results
	if len(noteIDs) == 0 {
		fmt.Fprintln(out, "[]")
		return nil
	}

	// Build entries
	entries := make([]noteEntry, len(noteIDs))
	for i, id := range noteIDs {
		entries[i] = noteEntry{id: id}
	}

	// Fetch note info if needed
	if needsNoteInfo(fields) {
		noteInfos, err := client.NotesInfo(noteIDs)
		if err != nil {
			return fmt.Errorf("failed to get note info: %w", err)
		}

		// Build a map for lookup
		infoMap := make(map[int64]*ankiconnect.NoteInfo)
		for i := range noteInfos {
			infoMap[noteInfos[i].NoteID] = &noteInfos[i]
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
			obj[f] = getNoteFieldJSON(e, f)
		}
		result = append(result, obj)
	}

	// Ensure empty array instead of null
	if result == nil {
		result = []map[string]interface{}{}
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// getNoteFieldJSON returns the JSON value for a note field.
func getNoteFieldJSON(e noteEntry, field string) interface{} {
	switch field {
	case "id":
		return e.id
	case "model":
		if e.info != nil {
			return e.info.ModelName
		}
		return ""
	case "tags":
		if e.info != nil && e.info.Tags != nil {
			return e.info.Tags
		}
		return []string{}
	case "fields":
		if e.info != nil && e.info.Fields != nil {
			return e.info.Fields
		}
		return map[string]ankiconnect.NoteFieldValue{}
	case "mod":
		if e.info != nil {
			return e.info.Mod
		}
		return int64(0)
	case "cards":
		if e.info != nil && e.info.Cards != nil {
			return e.info.Cards
		}
		return []int64{}
	}
	return nil
}

func init() {
	noteCreateCmd.Flags().StringP("deck", "d", "Default", "deck to add the note to")
	noteCreateCmd.Flags().StringP("model", "m", "Basic", "note type (model) to use")
	noteCreateCmd.Flags().StringP("front", "f", "", "front of the note (for Basic model)")
	noteCreateCmd.Flags().StringP("back", "b", "", "back of the note (for Basic model)")
	noteCreateCmd.Flags().StringArray("field", nil, `set a field value (format: "FieldName=value", repeatable)`)
	noteCreateCmd.Flags().StringSlice("tags", nil, "tags for the note (comma-separated or repeatable)")
	noteCreateCmd.Flags().Bool("allow-duplicate", false, "allow adding duplicate notes")
	noteCreateCmd.Flags().String("duplicate-scope", "", `scope for duplicate check: "deck" or empty for collection-wide`)
	noteCreateCmd.Flags().StringArray("audio", nil, `attach audio (format: "filename=name.mp3,path=/file.mp3,fields=Back")`)
	noteCreateCmd.Flags().StringArray("video", nil, `attach video (format: "filename=name.mp4,url=https://...,fields=Back")`)
	noteCreateCmd.Flags().StringArray("picture", nil, `attach picture (format: "filename=name.jpg,path=/file.jpg,fields=Front")`)
	noteCreateCmd.Flags().String("input-json", "", "Create note from JSON (cannot be combined with other field flags)")
	noteCreateCmd.Flags().Bool("schema", false, "Print JSON Schema for --input-json and exit")

	noteDeleteCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
	noteDeleteCmd.Flags().Bool("dry-run", false, "show what would be deleted without deleting")

	noteListCmd.Flags().Bool("json", false, "Output in JSON format")
	noteListCmd.Flags().StringP("fields", "f", "", "Comma-separated list of fields (available: id, model, tags, fields, mod, cards)")

	noteCmd.AddCommand(noteCreateCmd)
	noteCmd.AddCommand(noteDeleteCmd)
	noteCmd.AddCommand(noteListCmd)
	rootCmd.AddCommand(noteCmd)
}
