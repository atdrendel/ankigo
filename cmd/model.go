package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/atdrendel/ankigo/internal/ankiconnect"
	"github.com/spf13/cobra"
)

// modelListFields are the available fields for model list output.
var modelListFields = []string{"name", "id", "fields"}

// modelInfoFields are fields that require additional API calls beyond just names.
var modelInfoFields = []string{"id", "fields"}

// modelListOptions holds the options for the model list command.
type modelListOptions struct {
	json   bool
	fields []string
}

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage Anki note types (models)",
	Long:  `Commands for listing note types (models) in your Anki collection.`,
}

var modelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all note types",
	Long:  `List all note types (models) in your Anki collection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := ankiconnect.DefaultClient()
		jsonFlag, _ := cmd.Flags().GetBool("json")
		fieldsStr, _ := cmd.Flags().GetString("fields")

		fields, err := parseFields(fieldsStr, modelListFields)
		if err != nil {
			return err
		}

		opts := modelListOptions{json: jsonFlag, fields: fields}
		return runModelList(client, cmd.OutOrStdout(), opts)
	},
}

// needsModelInfo returns true if any of the fields require additional API calls.
func needsModelInfo(fields []string) bool {
	for _, f := range fields {
		if contains(modelInfoFields, f) {
			return true
		}
	}
	return false
}

// modelEntry holds model data for output.
type modelEntry struct {
	name   string
	id     int64
	fields []string
}

// runModelList is the testable implementation of model list.
func runModelList(client Client, out io.Writer, opts modelListOptions) error {
	// Validate fields first
	if opts.fields != nil {
		for _, f := range opts.fields {
			if !contains(modelListFields, f) {
				return fmt.Errorf("unknown field: %s", f)
			}
		}
	}

	if opts.json {
		return runModelListJSON(client, out, opts.fields)
	}
	return runModelListText(client, out, opts.fields)
}

func runModelListText(client Client, out io.Writer, fields []string) error {
	// If no fields specified, default to ["name"] for backwards compatibility
	if fields == nil {
		fields = []string{"name"}
	}

	wantsID := contains(fields, "id")
	wantsFields := contains(fields, "fields")

	var entries []modelEntry

	if wantsID {
		// Need IDs
		modelMap, err := client.ModelNamesAndIds()
		if err != nil {
			return fmt.Errorf("failed to get models: %w", err)
		}

		if len(modelMap) == 0 {
			fmt.Fprintln(out, "No models found")
			return nil
		}

		for name, id := range modelMap {
			entries = append(entries, modelEntry{name: name, id: id})
		}
	} else {
		// Only need names
		models, err := client.ModelNames()
		if err != nil {
			return fmt.Errorf("failed to get model names: %w", err)
		}

		if len(models) == 0 {
			fmt.Fprintln(out, "No models found")
			return nil
		}

		for _, name := range models {
			entries = append(entries, modelEntry{name: name})
		}
	}

	// Sort by name for consistent output
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	// Fetch field names if needed
	if wantsFields {
		for i := range entries {
			fieldNames, err := client.ModelFieldNames(entries[i].name)
			if err != nil {
				return fmt.Errorf("failed to get field names for %s: %w", entries[i].name, err)
			}
			entries[i].fields = fieldNames
		}
	}

	// Output
	for _, e := range entries {
		var vals []string
		for _, f := range fields {
			switch f {
			case "name":
				vals = append(vals, e.name)
			case "id":
				vals = append(vals, fmt.Sprintf("%d", e.id))
			case "fields":
				vals = append(vals, strings.Join(e.fields, ","))
			}
		}
		fmt.Fprintln(out, strings.Join(vals, "\t"))
	}
	return nil
}

func runModelListJSON(client Client, out io.Writer, fields []string) error {
	// If no fields specified, default to all fields
	if fields == nil {
		fields = modelListFields
	}

	wantsID := contains(fields, "id")
	wantsFields := contains(fields, "fields")

	var entries []modelEntry

	// For JSON, always need IDs if id field is requested
	if wantsID {
		modelMap, err := client.ModelNamesAndIds()
		if err != nil {
			return fmt.Errorf("failed to get models: %w", err)
		}

		for name, id := range modelMap {
			entries = append(entries, modelEntry{name: name, id: id})
		}
	} else {
		// Only need names
		models, err := client.ModelNames()
		if err != nil {
			return fmt.Errorf("failed to get model names: %w", err)
		}

		for _, name := range models {
			entries = append(entries, modelEntry{name: name})
		}
	}

	// Sort by name for consistent output
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	// Fetch field names if needed
	if wantsFields {
		for i := range entries {
			fieldNames, err := client.ModelFieldNames(entries[i].name)
			if err != nil {
				return fmt.Errorf("failed to get field names for %s: %w", entries[i].name, err)
			}
			entries[i].fields = fieldNames
		}
	}

	// Build output based on requested fields
	var result []map[string]interface{}
	for _, e := range entries {
		obj := make(map[string]interface{})
		for _, f := range fields {
			switch f {
			case "name":
				obj["name"] = e.name
			case "id":
				obj["id"] = e.id
			case "fields":
				obj["fields"] = e.fields
			}
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

// =============================================================================
// Model Create
// =============================================================================

// modelCreateOptions holds options for model create.
type modelCreateOptions struct {
	fields    []string
	templates []string
	css       string
	cssFile   string
	isCloze   bool
	inputJSON string
	schema    bool
}

// modelCreateSchemaJSON is the JSON Schema for --input-json on model create.
const modelCreateSchemaJSON = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "description": "Input schema for ankigo model create --input-json",
  "type": "object",
  "required": ["modelName", "fields", "templates"],
  "properties": {
    "modelName": { "type": "string", "description": "Unique name for the note type" },
    "fields": {
      "type": "array",
      "items": { "type": "string" },
      "minItems": 1,
      "description": "Field names in order"
    },
    "templates": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["front", "back"],
        "properties": {
          "name": { "type": "string", "description": "Template name (auto-generated if omitted)" },
          "front": { "type": "string", "description": "Front template HTML (use {{FieldName}} for substitution)" },
          "back": { "type": "string", "description": "Back template HTML" }
        }
      },
      "minItems": 1,
      "description": "Card templates"
    },
    "css": { "type": "string", "description": "Custom CSS styling" },
    "isCloze": { "type": "boolean", "description": "Create as cloze-deletion model" }
  }
}
`

