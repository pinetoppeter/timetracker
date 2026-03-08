#!/bin/bash

# Test script for metadata export integration
# Tests that metadata fields added via ttr meta are automatically included in export schema

set -e

echo "=== Testing Metadata Export Integration ==="

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

# Check initial export schema
echo "Checking initial export schema..."
# Use the containerized data folder directly
DATA_FOLDER="/home/testuser/.timetracker_data"

INITIAL_SCHEMA=$(cat "$DATA_FOLDER/export-schema.json")
echo "Initial schema has $(echo $INITIAL_SCHEMA | grep -o '"name"' | wc -l) columns"

# Create a test record (use a specific date to ensure it's in the current month)
echo "Creating test record..."
echo "" | $TTR_BIN start test-project
sleep 1
$TTR_BIN end

# Add various types of metadata
echo "Adding metadata fields..."
$TTR_BIN meta test-project client "Acme Corporation"
$TTR_BIN meta test-project priority 1
$TTR_BIN meta test-project billable true
$TTR_BIN meta test-project project_code "ACME-2024"
$TTR_BIN meta test-project estimated_hours 40

# Check that export schema was updated
echo "Checking updated export schema..."
UPDATED_SCHEMA=$(cat "$DATA_FOLDER/export-schema.json")
echo "Updated schema has $(echo $UPDATED_SCHEMA | grep -o '"name"' | wc -l) columns"

# Verify new fields were added to schema
if echo "$UPDATED_SCHEMA" | grep -q "client"; then
    echo "✅ Client field added to export schema"
else
    echo "❌ Client field not found in export schema"
    exit 1
fi

if echo "$UPDATED_SCHEMA" | grep -q "priority"; then
    echo "✅ Priority field added to export schema"
else
    echo "❌ Priority field not found in export schema"
    exit 1
fi

if echo "$UPDATED_SCHEMA" | grep -q "billable"; then
    echo "✅ Billable field added to export schema"
else
    echo "❌ Billable field not found in export schema"
    exit 1
fi

if echo "$UPDATED_SCHEMA" | grep -q "project_code"; then
    echo "✅ Project_code field added to export schema"
else
    echo "❌ Project_code field not found in export schema"
    exit 1
fi

if echo "$UPDATED_SCHEMA" | grep -q "estimated_hours"; then
    echo "✅ Estimated_hours field added to export schema"
else
    echo "❌ Estimated_hours field not found in export schema"
    exit 1
fi

# Verify field types are correct
if echo "$UPDATED_SCHEMA" | grep -A 5 '"priority"' | grep -q '"type": "number"'; then
    echo "✅ Priority field has correct type (number)"
else
    echo "❌ Priority field has incorrect type"
    exit 1
fi

if echo "$UPDATED_SCHEMA" | grep -A 5 '"billable"' | grep -q '"type": "boolean"'; then
    echo "✅ Billable field has correct type (boolean)"
else
    echo "❌ Billable field has incorrect type"
    exit 1
fi

if echo "$UPDATED_SCHEMA" | grep -A 5 '"client"' | grep -q '"type": "string"'; then
    echo "✅ Client field has correct type (string)"
else
    echo "❌ Client field has incorrect type"
    exit 1
fi

# Test that duplicate fields don't cause issues
echo "Testing duplicate field handling..."
$TTR_BIN meta test-project client "Updated Client"

FINAL_SCHEMA=$(cat "$DATA_FOLDER/export-schema.json")
CLIENT_COUNT=$(echo "$FINAL_SCHEMA" | grep -c '"client"' || true)
if [ "$CLIENT_COUNT" -eq 1 ]; then
    echo "✅ Duplicate field handling works (only one client field)"
else
    echo "❌ Duplicate field handling failed (found $CLIENT_COUNT client fields)"
    exit 1
fi

# Test export functionality
echo "Testing export with metadata..."
# Since the export command defaults to previous month, and we're creating data in current month,
# we need to explicitly specify the current month
CURRENT_MONTH=$(date +%m)
CURRENT_YEAR=$(date +%Y)
$TTR_BIN export << EOF
$CURRENT_MONTH
$CURRENT_YEAR

EOF

# Check if export file was created and contains metadata
# Debug: List all files in the data folder first
echo "Debug: Files in $DATA_FOLDER:"
ls -la "$DATA_FOLDER/" 2>/dev/null || echo "No files found or permission issue"

# Find the most recent export file (simplified approach)
echo "Looking for export files in $DATA_FOLDER..."
EXPORT_FILE=$(find "$DATA_FOLDER" -name "timetracker-export-*.csv" -type f 2>/dev/null | sort | tail -1)

echo "Debug: Found export file via find: $EXPORT_FILE"

# Also try explicit path based on current date
if [ -z "$EXPORT_FILE" ] || [ ! -f "$EXPORT_FILE" ]; then
    EXPLICIT_PATH="$DATA_FOLDER/timetracker-export-$(date +%Y)-$(date +%m).csv"
    echo "Debug: Trying explicit path: $EXPLICIT_PATH"
    if [ -f "$EXPLICIT_PATH" ]; then
        EXPORT_FILE="$EXPLICIT_PATH"
        echo "Debug: Using explicit path: $EXPORT_FILE"
    else
        echo "Debug: Explicit path not found either"
    fi
fi

# Final check
echo "Debug: Final EXPORT_FILE value: $EXPORT_FILE"
if [ -f "$EXPORT_FILE" ]; then
    echo "Debug: File exists and is readable"
    ls -la "$EXPORT_FILE"
else
    echo "Debug: File does not exist or is not readable"
fi

if [ -n "$EXPORT_FILE" ] && [ -f "$EXPORT_FILE" ]; then
    echo "✅ Export file created"
    echo "Export file: $EXPORT_FILE"
    
    # Check if export contains metadata columns
    if head -1 "$EXPORT_FILE" | grep -q "Client"; then
        echo "✅ Export contains Client column"
    else
        echo "❌ Export missing Client column"
        exit 1
    fi
    
    if head -1 "$EXPORT_FILE" | grep -q "Priority"; then
        echo "✅ Export contains Priority column"
    else
        echo "❌ Export missing Priority column"
        exit 1
    fi
    
    if head -1 "$EXPORT_FILE" | grep -q "Billable"; then
        echo "✅ Export contains Billable column"
    else
        echo "❌ Export missing Billable column"
        exit 1
    fi
    
    # Check if export contains metadata values (note: client was updated to "Updated Client")
    if grep "test-project" "$EXPORT_FILE" | grep -q "Updated Client"; then
        echo "✅ Export contains correct Client value"
    else
        echo "❌ Export missing correct Client value"
        echo "Actual line: $(grep "test-project" "$EXPORT_FILE")"
        exit 1
    fi
    
    if grep "test-project" "$EXPORT_FILE" | grep -q "1"; then
        echo "✅ Export contains correct Priority value"
    else
        echo "❌ Export missing correct Priority value"
        exit 1
    fi
    
    if grep "test-project" "$EXPORT_FILE" | grep -q "true"; then
        echo "✅ Export contains correct Billable value"
    else
        echo "❌ Export missing correct Billable value"
        exit 1
    fi
else
    echo "❌ No export file created"
    exit 1
fi

echo ""
echo "✅ All metadata export integration tests passed!"

# Clean up
rm -rf ~/.timetracker
rm -rf ~/.timetracker_data