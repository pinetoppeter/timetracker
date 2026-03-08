#!/bin/bash
cd /app

# Clean up
rm -rf ~/.timetracker

# Start a session with an unnamed record
echo "Starting session..."
echo "" | ./ttr start

# Switch to create another unnamed record
echo "Switching to create unnamed record..."
echo "" | ./ttr switch

# Try to end the session and name the records
echo "Ending session and naming records..."
echo -e "unnamed-record-1\nunnamed-record-2" | ./ttr end

# Check the session info
echo "Checking session info..."
echo "" | ./ttr info