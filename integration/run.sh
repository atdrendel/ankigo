#!/usr/bin/env bash
# Integration tests for ankigo CLI
# Requires: Anki running with anki-connect, jq installed

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

# Source helpers
source "$SCRIPT_DIR/lib/helpers.sh"

# Generate unique test prefix to avoid conflicts with user data
TEST_PREFIX="ANKIGO_TEST_$(date +%s)_$$"
export TEST_PREFIX

echo "=== ankigo Integration Tests ==="
echo "Test prefix: $TEST_PREFIX"
echo ""

# Cleanup on exit (even on error or Ctrl+C)
cleanup() {
    echo ""
    echo "[Cleanup]"
    local decks_to_delete
    decks_to_delete=$(./ankigo deck list 2>/dev/null | grep "^${TEST_PREFIX}" || true)
    if [[ -n "$decks_to_delete" ]]; then
        echo "$decks_to_delete" | while read -r deck; do
            ./ankigo deck delete "$deck" --force 2>/dev/null || true
            echo "  Deleted: $deck"
        done
    else
        echo "  No test decks to clean up"
    fi
    echo -e "${GREEN}✓${NC} cleanup complete"
}
trap cleanup EXIT

# =============================================================================
# Test Functions
# =============================================================================

test_prerequisites() {
    # Binary exists
    run_test "ankigo binary exists" test -x ./ankigo

    # Can connect to Anki
    run_test "anki-connect responding" assert_success ./ankigo deck list
}

test_version_help() {
    local version_output
    version_output=$(./ankigo version)
    run_test "version command returns output" assert_not_empty "$version_output"

    local full_version
    full_version=$(./ankigo version --full)
    run_test "version --full contains 'version:'" assert_contains "$full_version" "version:"

    run_test "root --help succeeds" assert_success ./ankigo --help

    run_test "deck --help succeeds" assert_success ./ankigo deck --help

    run_test "card --help succeeds" assert_success ./ankigo card --help
}

test_deck_list() {
    run_test "deck list succeeds" assert_success ./ankigo deck list

    local json_output
    json_output=$(./ankigo deck list --json)
    run_test "deck list --json is valid JSON array" assert_json_array "$json_output"

    run_test "deck list --fields id,name succeeds" \
        assert_success ./ankigo deck list --fields id,name

    run_test "deck list --fields invalid fails" \
        assert_failure ./ankigo deck list --fields invalid
}

test_deck_create() {
    local deck_name="${TEST_PREFIX}_Basic"
    local deck_id

    # Create deck and capture ID
    deck_id=$(./ankigo deck create "$deck_name")
    run_test "deck create returns numeric ID" assert_numeric "$deck_id"

    # Verify deck appears in list
    local deck_list
    deck_list=$(./ankigo deck list)
    run_test "created deck appears in list" assert_contains "$deck_list" "$deck_name"

    # Idempotent: same ID returned
    local deck_id2
    deck_id2=$(./ankigo deck create "$deck_name")
    run_test "deck create is idempotent (same ID)" test "$deck_id" = "$deck_id2"

    # Nested deck
    local nested_deck="${TEST_PREFIX}_Parent::Child"
    run_test "create nested deck" assert_success ./ankigo deck create "$nested_deck"

    # Unicode deck
    local unicode_deck="${TEST_PREFIX}_日本語"
    run_test "create unicode deck" assert_success ./ankigo deck create "$unicode_deck"
}

test_card_create_basic() {
    local deck_name="${TEST_PREFIX}_CardTest"
    ./ankigo deck create "$deck_name" >/dev/null

    local note_id
    note_id=$(./ankigo card create -d "$deck_name" -f "Test front" -b "Test back")
    run_test "card create returns numeric ID" assert_numeric "$note_id"

    # Verify card appears in search
    local search_result
    search_result=$(./ankigo card search "deck:$deck_name")
    run_test "created card appears in search" assert_not_empty "$search_result"

    # Card with tags
    run_test "card create with tags" \
        assert_success ./ankigo card create -d "$deck_name" \
            -f "Tagged Q" -b "Tagged A" --tags "${TEST_PREFIX}_tag1,${TEST_PREFIX}_tag2"

    # Unicode content
    run_test "card create with unicode" \
        assert_success ./ankigo card create -d "$deck_name" \
            -f "日本とは何ですか？" -b "Japan"
}

