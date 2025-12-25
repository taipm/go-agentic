#!/bin/bash

# Start the program in background
./quiz-exam &
PID=$!

# Wait up to 15 seconds
sleep 15

# Check if process is still running
if kill -0 $PID 2>/dev/null; then
    echo "ERROR: Process still running after 15 seconds - DEADLOCK DETECTED"
    kill -9 $PID
    exit 1
else
    echo "SUCCESS: Process completed normally"
    wait $PID
    exit $?
fi
