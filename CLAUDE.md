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

### Avoid Redundant Error Messages

When a command prints specific error messages (e.g., "Could not find X"), don't also return a generic error that gets printed by `main.go`. This creates noisy, redundant output:

```
# BAD: Redundant error message
Could not find deck-a
Could not find deck-b
Error: some decks were not found   ← Redundant! The user already knows.
```

```
# GOOD: Specific messages only, exit code signals failure
Could not find deck-a
Could not find deck-b
$ echo $?
1
```

Use `ErrSilent` when the command has already printed appropriate error messages:

```go
// cmd/errors.go defines sentinel errors:
// - ErrCancelled: user cancelled (e.g., answered "no" to confirmation)
// - ErrSilent: command failed but already printed specific error messages

// In your command:
for _, missing := range notFound {
    fmt.Fprintf(stderr, "Could not find %s\n", missing)
}
if len(notFound) > 0 {
    return ErrSilent  // Exit non-zero, but don't print another error
}
```

Before adding new error handling, check `cmd/errors.go` for existing patterns.

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

### Integration Tests

Integration tests live in `integration/` and run against a real Anki instance with anki-connect.

**When to update integration tests:**

- After completing a new feature, add tests covering the feature's happy path and key error cases
- After changing the behavior of an existing feature, update tests to reflect the new behavior
- After fixing a bug, consider adding a test that would have caught it

**CRITICAL SAFETY RULES:**

The integration tests run against the user's actual, production Anki database. You MUST follow these rules:

1. **NEVER read existing user data** — Don't search, list, or inspect decks/cards that weren't created by the test
2. **NEVER modify existing user data** — Don't update, move, or tag existing decks/cards
3. **NEVER delete existing user data** — Only delete decks/cards created by the test itself

**How safety is enforced:**

All test data uses a unique prefix: `ANKIGO_TEST_<timestamp>_<pid>`

```bash
TEST_PREFIX="ANKIGO_TEST_$(date +%s)_$$"

# GOOD: Create test deck with prefix
./ankigo deck create "${TEST_PREFIX}_MyTestDeck"

# GOOD: Search only in test deck
./ankigo card search "deck:${TEST_PREFIX}_MyTestDeck"

# GOOD: Delete only test decks
./ankigo deck delete "${TEST_PREFIX}_MyTestDeck" --force

# BAD: Never do these!
./ankigo deck list                    # Lists user's real decks
./ankigo card search "deck:Default"   # Searches user's real cards
./ankigo deck delete "Default"        # CATASTROPHIC: Deletes user data!
```

**Running integration tests:**

```bash
# Requires Anki running with anki-connect
./integration/run.sh
```

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
├── integration/              # Integration tests (run against real Anki)
│   ├── run.sh                # Main test runner
│   ├── lib/
│   │   └── helpers.sh        # Test utilities
│   └── README.md
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
- [ ] **Simulate the full user experience**: mentally run the command and read the complete output — check for redundant messages, unclear feedback, or missing information
- [ ] **Integration tests added** in `integration/run.sh` (using `$TEST_PREFIX` for all test data)
