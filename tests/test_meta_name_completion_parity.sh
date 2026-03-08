#!/bin/bash

# Test script to verify that ttr meta and ttr name have the same autocompletion behavior

set -e

echo "=== Testing Meta and Name Completion Parity ==="

# Clean up any existing timetracker data
rm -rf ~/.timetracker

# Copy the built binary to the tests directory (if it doesn't exist)
if [ ! -f /app/tests/ttr ]; then
    cp /app/ttr /app/tests/ttr
fi

# Make sure we're using the local binary
TTR_BIN="./ttr"

# Initialize timetracker
echo "Initializing TimeTracker..."
echo "" | $TTR_BIN setup

# Create some test records
echo "Creating test records..."
echo "" | $TTR_BIN start test-record-1
sleep 1
$TTR_BIN end

echo "" | $TTR_BIN start test-record-2
sleep 1
$TTR_BIN end

echo "" | $TTR_BIN start test-record-3
sleep 1
$TTR_BIN end

# Generate completion script
echo "Generating completion script..."
$TTR_BIN completion bash > /home/testuser/test_completion.bash

# Test that both name and meta commands complete the same record names
echo "Testing completion parity..."

# Source the completion script (don't need execute permission to source)
. /home/testuser/test_completion.bash 2>/dev/null || . /home/testuser/test_completion.bash

# Test the _ttr_complete function
# We'll simulate the completion environment

# Test name command completion
echo "Testing 'ttr name' completion..."
COMP_WORDS=("ttr" "name" "")
COMP_CWORD=2
cur="test"
. /home/testuser/test_completion.bash
_ttr_complete
NAME_COMPLETIONS="${COMPREPLY[*]}"
echo "Name completions for 'test': $NAME_COMPLETIONS"

# Test meta command completion
echo "Testing 'ttr meta' completion..."
COMP_WORDS=("ttr" "meta" "")
COMP_CWORD=2
cur="test"
. /home/testuser/test_completion.bash
_ttr_complete
META_COMPLETIONS="${COMPREPLY[*]}"
echo "Meta completions for 'test': $META_COMPLETIONS"

# Compare the completions
if [ "$NAME_COMPLETIONS" = "$META_COMPLETIONS" ]; then
    echo "✅ PASSED: Name and meta commands have identical completion behavior"
else
    echo "❌ FAILED: Name and meta commands have different completion behavior"
    echo "Name completions: $NAME_COMPLETIONS"
    echo "Meta completions: $META_COMPLETIONS"
    exit 1
fi

# Test that both commands complete all record names
echo "Testing full record name completion..."
COMP_WORDS=("ttr" "name" "")
COMP_CWORD=2
cur=""
. /home/testuser/test_completion.bash
_ttr_complete
NAME_ALL_COMPLETIONS="${COMPREPLY[*]}"

COMP_WORDS=("ttr" "meta" "")
COMP_CWORD=2
cur=""
. /home/testuser/test_completion.bash
_ttr_complete
META_ALL_COMPLETIONS="${COMPREPLY[*]}"

if [ "$NAME_ALL_COMPLETIONS" = "$META_ALL_COMPLETIONS" ]; then
    echo "✅ PASSED: Name and meta commands complete all record names identically"
    echo "Available record completions: $META_ALL_COMPLETIONS"
else
    echo "❌ FAILED: Name and meta commands complete different record sets"
    echo "Name completions: $NAME_ALL_COMPLETIONS"
    echo "Meta completions: $META_ALL_COMPLETIONS"
    exit 1
fi

# Test that .json extensions are stripped from both
echo "Testing .json extension stripping..."
if echo "$META_ALL_COMPLETIONS" | grep -q "\.json"; then
    echo "❌ FAILED: Meta completions contain .json extensions"
    exit 1
else
    echo "✅ PASSED: Meta completions have .json extensions stripped"
fi

if echo "$NAME_ALL_COMPLETIONS" | grep -q "\.json"; then
    echo "❌ FAILED: Name completions contain .json extensions"
    exit 1
else
    echo "✅ PASSED: Name completions have .json extensions stripped"
fi

echo ""
echo "✅ All completion parity tests passed!"

# Clean up
rm -f /home/testuser/test_completion.bash
rm -rf ~/.timetracker