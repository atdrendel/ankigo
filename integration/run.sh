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

    # Clean up test decks (this also deletes notes in those decks)
    local decks_to_delete
    decks_to_delete=$(./ankigo deck list 2>/dev/null | grep "^${TEST_PREFIX}" || true)
    if [[ -n "$decks_to_delete" ]]; then
        echo "$decks_to_delete" | while read -r deck; do
            ./ankigo deck delete "$deck" --force 2>/dev/null || true
            echo "  Deleted deck: $deck"
        done
    else
        echo "  No test decks to clean up"
    fi

    # Clean up test models (now empty after decks deleted)
    local models_to_prune
    models_to_prune=$(./ankigo model list 2>/dev/null | grep "^${TEST_PREFIX}" || true)
    if [[ -n "$models_to_prune" ]]; then
        echo "$models_to_prune" | while read -r model; do
            ./ankigo model prune "$model" --force 2>/dev/null || true
            echo "  Pruned model: $model"
        done
    else
        echo "  No test models to clean up"
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

    run_test "note --help succeeds" assert_success ./ankigo note --help
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

test_note_create_basic() {
    local deck_name="${TEST_PREFIX}_NoteTest"
    ./ankigo deck create "$deck_name" >/dev/null

    local note_id
    note_id=$(./ankigo note create -d "$deck_name" -m "$TEST_MODEL" -f "Test front" -b "Test back")
    run_test "note create returns numeric ID" assert_numeric "$note_id"

    # Verify card appears in search (notes create cards)
    local search_result
    search_result=$(./ankigo card search "deck:$deck_name")
    run_test "created note's card appears in search" assert_not_empty "$search_result"

    # Note with tags
    run_test "note create with tags" \
        assert_success ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" \
            -f "Tagged Q" -b "Tagged A" --tags "${TEST_PREFIX}_tag1,${TEST_PREFIX}_tag2"

    # Unicode content
    run_test "note create with unicode" \
        assert_success ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" \
            -f "日本とは何ですか？" -b "Japan"
}

test_note_create_validation() {
    local deck_name="${TEST_PREFIX}_ValidationTest"
    ./ankigo deck create "$deck_name" >/dev/null

    run_test "note create missing --front fails" \
        assert_failure ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" -b "back only"

    run_test "note create missing --back fails" \
        assert_failure ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" -f "front only"

    run_test "note create non-existent deck fails" \
        assert_failure ./ankigo note create -d "NONEXISTENT_DECK_${RANDOM}" -m "$TEST_MODEL" \
            -f "Q" -b "A"

    run_test "note create non-existent model fails" \
        assert_failure ./ankigo note create -d "$deck_name" \
            -m "NONEXISTENT_MODEL_${RANDOM}" --field "Text=test"
}

test_note_create_duplicates() {
    local deck_name="${TEST_PREFIX}_DuplicateTest"
    local deck_name2="${TEST_PREFIX}_DuplicateTest2"
    ./ankigo deck create "$deck_name" >/dev/null
    ./ankigo deck create "$deck_name2" >/dev/null

    # Create original note
    ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" -f "Duplicate Q ${TEST_PREFIX}" -b "Duplicate A" >/dev/null

    # Duplicate should fail
    run_test "duplicate note fails" \
        assert_failure ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" \
            -f "Duplicate Q ${TEST_PREFIX}" -b "Duplicate A"

    # Duplicate with --allow-duplicate succeeds
    run_test "duplicate with --allow-duplicate succeeds" \
        assert_success ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" \
            -f "Duplicate Q ${TEST_PREFIX}" -b "Duplicate A" --allow-duplicate

    # Duplicate in different deck with --duplicate-scope deck succeeds
    run_test "duplicate in different deck with scope succeeds" \
        assert_success ./ankigo note create -d "$deck_name2" -m "$TEST_MODEL" \
            -f "Duplicate Q ${TEST_PREFIX}" -b "Duplicate A" --duplicate-scope deck
}

test_note_create_cloze() {
    local deck_name="${TEST_PREFIX}_ClozeTest"
    ./ankigo deck create "$deck_name" >/dev/null

    run_test "cloze note create" \
        assert_success ./ankigo note create -d "$deck_name" \
            -m "Cloze" --field "Text=The {{c1::capital}} of France is {{c2::Paris}}"
}

