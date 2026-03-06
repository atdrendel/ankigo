# CLI Best Practices for Unix-Style Systems

This document defines conventions and best practices for building command-line interfaces. Follow these guidelines to create CLIs that are intuitive, composable, and consistent with Unix philosophy.

---

## Core Philosophy

1. **Do one thing well** — Each command should have a single, clear purpose
2. **Composability** — Output should be pipeable; accept stdin where sensible
3. **Least surprise** — Behave like other Unix tools users already know
4. **Silence is golden** — Don't output unnecessary information on success
5. **Fail fast and loud** — Report errors immediately with clear messages

---

## Command Structure

### Naming

- Use lowercase, short names (prefer `myapp` over `my-application`)
- Subcommands should be nouns or verbs: `myapp user create`, `myapp deploy`
- Avoid abbreviations unless universally understood (`ls`, `rm`, `cp` are okay; `crt` for "create" is not)
- Use hyphens for multi-word commands: `myapp get-config`, not `myapp getConfig`

### Hierarchy Pattern

```
myapp <resource> <action> [arguments] [flags]
```

Examples:
```bash
myapp user list
myapp user create --name "Alice"
myapp config get editor
myapp config set editor vim
```

For simple CLIs, flat structure is fine:
```bash
myapp init
myapp build
myapp deploy
```

### Default Commands

When a noun subcommand is invoked without an action, default to `list` if the resource is plural/listable:

```bash
myapp users          # equivalent to: myapp users list
myapp jobs           # equivalent to: myapp jobs list
```

For non-listable resources, show help:
```bash
myapp config         # shows help for config subcommands
```

---

## Standard Commands Every CLI Should Have

### `--help` / `-h`

Every command and subcommand must support `--help`. Output format:

```
<one-line description>

Usage:
  myapp <command> [flags]
  myapp <command> [subcommand]

Available Commands:
  init        Initialize a new project
  build       Build the project
  deploy      Deploy to production

Flags:
  -h, --help      Show this help message
  -v, --verbose   Enable verbose output
  -q, --quiet     Suppress non-error output
      --version   Show version information

Use "myapp <command> --help" for more information about a command.
```

Rules:
- First line is a brief description (no period, lowercase start unless proper noun)
- Group related flags together
- Show default values: `--timeout int   Request timeout in seconds (default: 30)`
- Show required flags clearly: `--name string   Project name (required)`
- List subcommands alphabetically
- Include examples for complex commands

### `--version` / `-V`

Output version and exit:

```
myapp 1.2.3
```

For verbose version info (optional `--version --verbose`):
```
myapp 1.2.3
  commit: a1b2c3d
  built:  2025-01-15T10:30:00Z
  go:     go1.23.0
```

Use semantic versioning. Exit code 0.

### `completion`

Provide shell completion scripts:

```bash
myapp completion bash
myapp completion zsh
myapp completion fish
myapp completion powershell
```

Output should be sourceable:
```bash
source <(myapp completion bash)
```

---

## The `list` Command

The `list` command is one of the most important commands. Get it right.

### Basic Behavior

```bash
myapp users list
myapp users        # same as above (list is default for plural nouns)
```

### Required Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output format: `table`, `json`, `yaml`, `csv`, `plain` |
| `--quiet` | `-q` | Output only IDs/names, one per line |
| `--no-header` | | Omit table header (useful for scripting) |

### Recommended Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--filter` | `-f` | Filter expression: `--filter "status=active"` |
| `--sort` | `-s` | Sort field and direction: `--sort "created:desc"` |
| `--limit` | `-n` | Maximum items to return |
| `--all` | `-a` | Include hidden/deleted/inactive items |
| `--columns` | | Columns to display: `--columns "id,name,status"` |

### Output Formats

**Table (default for TTY):**
```
ID       NAME      STATUS    CREATED
abc123   alice     active    2025-01-10
def456   bob       inactive  2025-01-08
```

**Plain (default for non-TTY/pipes):**
```
abc123  alice   active    2025-01-10
def456  bob     inactive  2025-01-08
```

