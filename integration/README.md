# Integration Tests

Manual integration tests for ankigo CLI against a real Anki instance.

## Prerequisites

1. **Anki desktop application** running
2. **anki-connect add-on** installed and enabled
3. **jq** installed (for JSON validation)
4. **Go toolchain** (to build ankigo)

## Running Tests

From the repository root:

```bash
./integration/run.sh
```

Or make it executable first:

```bash
chmod +x ./integration/run.sh
./integration/run.sh
```

## Safety

Tests use a unique prefix (`ANKIGO_TEST_<timestamp>_<pid>`) for all created decks.
**Only test-created decks are modified or deleted.** Your existing Anki data is never touched.

Cleanup runs automatically on exit (including Ctrl+C).

## What Gets Tested

- Version and help commands
- Deck listing (with various output formats)
- Deck creation (basic, nested, unicode)
- Card creation (basic, tags, unicode, cloze)
- Card creation validation (missing fields, non-existent deck/model)
- Duplicate card handling
- Card search (by deck, tag, with JSON output)
- Deck deletion (dry-run, force, by ID, multiple)
- Exit codes

## Adding Tests

1. Add test functions to `run.sh`
2. Use helpers from `lib/helpers.sh`
3. **Always prefix test data with `$TEST_PREFIX`**

## Troubleshooting

### "connection refused" errors

Ensure Anki is running and anki-connect is enabled. The default URL is `http://localhost:8765`.

### Tests fail to clean up

If tests are interrupted abnormally, you may have leftover `ANKIGO_TEST_*` decks.
Delete them manually or run:

```bash
./ankigo deck list | grep "^ANKIGO_TEST_" | xargs -I{} ./ankigo deck delete "{}" --force
```