test_note_create_media() {
    local deck_name="${TEST_PREFIX}_MediaTest"
    ./ankigo deck create "$deck_name" >/dev/null

    # Use RELATIVE paths to test that ankigo converts them to absolute paths
    # (anki-connect requires absolute paths, so ankigo must handle this)
    local test_audio="integration/testdata/test.mp3"
    local test_image="integration/testdata/test.png"
    local test_video="integration/testdata/test.mp4"

    # Test: Note with audio from relative path
    run_test "note create with audio (relative path)" \
        assert_success ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" \
            -f "Audio Q" -b "Audio A" \
            --audio "filename=${TEST_PREFIX}_audio.mp3,path=$test_audio,fields=Back"

    # Test: Note with picture from relative path
    run_test "note create with picture (relative path)" \
        assert_success ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" \
            -f "Picture Q" -b "Picture A" \
            --picture "filename=${TEST_PREFIX}_image.png,path=$test_image,fields=Front"

    # Test: Note with video from relative path
    run_test "note create with video (relative path)" \
        assert_success ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" \
            -f "Video Q" -b "Video A" \
            --video "filename=${TEST_PREFIX}_video.mp4,path=$test_video,fields=Back"

    # Test: Note with multiple media types
    run_test "note create with multiple media" \
        assert_success ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" \
            -f "Multi Q" -b "Multi A" \
            --audio "filename=${TEST_PREFIX}_a2.mp3,path=$test_audio,fields=Back" \
            --picture "filename=${TEST_PREFIX}_i2.png,path=$test_image,fields=Front"

    # Test: Note with media attached to multiple fields
    run_test "note create with media on multiple fields" \
        assert_success ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" \
            -f "Both Q" -b "Both A" \
            --picture "filename=${TEST_PREFIX}_both.png,path=$test_image,fields=Front;Back"

    # Test: Note with absolute path (should also work)
    local abs_audio="$SCRIPT_DIR/testdata/test.mp3"
    run_test "note create with audio (absolute path)" \
        assert_success ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" \
            -f "Abs Audio Q" -b "Abs Audio A" \
            --audio "filename=${TEST_PREFIX}_abs_audio.mp3,path=$abs_audio,fields=Back"
}

test_note_create_media_errors() {
    local deck_name="${TEST_PREFIX}_MediaErrorTest"
    ./ankigo deck create "$deck_name" >/dev/null

    # Test: Missing filename
    run_test "note create media missing filename fails" \
        assert_failure ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" \
            -f "Q" -b "A" --audio "path=/tmp/test.mp3,fields=Back"

    # Test: Missing source (no path/url/data)
    run_test "note create media missing source fails" \
        assert_failure ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" \
            -f "Q" -b "A" --audio "filename=test.mp3,fields=Back"

    # Test: Missing fields
    run_test "note create media missing fields fails" \
        assert_failure ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" \
            -f "Q" -b "A" --audio "filename=test.mp3,path=/tmp/test.mp3"

    # Test: Invalid spec format
    run_test "note create media invalid format fails" \
        assert_failure ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" \
            -f "Q" -b "A" --audio "invalid"
}

test_note_delete() {
    local deck_name="${TEST_PREFIX}_NoteDeleteTest"
    ./ankigo deck create "$deck_name" >/dev/null

    # Create a note to delete
    local note_id
    note_id=$(./ankigo note create -d "$deck_name" -m "$TEST_MODEL" -f "Delete me" -b "Answer")
    run_test "note create returns numeric ID" assert_numeric "$note_id"

    # Verify note exists via card search
    local search_before
    search_before=$(./ankigo card search "deck:$deck_name")
    run_test "created note has cards" assert_not_empty "$search_before"

    # Delete the note
    run_test "note delete succeeds" \
        assert_success ./ankigo note delete "$note_id" --force

    # Verify note is gone (no cards found)
    local search_after
    search_after=$(./ankigo card search "deck:$deck_name" 2>/dev/null || true)
    run_test "deleted note has no cards" \
        test "$search_after" = "No cards found" -o -z "$search_after"
}

