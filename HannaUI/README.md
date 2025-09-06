# HannaUI - AI Chat Interface

A beautiful, modern AI chat interface featuring animated orbs, multiple themes, full markdown support, and secure authentication.

## 🚀 Features

### 🔐 **Authentication System**
- Secure login and registration pages
- User session management with localStorage
- Access control to protect chat interface
- Demo credentials: `admin` / `admin123`
- Remember me functionality
- Password visibility toggle
- Real-time form validation

### 💬 **Chat Interface with Markdown**
- Modern, responsive chat design
- Full **markdown rendering** support including:
  - Headers (H1-H6) with proper sizing
  - **Bold** and *italic* text formatting
  - `Inline code` and code blocks with syntax highlighting
  - Tables with borders and alternating rows
  - Ordered and unordered lists
  - Blockquotes and horizontal rules
  - Links with hover effects
- Real-time message timestamps
- Animated message appearance
- Mobile-responsive design
- Personalized welcome messages

### 🌟 Animated Orbs  
- [x] 10 individual orb elements with glowing effects
- [x] Firefly-style idle animation with natural drift
- [x] Orbs occasionally cluster together
- [x] Mouse avoidance behavior
- [x] Smooth floating animations
- [x] Dynamic color system tied to themes

### 🎭 Orb Interaction & AI Animations
- [x] **Typing** - Orbs swirl together when user focuses input
- [x] **Network** - Flowing patterns during AI "thinking"  
- [x] **DNA** - Double helix intertwining pattern
- [x] **Loading** - Circular pulsing loading indicator
- [x] **Racetrack** - Oval track following pattern
- [x] **Wave** - Smooth sine-wave movements
- [x] **Bloom** - Expanding flower pattern
- [x] **Pulse** - Rhythmic brightness pulsing
- [x] **Constellation** - Connected star formations
- [x] **Cascade** - Chain following movements
- [x] Smooth transitions between states

### 🎨 Themes
- [x] 5 complete theme system:
  - **Dark** - Deep blues and purples (default)
  - **Blue** - Ocean blues and cyans  
  - **Purple** - Rich purples and violets
  - **Pink** - Warm pinks and magentas
  - **Cute** - Bright pinks and purples with extra sparkle
- [x] Dynamic UI updates on theme change
- [x] Persistent theme preferences
- [x] Intuitive theme selector

### ✨ Effects & Aesthetics  
- [x] Glowing orb effects with multiple colors
- [x] Smooth, fluid animations (60fps)
- [x] Modern AI color palette
- [x] Backdrop blur effects
- [x] Performance optimized rendering
- [x] CSS animations with hardware acceleration

## 🚀 Quick Start

### 🐳 **Docker Deployment (Recommended)**

The easiest way to deploy HannaUI is using Docker:

#### **1. Clone and Configure**
```bash
# Clone or navigate to HannaUI directory
cd HannaUI

# Copy environment template
cp .env.example .env

# Edit .env file to set your n8n endpoint
nano .env
```

#### **2. Set n8n Endpoint**
Edit `.env` file:
```bash
# For HTTP streaming (Server-Sent Events)
HANNA_N8N_ENDPOINT=https://your-n8n-instance.com/webhook/hanna-chat

# For WebSocket streaming (Real-time bidirectional)  
HANNA_N8N_ENDPOINT=wss://your-n8n-instance.com/webhook/hanna-chat

# For local development
HANNA_N8N_ENDPOINT=http://localhost:5678/webhook/hanna-chat
HANNA_N8N_ENDPOINT=ws://localhost:5678/webhook/hanna-chat

# Leave empty for demo mode
# HANNA_N8N_ENDPOINT=
```

#### **3. Deploy with Docker Compose**
```bash
# Build and start HannaUI
docker-compose up -d

# View logs
docker-compose logs -f hannaui

# Stop the service
docker-compose down
```

#### **4. Access the Application**
- Visit `http://localhost:3000`
- Login with: `admin` / `admin123`
- Start chatting with Hanna! ✨

### 🔧 **Manual Deployment**

For development or custom setups:

#### **1. Set Environment Variable**
```bash
# Set n8n endpoint
export HANNA_N8N_ENDPOINT="http://your-n8n-url/webhook/hanna-chat"

# Or for demo mode
unset HANNA_N8N_ENDPOINT
```

#### **2. Deploy with Docker (Recommended)**
```bash
# Navigate to the project directory
cd HannaUI

# Start with Docker (includes CORS proxy for n8n)
docker-compose up -d

# Or with custom n8n endpoint
HANNA_N8N_ENDPOINT="https://your-n8n-url/webhook/chat" docker-compose up -d
```

#### **3. Access**
- Visit `http://localhost:8081`
- Use demo credentials: `admin` / `admin123`

#### **Alternative: Local Development Server**
```bash
# For development without Docker
python3 -m http.server 8000
# OR
npx serve .
# Then visit http://localhost:8000
```

## Usage

### Basic Usage
1. Open `index.html` in a modern web browser
2. Start chatting with Hanna AI
3. Watch orbs respond to your interactions
4. Switch themes using buttons in top-right

### Animation Controls (Debug Mode)
- Press `Ctrl+C` to toggle animation controls
- Click animation buttons to test different orb behaviors
- Useful for development and demonstration