**JSON:**
```json
[
  {"id": "abc123", "name": "alice", "status": "active", "created": "2025-01-10"},
  {"id": "def456", "name": "bob", "status": "inactive", "created": "2025-01-08"}
]
```

Output valid JSON to stdout. Use JSON Lines (`--output jsonl`) for streaming large datasets:
```
{"id": "abc123", "name": "alice", "status": "active"}
{"id": "def456", "name": "bob", "status": "inactive"}
```

**Quiet mode:**
```
abc123
def456
```

One identifier per line. Essential for piping:
```bash
myapp users list -q | xargs -I{} myapp user delete {}
```

### TTY Detection

Automatically adjust output based on whether stdout is a terminal:

| Condition | Behavior |
|-----------|----------|
| stdout is TTY | Colors, table formatting, progress indicators |
| stdout is pipe | Plain text, no colors, no progress spam |
| stderr is TTY | Show warnings/progress on stderr |

Allow override with `--color=always|never|auto` and `--no-progress`.

---

## Common CRUD Commands

### `create` / `new` / `add`

```bash
myapp user create --name "Alice" --email "alice@example.com"
myapp user create -f user.yaml    # from file
cat user.json | myapp user create # from stdin
```

Flags:
- Accept `--file` / `-f` for input from file
- Accept `-` to read from stdin: `myapp user create -f -`
- Use `--dry-run` to validate without creating
- Output the created resource ID (or full resource with `-o json`)

On success:
```
Created user abc123
```

Or with `--quiet`:
```
abc123
```

### `get` / `show` / `describe`

```bash
myapp user get abc123
myapp user get --name alice
```

- Accept ID as positional argument
- Support lookup by other unique fields via flags
- Output full resource details
- Support `--output` flag for format control

### `update` / `edit` / `set`

```bash
myapp user update abc123 --email "new@example.com"
myapp user edit abc123              # opens $EDITOR
myapp user update abc123 -f patch.yaml
```

- Show what changed (unless `--quiet`)
- Support `--dry-run`
- Support partial updates (PATCH semantics), not just full replacement

### `delete` / `remove` / `rm`

```bash
myapp user delete abc123
myapp user delete abc123 def456     # multiple
myapp users list -q | xargs myapp user delete  # from pipe
```

**Critical: Destructive actions require confirmation**

```bash
$ myapp user delete abc123
Delete user abc123 (alice)? [y/N]: y
Deleted user abc123
```

Flags:
- `--force` / `-f` — Skip confirmation (for scripting)
- `--dry-run` — Show what would be deleted
- `--quiet` — No output on success

Never require `--force` for non-destructive operations.

---

## Flags and Arguments

### Positional Arguments

- Use for required, obvious arguments: `myapp get <id>`
- Limit to 1-2 positional arguments maximum
- More complex input should use flags

### Flag Conventions

| Convention | Example |
|------------|---------|
| Long flags use `--` | `--verbose`, `--output` |
| Short flags use `-` | `-v`, `-o` |
| Boolean flags don't take values | `--verbose`, not `--verbose=true` |
| Negation with `--no-` prefix | `--no-cache`, `--no-color` |
| Values with `=` or space | `--output=json` or `--output json` |
| Combined short flags | `-xvf` means `-x -v -f` |

### Universal Flags

Every CLI should support these at the root level:

| Flag | Short | Description |
|------|-------|-------------|
| `--help` | `-h` | Show help |
| `--version` | `-V` | Show version (use capital V to avoid conflict with verbose) |
| `--verbose` | `-v` | Increase verbosity (can stack: `-vvv`) |
| `--quiet` | `-q` | Suppress non-error output |
| `--config` | `-c` | Config file path |
| `--color` | | Color output: `always`, `never`, `auto` |

### Flag Types

**Strings:**
```bash
--name "Alice"
--name="Alice"
```

**Integers:**
```bash
--count 10
--timeout 30s    # support duration strings where applicable
```

**Booleans:**
```bash
--verbose        # true
--no-verbose     # false (for flags that default to true)
--verbose=false  # explicit false
```

**Arrays (repeatable flags):**
```bash
--tag foo --tag bar --tag baz
--tag foo,bar,baz              # also support comma-separated
```

