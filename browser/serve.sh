#!/bin/bash

# Start a local web server to test the browser
# Usage: ./serve.sh

PORT=8080
DIR="testpage"

echo "Starting web server..."
echo "Open browser with: go run . http://localhost:$PORT"
echo "Press Ctrl+C to stop"
echo ""

cd "$DIR"
python3 -m http.server $PORT
