#!/bin/bash
cd /home/peter/repos/go/timetracker

# Clean up
rm -rf ~/.timetracker

# Start a session with an unnamed record
echo "Starting session..."
./ttr start

# Create another unnamed record by switching without a name
echo "Creating unnamed record..."
printf "\n" | ./ttr switch

# Check current state
echo "Current session state:"
./ttr info

# Now test the end command with input
echo "Testing end command with input..."
printf "test-name-1\ntest-name-2\n" | ./ttr end

# Check final state
echo "Final session state:"
./ttr info