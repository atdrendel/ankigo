# ankigo

A command-line interface for managing Anki flashcards via [anki-connect](https://foosoft.net/projects/anki-connect/).

## Installation

### From source

```bash
go install github.com/atdrendel/ankigo@latest
```

### Build locally

```bash
git clone https://github.com/atdrendel/ankigo.git
cd ankigo
go build .
```

### Build with version info

```bash
go build -ldflags "-X github.com/atdrendel/ankigo/internal/version.Version=1.0.0 \
  -X github.com/atdrendel/ankigo/internal/version.Commit=$(git rev-parse --short HEAD) \
  -X github.com/atdrendel/ankigo/internal/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" .
```

## Usage

```bash
# Show help
ankigo --help

# Show version
ankigo version
ankigo version --full

# Deck commands
ankigo deck list
ankigo deck create "My New Deck"

# Card commands
ankigo card add --front "Question" --back "Answer"
ankigo card add -f "Question" -b "Answer" -d "My Deck"
ankigo card search "tag:vocabulary"

# Note commands
ankigo note create --input-json '{
  "deckName": "Default",
  "modelName": "Basic",
  "fields": {"Front": "Question", "Back": "Answer"}
}'
ankigo note list "deck:Default"
ankigo note delete 1234567890

# Model commands
ankigo model create --input-json '{
  "modelName": "My Model",
  "fields": ["Front", "Back"],
  "templates": [{"name": "Card 1", "front": "{{Front}}", "back": "{{Back}}"}]
}'
ankigo model list
ankigo model prune

# Enable verbose output
ankigo --verbose deck list
```

### Agent Usage

Commands with `--input-json` support `--schema` to output the JSON Schema describing accepted input. This lets AI agents discover the expected format programmatically:

```bash
# Discover what note create accepts
ankigo note create --schema

# Discover what model create accepts
ankigo model create --schema
```

When invoked non-interactively (stdin is not a TTY), confirmation prompts are skipped automatically. Use `--force` for destructive actions in scripts and agent pipelines.

## Shell Completion

### Bash

```bash
# Add to ~/.bashrc
source <(ankigo completion bash)

# Or install permanently (Linux)
ankigo completion bash > /etc/bash_completion.d/ankigo

# Or install permanently (macOS with Homebrew)
ankigo completion bash > $(brew --prefix)/etc/bash_completion.d/ankigo
```

### Zsh

```bash
# Enable completion (add to ~/.zshrc if not already enabled)
autoload -U compinit; compinit

# Install completion
ankigo completion zsh > "${fpath[1]}/_ankigo"
```

### Fish

```bash
ankigo completion fish > ~/.config/fish/completions/ankigo.fish
```

### PowerShell

```powershell
ankigo completion powershell | Out-String | Invoke-Expression
```

## Development

### Prerequisites

- Go 1.21 or later
- Anki with [anki-connect](https://ankiweb.net/shared/info/2055492159) plugin installed

### Build

```bash
go build .
```

### Test

```bash
go test ./...
```

### Run

```bash
go run .
```

## Project Structure

```
ankigo/
├── main.go                 # Entry point
├── cmd/                    # CLI commands
│   ├── root.go             # Root command
│   ├── version.go          # Version subcommand
│   ├── deck.go             # Deck subcommands
│   ├── card.go             # Card subcommands
│   ├── note.go             # Note subcommands
│   ├── model.go            # Model subcommands
│   ├── confirm.go          # Confirmation prompts
│   ├── errors.go           # Sentinel errors
│   └── completion.go       # Shell completion
├── integration/            # Integration tests (run against real Anki)
│   ├── run.sh
│   └── lib/
│       └── helpers.sh
└── internal/
    ├── ankiconnect/        # anki-connect client
    └── version/            # Build-time version info
```

## License

MIT
