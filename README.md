# Hanna - IRC Bot with REST API

[![npm version](https://badge.fury.io/js/n8n-nodes-hanna.svg)](https://www.npmjs.com/package/n8n-nodes-hanna)

A robust, self-contained Go IRC bot that connects over TLS and exposes a secure, token-authenticated REST API for remote control. Perfect for automation, monitoring, and integration with other services.

## 🚀 Features

- **Secure IRC Connection**: TLS-enabled IRC connections with optional server password support
- **SASL Authentication**: Optional SASL PLAIN authentication for IRC networks that require it
- **Auto-Reconnect**: Intelligent reconnection with exponential backoff for maximum uptime
- **REST API**: Token-protected HTTP/HTTPS endpoints for complete bot control
- **IRC Network Discovery**: List channels and get user information via LIST and WHOIS commands
- **Flexible Event System**: Advanced trigger system supporting multiple IRC events (mentions, joins, parts, mode changes, etc.)
- **n8n Integration**: Comprehensive n8n node package with both action and trigger nodes
- **Multiple Webhooks**: Support for multiple trigger endpoints with filtering and authentication
- **Channel Management**: Join, part, and track channels programmatically
- **Message Control**: Send messages, notices, and raw IRC commands via API
- **Graceful Shutdown**: Clean disconnection and resource cleanup
- **Zero Dependencies**: Self-contained binary with no external dependencies
- **Production Ready**: Comprehensive logging, error handling, and monitoring endpoints

## 📋 Requirements

- Go 1.24 or later
- IRC server with TLS support
- (Optional) TLS certificates for HTTPS API

## 🛠️ Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/h4ks-com/hanna
cd hanna

# Build from source (no need for go mod tidy - dependencies are all local)
go build -o hanna .

# Or build with specific flags for production
CGO_ENABLED=0 go build -ldflags="-w -s" -o hanna .
```

### Docker Build

```bash
# Build using Docker
docker build -t hanna-bot .

# Or using docker-compose
docker-compose build hanna-bot
```

### Project Structure

The project is organized as a proper Go module:
```
hanna/
├── main.go              # Main application entry point
├── irc/                 # IRC client package
│   ├── client.go        # IRC client implementation
│   └── *_test.go        # Tests for IRC functionality
├── go.mod               # Go module definition
└── Dockerfile           # Docker build configuration
```

### Quick Start

```bash
# Basic setup with HTTP API
API_TOKEN=your_secret_token \
IRC_ADDR=irc.libera.chat:6697 \
IRC_NICK=YourBotName \
./hanna
```

## ⚙️ Configuration

All configuration is done via environment variables:

### IRC Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `IRC_ADDR` | IRC server address (host:port) | - | ✅ |
| `IRC_TLS` | Enable TLS connection | `1` | ❌ |
| `IRC_TLS_INSECURE` | Skip TLS certificate verification | `0` | ❌ |
| `IRC_PASS` | Server password | - | ❌ |
| `IRC_NICK` | Bot nickname | `goircbot` | ❌ |
| `IRC_USER` | Username/ident | `goircbot` | ❌ |
| `IRC_NAME` | Real name/GECOS | `Go IRC Bot` | ❌ |
| `SASL_USER` | SASL authentication username | - | ❌ |
| `SASL_PASS` | SASL authentication password | - | ❌ |
| `AUTOJOIN` | Comma-separated channels to auto-join | - | ❌ |

### API Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `API_ADDR` | HTTP/HTTPS listen address | `:8080` | ❌ |
| `API_TOKEN` | Bearer token for API authentication | - | ⚠️ |
| `API_TLS` | Enable HTTPS | `0` | ❌ |
| `API_CERT` | Path to TLS certificate file | - | ⚠️* |
| `API_KEY` | Path to TLS private key file | - | ⚠️* |

### n8n Integration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `N8N_WEBHOOK` | Legacy webhook URL for chat integration | - | ❌ |
| `TRIGGER_CONFIG` | JSON configuration for multiple trigger endpoints | - | ❌ |

### Event Trigger Configuration

The bot supports a flexible trigger system for sending IRC events to multiple endpoints:

**Legacy Mode (backward compatible):**
```bash
export N8N_WEBHOOK="https://n8n.example.com/webhook/irc-bot"
```

**Advanced Mode (recommended):**
```bash
export TRIGGER_CONFIG='{
  "endpoints": {
    "mentions": {
      "url": "https://n8n.example.com/webhook/mentions",
      "token": "secret-token-123",
      "events": ["mention"],
      "channels": ["#support", "#general"]
    },
    "moderation": {
      "url": "https://n8n.example.com/webhook/moderation", 
      "token": "mod-token-456",
      "events": ["join", "part", "kick", "mode"],
      "channels": ["#general"]
    }
  }
}'
```

**Supported Event Types:**
- `mention` - When the bot is mentioned
- `privmsg` - All channel/private messages
- `join` - User joins a channel
- `part` - User leaves a channel  
- `quit` - User quits IRC
- `kick` - User is kicked from channel
- `mode` - Mode changes (op, voice, etc.)
- `nick` - Nickname changes
- `topic` - Channel topic changes
- `notice` - IRC notices

*Required when `API_TLS=1`  
⚠️ Highly recommended for security

## 🔒 HTTPS Setup

### Using Let's Encrypt

If you have Let's Encrypt certificates for your domain:

```bash
export API_TLS=1
export API_CERT=/etc/letsencrypt/live/yourdomain.com/fullchain.pem
export API_KEY=/etc/letsencrypt/live/yourdomain.com/privkey.pem
```

### Self-Signed Certificates (Development)

```bash
# Generate self-signed certificate
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Use with bot
export API_TLS=1
export API_CERT=cert.pem
export API_KEY=key.pem
```

## 🌐 API Reference

### Authentication

All API endpoints (except `/health`) require a Bearer token:

```bash
curl -H "Authorization: Bearer your_secret_token" https://your-server:8080/api/state
```

### Endpoints

#### Health Check
```http
GET /health
```
Returns bot connection status and current nickname.

**Response:**
```json
{
  "ok": true,
  "nick": "YourBot"
}
```

#### Bot State
```http
GET /api/state
Authorization: Bearer <token>
```
Returns comprehensive bot status including connected channels.

**Response:**
```json
{
  "connected": true,
  "nick": "YourBot",
  "channels": ["#general", "#bots"]
}
```

#### Join Channel
```http
POST /api/join
Authorization: Bearer <token>
Content-Type: application/json

{
  "channel": "#example"
}
```

#### Leave Channel
```http
POST /api/part
Authorization: Bearer <token>
Content-Type: application/json

{
  "channel": "#example",
  "reason": "Goodbye!"
}
```

#### Send Message
```http
POST /api/send
Authorization: Bearer <token>
Content-Type: application/json

{
  "target": "#example",
  "message": "Hello, world!"
}
```

#### Send Notice
```http
POST /api/notice
Authorization: Bearer <token>
Content-Type: application/json

{
  "target": "#example",
  "message": "This is a notice"
}
```

#### Change Nickname
```http
POST /api/nick
Authorization: Bearer <token>
Content-Type: application/json

{
  "nick": "NewBotName"
}
```

#### Send Raw IRC Command
```http
POST /api/raw
Authorization: Bearer <token>
Content-Type: application/json

{
  "line": "PRIVMSG #channel :Custom message"
}
```

#### List IRC Channels
```http
GET /api/list
Authorization: Bearer <token>
```

Response:
```json
{
  "channels": [
    {
      "channel": "#general",
      "users": "42",
      "topic": "Welcome to the general discussion channel"
    },
    {
      "channel": "#bots",
      "users": "15", 
      "topic": "Bot testing and development"
    }
  ],
  "count": 2
}
```

#### Get User Information (WHOIS)
```http
POST /api/whois
Authorization: Bearer <token>
Content-Type: application/json

{
  "nick": "username"
}
```

Response:
```json
{
  "nick": "username",
  "user": "user",
  "host": "example.com",
  "real_name": "Real Name",
  "server": "irc.libera.chat",
  "server_info": "Stockholm, Sweden",
  "operator": false,
  "idle_seconds": "42",
  "idle_info": "seconds idle",
  "channels": "@#ops +#general #random",
  "raw_data": [
    {
      "type": "user",
      "nick": "username",
      "user": "user",
      "host": "example.com",
      "real_name": "Real Name"
    }
  ]
}
```

## 📝 Usage Examples

### Basic Bot Setup

```bash
#!/bin/bash

# Production setup with HTTPS and advanced n8n trigger integration
export API_TOKEN="$(openssl rand -hex 32)"
export API_TLS=1
export API_CERT="/etc/letsencrypt/live/bot.example.com/fullchain.pem"
export API_KEY="/etc/letsencrypt/live/bot.example.com/privkey.pem"
export API_ADDR=":443"

export IRC_ADDR="irc.libera.chat:6697"
export IRC_NICK="MyAwesomeBot"
export IRC_USER="mybot"
export IRC_NAME="My Awesome IRC Bot"
export AUTOJOIN="#general,#bots"

# Advanced trigger configuration for multiple n8n workflows
export TRIGGER_CONFIG='{
  "endpoints": {
    "mentions": {
      "url": "https://n8n.example.com/webhook/bot-mentions",
      "token": "mention-token-123",
      "events": ["mention"],
      "channels": ["#general", "#support"]
    },
    "moderation": {
      "url": "https://n8n.example.com/webhook/moderation",
      "token": "mod-token-456", 
      "events": ["join", "part", "kick", "mode"],
      "channels": ["#general"]
    },
    "private-messages": {
      "url": "https://n8n.example.com/webhook/private-chat",
      "token": "private-token-789",
      "events": ["privmsg"],
      "channels": []
    }
  }
}'