**Key-value pairs:**
```bash
--label env=prod --label team=backend
--env KEY=VALUE
```

---

## Output and Formatting

### Stdout vs Stderr

| Stream | Use for |
|--------|---------|
| stdout | Primary output (data, results) — must be parseable |
| stderr | Progress, warnings, errors, prompts — human-readable |

```bash
# Good: data to stdout, progress to stderr
$ myapp download large-file.zip
Downloading... 45% [=====>    ]    # stderr
# file contents go to stdout or file

# This allows:
$ myapp users list > users.json    # progress still visible
$ myapp users list 2>/dev/null     # suppress progress
```

### Structured Output

Always support machine-readable output:

```bash
myapp users list -o json    # full JSON
myapp users list -o jsonl   # JSON Lines (for streaming)
myapp users list -o csv     # CSV with header
myapp users list -o yaml    # YAML
myapp users list -q         # IDs only, newline-separated
```

JSON output must be valid and complete. Don't mix prose with JSON:

```bash
# Bad
Found 3 users:
[{"id": "abc"}, ...]

# Good
[{"id": "abc"}, ...]
```

### Tables

For human-readable table output:
- Align columns
- Truncate long values with `...`
- Show header by default, support `--no-header`
- Use consistent column widths

```
ID        NAME                 STATUS    CREATED
abc123    alice                active    2025-01-10
def456    bob-with-a-very...   inactive  2025-01-08
```

### Colors

Use color meaningfully:
- **Red** — Errors, failures, destructive actions
- **Yellow** — Warnings
- **Green** — Success, creation
- **Blue/Cyan** — Info, highlights
- **Dim/Gray** — Secondary information

Respect `NO_COLOR` environment variable and `--no-color` flag.

### Progress Indicators

For long operations:
- Show progress on stderr
- Use spinners for indeterminate progress
- Use progress bars for measurable progress
- Update in place (carriage return), don't spam lines
- Degrade gracefully when not a TTY (periodic updates or silence)

```bash
$ myapp deploy
Deploying... ━━━━━━━━━━━━━━━━━━━━ 100% (15s)
✓ Deployed successfully
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Misuse of command (bad flags, missing args) |
| 64-78 | BSD-standard codes (see sysexits.h) |
| 124 | Timeout |
| 125 | Command itself failed (not the operation) |
| 126 | Command found but not executable |
| 127 | Command not found |
| 128+N | Killed by signal N |
| 130 | Interrupted (Ctrl+C, SIGINT) |

Be consistent. Document non-zero exit codes in help.

---

## Error Handling

### Error Message Format

```
myapp: error: <what went wrong>
```

Or with context:
```
myapp: error: failed to create user: email already exists
```

Include:
- What operation failed
- Why it failed (specific reason)
- How to fix it (when possible)

```
myapp: error: config file not found: ~/.myapp/config.yaml
       hint: run 'myapp init' to create a default config
```

### Validation Errors

Validate all input before starting operations:

```bash
$ myapp user create
myapp: error: missing required flag: --name

$ myapp user create --name "" --email "invalid"
myapp: error: validation failed:
  --name: cannot be empty
  --email: invalid email format
```

### Warnings

Use stderr, prefix with "warning:":

```
myapp: warning: config file has insecure permissions
```

Don't exit non-zero for warnings alone.

---

## Configuration

### Precedence (highest to lowest)

1. Command-line flags
2. Environment variables
3. Project config file (`./.myapp.yaml`)
4. User config file (`~/.config/myapp/config.yaml`)
5. System config file (`/etc/myapp/config.yaml`)
6. Built-in defaults

### Environment Variables

- Prefix with app name: `MYAPP_`
- Map to flags: `--api-key` → `MYAPP_API_KEY`
- Document all environment variables in `--help`

```
Environment:
  MYAPP_API_KEY      API authentication key
  MYAPP_CONFIG       Path to config file
  MYAPP_NO_COLOR     Disable color output (set to any value)
