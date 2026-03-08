#!/bin/bash

# Test script for metadata persistence across sessions
# Tests that metadata is preserved when creating new sessions with the same record name

set -e

echo "=== Testing Metadata Persistence Across Sessions ==="

# Clean up any existing timetracker data
rm -rf ~/.timetracker
rm -rf ~/.timetracker_data

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

# Test 1: Create first session and add metadata
echo "Test 1: Creating first session and adding metadata..."
echo "" | $TTR_BIN start project-alpha
sleep 1
$TTR_BIN end

$TTR_BIN meta project-alpha client "Acme Corporation"
$TTR_BIN meta project-alpha priority 1
$TTR_BIN meta project-alpha project_manager "John Doe"

# Verify metadata was saved
METADATA_1=$(./ttr meta list project-alpha)
if echo "$METADATA_1" | grep -q "client: Acme Corporation"; then
    echo "✅ First session metadata saved correctly"
else
    echo "✗ First session metadata not saved"
    exit 1
fi

# Test 2: Create new session with same record name
echo "Test 2: Creating new session with same record name..."
$TTR_BIN start project-alpha
sleep 1
$TTR_BIN end

# Verify metadata is still there
METADATA_2=$(./ttr meta list project-alpha)
if echo "$METADATA_2" | grep -q "client: Acme Corporation"; then
    echo "✅ Metadata preserved after new session creation"
else
    echo "✗ Metadata lost after new session creation"
    echo "Metadata after new session: $METADATA_2"
    exit 1
fi

if echo "$METADATA_2" | grep -q "priority: 1"; then
    echo "✅ Priority metadata preserved"
else
    echo "✗ Priority metadata lost"
    exit 1
fi

if echo "$METADATA_2" | grep -q "project_manager: John Doe"; then
    echo "✅ Project manager metadata preserved"
else
    echo "✗ Project manager metadata lost"
    exit 1
fi

# Test 3: Add additional metadata to existing record
echo "Test 3: Adding additional metadata..."
$TTR_BIN meta project-alpha estimated_hours 40
$TTR_BIN meta project-alpha status "in-progress"

# Verify all metadata is present
METADATA_3=$(./ttr meta list project-alpha)
if echo "$METADATA_3" | grep -q "estimated_hours: 40"; then
    echo "✅ New metadata added successfully"
else
    echo "✗ New metadata not added"
    exit 1
fi

if echo "$METADATA_3" | grep -q "status: in-progress"; then
    echo "✅ Status metadata added successfully"
else
    echo "✗ Status metadata not added"
    exit 1
fi

# Verify old metadata is still there
if echo "$METADATA_3" | grep -q "client: Acme Corporation"; then
    echo "✅ Original metadata still preserved after adding new metadata"
else
    echo "✗ Original metadata lost after adding new metadata"
    exit 1
fi

# Test 4: Create another new session and verify all metadata persists
echo "Test 4: Creating another new session..."
$TTR_BIN start project-alpha
sleep 1
$TTR_BIN end

METADATA_4=$(./ttr meta list project-alpha)

# Count metadata fields
FIELD_COUNT=$(echo "$METADATA_4" | grep -c ":" || true)
if [ "$FIELD_COUNT" -ge 5 ]; then
    echo "✅ All metadata fields preserved across multiple sessions"
else
    echo "✗ Some metadata fields lost (found $FIELD_COUNT, expected 5+)"
    echo "Final metadata: $METADATA_4"
    exit 1
fi

# Test 5: Test with different record names to ensure isolation
echo "Test 5: Testing metadata isolation between different records..."
$TTR_BIN start project-beta
sleep 1
$TTR_BIN end

$TTR_BIN meta project-beta client "Different Client"
$TTR_BIN meta project-beta priority 2

# Verify project-alpha metadata is unchanged
METADATA_ALPHA=$(./ttr meta list project-alpha)
METADATA_BETA=$(./ttr meta list project-beta)

if echo "$METADATA_ALPHA" | grep -q "client: Acme Corporation"; then
    echo "✅ Project-alpha metadata unchanged"
else
    echo "✗ Project-alpha metadata affected by project-beta changes"
    exit 1
fi

if echo "$METADATA_BETA" | grep -q "client: Different Client"; then
    echo "✅ Project-beta has its own metadata"
else
    echo "✗ Project-beta metadata not saved correctly"
    exit 1
fi

# Test 6: Test export functionality with preserved metadata
echo "Test 6: Testing export with preserved metadata..."
CURRENT_MONTH=$(date +%m)
CURRENT_YEAR=$(date +%Y)
$TTR_BIN export << EOF
$CURRENT_MONTH
$CURRENT_YEAR

EOF

# Check if export file was created and contains metadata
# With new data folder structure, exports go to data folder
DATA_FOLDER="/home/testuser/.timetracker_data"
if [ -z "$DATA_FOLDER" ]; then
    DATA_FOLDER="/home/testuser/.timetracker_data"
fi

# Debug: List all files in the data folder
echo "Files in $DATA_FOLDER:"
ls -la "$DATA_FOLDER/" 2>/dev/null || echo "No files found or permission issue"

# Wait a moment for file system to sync
sleep 1

# Find the most recent export file (with more robust search)
echo "Looking for export files in $DATA_FOLDER..."
EXPORT_FILE=$(find "$DATA_FOLDER" -name "timetracker-export-*.csv" -type f 2>/dev/null | sort | tail -1)

echo "Found export file: $EXPORT_FILE"

# If still not found, try explicit path based on export message
echo "Trying explicit path..."
EXPLICIT_PATH="/home/testuser/.timetracker_data/timetracker-export-$(date +%Y)-$(date +%m).csv"
if [ -f "$EXPLICIT_PATH" ]; then
    EXPORT_FILE="$EXPLICIT_PATH"
    echo "Found export file at explicit path: $EXPORT_FILE"
fi

# Also check current directory as fallback
if [ -z "$EXPORT_FILE" ] || [ ! -f "$EXPORT_FILE" ]; then
    echo "Checking current directory for export files..."
    EXPORT_FILE=$(find . -name "timetracker-export-*.csv" -type f 2>/dev/null | sort | tail -1)
    echo "Fallback export file: $EXPORT_FILE"
fi

if [ -n "$EXPORT_FILE" ] && [ -f "$EXPORT_FILE" ]; then

    
    if grep "project-alpha" "$EXPORT_FILE" | grep -q "Acme Corporation"; then
        echo "✅ Export contains preserved metadata"
    else
        echo "✗ Export missing preserved metadata"
        echo "Export line: $(grep "project-alpha" "$EXPORT_FILE")"
        exit 1
    fi
else
    echo "❌ No export file created in $DATA_FOLDER"
    exit 1
fi

echo ""
echo "✅ All metadata persistence tests passed!"

# Clean up
rm -rf ~/.timetracker
rm -rf ~/.timetracker_data