test_note_delete_dryrun() {
    local deck_name="${TEST_PREFIX}_NoteDeleteDryRun"
    ./ankigo deck create "$deck_name" >/dev/null

    local note_id
    note_id=$(./ankigo note create -d "$deck_name" -m "$TEST_MODEL" -f "Dry run Q" -b "Dry run A")

    # Dry run should NOT delete
    ./ankigo note delete "$note_id" --dry-run >/dev/null 2>&1 || true

    # Verify note still exists
    local search_result
    search_result=$(./ankigo card search "deck:$deck_name")
    run_test "note delete --dry-run doesn't delete" assert_not_empty "$search_result"
}

test_note_delete_multiple() {
    local deck_name="${TEST_PREFIX}_NoteDeleteMultiple"
    ./ankigo deck create "$deck_name" >/dev/null

    local note1 note2
    note1=$(./ankigo note create -d "$deck_name" -m "$TEST_MODEL" -f "Q1" -b "A1")
    note2=$(./ankigo note create -d "$deck_name" -m "$TEST_MODEL" -f "Q2" -b "A2")

    run_test "note delete multiple succeeds" \
        assert_success ./ankigo note delete "$note1" "$note2" --force

    local search_result
    search_result=$(./ankigo card search "deck:$deck_name" 2>/dev/null || true)
    run_test "all notes deleted" \
        test "$search_result" = "No cards found" -o -z "$search_result"
}

test_note_delete_validation() {
    run_test "note delete no args fails" \
        assert_failure ./ankigo note delete

    run_test "note delete invalid ID fails" \
        assert_failure ./ankigo note delete "not-a-number"
}

test_note_list() {
    local deck_name="${TEST_PREFIX}_NoteListTest"
    ./ankigo deck create "$deck_name" >/dev/null

    # Create test notes with distinct tags for filtering
    local note1 note2
    note1=$(./ankigo note create -d "$deck_name" -m "$TEST_MODEL" -f "Q1" -b "A1" --tags "${TEST_PREFIX}_tag1,${TEST_PREFIX}_tag2")
    note2=$(./ankigo note create -d "$deck_name" -m "$TEST_MODEL" -f "Q2" -b "A2" --tags "${TEST_PREFIX}_tag2,${TEST_PREFIX}_tag3")

    # Basic list returns note IDs
    local result
    result=$(./ankigo note list "deck:$deck_name")
    run_test "note list returns results" assert_not_empty "$result"
    run_test "note list includes note1" assert_contains "$result" "$note1"
    run_test "note list includes note2" assert_contains "$result" "$note2"

    # List with tag filter
    result=$(./ankigo note list "deck:$deck_name tag:${TEST_PREFIX}_tag1")
    run_test "note list with tag filter returns note1" assert_contains "$result" "$note1"

    # List with fields
    result=$(./ankigo note list "deck:$deck_name" --fields id,model)
    run_test "note list with fields contains model" assert_contains "$result" "$TEST_MODEL"

    # JSON output
    result=$(./ankigo note list "deck:$deck_name" --json)
    run_test "note list JSON is valid" assert_json_array "$result"

    # Empty result (non-existent deck)
    result=$(./ankigo note list "deck:NonExistentDeck_${TEST_PREFIX}" 2>&1 || true)
    run_test "note list empty result" assert_contains "$result" "No notes found"
}

test_note_list_validation() {
    run_test "note list invalid field fails" \
        assert_failure ./ankigo note list "deck:Default" --fields invalid_field
}

test_note_list_all() {
    # Test listing all notes (no query) - only verifies command succeeds
    # and returns valid output (we don't check content since we don't
    # want to depend on or inspect user's existing notes)
    local deck_name="${TEST_PREFIX}_NoteListAllTest"
    ./ankigo deck create "$deck_name" >/dev/null

    # Create a test note so we have at least one
    local note_id
    note_id=$(./ankigo note create -d "$deck_name" -m "$TEST_MODEL" -f "AllQ" -b "AllA")

    # List all notes should succeed and include our test note
    local result
    result=$(./ankigo note list)
    run_test "note list (no query) succeeds" assert_not_empty "$result"
    run_test "note list includes test note" assert_contains "$result" "$note_id"
}