```

Common standard variables to respect:
- `NO_COLOR` — Disable colors
- `FORCE_COLOR` — Force colors
- `DEBUG` — Enable debug mode
- `EDITOR` / `VISUAL` — Default editor
- `HOME` — User home directory
- `XDG_CONFIG_HOME` — Config directory (~/.config)
- `XDG_DATA_HOME` — Data directory (~/.local/share)
- `XDG_CACHE_HOME` — Cache directory (~/.cache)

### Config File Format

Prefer YAML or TOML for human-edited configs. Support JSON for machine-generated configs.

```yaml
# ~/.config/myapp/config.yaml
api_key: xxx
output: json
defaults:
  region: us-east-1
  timeout: 30s
```

Provide commands to manage config:

```bash
myapp config show                  # display current config
myapp config get <key>             # get specific value
myapp config set <key> <value>     # set value
myapp config edit                  # open in $EDITOR
myapp config path                  # show config file location
myapp config init                  # create default config
```

---

## Interactive Features

### Prompts

When requiring input:

```bash
$ myapp login
Username: alice
Password: ******
```

- Use readline for line editing where possible
- Hide password input
- Provide defaults in brackets: `Region [us-east-1]:`
- Support `--yes` / `-y` to auto-confirm all prompts

### Confirmation

For destructive actions:

```bash
$ myapp database drop production
This will permanently delete the 'production' database.
Type the database name to confirm: production
Dropped database 'production'
```

For less severe confirmations:
```bash
Delete 15 files? [y/N]: y
```

The capitalized option is the default. Press Enter to accept default.

### Interactive Mode

For complex CLIs, consider an interactive/REPL mode:

```bash
$ myapp --interactive
myapp> users list
...
myapp> user create --name test
...
myapp> exit
```

---

## Signals and Interrupts

Handle these signals gracefully:

| Signal | Action |
|--------|--------|
| SIGINT (Ctrl+C) | Cancel current operation, clean up, exit 130 |
| SIGTERM | Graceful shutdown |
| SIGPIPE | Stop writing, exit quietly |
| SIGHUP | Reload config (for daemons) |

On interrupt:
- Stop current operation
- Clean up temporary files
- Release locks
- Print brief message: `^C Interrupted`
- Exit with appropriate code

---

## Subcommand Patterns

### `init` / `setup`

```bash
myapp init
myapp init --template basic
myapp init ./my-project
```

- Create configuration files
- Set up directory structure
- Be idempotent (safe to run twice)
- Don't overwrite existing files without `--force`
- Show what was created

### `run` / `exec` / `start`

```bash
myapp run
myapp run --watch
myapp run --port 8080
```

- Run in foreground by default
- Support `--daemon` / `-d` for background
- Support `--watch` for file watching
- Show clear output about what's running

### `status` / `info`

```bash
myapp status
myapp info
```

Show current state:
```
myapp v1.2.3

Config:    ~/.config/myapp/config.yaml
Data:      ~/.local/share/myapp/
Logged in: alice (alice@example.com)
Project:   my-project (./myapp.yaml)

Services:
  api      running  http://localhost:8080
  worker   stopped
```

### `logs` / `log`

```bash
myapp logs
myapp logs --follow
myapp logs --since 1h
myapp logs --tail 100
myapp logs service-name
```

Standard flags:
- `--follow` / `-f` — Stream logs
- `--tail` / `-n` — Number of lines
- `--since` — Time filter (e.g., `1h`, `2024-01-01`)
- `--until` — End time
- `--timestamps` — Show timestamps
- `--no-color` — Disable coloring

### `login` / `logout` / `auth`

```bash
myapp login
myapp login --token xxx
myapp logout
myapp auth status
```

- Support interactive browser-based OAuth
- Support `--token` for CI/scripting
- Store credentials securely (system keychain when possible)
- Show login status with `auth status`

### `doctor` / `check`

```bash
myapp doctor
```

Verify installation and dependencies:
```
Checking myapp installation...
✓ Config file exists
✓ Credentials valid
✓ API reachable
✗ Missing dependency: ffmpeg
  hint: install with 'brew install ffmpeg'

