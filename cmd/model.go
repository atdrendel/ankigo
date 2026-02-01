package cmd

import (
	"bufio"
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
    --template "Cloze,{{cloze:Text}},{{Text}}<br>{{Extra}}"

  # With custom CSS
  ankigo model create "Styled" --field Q --field A \
    --template "Card 1,{{Q}},{{A}}" --css ".card { font-size: 24px; }"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := ankiconnect.DefaultClient()
		fieldFlags, _ := cmd.Flags().GetStringArray("field")
		templateFlags, _ := cmd.Flags().GetStringArray("template")
		css, _ := cmd.Flags().GetString("css")
		cssFile, _ := cmd.Flags().GetString("css-file")
		isCloze, _ := cmd.Flags().GetBool("cloze")

		opts := modelCreateOptions{
			fields:    fieldFlags,
			templates: templateFlags,
			css:       css,
			cssFile:   cssFile,
			isCloze:   isCloze,
		}

		return runModelCreate(client, cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0], opts)
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
	Use:   "prune [model-names...]",
	Short: "Remove empty note types (models)",
	Long: `Remove note types (models) that have no notes using them.

If model names are provided, only those models are pruned (if empty).
If no names are provided, all empty models are pruned.

Requires confirmation or --force flag.`,
	Example: `  # Remove all empty models (with confirmation)
  ankigo model prune

  # Remove without confirmation
  ankigo model prune --force

  # Preview what would be removed
  ankigo model prune --dry-run

  # Remove specific empty models
  ankigo model prune "Unused Model 1" "Unused Model 2" --force`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Silence usage for errors that happen during execution (not arg validation)
		cmd.SilenceUsage = true

		client := ankiconnect.DefaultClient()
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		opts := modelPruneOptions{force: force, dryRun: dryRun}
		return runModelPrune(client, os.Stdin, cmd.OutOrStdout(), cmd.ErrOrStderr(), args, opts)
	},
}

// runModelPrune is the testable implementation of model prune.
func runModelPrune(client Client, stdin io.Reader, stdout, stderr io.Writer, names []string, opts modelPruneOptions) error {
	// Get all model names
	allModels, err := client.ModelNames()
	if err != nil {
		return fmt.Errorf("failed to get model names: %w", err)
	}

	// Build set of existing models
	existingModels := make(map[string]bool)
	for _, name := range allModels {
		existingModels[name] = true
	}

	// Determine which models to check
	var modelsToCheck []string
	if len(names) > 0 {
		// Validate requested models exist
		for _, name := range names {
			if !existingModels[name] {
				return fmt.Errorf("model not found: %s", name)
			}
		}
		modelsToCheck = names
	} else {
		modelsToCheck = allModels
	}

	// Find empty models (models with no notes)
	var emptyModels []string
	for _, name := range modelsToCheck {
		// Query for notes using this model
		noteIDs, err := client.FindNotes(fmt.Sprintf("note:%q", name))
		if err != nil {
			return fmt.Errorf("failed to check notes for %s: %w", name, err)
		}

		if len(noteIDs) == 0 {
			emptyModels = append(emptyModels, name)
		} else if len(names) > 0 {
			// Only report skipped if specific models were requested
			fmt.Fprintf(stderr, "Skipped %s: has %d notes\n", name, len(noteIDs))
		}
	}

	// Check if there's anything to do
	if len(emptyModels) == 0 {
		if len(names) == 0 {
			fmt.Fprintln(stderr, "No empty models found")
		}
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

	// model prune flags
	modelPruneCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	modelPruneCmd.Flags().Bool("dry-run", false, "Show what would be removed without executing")

	// Register commands
	modelCmd.AddCommand(modelListCmd)
	modelCmd.AddCommand(modelCreateCmd)
	modelCmd.AddCommand(modelPruneCmd)
	rootCmd.AddCommand(modelCmd)
}
