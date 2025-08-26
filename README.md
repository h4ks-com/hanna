# Hanna - IRC Bot with REST API

A robust, self-contained Go IRC bot that connects over TLS and exposes a secure, token-authenticated REST API for remote control. Perfect for automation, monitoring, and integration with other services.

## 🚀 Features

- **Secure IRC Connection**: TLS-enabled IRC connections with optional server password support
- **SASL Authentication**: Optional SASL PLAIN authentication for IRC networks that require it
- **Auto-Reconnect**: Intelligent reconnection with exponential backoff for maximum uptime
- **REST API**: Token-protected HTTP/HTTPS endpoints for complete bot control
- **n8n Integration**: Automatic webhook calls when mentioned in IRC channels
- **Channel Management**: Join, part, and track channels programmatically
- **Message Control**: Send messages, notices, and raw IRC commands via API
- **Graceful Shutdown**: Clean disconnection and resource cleanup
- **Zero Dependencies**: Self-contained binary with no external dependencies
- **Production Ready**: Comprehensive logging, error handling, and monitoring endpoints

## 📋 Requirements

- Go 1.21 or later
- IRC server with TLS support
- (Optional) TLS certificates for HTTPS API

## 🛠️ Installation

### From Source

```bash
git clone https://github.com/h4ks-com/hanna
cd hanna
go mod tidy
go build -o hanna
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
| `N8N_WEBHOOK` | n8n webhook URL for chat integration | - | ❌ |

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

## 📝 Usage Examples

### Basic Bot Setup

```bash
#!/bin/bash

# Production setup with HTTPS and n8n integration
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
export N8N_WEBHOOK="https://n8n.example.com/webhook/irc-bot"

./hanna
```

### n8n Integration

When someone mentions the bot in IRC with `@botname message`, the bot will automatically call the configured n8n webhook with a JSON payload:

```json
{
  "sender": "username",
  "target": "#channel",
  "message": "hello bot",
  "fullMessage": "@botname hello bot",
  "botNick": "MyAwesomeBot",
  "timestamp": 1692345678
}
```

#### Example n8n Workflow Setup

1. Create a new workflow in n8n
2. Add a "Webhook" trigger node
3. Set the webhook URL in the `N8N_WEBHOOK` environment variable
4. Process the incoming IRC data and respond as needed
5. Optionally use the Hanna API to send responses back to IRC

**Mention Examples:**
- `@MyAwesomeBot what's the weather?` → Triggers webhook
- `@MyAwesomeBot help` → Triggers webhook  
- `MyAwesomeBot hello` → Does NOT trigger (missing @)
- `@DifferentBot hello` → Does NOT trigger (wrong bot name)

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

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o hanna

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/hanna .

EXPOSE 8080
CMD ["./hanna"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  hanna:
    build: .
    ports:
      - "8080:8080"
    environment:
      - API_TOKEN=your_secret_token
      - IRC_ADDR=irc.libera.chat:6697
      - IRC_NICK=DockerBot
      - AUTOJOIN=#general,#bots
    volumes:
      - ./certs:/certs:ro  # Mount certificates if using HTTPS
    restart: unless-stopped
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

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

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