### Theme Switching
- Click theme emoji buttons (🌙💙💜🌸🦄) in top-right corner
- Themes are automatically saved to localStorage
- Orb colors and UI elements update instantly

## Technical Details

### 🐳 **Docker Architecture**

HannaUI uses a lightweight nginx-alpine container with environment variable injection:

```
Docker Container:
├── nginx:alpine           # Web server base
├── /usr/share/nginx/html/ # Web files
├── /etc/nginx/conf.d/     # Custom nginx config
└── /docker-entrypoint.sh  # Environment injection script
```

**Environment Variables:**
- `HANNA_N8N_ENDPOINT` - n8n webhook URL for real AI
- `HANNA_VERSION` - Application version
- `HANNA_BUILD_TIME` - Container build timestamp

**Key Features:**
- Automatic environment variable injection into JavaScript
- Health check endpoint at `/health`
- Security headers and gzip compression
- Graceful fallback to demo mode

### 📂 **File Structure**
```
HannaUI/
├── index.html          # Main chat interface (protected)
├── login.html          # Login page
├── register.html       # Registration page
├── styles.css          # Main styling
├── auth.css           # Authentication page styles
├── chat.js            # Chat functionality with n8n integration
├── auth.js            # Authentication system
├── orbs.js            # Orb physics and rendering
├── animations.js      # Animation patterns
├── themes.js          # Theme system
├── env.js             # Environment configuration
├── Dockerfile         # Docker container configuration
├── docker-compose.yml # Docker deployment
├── nginx.conf         # Web server configuration
├── docker-entrypoint.sh # Environment injection script
├── .env.example       # Environment template
└── README.md          # Documentation
```

### 🤖 **n8n Integration with Streaming**

HannaUI connects to a real AI through n8n with **streaming support**:

**🌊 Streaming Protocols Supported:**
- **HTTP Server-Sent Events** (`http://` or `https://`)
- **WebSocket** (`ws://` or `wss://`)

**📡 HTTP Request Headers:**
```
POST /webhook/hanna-chat
Accept: text/event-stream
Accept-Encoding: gzip, deflate, br, zstd
Accept-Language: en-GB,en;q=0.5
Connection: keep-alive
Content-Type: application/json
```

**📦 Payload to n8n:**
```json
{
  "message": "User's message text",
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

**📤 Expected Response Format:**
Newline-delimited JSON objects for streaming:
```json
{"text": "Hello"}
{"text": " there!"}
{"text": " How"}
{"text": " can I help you?"}
{"type": "complete"}
```

**Alternative single response:**
```json
{"response": "Hello there! How can I help you?"}
```

**🔄 Streaming Chunk Types:**
- `{"text": "content"}` - Text content to append
- `{"content": "content"}` - Alternative text field
- `{"message": "content"}` - Alternative text field  
- `{"type": "thinking"}` - AI is processing (triggers network animation)
- `{"type": "typing"}` - AI is typing (triggers typing animation)
- `{"type": "complete"}` - Response finished (triggers pulse animation)
- `{"done": true}` - WebSocket end signal

**Connection States:**
- 🟢 **Connected** - Real AI responses via n8n
- 🟡 **Offline** - Demo mode with simulated responses  
- 🔴 **Error** - Connection issues, fallback to demo
- ⚪ **No Endpoint** - No n8n URL configured

### 🏗️ **Key Classes**
- `ChatManager` - Handles chat, n8n integration, and AI responses
- `AuthManager` - User authentication and session management
- `OrbManager` - Orb positioning, physics, and mouse interactions
- `AnimationController` - Complex animation pattern management  
- `ThemeManager` - Theme switching and persistence
- `EnvironmentConfig` - Environment variable management

### Responsive Design
- Desktop: Chat panel on right side with full orb field
- Mobile: Full-screen chat with adapted orb container
- Tablet: Responsive layout adapts smoothly

## Customization

### Adding New Themes
1. Add theme object to `themes.js`
2. Include all CSS variable definitions
3. Add theme button to HTML
4. Test orb color coordination

### Creating New Animations
1. Add animation function to `AnimationController`
2. Include smooth interpolation for orb movement
3. Add control button for testing
4. Consider performance impact

### Orb Behavior Tuning
- Adjust `baseSpeed` in `OrbManager` for movement speed
- Modify mouse avoidance threshold (100px default)
- Change clustering probability (0.001 default)
- Tune animation smoothing factors (0.03-0.1)

## Browser Compatibility
- Modern browsers with ES6+ support
- Chrome/Edge 80+
- Firefox 75+  
- Safari 13+
- Mobile browsers supported

## Performance Notes
- Optimized for 60fps animation
- Hardware-accelerated CSS transforms
- Efficient requestAnimationFrame usage
- Memory-conscious orb management
- Responsive to device capabilities

## Development
To modify or extend HannaUI:

1. Edit files directly (no build process required)
2. Test in multiple browsers and screen sizes
3. Use animation controls for debugging
4. Monitor performance with browser dev tools

## Future Enhancements
- [ ] Sound effects for orb interactions
- [ ] Particle trail effects
- [ ] Orb constellation line connections  
- [ ] Voice chat integration
- [ ] More animation patterns
- [ ] Accessibility improvements
- [ ] PWA capabilities

---

**HannaUI** - Where AI conversation meets beautiful interaction design ✨
