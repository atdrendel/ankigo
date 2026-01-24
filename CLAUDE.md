# ankigo Development Guidelines

This document defines conventions for developing the ankigo CLI. Follow these guidelines to ensure consistency with Unix CLI best practices.

## Core Philosophy

1. **Do one thing well** — Each command should have a single, clear purpose
2. **Composability** — Output should be pipeable; accept stdin where sensible
3. **Least surprise** — Behave like other Unix tools users already know
4. **Silence is golden** — Don't output unnecessary information on success
5. **Fail fast and loud** — Report errors immediately with clear messages
6. **Test-driven development** — Always write failing tests first, then implement

## Command Structure

Follow the pattern: `ankigo <resource> <action> [arguments] [flags]`

```bash
ankigo deck list
ankigo deck create "My Deck"
ankigo card add --deck "Default" --front "Q" --back "A"
ankigo card search "tag:japanese"
```

### Naming Rules

- Use **singular** resource names (following `gh` CLI convention): `deck`, `card`, not `decks`, `cards`
- Always require explicit actions: `deck list`, not `deck` defaulting to list
- Use lowercase for commands
- Use hyphens for multi-word commands: `get-stats`, not `getStats`
- Subcommands should be nouns (`deck`, `card`) or verbs (`sync`)

This approach ensures CRUD operations read naturally:
- `ankigo deck create` (not `decks create`)
- `ankigo deck delete` (not `decks delete`)
- `ankigo card add` (not `cards add`)

## Output Conventions

### Stdout vs Stderr

| Stream | Use for |
|--------|---------|
| stdout | Primary output (data, results) — must be parseable |
| stderr | Progress, warnings, errors, prompts — human-readable |

### Default Output

For `list` commands, output one item per line (plain text):
```
Default
Japanese::JLPT N3
```

This enables piping:
```bash
ankigo deck list | grep Japanese | xargs -I{} ankigo card search "deck:{}"
```

### Empty Results

When a list is empty, output a human-friendly message:
```
No decks found
```

### Future: Structured Output Flags

When implementing output format options, support:
- `--output json` / `-o json` — JSON array
- `--output jsonl` — JSON Lines (one object per line)
- `--quiet` / `-q` — IDs/names only, one per line

## Error Handling

### Error Message Format

```
ankigo: error: <what went wrong>
```

With context:
```
ankigo: error: failed to connect to Anki: connection refused
       hint: ensure Anki is running with anki-connect installed
```

### Error Output

- Errors go to stderr
- Include: what failed, why, and how to fix (when possible)
- Exit with non-zero code

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (API error, connection failed) |
| 2 | Misuse of command (bad flags, missing args) |

## Flags

### Universal Flags (root level)

| Flag | Short | Description |
|------|-------|-------------|
| `--help` | `-h` | Show help |
| `--version` | | Show version |
| `--verbose` | `-v` | Enable verbose output |

### Flag Conventions

- Long flags use `--`: `--verbose`, `--output`
- Short flags use `-`: `-v`, `-o`
- Boolean flags don't take values: `--verbose`, not `--verbose=true`
- Show defaults in help: `--timeout int  Request timeout (default: 30)`
- Mark required flags: `--deck string  Target deck (required)`

## CRUD Command Patterns

### `list`

```bash
ankigo deck list
ankigo card search "query"
```

- Output one item per line by default
- Support `--output` for format control (future)
- Support `--quiet` for IDs only (future)

### `create` / `add`

```bash
ankigo deck create "Deck Name"
ankigo card add --deck "Default" --front "Q" --back "A"
```

- On success, output the created resource identifier
- Support `--quiet` for scripts (output only ID)

### `delete` / `remove`

```bash
ankigo deck delete "Deck Name"
```

- **Require confirmation for destructive actions**
- Support `--force` / `-f` to skip confirmation (for scripting)
- Support `--dry-run` to show what would be deleted

## Testing Requirements

Use TDD. For each feature:

1. Write failing tests first
2. Implement to pass tests
3. Refactor if needed

### Test Coverage

- Unit tests for business logic
- Mock the anki-connect client for command tests
- Test error cases (connection refused, API errors, invalid input)
- Test empty results
- **Test misaligned data between related API calls** (see below)

### Test Patterns

```go
// Mock client for testing
type mockClient struct {
    decks []string
    err   error
}

func (m *mockClient) DeckNames() ([]string, error) {
    return m.decks, m.err
}

// Testable function signature
func runDeckList(client Client, out io.Writer) error
```

### Testing Related API Calls

When a command calls multiple APIs where one depends on another (e.g., fetching deck IDs then looking up stats by ID), **always test the case where the secondary lookup fails to find matching data**:

```go
// GOOD: Test when stats lookup returns no matching entries
mock := &mockClient{
    deckIDs:   map[string]int64{"Default": 1},
    deckStats: map[int64]DeckStats{},  // No stats for deck ID 1
}
// Verify: should handle gracefully (e.g., output zeros), not panic
```

This catches nil pointer issues and ensures the code handles incomplete API responses gracefully. Real-world APIs may return mismatched data due to timing, permissions, or data inconsistencies.

## anki-connect Integration

### Client Interface

Define interfaces for testability:

```go
type Client interface {
    DeckNames() ([]string, error)
    // Add methods as needed
}
```

### Error Wrapping

Wrap errors with context:

```go
return fmt.Errorf("failed to get deck names: %w", err)
```

### Default URL

Use `http://localhost:8765` as the default anki-connect URL.

## Code Organization

```
ankigo/
├── cmd/                      # Cobra commands
│   ├── root.go
│   ├── deck.go
│   ├── deck_test.go
│   ├── card.go
│   └── card_test.go
├── internal/
│   └── ankiconnect/          # anki-connect client
│       ├── client.go
│       └── client_test.go
└── main.go
```

## Checklist for New Commands

- [ ] `--help` works and is accurate
- [ ] Errors go to stderr with clear messages
- [ ] Returns correct exit codes (0 success, 1 error, 2 misuse)
- [ ] Works in pipes (non-TTY mode)
- [ ] Has unit tests with mock client
- [ ] Destructive actions require confirmation (unless `--force`)
