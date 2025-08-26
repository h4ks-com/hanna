# n8n Integration Example

This example demonstrates how to set up n8n to receive IRC mentions from Hanna and respond back.

## Webhook Payload Structure

When someone mentions your bot in IRC like `@hanna what's the weather?`, the bot sends this JSON payload to your n8n webhook:

```json
{
  "sender": "alice",
  "target": "#general", 
  "message": "what's the weather?",
  "fullMessage": "@hanna what's the weather?",
  "botNick": "hanna",
  "timestamp": 1692345678
}
```

## Basic n8n Workflow

1. **Webhook Trigger**: Receives the IRC mention data
2. **Function Node**: Process the message and generate a response
3. **HTTP Request**: Send response back to Hanna's API

### Example n8n Function Node

```javascript
// Extract the message content
const message = $json.message.toLowerCase();
const sender = $json.sender;
const target = $json.target;

let response = "";

// Simple command handling
if (message.includes("weather")) {
  response = `@${sender} Sorry, I don't have weather data yet!`;
} else if (message.includes("help")) {
  response = `@${sender} Available commands: weather, help, time`;
} else if (message.includes("time")) {
  response = `@${sender} Current time: ${new Date().toISOString()}`;
} else {
  response = `@${sender} I heard you! You said: ${message}`;
}

// Return data for the HTTP request node
return {
  json: {
    target: target,
    message: response,
    originalSender: sender,
    originalMessage: message
  }
};
```

### Example HTTP Request Node Configuration

- **Method**: POST
- **URL**: `https://your-bot-domain.com/api/send`
- **Headers**: 
  - `Authorization`: `Bearer your_secret_token`
  - `Content-Type`: `application/json`
- **Body**:
```json
{
  "target": "{{ $json.target }}",
  "message": "{{ $json.message }}"
}
```

## Advanced Example: ChatGPT Integration

```javascript
// Function node to prepare ChatGPT request
const message = $json.message;
const sender = $json.sender;

return {
  json: {
    model: "gpt-3.5-turbo",
    messages: [
      {
        role: "system", 
        content: "You are a helpful IRC bot. Keep responses brief and friendly."
      },
      {
        role: "user",
        content: message
      }
    ],
    max_tokens: 150,
    originalTarget: $json.target,
    originalSender: sender
  }
};
```

Then add an OpenAI node and finally an HTTP Request node to send the ChatGPT response back to IRC.

## Environment Setup

```bash
export N8N_WEBHOOK="https://your-n8n-instance.com/webhook/irc-hanna"
export API_TOKEN="your_secret_bot_token"
```

## Testing

1. Start your bot with the n8n webhook configured
2. Join a channel where the bot is present  
3. Type: `@hanna hello there`
4. Check your n8n workflow execution log
5. Verify the bot responds in IRC

## Security Notes

- Use HTTPS for both your bot API and n8n webhook
- Keep your API tokens secure
- Consider rate limiting in your n8n workflow
- Validate incoming webhook data in n8n