test_card_create_validation() {
    local deck_name="${TEST_PREFIX}_ValidationTest"
    ./ankigo deck create "$deck_name" >/dev/null

    run_test "card create missing --front fails" \
        assert_failure ./ankigo card create -d "$deck_name" -b "back only"

    run_test "card create missing --back fails" \
        assert_failure ./ankigo card create -d "$deck_name" -f "front only"

    run_test "card create non-existent deck fails" \
        assert_failure ./ankigo card create -d "NONEXISTENT_DECK_${RANDOM}" \
            -f "Q" -b "A"

    run_test "card create non-existent model fails" \
        assert_failure ./ankigo card create -d "$deck_name" \
            -m "NONEXISTENT_MODEL_${RANDOM}" --field "Text=test"
}

test_card_create_duplicates() {
    local deck_name="${TEST_PREFIX}_DuplicateTest"
    local deck_name2="${TEST_PREFIX}_DuplicateTest2"
    ./ankigo deck create "$deck_name" >/dev/null
    ./ankigo deck create "$deck_name2" >/dev/null

    # Create original card
    ./ankigo card create -d "$deck_name" -f "Duplicate Q ${TEST_PREFIX}" -b "Duplicate A" >/dev/null

    # Duplicate should fail
    run_test "duplicate card fails" \
        assert_failure ./ankigo card create -d "$deck_name" \
            -f "Duplicate Q ${TEST_PREFIX}" -b "Duplicate A"

    # Duplicate with --allow-duplicate succeeds
    run_test "duplicate with --allow-duplicate succeeds" \
        assert_success ./ankigo card create -d "$deck_name" \
            -f "Duplicate Q ${TEST_PREFIX}" -b "Duplicate A" --allow-duplicate

    # Duplicate in different deck with --duplicate-scope deck succeeds
    run_test "duplicate in different deck with scope succeeds" \
        assert_success ./ankigo card create -d "$deck_name2" \
            -f "Duplicate Q ${TEST_PREFIX}" -b "Duplicate A" --duplicate-scope deck
}

test_card_create_cloze() {
    local deck_name="${TEST_PREFIX}_ClozeTest"
    ./ankigo deck create "$deck_name" >/dev/null

    run_test "cloze card create" \
        assert_success ./ankigo card create -d "$deck_name" \
            -m "Cloze" --field "Text=The {{c1::capital}} of France is {{c2::Paris}}"
}

test_card_search() {
    local deck_name="${TEST_PREFIX}_SearchTest"
    ./ankigo deck create "$deck_name" >/dev/null
    ./ankigo card create -d "$deck_name" -f "Search Q" -b "Search A" \
        --tags "${TEST_PREFIX}_searchtag" >/dev/null

    run_test "card search by deck" \
        assert_success ./ankigo card search "deck:$deck_name"

    run_test "card search by tag" \
        assert_success ./ankigo card search "tag:${TEST_PREFIX}_searchtag"

    local json_output
    json_output=$(./ankigo card search "deck:$deck_name" --json)
    run_test "card search --json is valid JSON array" assert_json_array "$json_output"

    run_test "card search --fields id,deck" \
        assert_success ./ankigo card search "deck:$deck_name" --fields id,deck

    # No results returns empty (not error)
    run_test "card search no results succeeds" \
        assert_success ./ankigo card search "deck:NONEXISTENT_${RANDOM}"

    run_test "card search invalid field fails" \
        assert_failure ./ankigo card search "deck:$deck_name" --fields invalid
}

