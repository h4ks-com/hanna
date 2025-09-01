# Hanna IRC Bot Trigger System

The Hanna IRC bot now supports a flexible trigger system that can send IRC events to multiple n8n trigger endpoints or other webhook endpoints.

## Configuration

### Environment Variables

The bot can be configured using two methods:

1. **Legacy N8N_WEBHOOK** (backward compatibility)
   ```bash
   N8N_WEBHOOK=https://your-n8n.example.com/webhook/your-webhook-id
   ```

2. **New TRIGGER_CONFIG** (JSON format for multiple triggers)
   ```bash
   TRIGGER_CONFIG='{"endpoints": {"mentions": {"url": "https://n8n.example.com/webhook/mentions", "token": "your-auth-token", "events": ["mention"], "channels": ["#general", "#dev"]}, "all-events": {"url": "https://n8n.example.com/webhook/all", "token": "another-token", "events": ["privmsg", "join", "part", "quit", "kick", "mode", "topic", "notice", "nick"]}}}'
   ```

### TRIGGER_CONFIG Structure

```json
{
  "endpoints": {
    "endpoint-name": {
      "url": "https://webhook-url",
      "token": "authentication-token",
      "events": ["mention", "privmsg", "join", "part"],
      "channels": ["#channel1", "#channel2"],  // optional filter
      "users": ["user1", "user2"]              // optional filter
    }
  }
}
```

### Supported Events

- `mention` - When the bot is mentioned in a message
- `privmsg` - All private messages (including channel messages)
- `notice` - IRC notices received
- `join` - When someone joins a channel
- `part` - When someone leaves a channel
- `quit` - When someone quits the IRC server
- `kick` - When someone is kicked from a channel
- `mode` - Channel or user mode changes
- `nick` - When someone changes their nickname
- `topic` - When channel topic is changed

### Filters

- `channels`: Only trigger for events in specified channels (optional)
- `users`: Only trigger for events from specified users (optional)

## n8n Trigger Node

The n8n package includes a new "Hanna Bot Trigger" node that:

1. Provides a webhook endpoint
2. Validates authentication tokens
3. Filters events based on type, channels, and users
4. Passes IRC event data to your workflow

### Setting up the Trigger Node

1. Add the "Hanna Bot Trigger" node to your workflow
2. Configure the authentication token (this should match the token in TRIGGER_CONFIG)
3. Select which IRC events you want to monitor
4. Optionally filter by channels or users
5. Copy the webhook URL from the node
6. Add this URL and token to your TRIGGER_CONFIG

### Payload Structure

The trigger node receives this data structure:

```json
{
  "eventType": "mention",
  "sender": "username",
  "target": "#channel",
  "message": "Hello @botname how are you?",
  "chatInput": "Hello @botname how are you?",
  "botNick": "botname",
  "sessionId": "IRC",
  "timestamp": 1693526400,
  "messageTags": {
    "time": "2023-09-01T12:00:00.000Z"
  }
}
```

## Example Configurations

### Simple Mention Handling
```json
{
  "endpoints": {
    "mentions": {
      "url": "https://n8n.example.com/webhook/bot-mentions",
      "token": "secret-token-123",
      "events": ["mention"]
    }
  }
}
```

### Multi-Channel Monitoring
```json
{
  "endpoints": {
    "support-channels": {
      "url": "https://n8n.example.com/webhook/support",
      "token": "support-token",
      "events": ["mention", "privmsg"],
      "channels": ["#support", "#help"]
    },
    "moderation": {
      "url": "https://n8n.example.com/webhook/moderation",
      "token": "mod-token",
      "events": ["kick", "mode", "join", "part"],
      "channels": ["#general", "#offtopic"]
    }
  }
}
```

### User-Specific Monitoring
```json
{
  "endpoints": {
    "admin-actions": {
      "url": "https://n8n.example.com/webhook/admin",
      "token": "admin-token",
      "events": ["mode", "kick", "topic"],
      "users": ["admin1", "admin2", "moderator"]
    }
  }
}
```

## Migration from Legacy N8N_WEBHOOK

If you're currently using `N8N_WEBHOOK`, the bot will automatically create a legacy endpoint configuration that listens for "mention" events only. To take advantage of the new features, migrate to `TRIGGER_CONFIG` format:

**Old:**
```bash
N8N_WEBHOOK=https://n8n.example.com/webhook/abc123
```

**New:**
```bash
TRIGGER_CONFIG='{"endpoints": {"legacy": {"url": "https://n8n.example.com/webhook/abc123", "events": ["mention"]}}}'
```

## Security

- Always use HTTPS for webhook URLs in production
- Use strong, unique authentication tokens for each endpoint
- Consider using different tokens for different types of events
- The bot validates tokens using Bearer authentication or X-Auth-Token header
