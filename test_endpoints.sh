#!/bin/bash

# Test script for new LIST and WHOIS endpoints
# Make sure the bot is running and connected to IRC

API_TOKEN="${API_TOKEN:-test_token}"
API_BASE="${API_BASE:-http://localhost:8080}"

echo "Testing Hanna Bot API endpoints..."
echo "API Base: $API_BASE"
echo "Token: $API_TOKEN"
echo ""

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s "$API_BASE/health" | jq .
echo ""

# Test bot state
echo "2. Testing bot state..."
curl -s -H "Authorization: Bearer $API_TOKEN" "$API_BASE/api/state" | jq .
echo ""

# Test LIST command
echo "3. Testing LIST channels..."
curl -s -H "Authorization: Bearer $API_TOKEN" "$API_BASE/api/list" | jq .
echo ""

# Test WHOIS command (replace 'testuser' with an actual nick on your IRC network)
echo "4. Testing WHOIS command..."
read -p "Enter a nick to WHOIS (or press Enter to skip): " nick
if [ ! -z "$nick" ]; then
    curl -s -H "Authorization: Bearer $API_TOKEN" \
         -H "Content-Type: application/json" \
         -d "{\"nick\": \"$nick\"}" \
         "$API_BASE/api/whois" | jq .
else
    echo "Skipped WHOIS test"
fi
echo ""

echo "Test completed!"