test_deck_delete_dryrun() {
    local deck_name="${TEST_PREFIX}_DryRunDelete"
    ./ankigo deck create "$deck_name" >/dev/null

    # Dry run should not delete
    ./ankigo deck delete "$deck_name" --dry-run >/dev/null 2>&1 || true
    local deck_list
    deck_list=$(./ankigo deck list)
    run_test "deck delete --dry-run doesn't delete" assert_contains "$deck_list" "$deck_name"
}

test_deck_delete_force() {
    local deck_name="${TEST_PREFIX}_ForceDelete"
    ./ankigo deck create "$deck_name" >/dev/null

    run_test "deck delete --force succeeds" \
        assert_success ./ankigo deck delete "$deck_name" --force

    # Verify deck no longer exists
    local deck_list
    deck_list=$(./ankigo deck list)
    run_test "deleted deck no longer in list" assert_not_contains "$deck_list" "$deck_name"
}

test_deck_delete_by_id() {
    local deck_name="${TEST_PREFIX}_DeleteByID"
    local deck_id
    deck_id=$(./ankigo deck create "$deck_name")

    run_test "deck delete --id succeeds" \
        assert_success ./ankigo deck delete --id "$deck_id" --force

    local deck_list
    deck_list=$(./ankigo deck list)
    run_test "deck deleted by ID no longer in list" assert_not_contains "$deck_list" "$deck_name"
}

test_deck_delete_nonexistent() {
    run_test "deck delete non-existent fails" \
        assert_failure ./ankigo deck delete "NONEXISTENT_${RANDOM}" --force
}

test_deck_delete_multiple() {
    local deck1="${TEST_PREFIX}_Multi1"
    local deck2="${TEST_PREFIX}_Multi2"
    ./ankigo deck create "$deck1" >/dev/null
    ./ankigo deck create "$deck2" >/dev/null

    run_test "deck delete multiple succeeds" \
        assert_success ./ankigo deck delete "$deck1" "$deck2" --force

    local deck_list
    deck_list=$(./ankigo deck list)
    run_test "first deleted deck no longer in list" assert_not_contains "$deck_list" "$deck1"
    run_test "second deleted deck no longer in list" assert_not_contains "$deck_list" "$deck2"
}

test_exit_codes() {
    local exit_code

    ./ankigo version >/dev/null 2>&1
    exit_code=$?
    run_test "exit code 0 on success" test "$exit_code" = "0"

    ./ankigo deck delete "NONEXISTENT_${RANDOM}" --force 2>/dev/null || exit_code=$?
    run_test "exit code 1 on API error" test "$exit_code" = "1"

    ./ankigo unknowncommand 2>/dev/null || exit_code=$?
    run_test "exit code non-zero on unknown command" test "$exit_code" != "0"
}

# =============================================================================
# Main Execution
# =============================================================================

# Build binary
echo "[Build]"
go build -o ankigo . || { echo "Build failed"; exit 1; }
echo -e "${GREEN}✓${NC} built ankigo binary"
echo ""

# Run all tests
echo "[Prerequisites]"
test_prerequisites

echo ""
echo "[Version & Help]"
test_version_help

echo ""
echo "[Deck List]"
test_deck_list

echo ""
echo "[Deck Create]"
test_deck_create

echo ""
echo "[Card Create - Basic]"
test_card_create_basic

echo ""
echo "[Card Create - Validation]"
test_card_create_validation

echo ""
echo "[Card Create - Duplicates]"
test_card_create_duplicates

echo ""
echo "[Card Create - Cloze]"
test_card_create_cloze

echo ""
echo "[Card Search]"
test_card_search

echo ""
echo "[Deck Delete - Dry Run]"
test_deck_delete_dryrun

echo ""
echo "[Deck Delete - Force]"
test_deck_delete_force

echo ""
echo "[Deck Delete - By ID]"
test_deck_delete_by_id

echo ""
echo "[Deck Delete - Non-Existent]"
test_deck_delete_nonexistent

echo ""
echo "[Deck Delete - Multiple]"
test_deck_delete_multiple

echo ""
echo "[Exit Codes]"
test_exit_codes

# Print summary and exit with appropriate code
print_summary
