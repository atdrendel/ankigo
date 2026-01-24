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

# Enable verbose output
ankigo --verbose deck list
```

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
│   └── completion.go       # Shell completion
└── internal/
    └── version/            # Build-time version info
```

## License

MIT
