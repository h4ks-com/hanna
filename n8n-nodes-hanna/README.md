# n8n-nodes-hanna

This is an n8n community node for integrating with the Hanna IRC Bot API.

[Hanna](https://github.com/h4ks-com/hanna) is a self-contained Go IRC bot that connects over TLS and exposes a token-authenticated REST API for remote control.

## Installation

Follow the [installation guide](https://docs.n8n.io/integrations/community-nodes/installation/) in the n8n community nodes documentation.

## Operations

The Hanna Bot node supports the following operations:

- **Send Message**: Send a message to a channel or user
- **Send Notice**: Send a notice to a channel or user  
- **Join Channel**: Join an IRC channel
- **Part Channel**: Leave an IRC channel
- **Get Status**: Get bot connection status and channels
- **Change Nick**: Change bot nickname
- **Send Raw Command**: Send raw IRC command

## Credentials

You need to configure the following credentials:

- **API URL**: Base URL of your Hanna Bot instance (e.g., `https://bot.example.com`)
- **API Token**: Bearer token for API authentication (set as `API_TOKEN` when running Hanna Bot)

## Usage with AI Agents

This node is marked as `usableAsTool: true`, which means AI agents can automatically use it to:

- Send messages to IRC channels based on chat interactions
- Respond to IRC mentions by calling webhooks
- Manage channel membership
- Send automated notifications

## Example Workflow

1. **Webhook Trigger**: Receives IRC mention from Hanna Bot
2. **Function Node**: Process the message and generate response
3. **Hanna Bot Node**: Send response back to IRC channel

```javascript
// Example function node processing IRC mention
const message = $json.message.toLowerCase();
const sender = $json.sender;
const target = $json.target;

if (message.includes('weather')) {
  return {
    target: target,
    message: `@${sender} Sorry, I don't have weather data yet!`
  };
}
```

## Compatibility

This node is compatible with n8n version 1.0 and above.

## License

MIT