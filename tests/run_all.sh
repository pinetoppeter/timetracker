#!/bin/bash

# Master test script to run all TimeTracker test suites
# Optimized for containerized testing

set -e

echo "🚀 Running all TimeTracker tests in container..."
echo "================================================"
echo ""

# Set up containerized test environment
export TTR_BASE_DIR="/home/testuser/.timetracker_test"
export HOME="/home/testuser"
export TEST_ENVIRONMENT="containerized"

# Create necessary directories
mkdir -p "$TTR_BASE_DIR/records"
mkdir -p "$TTR_BASE_DIR/sessions"

# Run Go tests first
echo "🧪 Running Go unit tests..."
# Set GOMODCACHE to a writable location for non-root user
export GOMODCACHE="/home/testuser/.go-mod-cache"
mkdir -p "$GOMODCACHE"
if go test ./...; then
    echo "✅ Go tests passed"
    echo ""
else
    echo "❌ Go tests failed"
    echo ""
    exit 1
fi

# Define important test files to run (focused on core functionality)
TEST_FILES=(
    "tests/test_autocomplete.sh"
    "tests/test_metadata_persistence.sh"
    "tests/test_meta_export_integration.sh"
    "tests/test_meta_name_completion_parity.sh"
    "tests/test_meta_specific_record.sh"
    "tests/test_end.sh"
)
TOTAL_TESTS=${#TEST_FILES[@]}
echo "Running $TOTAL_TESTS important test files"
echo ""

# Run important test scripts
FAILED_TESTS=0
PASSED_TESTS=0

for test_script in ${TEST_FILES[@]}; do
    if [ -f "$test_script" ]; then
        echo "📋 Running $test_script..."
        if ! ./"/$test_script"; then
            echo "❌ $test_script failed"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        else
            echo "✅ $test_script passed"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        fi
        echo ""
    else
        echo "⚠️  $test_script not found, skipping"
        echo ""
    fi
done

echo "================================================"
echo "📊 Test Results:"
echo "   Passed: $PASSED_TESTS/$TOTAL_TESTS"
echo "   Failed: $FAILED_TESTS/$TOTAL_TESTS"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo "🎉 All tests completed successfully!"
    exit 0
else
    echo "❌ Some tests failed"
    exit 1
fi