test_card_search() {
    local deck_name="${TEST_PREFIX}_SearchTest"
    ./ankigo deck create "$deck_name" >/dev/null
    ./ankigo note create -d "$deck_name" -m "$TEST_MODEL" -f "Search Q" -b "Search A" \
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

test_model_list() {
    # Test that model list runs successfully
    run_test "model list succeeds" assert_success ./ankigo model list

    # Verify test model appears (created in setup)
    local result
    result=$(./ankigo model list)
    run_test "model list includes test model" assert_contains "$result" "$TEST_MODEL"

    # Test with fields flag
    run_test "model list --fields name,id succeeds" \
        assert_success ./ankigo model list --fields name,id

    # Test with fields showing field names
    local fields_output
    fields_output=$(./ankigo model list --fields name,fields)
    run_test "model list shows field names" assert_contains "$fields_output" "Front"

    # Test JSON output
    local json_output
    json_output=$(./ankigo model list --json)
    run_test "model list --json is valid JSON" assert_json_array "$json_output"

    # Invalid field should fail
    run_test "model list invalid field fails" \
        assert_failure ./ankigo model list --fields invalid_field

}

test_model_create() {
    local model_name="${TEST_PREFIX}_TestModel"

    # Create basic model
    run_test "model create basic succeeds" \
        assert_success ./ankigo model create "$model_name" \
            --field Front --field Back \
            --template "Card 1,{{Front}},{{Back}}"

    # Verify model appears in list
    local model_list
    model_list=$(./ankigo model list)
    run_test "created model appears in list" assert_contains "$model_list" "$model_name"

    # Verify fields
    local fields_output
    fields_output=$(./ankigo model list --fields name,fields | grep "$model_name")
    run_test "model has correct fields" assert_contains "$fields_output" "Front,Back"

    # Create cloze model
    local cloze_name="${TEST_PREFIX}_ClozeModel"
    run_test "create cloze model" assert_success \
        ./ankigo model create "$cloze_name" \
            --field Text --field Extra --cloze \
            --template "Cloze,{{cloze:Text}},{{cloze:Text}}<br>{{Extra}}"

    # Create model with multiple templates
    local multi_name="${TEST_PREFIX}_MultiTemplate"
    run_test "create model with multiple templates" assert_success \
        ./ankigo model create "$multi_name" \
            --field Front --field Back \
            --template "Forward,{{Front}},{{Back}}" \
            --template "Reverse,{{Back}},{{Front}}"

    # Create model with CSS
    local styled_name="${TEST_PREFIX}_StyledModel"
    run_test "create model with CSS" assert_success \
        ./ankigo model create "$styled_name" \
            --field Q --field A \
            --template "Card 1,{{Q}},{{A}}" \
            --css ".card { font-size: 20px; }"

    # Error: duplicate name
    run_test "model create duplicate fails" assert_failure \
        ./ankigo model create "$model_name" --field X --template "T,{{X}},{{X}}"

    # Error: missing fields
    run_test "model create missing fields fails" assert_failure \
        ./ankigo model create "${TEST_PREFIX}_NoFields" --template "T,X,Y"

    # Error: missing templates
    run_test "model create missing templates fails" assert_failure \
        ./ankigo model create "${TEST_PREFIX}_NoTemplates" --field X

    # Error: invalid template format
    run_test "model create invalid template fails" assert_failure \
        ./ankigo model create "${TEST_PREFIX}_BadTemplate" --field X --template "invalid"
}

test_model_prune() {
    # Create an empty test model
    local model1="${TEST_PREFIX}_PruneMe1"
    ./ankigo model create "$model1" \
        --field Front --field Back \
        --template "Card 1,{{Front}},{{Back}}" >/dev/null

    # Dry run doesn't remove
    ./ankigo model prune --dry-run >/dev/null 2>&1
    local list_after_dry
    list_after_dry=$(./ankigo model list)
    run_test "model prune --dry-run doesn't remove" assert_contains "$list_after_dry" "$model1"

    # Prune all empty models (with --force to skip confirmation)
    run_test "model prune succeeds" assert_success \
        ./ankigo model prune --force

    # Verify model is gone
    local list_after_prune
    list_after_prune=$(./ankigo model list)
    run_test "pruned model not in list" assert_not_contains "$list_after_prune" "$model1"
}

test_note_create_schema() {
    local schema_output
    schema_output=$(./ankigo note create --schema)
    run_test "note create --schema outputs valid JSON" assert_valid_json "$schema_output"
    run_test "note create --schema contains deckName" assert_contains "$schema_output" "deckName"
    run_test "note create --schema contains modelName" assert_contains "$schema_output" "modelName"
    run_test "note create --schema contains fields" assert_contains "$schema_output" "fields"
}

test_model_create_schema() {
    local schema_output
    schema_output=$(./ankigo model create --schema)
    run_test "model create --schema outputs valid JSON" assert_valid_json "$schema_output"
    run_test "model create --schema contains modelName" assert_contains "$schema_output" "modelName"
    run_test "model create --schema contains fields" assert_contains "$schema_output" "fields"
    run_test "model create --schema contains templates" assert_contains "$schema_output" "templates"
}

test_note_create_input_json() {
    local deck_name="${TEST_PREFIX}_InputJSONNote"
    ./ankigo deck create "$deck_name" >/dev/null

    # Happy path: create note via --input-json
    local note_id
    note_id=$(./ankigo note create --input-json "{
        \"deckName\": \"$deck_name\",
        \"modelName\": \"$TEST_MODEL\",
        \"fields\": {\"Front\": \"JSON Q\", \"Back\": \"JSON A\"},
        \"tags\": [\"${TEST_PREFIX}_json\"]
    }")
    run_test "note create --input-json returns numeric ID" assert_numeric "$note_id"

    # Verify note exists
    local search_result
    search_result=$(./ankigo card search "deck:$deck_name")
    run_test "note created via --input-json appears in search" assert_not_empty "$search_result"

    # Error: invalid JSON
    run_test "note create --input-json invalid JSON fails" \
        assert_failure ./ankigo note create --input-json "not json"

    # Error: conflict with --front flag
    run_test "note create --input-json with --front fails" \
        assert_failure ./ankigo note create --input-json '{"deckName":"x","modelName":"Basic","fields":{"Front":"Q"}}' \
            -f "conflict"
}

test_model_create_input_json() {
    local model_name="${TEST_PREFIX}_InputJSONModel"

    # Happy path: create model via --input-json (name from JSON, no positional arg)
    run_test "model create --input-json succeeds" assert_success \
        ./ankigo model create --input-json "{
            \"modelName\": \"$model_name\",
            \"fields\": [\"Question\", \"Answer\"],
            \"templates\": [{\"name\": \"Card 1\", \"front\": \"{{Question}}\", \"back\": \"{{Answer}}\"}]
        }"

    # Verify model appears in list
    local model_list
    model_list=$(./ankigo model list)
    run_test "model created via --input-json appears in list" assert_contains "$model_list" "$model_name"

    # Verify fields
    local fields_output
    fields_output=$(./ankigo model list --fields name,fields | grep "$model_name")
    run_test "model --input-json has correct fields" assert_contains "$fields_output" "Question,Answer"

    # Error: invalid JSON
    run_test "model create --input-json invalid JSON fails" \
        assert_failure ./ankigo model create --input-json "not json"
}

