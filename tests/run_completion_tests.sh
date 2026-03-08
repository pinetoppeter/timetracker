#!/bin/bash

# Master script to run all autocompletion tests

set -e

echo "🧪 Running Comprehensive Autocompletion Test Suite"
echo "================================================"

# Run all individual tests
TESTS=(
    "test_completion_basic.sh"
    "test_completion_records.sh" 
    "test_completion_edge_cases.sh"
    "test_completion_setup.sh"
)

PASSED=0
FAILED=0

for test in "${TESTS[@]}"; do
    echo ""
    echo "Running $test..."
    if ./"$test"; then
        echo "✅ $test PASSED"
        ((PASSED++))
    else
        echo "❌ $test FAILED"
        ((FAILED++))
    fi
done

echo ""
echo "================================================"
echo "Test Results Summary:"
echo "  Passed: $PASSED"
echo "  Failed: $FAILED"
echo "  Total:  $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "🎉 All autocompletion tests passed!"
    exit 0
else
    echo "💥 Some tests failed. Please check the output above."
    exit 1
fi