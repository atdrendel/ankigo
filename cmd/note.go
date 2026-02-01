package cmd

import (
	"bufio"
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
		client := ankiconnect.DefaultClient()

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
		}

		return runNoteCreate(client, cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
	},
}

// runNoteCreate is the testable implementation of note create.
func runNoteCreate(client Client, stdout, stderr io.Writer, opts noteCreateOptions) error {
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

	// Create the note
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
			return fmt.Errorf("model %q not found", opts.model)
		}
		if strings.Contains(errMsg, "deck was not found") {
			return fmt.Errorf("deck %q not found", opts.deck)
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

		return runNoteDelete(client, os.Stdin, cmd.OutOrStdout(), cmd.ErrOrStderr(), noteIDs, force, dryRun)
	},
}

// runNoteDelete is the testable implementation of note delete.
func runNoteDelete(client Client, stdin io.Reader, stdout, stderr io.Writer, noteIDs []int64, force, dryRun bool) error {
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
		fmt.Fprint(stderr, "Continue? [y/N] ")

		reader := bufio.NewReader(stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return ErrCancelled
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return ErrCancelled
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

	noteDeleteCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
	noteDeleteCmd.Flags().Bool("dry-run", false, "show what would be deleted without deleting")

	noteCmd.AddCommand(noteCreateCmd)
	noteCmd.AddCommand(noteDeleteCmd)
	rootCmd.AddCommand(noteCmd)
}