test_non_tty_confirmation() {
    local deck_name="${TEST_PREFIX}_NonTTYTest"
    ./ankigo deck create "$deck_name" >/dev/null

    # deck delete without --force in non-TTY mode should fail
    local output
    output=$(./ankigo deck delete "$deck_name" </dev/null 2>&1) || true
    run_test "deck delete non-TTY without --force fails" \
        assert_contains "$output" "use --force"

    # Verify deck still exists
    local deck_list
    deck_list=$(./ankigo deck list)
    run_test "deck not deleted in non-TTY mode" assert_contains "$deck_list" "$deck_name"

    # note delete without --force in non-TTY mode
    local note_id
    note_id=$(./ankigo note create -d "$deck_name" -m "$TEST_MODEL" -f "NonTTY Q" -b "NonTTY A")
    output=$(./ankigo note delete "$note_id" </dev/null 2>&1) || true
    run_test "note delete non-TTY without --force fails" \
        assert_contains "$output" "use --force"

    # model prune without --force in non-TTY mode (model must be empty)
    local model_name="${TEST_PREFIX}_NonTTYModel"
    ./ankigo model create "$model_name" \
        --field Front --field Back \
        --template "Card 1,{{Front}},{{Back}}" >/dev/null
    output=$(./ankigo model prune "$model_name" </dev/null 2>&1) || true
    run_test "model prune non-TTY without --force fails" \
        assert_contains "$output" "use --force"
}