Found 1 issue
```

---

## Documentation

### In-app Help

- Every command has `--help`
- Help is generated from code (single source of truth)
- Include examples in help for complex commands

### Man Pages

For widely-distributed CLIs, provide man pages:
- `man myapp` — Overview
- `man myapp-<subcommand>` — Per-command docs

### Examples

Provide an `examples` subcommand or section in help:

```bash
myapp examples
myapp user create --help

Examples:
  # Create a user interactively
  myapp user create

  # Create a user with all fields
  myapp user create --name alice --email alice@example.com --role admin

  # Create multiple users from a file
  myapp user create -f users.yaml

  # Pipe from another command
  echo '{"name": "alice"}' | myapp user create -f -
```

---

## Testing Your CLI

Ensure your CLI works correctly:

1. **Unit tests** — Test individual functions
2. **Integration tests** — Test full commands
3. **Golden tests** — Compare output against expected files
4. **TTY tests** — Test both TTY and non-TTY behavior

Test these scenarios:
- Help output for all commands
- Version output
- Missing required arguments
- Invalid arguments
- Stdin input
- Piped output (non-TTY)
- Exit codes
- Signal handling (Ctrl+C)
- Very long inputs
- Unicode handling
- Empty results

---

## Distribution

### Single Binary

Distribute as a single static binary when possible. No runtime dependencies.

### Naming

- Binary name: lowercase, no extension
- Archive: `myapp-v1.2.3-linux-amd64.tar.gz`
- Supported patterns: `{os}-{arch}`, e.g., `darwin-arm64`, `linux-amd64`, `windows-amd64`

### Installation Methods

Support multiple installation methods:

```bash
# Homebrew (macOS/Linux)
brew install myapp

# apt (Debian/Ubuntu)
apt install myapp

# Direct download
curl -fsSL https://example.com/install.sh | sh

# Go
go install example.com/myapp@latest

# npm (if applicable)
npm install -g myapp
```

### Update Mechanism

```bash
myapp update              # self-update
myapp update --check      # check for updates only
myapp version --check     # show current and latest version
```

---

## AI Agent Compatibility

AI agents are increasingly the primary consumers of CLIs. Human DX optimizes for discoverability and forgiveness; Agent DX optimizes for predictability and defense-in-depth. A well-designed CLI serves both.

### Raw JSON Input

Agents prefer structured input over bespoke flags. A flag like `--title "My Doc"` can't express nested structures without custom abstractions. Support a raw JSON payload path alongside convenience flags:

```bash
# Human-friendly flags
myapp user create --name "Alice" --email "alice@example.com"

# Agent-friendly JSON (maps directly to the API schema)
myapp user create --input-json '{"name": "Alice", "email": "alice@example.com", "roles": ["admin"]}'
```

The `--input-json` (or `--json`) flag should accept the full API payload. This eliminates translation loss between what the agent generates and what the API accepts.

### Schema Introspection

Agents can't google the docs without blowing up their token budget. Make the CLI self-describing at runtime:

```bash
myapp user create --schema
# Outputs JSON Schema describing what --input-json accepts
```

This lets agents discover what a command accepts programmatically — no pre-stuffed documentation needed. Use standard JSON Schema so agents can validate input before sending it.

### Input Hardening Against Hallucinations

Humans typo. Agents hallucinate. The failure modes are different:

| Input type | Human mistake | Agent mistake |
|-----------|---------------|---------------|
| File paths | Misspelling | Path traversal (`../../.ssh`) |
| Control chars | Copy-paste garbage | Invisible characters in output |
| Resource IDs | Misspelled ID | Embedded query params (`id?fields=name`) |
| URL encoding | Rarely pre-encode | Double-encoding (`%2e%2e` for `..`) |

Defend at the CLI boundary:
- **Canonicalize and sandbox file paths** to the working directory
- **Reject control characters** below ASCII 0x20 in string inputs
- **Reject `?`, `#`, and `%`** in resource identifiers
- **Validate before sending** — the CLI is the last line of defense

### Context Window Discipline

API responses can consume a meaningful fraction of an agent's context window. Agents pay per token and lose reasoning capacity with every irrelevant field.

