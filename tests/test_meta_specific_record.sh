#!/bin/bash

# Test script for ttr meta command with specific record functionality
# This tests the new feature: ttr meta <recordname> <meta-field-name> <meta-field-value>

set -e

# Clean up any existing timetracker data
rm -rf ~/.timetracker

# The binary should already be in /home/testuser/ttr from Dockerfile
# If not, try to copy it
if [ ! -f /home/testuser/ttr ]; then
    cp /app/ttr /home/testuser/ttr 2>/dev/null || true
fi

# Make sure we're using the local binary
TTR_BIN="/home/testuser/ttr"

# Initialize timetracker
echo "Initializing TimeTracker..."
echo "" | $TTR_BIN setup

# Test 1: Create some records first
echo "Test 1: Creating test records..."
echo "" | $TTR_BIN start test-record-1
sleep 1
$TTR_BIN end

$TTR_BIN start test-record-2
sleep 1
$TTR_BIN end

# Test 2: Add metadata to a specific record
echo "Test 2: Adding metadata to specific record..."
$TTR_BIN meta test-record-1 project "web-development"
$TTR_BIN meta test-record-1 priority high
$TTR_BIN meta test-record-1 billable true

# Test 3: Add metadata to another specific record
echo "Test 3: Adding metadata to another specific record..."
$TTR_BIN meta test-record-2 project "mobile-app"
$TTR_BIN meta test-record-2 priority low
$TTR_BIN meta test-record-2 client "acme-corp"

# Test 4: List metadata for specific records
echo "Test 4: Listing metadata for specific records..."
echo "Metadata for test-record-1:"
$TTR_BIN meta list test-record-1

echo "Metadata for test-record-2:"
$TTR_BIN meta list test-record-2

# Test 5: Verify metadata is correctly stored
echo "Test 5: Verifying metadata storage..."
METADATA_1=$(./ttr meta list test-record-1)
METADATA_2=$(./ttr meta list test-record-2)

if echo "$METADATA_1" | grep -q "project: web-development"; then
    echo "✓ test-record-1 has correct project metadata"
else
    echo "✗ test-record-1 project metadata not found"
    exit 1
fi

if echo "$METADATA_1" | grep -q "priority: high"; then
    echo "✓ test-record-1 has correct priority metadata"
else
    echo "✗ test-record-1 priority metadata not found"
    exit 1
fi

if echo "$METADATA_1" | grep -q "billable: true"; then
    echo "✓ test-record-1 has correct billable metadata"
else
    echo "✗ test-record-1 billable metadata not found"
    exit 1
fi

if echo "$METADATA_2" | grep -q "project: mobile-app"; then
    echo "✓ test-record-2 has correct project metadata"
else
    echo "✗ test-record-2 project metadata not found"
    exit 1
fi

if echo "$METADATA_2" | grep -q "client: acme-corp"; then
    echo "✓ test-record-2 has correct client metadata"
else
    echo "✗ test-record-2 client metadata not found"
    exit 1
fi

# Test 6: Test backward compatibility (current record metadata)
echo "Test 6: Testing backward compatibility..."
$TTR_BIN start current-test-record
$TTR_BIN meta current-field current-value
CURRENT_METADATA=$(./ttr meta list)

if echo "$CURRENT_METADATA" | grep -q "current-field: current-value"; then
    echo "✓ Current record metadata works (backward compatibility)"
else
    echo "✗ Current record metadata failed"
    exit 1
fi

$TTR_BIN end

# Test 7: Test different data types
echo "Test 7: Testing different data types..."
$TTR_BIN meta test-record-1 numeric-value 42
$TTR_BIN meta test-record-1 float-value 3.14
$TTR_BIN meta test-record-1 bool-false false

UPDATED_METADATA=$(./ttr meta list test-record-1)

if echo "$UPDATED_METADATA" | grep -q "numeric-value: 42"; then
    echo "✓ Numeric value stored correctly"
else
    echo "✗ Numeric value not found"
    exit 1
fi

if echo "$UPDATED_METADATA" | grep -q "float-value: 3.14"; then
    echo "✓ Float value stored correctly"
else
    echo "✗ Float value not found"
    exit 1
fi

if echo "$UPDATED_METADATA" | grep -q "bool-false: false"; then
    echo "✓ Boolean false value stored correctly"
else
    echo "✗ Boolean false value not found"
    exit 1
fi

# Test 8: Test that metadata can be added to non-existent records (creates new record metadata)
echo "Test 8: Testing metadata creation for new records..."
$TTR_BIN meta non-existent-record field value

# Verify the metadata was created
NEW_RECORD_METADATA=$(./ttr meta list non-existent-record)

if echo "$NEW_RECORD_METADATA" | grep -q "field: value"; then
    echo "✓ Metadata can be added to non-existent records (creates new metadata file)"
else
    echo "✗ Metadata creation for new records failed"
    exit 1
fi

echo ""
echo "All tests passed! ✓"

# Clean up
rm -rf ~/.timetracker