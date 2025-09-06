# HannaUI - Docker Integration Complete! 🐳

## ✅ **What We've Built**

You now have a complete Docker-based deployment system for HannaUI that integrates with your n8n Hanna AI endpoint!

### 🚀 **Key Features Added**

1. **🐳 Full Docker Support**
   - Lightweight nginx-alpine container
   - Automatic environment variable injection
   - Health checks and security headers
   - Production-ready configuration

2. **🔗 n8n Integration**
   - Environment variable configuration: `HANNA_N8N_ENDPOINT`
   - Automatic fallback to demo mode if no endpoint
   - Real-time connection status indicator
   - Comprehensive error handling

3. **⚙️ Easy Configuration**
   - `.env` file for environment variables
   - `docker-compose.yml` for simple deployment
   - `start.sh` interactive script for management
   - Multiple deployment options

## 🚀 **Quick Deployment**

### **Option 1: Docker Compose (Recommended)**
```bash
# 1. Configure environment
cp .env.example .env
nano .env  # Set HANNA_N8N_ENDPOINT

# 2. Deploy
docker-compose up -d

# 3. Access
# Visit http://localhost:3000
# Login: admin / admin123
```

### **Option 2: Interactive Script**
```bash
# Run the interactive setup
./start.sh

# Follow the prompts to:
# - Configure n8n endpoint  
# - Build and start containers
# - View logs and manage service
```

### **Option 3: Direct Docker**
```bash
# Build image
docker build -t hannaui .

# Run with environment variable
docker run -d -p 3000:80 \
  -e HANNA_N8N_ENDPOINT="http://your-n8n-url/webhook/hanna-chat" \
  --name hannaui \
  hannaui:latest
```

## 🔧 **n8n Endpoint Configuration**

### **Environment Variable**
Set `HANNA_N8N_ENDPOINT` to your n8n webhook URL with streaming support:

**HTTP Streaming (Server-Sent Events):**
```bash
HANNA_N8N_ENDPOINT=https://your-n8n-instance.com/webhook/hanna-chat
```

**WebSocket Streaming (Real-time):**
```bash
HANNA_N8N_ENDPOINT=wss://your-n8n-instance.com/webhook/hanna-chat
```

**Local Development:**
```bash
HANNA_N8N_ENDPOINT=http://localhost:5678/webhook/hanna-chat
HANNA_N8N_ENDPOINT=ws://localhost:5678/webhook/hanna-chat
```

**Demo Mode:**
```bash
# Leave empty or unset for demo mode
HANNA_N8N_ENDPOINT=
```

### **Expected n8n Payload**
HannaUI sends this to your n8n webhook:
```json
{
  "message": "User's message",
  "user": {
    "username": "admin",
    "email": "user@example.com",
    "timestamp": "2025-09-06T15:30:00.000Z"
  },
  "context": {
    "source": "hannaui-web",
    "version": "1.0.0",
    "sessionId": "session_1234567890_abc123"
  }
}
```

### **Expected n8n Response**
Your n8n workflow should return **streaming newline-delimited JSON**:

**Streaming Format:**
```json
{"text": "Hello"}
{"text": " there!"}
{"text": " How can I help you?"}
{"type": "complete"}
```

**Headers sent to n8n:**
```
Accept: text/event-stream
Accept-Encoding: gzip, deflate, br, zstd
Accept-Language: en-GB,en;q=0.5
Connection: keep-alive
```

**Alternative single response:**
- Simple string: `"Hello! How can I help you?"`
- Object with response field: `{"response": "Hello! How can I help?"}`
- Object with message field: `{"message": "Hello! How can I help?"}`

## 📊 **Connection Status**

The AI status indicator shows:
- 🟢 **Online** - Connected to real Hanna AI via n8n
- 🟡 **Offline (Demo Mode)** - n8n unreachable, using demo responses
- 🔴 **Connection Error** - Network/configuration issues
- ⚪ **Demo Mode** - No n8n endpoint configured

## 🛠️ **Development Tools**

### **Browser Console Commands**
```javascript
// Set n8n endpoint dynamically
setN8nEndpoint("http://localhost:5678/webhook/hanna-chat");

// Check current configuration
showEndpointInfo();

// Get endpoint URL
getN8nEndpoint();
```

### **Container Management**
```bash
# View logs
docker-compose logs -f hannaui

# Restart with new environment
docker-compose down
docker-compose up -d

# Access container shell
docker exec -it hannaui sh
```

## 📁 **New Files Added**

- `Dockerfile` - Container configuration
- `docker-compose.yml` - Deployment orchestration  
- `docker-entrypoint.sh` - Environment injection script
- `nginx.conf` - Web server configuration
- `env.js` - Browser environment management
- `.env.example` - Environment template
- `start.sh` - Interactive deployment script
- `.dockerignore` - Build optimization

## 🎯 **Next Steps**

1. **Configure your n8n endpoint** in `.env` file
2. **Deploy with Docker Compose**: `docker-compose up -d`
3. **Test the connection** by sending messages
4. **Monitor logs** to verify n8n integration
5. **Customize themes and animations** as needed

## 🐛 **Troubleshooting**

### **Container Won't Start**
```bash
# Check logs
docker-compose logs hannaui

# Rebuild container
docker-compose down
docker-compose up -d --build
```

### **n8n Connection Issues**
- Verify `HANNA_N8N_ENDPOINT` URL is correct
- Check n8n webhook is accessible
- Ensure CORS is configured if needed
- Test webhook manually with curl

### **Browser Console Errors**
- Check network tab for failed requests
- Verify env-config.js is loading
- Use `showEndpointInfo()` to debug configuration

---

**🎉 Your HannaUI is now ready for production deployment with full n8n integration!**

Simply configure your n8n endpoint and deploy with Docker for a complete AI chat experience! ✨
