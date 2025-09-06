#!/bin/bash
set -e

echo "🚀 Starting n8n with auto-import..."

# Start n8n in the background
/docker-entrypoint.sh "$@" &
N8N_PID=$!

# Wait a bit for n8n to start, then run import
(
    sleep 10
    /auto-import.sh
) &

# Wait for n8n process
wait $N8N_PID