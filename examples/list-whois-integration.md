# IRC Channel List and WHOIS Integration Example

This example demonstrates how to use the new LIST and WHOIS endpoints to gather IRC network information.

## Channel Discovery and User Analysis

```bash
#!/bin/bash

# Configuration
API_TOKEN="your_secure_token_here"
API_BASE="http://localhost:8080"

# Function to make authenticated API calls
api_call() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    
    if [ "$method" = "GET" ]; then
        curl -s -H "Authorization: Bearer $API_TOKEN" "$API_BASE$endpoint"
    else
        curl -s -H "Authorization: Bearer $API_TOKEN" \
             -H "Content-Type: application/json" \
             -d "$data" \
             -X "$method" \
             "$API_BASE$endpoint"
    fi
}

# Get list of all channels on the IRC network
echo "Fetching channel list..."
CHANNELS=$(api_call "GET" "/api/list")
echo "$CHANNELS" | jq '.channels[] | select(.users | tonumber > 10) | .channel'

# Get information about a specific user
echo "Getting user information..."
USER_INFO=$(api_call "POST" "/api/whois" '{"nick": "someuser"}')
echo "$USER_INFO" | jq '{nick, user, host, real_name, channels}'

# Example: Find channels with more than 50 users
echo "Large channels (>50 users):"
echo "$CHANNELS" | jq '.channels[] | select(.users | tonumber > 50) | {channel, users, topic}'
```

## n8n Workflow Integration

### Channel Monitoring Workflow

1. **Trigger**: Use Hanna Bot Trigger node for "join" events
2. **List Channels**: Use HTTP Request node to call `/api/list`
3. **User Info**: Use HTTP Request node to call `/api/whois` for new users
4. **Action**: Send welcome message or log user activity

```json
{
  "nodes": [
    {
      "name": "Hanna Bot Trigger",
      "type": "n8n-nodes-hanna.hannaBotTrigger",
      "parameters": {
        "events": ["join"],
        "channels": ["#general", "#welcome"]
      }
    },
    {
      "name": "Get User Info",
      "type": "n8n-nodes-base.httpRequest",
      "parameters": {
        "url": "http://localhost:8080/api/whois",
        "method": "POST",
        "headers": {
          "Authorization": "Bearer YOUR_TOKEN"
        },
        "body": {
          "nick": "={{$json.sender}}"
        }
      }
    },
    {
      "name": "Welcome Message",
      "type": "n8n-nodes-hanna.hannaBot",
      "parameters": {
        "action": "send",
        "target": "={{$json.target}}",
        "message": "Welcome {{$json.sender}}! You're from {{$node['Get User Info'].json.host}}"
      }
    }
  ]
}
```

## Use Cases

### 1. Channel Discovery
- Find active channels by user count
- Discover channels by topic keywords
- Monitor channel growth over time

### 2. User Authentication
- Verify user hostmasks for access control
- Check if users are IRC operators
- Identify user's other channels for context

### 3. Moderation Tools
- Check new users' information before allowing access
- Identify potential spam accounts by hostname patterns
- Monitor operator status changes

### 4. Analytics
- Track channel popularity over time
- Analyze user distribution across channels
- Generate network statistics reports

## Rate Limiting Considerations

Both LIST and WHOIS commands can be resource-intensive on IRC servers:

- **LIST**: Some networks limit how often you can request the full channel list
- **WHOIS**: Multiple rapid WHOIS requests may trigger flood protection

Best practices:
- Cache LIST results for several minutes
- Implement request queuing for multiple WHOIS requests
- Use appropriate delays between requests
- Monitor for IRC server error responses

## Error Handling

The endpoints return appropriate HTTP status codes:

- `200`: Success
- `400`: Bad request (missing nick for WHOIS)
- `500`: IRC command timeout or server error
- `503`: Bot not connected to IRC

Example error response:
```json
{
  "error": "whois request failed: request timed out"
}
```