- **Support field masks** (`--fields "id,name,status"`) to limit what's returned
- **Support NDJSON** (`--output jsonl`) for streaming large result sets without buffering a top-level array
- **Default to minimal output** when stdout is not a TTY

### Agent Auth

Agents can't do browser-based OAuth. Support headless authentication:

- **Environment variables** for tokens: `MYAPP_TOKEN`, `MYAPP_CREDENTIALS_FILE`
- **Service accounts** where possible
- Avoid flows that require a browser redirect

### Agent Context Files

Agents learn through context injected at conversation start, not `--help` and Stack Overflow. Ship machine-readable guidance alongside the CLI:

- **`CONTEXT.md`** — Agent-specific guidance: which flags to always use, what to confirm before executing, common pitfalls
- **Skill files** — Structured Markdown (with YAML frontmatter) describing workflows, one per API surface or task

These encode invariants that agents can't intuit from `--help` alone, such as "always use `--dry-run` for mutating operations" or "always add `--fields` to list calls."

### MCP (Model Context Protocol)

If the CLI wraps a structured API, consider exposing it as typed JSON-RPC tools over stdio via MCP. This eliminates shell escaping, argument parsing ambiguity, and output parsing — the agent calls a typed function instead of constructing a command string.

### Non-TTY Behavior

When stdout is not a terminal:
- **Skip interactive prompts** — don't hang waiting for input that will never come
- **Suppress colors and progress indicators** — they're noise for machines
- **Default to machine-readable output** — plain text or JSON instead of tables

When stdin is not a terminal:
- **Decline confirmations by default** — require `--force` for destructive actions
- **Accept piped input** where sensible

## Security Considerations

- Never log secrets or tokens
- Mask sensitive values in debug output
- Use secure credential storage (keychain/keyring)
- Validate all input (assume hostile — whether from humans or AI agents)
- Be careful with shell expansion in generated commands
- Warn about insecure file permissions on config files
- Support `--dry-run` for destructive operations
- **Response sanitization** — Consider that API responses may contain prompt injection attempts (e.g., a malicious email body saying "Ignore previous instructions..."). If your CLI's output will be consumed by an AI agent, sanitize or flag suspicious content in API responses before returning them

---

## Quick Reference

### Command Checklist

For every command, verify:

- [ ] `--help` works and is accurate
- [ ] Supports `--output` / `-o` for machine-readable output (where applicable)
- [ ] Supports `--quiet` / `-q` for minimal output
- [ ] Returns correct exit codes
- [ ] Errors go to stderr with clear messages
- [ ] Works in pipes (non-TTY mode)
- [ ] Handles interrupts gracefully
- [ ] Destructive actions require confirmation (unless `--force`)
- [ ] Works in non-TTY mode (no hanging prompts, machine-readable output)
- [ ] Complex create commands support `--input-json` for raw payloads
- [ ] Commands with `--input-json` support `--schema` for introspection

### Flag Checklist

- [ ] Long form: `--flag`
- [ ] Short form for common flags: `-f`
- [ ] Boolean flags don't require values
- [ ] Default values documented in help
- [ ] Required flags marked in help
- [ ] Environment variable alternatives documented

### Output Checklist

- [ ] Data goes to stdout
- [ ] Progress/errors go to stderr
- [ ] No color when piped (unless `--color=always`)
- [ ] JSON output is valid and complete
- [ ] Tables are readable and aligned
- [ ] Quiet mode outputs only essential identifiers

---

## Summary of Defaults

| Aspect | Default |
|--------|---------|
| `list` command for plural nouns | Yes, `myapp users` = `myapp users list` |
| Output format | `table` for TTY, `plain` for pipe |
| Colors | Enabled for TTY, disabled for pipe |
| Confirmation for destructive actions | Yes, skip with `--force` |
| Help flag | `-h`, `--help` |
| Version flag | `-V`, `--version` |
| Quiet flag | `-q`, `--quiet` |
| Verbose flag | `-v`, `--verbose` |
| Output format flag | `-o`, `--output` |
| Config file flag | `-c`, `--config` |

---

*This document follows conventions from POSIX, GNU, and modern CLI tools like `kubectl`, `docker`, `gh`, `aws`, `git`, and `cargo`.*