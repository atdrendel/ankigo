#!/usr/bin/env bash
# Integration test helper functions

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Run a test and track pass/fail
# Usage: run_test "test name" command args...
run_test() {
    local name="$1"
    shift
    if "$@"; then
        echo -e "${GREEN}✓${NC} $name"
        ((TESTS_PASSED++)) || true
    else
        echo -e "${RED}✗${NC} $name"
        ((TESTS_FAILED++)) || true
    fi
}

# Assert command succeeds (exit code 0)
assert_success() {
    "$@" >/dev/null 2>&1
}

# Assert command fails (non-zero exit code)
assert_failure() {
    ! "$@" >/dev/null 2>&1
}

# Assert output contains expected string
# Usage: assert_contains "$output" "expected"
assert_contains() {
    local output="$1"
    local expected="$2"
    [[ "$output" == *"$expected"* ]]
}

# Assert output does not contain string
# Usage: assert_not_contains "$output" "unexpected"
assert_not_contains() {
    local output="$1"
    local unexpected="$2"
    [[ "$output" != *"$unexpected"* ]]
}

# Assert output is a valid integer
assert_numeric() {
    [[ "$1" =~ ^[0-9]+$ ]]
}

# Assert output is valid JSON (any type)
assert_valid_json() {
    echo "$1" | jq -e '.' >/dev/null 2>&1
}

# Assert output is valid JSON array
assert_json_array() {
    echo "$1" | jq -e 'type == "array"' >/dev/null 2>&1
}

# Assert output is non-empty
assert_not_empty() {
    [[ -n "$1" ]]
}

# Print test summary and return appropriate exit code
print_summary() {
    echo ""
    echo "================================="
    echo -e "Tests: ${GREEN}${TESTS_PASSED} passed${NC}, ${RED}${TESTS_FAILED} failed${NC}"
    echo "================================="
    [[ $TESTS_FAILED -eq 0 ]]
}