// modelCreateInput is the JSON input structure for model create.
type modelCreateInput struct {
	ModelName string           `json:"modelName"`
	Fields    []string         `json:"fields"`
	Templates []templateInput  `json:"templates"`
	CSS       string           `json:"css,omitempty"`
	IsCloze   bool             `json:"isCloze,omitempty"`
}

// templateInput is the JSON input structure for a card template.
type templateInput struct {
	Name  string `json:"name"`
	Front string `json:"front"`
	Back  string `json:"back"`
}

var modelCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new note type (model)",
	Long:  `Create a new note type (model) with custom fields and card templates.`,
	Example: `  # Basic model
  ankigo model create "Vocabulary" --field Front --field Back \
    --template "Card 1,{{Front}},{{Back}}"

  # Model with multiple templates (reversed cards)
  ankigo model create "Bidirectional" --field Front --field Back \
    --template "Forward,{{Front}},{{Back}}" \
    --template "Reverse,{{Back}},{{Front}}"

  # Cloze model
  ankigo model create "My Cloze" --field Text --field Extra --cloze \
    --template "Cloze,{{cloze:Text}},{{cloze:Text}}<br>{{Extra}}"

  # With custom CSS
  ankigo model create "Styled" --field Q --field A \
    --template "Card 1,{{Q}},{{A}}" --css ".card { font-size: 24px; }"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		schema, _ := cmd.Flags().GetBool("schema")
		if schema {
			return runModelCreate(nil, cmd.OutOrStdout(), cmd.ErrOrStderr(), "", modelCreateOptions{schema: true})
		}

		client := ankiconnect.DefaultClient()
		inputJSON, _ := cmd.Flags().GetString("input-json")
		fieldFlags, _ := cmd.Flags().GetStringArray("field")
		templateFlags, _ := cmd.Flags().GetStringArray("template")
		css, _ := cmd.Flags().GetString("css")
		cssFile, _ := cmd.Flags().GetString("css-file")
		isCloze, _ := cmd.Flags().GetBool("cloze")

		// Require either positional arg or --input-json
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		if name == "" && inputJSON == "" {
			return fmt.Errorf("model name is required (as argument or in --input-json)")
		}

		opts := modelCreateOptions{
			fields:    fieldFlags,
			templates: templateFlags,
			css:       css,
			cssFile:   cssFile,
			isCloze:   isCloze,
			inputJSON: inputJSON,
		}

		return runModelCreate(client, cmd.OutOrStdout(), cmd.ErrOrStderr(), name, opts)
	},
}

// parseTemplateSpec parses a template specification string into a CardTemplate.
// Format: "Name,FrontHTML,BackHTML"
func parseTemplateSpec(spec string) (ankiconnect.CardTemplate, error) {
	parts := strings.SplitN(spec, ",", 3)
	if len(parts) != 3 {
		return ankiconnect.CardTemplate{}, fmt.Errorf("invalid template format %q: expected Name,Front,Back", spec)
	}

	return ankiconnect.CardTemplate{
		Name:  strings.TrimSpace(parts[0]),
		Front: strings.TrimSpace(parts[1]),
		Back:  strings.TrimSpace(parts[2]),
	}, nil
}

// runModelCreate is the testable implementation of model create.
func runModelCreate(client Client, stdout, stderr io.Writer, name string, opts modelCreateOptions) error {
	// Handle --schema: output JSON Schema and return
	if opts.schema {
		fmt.Fprint(stdout, modelCreateSchemaJSON)
		return nil
	}

	// Handle --input-json
	if opts.inputJSON != "" {
		var input modelCreateInput
		if err := json.Unmarshal([]byte(opts.inputJSON), &input); err != nil {
			return fmt.Errorf("invalid JSON input: %w", err)
		}

		// Use name from JSON if not provided as positional arg
		if name == "" {
			name = input.ModelName
		}

		// Convert templates
		var cardTemplates []ankiconnect.CardTemplate
		for _, t := range input.Templates {
			cardTemplates = append(cardTemplates, ankiconnect.CardTemplate{
				Name:  t.Name,
				Front: t.Front,
				Back:  t.Back,
			})
		}

		params := ankiconnect.CreateModelParams{
			ModelName:     name,
			Fields:        input.Fields,
			CardTemplates: cardTemplates,
			CSS:           input.CSS,
			IsCloze:       input.IsCloze,
		}

		result, err := client.CreateModel(params)
		if err != nil {
			return fmt.Errorf("failed to create model: %w", err)
		}
		if id, ok := result["id"]; ok {
			fmt.Fprintln(stdout, id)
		}
		return nil
	}

	// Validate name
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	// Validate fields
	if len(opts.fields) == 0 {
		return fmt.Errorf("at least one --field is required")
	}

	// Validate templates
	if len(opts.templates) == 0 {
		return fmt.Errorf("at least one --template is required")
	}

	// Validate mutually exclusive CSS flags
	if opts.css != "" && opts.cssFile != "" {
		return fmt.Errorf("--css and --css-file are mutually exclusive")
	}

	// Parse templates
	var cardTemplates []ankiconnect.CardTemplate
	for _, spec := range opts.templates {
		tmpl, err := parseTemplateSpec(spec)
		if err != nil {
			return err
		}
		cardTemplates = append(cardTemplates, tmpl)
	}

	// Read CSS from file if specified
	css := opts.css
	if opts.cssFile != "" {
		data, err := os.ReadFile(opts.cssFile)
		if err != nil {
			return fmt.Errorf("failed to read CSS file: %w", err)
		}
		css = string(data)
	}

	// Build params
	params := ankiconnect.CreateModelParams{
		ModelName:     name,
		Fields:        opts.fields,
		CardTemplates: cardTemplates,
		CSS:           css,
		IsCloze:       opts.isCloze,
	}

	// Create the model
	result, err := client.CreateModel(params)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Output the model ID (for scripts)
	if id, ok := result["id"]; ok {
		fmt.Fprintln(stdout, id)
	}

	return nil
}

// =============================================================================
// Model Prune
// =============================================================================

// modelPruneOptions holds options for model prune.
type modelPruneOptions struct {
	force  bool
	dryRun bool
}

var modelPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove empty note types (models)",
	Long: `Remove note types (models) that have no notes using them.