./hanna
```

### n8n Integration & Event Triggers

The bot provides comprehensive integration with n8n through both legacy webhooks and the new trigger system.

#### Legacy Webhook Integration (Simple)

When someone mentions the bot in IRC with `@botname message`, the bot will automatically call the configured n8n webhook:

```json
{
  "eventType": "mention",
  "sender": "username",
  "target": "#channel", 
  "message": "hello bot",
  "chatInput": "@botname hello bot",
  "botNick": "MyAwesomeBot",
  "sessionId": "IRC",
  "timestamp": 1692345678,
  "messageTags": {
    "time": "2023-09-01T12:00:00.000Z"
  }
}
```

#### Advanced Trigger System (Recommended)

The new trigger system supports multiple IRC events and endpoints:

**1. Install the n8n Node Package**

```bash
# In your n8n installation
npm install n8n-nodes-hanna
```

**2. Create Trigger Workflows**

- Add "Hanna Bot Trigger" nodes to your workflows
- Configure authentication tokens and event filters
- Copy webhook URLs to your `TRIGGER_CONFIG`

**3. Event Examples:**

**User Joins Channel:**
```json
{
  "eventType": "join",
  "sender": "newuser",
  "target": "#general",
  "message": "",
  "botNick": "MyBot",
  "timestamp": 1692345678
}
```

**Channel Mode Change:**
```json
{
  "eventType": "mode", 
  "sender": "operator",
  "target": "#general",
  "message": "Mode #general +o newop",
  "timestamp": 1692345678
}
```

**Private Message:**
```json
{
  "eventType": "privmsg",
  "sender": "user123",
  "target": "MyBot",
  "message": "Hello bot, how are you?",
  "chatInput": "Hello bot, how are you?",
  "timestamp": 1692345678
}
```

#### Example n8n Workflow Setup

**Using the New Trigger Node (Recommended):**

1. Install `n8n-nodes-hanna` package in your n8n instance
2. Create a new workflow in n8n
3. Add a "Hanna Bot Trigger" node
4. Configure:
   - Authentication token (create a secure random token)
   - Select events you want to monitor (mention, join, part, etc.)
   - Optional: Filter by channels or users
5. Copy the webhook URL from the trigger node
6. Add the URL and token to your bot's `TRIGGER_CONFIG`
7. Add processing nodes (OpenAI, database, notifications, etc.)
8. Optionally add "Hanna Bot" action node to send responses back to IRC

**Using Legacy Webhook:**

1. Create a new workflow in n8n
2. Add a "Webhook" trigger node  
3. Set the webhook URL in the `N8N_WEBHOOK` environment variable
4. Process the incoming IRC data and respond as needed
5. Optionally use the Hanna API to send responses back to IRC

#### Advanced Workflow Examples

**AI-Powered Chat Bot:**
```
Hanna Bot Trigger (mention events) 
  → OpenAI Chat Completion
  → Hanna Bot Send Message
