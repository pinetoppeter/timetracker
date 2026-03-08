#!/bin/bash
cd /app

# Clean up
rm -rf ~/.timetracker

# Create some existing named records
echo "Creating existing records..."
echo "" | ./ttr start "project-meeting" && \
echo "" | ./ttr switch "code-review" && \
echo "" | ./ttr switch "documentation"

# Create an unnamed record
echo "Creating unnamed record..."
./ttr start

# Test autocomplete by selecting suggestion #2
echo "Testing autocomplete with suggestion selection..."
echo "" | echo "2" | ./ttr end

# Check the result
echo "Final session state:"
echo "" | ./ttr info