test_model_prune_skips_nonempty() {
    # Create model with a note (should be skipped by prune)
    local model_name="${TEST_PREFIX}_HasNotes"
    local deck_name="${TEST_PREFIX}_PruneDeck"

    ./ankigo model create "$model_name" \
        --field Front --field Back \
        --template "Card 1,{{Front}},{{Back}}" >/dev/null
    ./ankigo deck create "$deck_name" >/dev/null

    # Create a note using this model
    ./ankigo note create --deck "$deck_name" --model "$model_name" \
        --field "Front=Q" --field "Back=A" >/dev/null

    # Prune should skip this model (it has notes)
    local prune_output
    prune_output=$(./ankigo model prune "$model_name" 2>&1)
    run_test "model prune skips non-empty" assert_contains "$prune_output" "Skipped"

    # Model should still exist
    local list_after
    list_after=$(./ankigo model list)
    run_test "non-empty model still exists" assert_contains "$list_after" "$model_name"
}

# =============================================================================
# Main Execution
# =============================================================================

# Build binary
echo "[Build]"
go build -o ankigo . || { echo "Build failed"; exit 1; }
echo -e "${GREEN}✓${NC} built ankigo binary"

# Create shared test model (replaces dependency on built-in "Basic")
echo ""
echo "[Setup]"
TEST_MODEL="${TEST_PREFIX}_Model"
./ankigo model create "$TEST_MODEL" \
    --field Front --field Back \
    --template "Card 1,{{Front}},{{Back}}" >/dev/null
echo -e "${GREEN}✓${NC} created test model: $TEST_MODEL"
echo ""

# Run all tests
echo "[Prerequisites]"
test_prerequisites

echo ""
echo "[Version & Help]"
test_version_help

echo ""
echo "[Note Create - Schema]"
test_note_create_schema

echo ""
echo "[Model Create - Schema]"
test_model_create_schema

echo ""
echo "[Deck List]"
test_deck_list

echo ""
echo "[Model List]"
test_model_list

echo ""
echo "[Model Create]"
test_model_create

echo ""
echo "[Model Create - Input JSON]"
test_model_create_input_json

echo ""
echo "[Model Prune]"
test_model_prune

echo ""
echo "[Model Prune - Skips Non-Empty]"
test_model_prune_skips_nonempty

echo ""
echo "[Deck Create]"
test_deck_create

echo ""
echo "[Note Create - Basic]"
test_note_create_basic

echo ""
echo "[Note Create - Validation]"
test_note_create_validation

echo ""
echo "[Note Create - Duplicates]"
test_note_create_duplicates

echo ""
echo "[Note Create - Cloze]"
test_note_create_cloze

echo ""
echo "[Note Create - Media]"
test_note_create_media

echo ""
echo "[Note Create - Media Errors]"
test_note_create_media_errors

echo ""
echo "[Note Create - Input JSON]"
test_note_create_input_json

echo ""
echo "[Note Delete]"
test_note_delete

echo ""
echo "[Note Delete - Dry Run]"
test_note_delete_dryrun

echo ""
echo "[Note Delete - Multiple]"
test_note_delete_multiple

echo ""
echo "[Note Delete - Validation]"
test_note_delete_validation

echo ""
echo "[Note List]"
test_note_list

echo ""
echo "[Note List - Validation]"
test_note_list_validation

echo ""
echo "[Note List - All]"
test_note_list_all

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
echo "[Non-TTY Confirmation]"
test_non_tty_confirmation

echo ""
echo "[Exit Codes]"
test_exit_codes

# Print summary and exit with appropriate code
print_summary