```

**Channel Moderation:**
```
Hanna Bot Trigger (kick, mode events)
  → Log to Database
  → Send Alert if Spam Detected
  → Auto-Ban Repeat Offenders
```

**Welcome System:**
```  
Hanna Bot Trigger (join events)
  → Check User Database
  → Send Welcome Message for New Users
  → Assign Roles Based on Rules
```

**Multi-Channel Bot:**
```
Trigger 1: Support Channel (mention, privmsg in #support)
  → Route to Support AI
  → Log Support Ticket
  
Trigger 2: General Chat (mention in #general)  
  → Route to Fun AI
  → Random Responses
```

**Mention Examples:**
- `@MyAwesomeBot what's the weather?` → Triggers mention event
- `@MyAwesomeBot help` → Triggers mention event  
- `MyAwesomeBot hello` → Does NOT trigger mention (missing @)
- `@DifferentBot hello` → Does NOT trigger mention (wrong bot name)

**Other Event Examples:**
- User joins `#general` → Triggers join event  
- User types `hello everyone` in `#general` → Triggers privmsg event
- Operator gives voice to user → Triggers mode event
- User gets kicked from channel → Triggers kick event

### Send Message via API

```bash
# Send a message to a channel
curl -X POST \
  -H "Authorization: Bearer your_secret_token" \
  -H "Content-Type: application/json" \
  -d '{"target": "#general", "message": "Hello from the API!"}' \
  https://bot.example.com/api/send

# Check bot status
curl -H "Authorization: Bearer your_secret_token" \
  https://bot.example.com/api/state

# Join a new channel
curl -X POST \
  -H "Authorization: Bearer your_secret_token" \
  -H "Content-Type: application/json" \
  -d '{"channel": "#newchannel"}' \
  https://bot.example.com/api/join
```

### Integration Examples

#### Python Integration

```python
import requests
import json

class IRCBotAPI:
    def __init__(self, base_url, token):
        self.base_url = base_url.rstrip('/')
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
    
    def send_message(self, target, message):
        data = {'target': target, 'message': message}
        response = requests.post(
            f'{self.base_url}/api/send',
            headers=self.headers,
            json=data
        )
        return response.json()
    
    def get_status(self):
        response = requests.get(
            f'{self.base_url}/api/state',
            headers=self.headers
        )
        return response.json()

# Usage
bot = IRCBotAPI('https://bot.example.com', 'your_secret_token')
bot.send_message('#general', 'Hello from Python!')
status = bot.get_status()
print(f"Bot is {'connected' if status['connected'] else 'disconnected'}")
```

#### Shell Script Integration

```bash
#!/bin/bash

BOT_URL="https://bot.example.com"
BOT_TOKEN="your_secret_token"

send_irc_message() {
    local channel="$1"
    local message="$2"
    
    curl -s -X POST \
        -H "Authorization: Bearer $BOT_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"target\": \"$channel\", \"message\": \"$message\"}" \
        "$BOT_URL/api/send"
}

# Send system status to IRC
LOAD=$(uptime | awk -F'load average:' '{print $2}')
send_irc_message "#monitoring" "Server load: $LOAD"
```

## 🐳 Docker Deployment

### Quick Start with Docker Compose

The easiest way to run Hanna is using Docker Compose with the included configuration:

1. **Copy environment template:**
   ```bash
   cp .env.example .env
   ```

2. **Configure your bot:**
   Edit `.env` and set at least:
   ```bash
   IRC_ADDR=irc.libera.chat:6697
   IRC_NICK=YourBotName
   API_TOKEN=your-secure-token-here
   AUTOJOIN=#your-channels
   ```

3. **Start the services:**
   ```bash
   docker compose up -d
   ```

This will start both Hanna IRC bot and n8n for workflow automation. The bot will be available at `http://localhost:8080` and n8n at `http://localhost:5678`.

### Docker Compose Configuration

The included `compose.yaml` provides:

- **Hanna IRC Bot**: Main bot service with API
- **n8n**: Workflow automation platform for chat processing
- **n8n Workflow Loader**: Sidecar container that automatically loads workflows on startup
- **Automatic networking**: Bot can communicate with n8n via `http://n8n:5678`
- **Persistent storage**: n8n workflows and data are preserved
- **Environment-based config**: All settings from `.env` file

### Automatic Workflow Loading

The Docker Compose setup includes a workflow loader that automatically imports workflows into n8n:

1. **Default Echo Bot**: A pre-configured workflow that echoes IRC messages
2. **Auto-Import**: Workflows in `n8n-workflow-loader/workflows/` are loaded on startup
3. **Skip Existing**: Already-imported workflows are detected and skipped
4. **Auto-Activation**: Imported workflows are automatically activated
5. **Trigger Integration**: The bot's `TRIGGER_CONFIG` is automatically configured

**Adding Custom Workflows:**
1. Place workflow JSON files in `n8n-workflow-loader/workflows/`
2. Remove the `id` field from the workflow JSON
3. Restart with `docker-compose up -d`

The workflow loader uses the n8n REST API and includes error handling and duplicate detection.

### Manual Docker Build

If you prefer to build and run manually:

```bash
# Build the image
docker build -t hanna-bot .

# Run with environment variables
docker run -d \
  --name hanna \
  -p 8080:8080 \
  -e API_TOKEN=your_secret_token \
  -e IRC_ADDR=irc.libera.chat:6697 \
  -e IRC_NICK=DockerBot \
  -e AUTOJOIN="#general,#bots" \
  hanna-bot
```

## 🔧 Systemd Service

Create `/etc/systemd/system/hanna.service`:

```ini
[Unit]
Description=Hanna IRC Bot
After=network.target

[Service]
Type=simple
User=ircd
Group=ircd
WorkingDirectory=/opt/hanna
ExecStart=/opt/hanna/hanna
EnvironmentFile=/opt/hanna/.env
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Environment file `/opt/hanna/.env`:

```bash
API_TOKEN=your_secret_token
API_TLS=1
API_CERT=/etc/letsencrypt/live/bot.example.com/fullchain.pem
API_KEY=/etc/letsencrypt/live/bot.example.com/privkey.pem
IRC_ADDR=irc.libera.chat:6697
IRC_NICK=HannaBot
AUTOJOIN=#general

# Advanced trigger configuration
TRIGGER_CONFIG={"endpoints":{"mentions":{"url":"https://n8n.example.com/webhook/mentions","token":"secure-token-123","events":["mention"],"channels":["#general","#support"]},"moderation":{"url":"https://n8n.example.com/webhook/moderation","token":"mod-token-456","events":["join","part","kick","mode"],"channels":["#general"]}}}
```

Enable and start:

```bash
sudo systemctl enable hanna
sudo systemctl start hanna
sudo systemctl status hanna
```

## 📊 Monitoring

### Health Check Endpoint

The `/health` endpoint provides a simple way to monitor bot status:

```bash
# Check if bot is connected
curl -f https://bot.example.com/health || echo "Bot is down!"
```

### Prometheus Metrics (Future Enhancement)

Consider extending the bot with Prometheus metrics for comprehensive monitoring.

## 🛡️ Security Considerations

1. **API Token**: Use a strong, randomly generated token
2. **HTTPS**: Always use HTTPS in production
3. **Firewall**: Restrict API access to trusted networks
4. **File Permissions**: Protect certificate files (0600 permissions)
5. **Regular Updates**: Keep dependencies and certificates updated

## 📦 n8n Node Package

Hanna includes a comprehensive n8n node package (`n8n-nodes-hanna`) that provides both action and trigger nodes for complete IRC automation.

### Installation

```bash
# Install in your n8n instance
npm install n8n-nodes-hanna

# Or install globally
npm install -g n8n-nodes-hanna
```

### Available Nodes

#### 1. Hanna Bot (Action Node)
- Send messages and notices to IRC channels
- Join and part channels programmatically  
- Change bot nickname
- Send raw IRC commands
- Get bot status and connected channels

#### 2. Hanna Bot Trigger (Trigger Node)
- Webhook-based trigger for IRC events
- Configurable event filtering (mention, join, part, etc.)
- Channel and user filtering
- Authentication token validation
- Support for multiple simultaneous triggers

### Node Configuration

#### Action Node Setup:
1. Add "Hanna Bot" node to your workflow
2. Configure credentials:
   - API URL: `https://your-bot.example.com`
   - API Token: Your bot's API token
3. Select operation (Send Message, Join Channel, etc.)
4. Configure target and message parameters

#### Trigger Node Setup:
1. Add "Hanna Bot Trigger" node to your workflow
2. Set authentication token (secure random string)
3. Select IRC events to monitor
4. Optional: Set channel/user filters
5. Copy webhook URL to bot's `TRIGGER_CONFIG`

### Example Workflows

See the [Advanced Workflow Examples](#advanced-workflow-examples) section above for complete workflow configurations using both trigger and action nodes.

## 🧪 Testing

```bash
# Run all tests
go test -v
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Run `go test -v` to ensure all tests pass
6. Submit a pull request

## 📄 License

This project is licensed under the GPLv3 License - see the LICENSE file for details.

## 🐛 Troubleshooting

### Common Issues

**Bot won't connect to IRC:**
- Check `IRC_ADDR` format (must include port)
- Verify TLS settings match server requirements
- Check firewall/network connectivity

**API returns 401 Unauthorized:**
- Verify `API_TOKEN` is set and matches request
- Check `Authorization` header format: `Bearer <token>`

**HTTPS certificate errors:**
- Verify certificate and key file paths
- Check file permissions
- Ensure certificate is valid and not expired

**Bot keeps reconnecting:**
- Check IRC server logs for connection issues
- Verify SASL credentials if using authentication
- Check for nick conflicts

### Debug Mode

For detailed logging, you can modify the log level in the source code or redirect output:

```bash
./hanna 2>&1 | tee bot.log
```

---

*Built with ❤️ in Go*