All empty models are pruned. Requires confirmation or --force flag.`,
	Example: `  # Remove all empty models (with confirmation)
  ankigo model prune

  # Remove without confirmation
  ankigo model prune --force

  # Preview what would be removed
  ankigo model prune --dry-run`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Silence usage for errors that happen during execution (not arg validation)
		cmd.SilenceUsage = true

		client := ankiconnect.DefaultClient()
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		opts := modelPruneOptions{force: force, dryRun: dryRun}
		return runModelPrune(client, os.Stdin, cmd.OutOrStdout(), cmd.ErrOrStderr(), opts, isStdinTerminal)
	},
}

// runModelPrune is the testable implementation of model prune.
func runModelPrune(client Client, stdin io.Reader, stdout, stderr io.Writer, opts modelPruneOptions, isTerminal func() bool) error {
	// Get all model names
	allModels, err := client.ModelNames()
	if err != nil {
		return fmt.Errorf("failed to get model names: %w", err)
	}

	// Find empty models (models with no notes)
	var emptyModels []string
	for _, name := range allModels {
		// Query for notes using this model
		noteIDs, err := client.FindNotes(fmt.Sprintf("note:%q", name))
		if err != nil {
			return fmt.Errorf("failed to check notes for %s: %w", name, err)
		}

		if len(noteIDs) == 0 {
			emptyModels = append(emptyModels, name)
		}
	}

	// Check if there's anything to do
	if len(emptyModels) == 0 {
		fmt.Fprintln(stderr, "No empty models found")
		return nil
	}

	// Dry run - just report what would be removed
	if opts.dryRun {
		fmt.Fprintln(stderr, "Would remove the following empty models:")
		for _, name := range emptyModels {
			fmt.Fprintln(stdout, name)
		}
		return nil
	}

	// Confirmation prompt (unless --force)
	if !opts.force {
		fmt.Fprintln(stderr, "The following empty models will be removed:")
		for _, name := range emptyModels {
			fmt.Fprintf(stderr, "  - %s\n", name)
		}
		if err := requireConfirmation(stdin, stderr, isTerminal); err != nil {
			return err
		}
	}

	// Remove empty models
	if err := client.RemoveEmptyNotes(); err != nil {
		return fmt.Errorf("failed to remove empty models: %w", err)
	}

	// Report what was removed
	for _, name := range emptyModels {
		fmt.Fprintf(stderr, "Removed %s\n", name)
	}

	return nil
}

func init() {
	// model list flags
	modelListCmd.Flags().Bool("json", false, "Output in JSON format")
	modelListCmd.Flags().StringP("fields", "f", "", "Comma-separated list of fields (available: name, id, fields)")

	// model create flags
	modelCreateCmd.Flags().StringArrayP("field", "f", nil, "Field name (can be repeated)")
	modelCreateCmd.Flags().StringArrayP("template", "t", nil, "Template spec: Name,FrontHTML,BackHTML (can be repeated)")
	modelCreateCmd.Flags().String("css", "", "Custom CSS styling")
	modelCreateCmd.Flags().String("css-file", "", "Read CSS from file")
	modelCreateCmd.Flags().Bool("cloze", false, "Create a Cloze-type model")
	modelCreateCmd.Flags().String("input-json", "", "Create model from JSON (provides all parameters as structured input)")
	modelCreateCmd.Flags().Bool("schema", false, "Print JSON Schema for --input-json and exit")

	// model prune flags
	modelPruneCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	modelPruneCmd.Flags().Bool("dry-run", false, "Show what would be removed without executing")

	// Register commands
	modelCmd.AddCommand(modelListCmd)
	modelCmd.AddCommand(modelCreateCmd)
	modelCmd.AddCommand(modelPruneCmd)
	rootCmd.AddCommand(modelCmd)
}
