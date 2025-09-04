#!/bin/bash

# Test script for comprehensive IRC state tracking
echo "Testing IRC Comprehensive State Tracking"

# Set up test environment
export IRC_ADDR="chat.freenode.net:6697"
export IRC_TLS="1"
export IRC_NICK="HannaTestBot$$"
export IRC_USER="HannaTest"
export IRC_NAME="Hanna Test Bot"
export API_TOKEN="test-token-123"
export API_PORT="8081"

# Start the bot in background (simulated test mode)
echo "Starting enhanced bot..."
timeout 10s ./hanna &
BOT_PID=$!
sleep 2

# Test the new API endpoints
echo "Testing new API endpoints..."

# Test server info endpoint
echo "1. Testing /api/server endpoint..."
curl -s -H "Authorization: Bearer test-token-123" http://localhost:8081/api/server | jq '.' > server_info.json 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Server info endpoint working"
else
    echo "✗ Server info endpoint failed"
fi

# Test comprehensive state endpoint
echo "2. Testing /api/comprehensive-state endpoint..."
curl -s -H "Authorization: Bearer test-token-123" http://localhost:8081/api/comprehensive-state | jq '.' > comprehensive_state.json 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Comprehensive state endpoint working"
else
    echo "✗ Comprehensive state endpoint failed"
fi

# Test users endpoint
echo "3. Testing /api/users endpoint..."
curl -s -H "Authorization: Bearer test-token-123" http://localhost:8081/api/users | jq '.' > users.json 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Users endpoint working"
else
    echo "✗ Users endpoint failed"
fi

# Test stats endpoint
echo "4. Testing /api/stats endpoint..."
curl -s -H "Authorization: Bearer test-token-123" http://localhost:8081/api/stats | jq '.' > stats.json 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Stats endpoint working"
else
    echo "✗ Stats endpoint failed"
fi

# Test errors endpoint
echo "5. Testing /api/errors endpoint..."
curl -s -H "Authorization: Bearer test-token-123" http://localhost:8081/api/errors | jq '.' > errors.json 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Errors endpoint working"
else
    echo "✗ Errors endpoint failed"
fi

# Cleanup
kill $BOT_PID 2>/dev/null
wait $BOT_PID 2>/dev/null

echo "Test completed. Check the generated JSON files for detailed output."
echo "Generated files:"
ls -la *.json 2>/dev/null | grep -E "\.(json)$" || echo "No JSON files generated"
