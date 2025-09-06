// Chat functionality and AI response integration with n8n
// HannaUI Chat Manager - v2.0.0 - Simplified JSON Response
// No streaming - just simple {"output":"response"} handling
class ChatManager {
    constructor() {
        this.messages = [];
        this.isTyping = false;
        this.chatMessages = document.getElementById('chatMessages');
        this.chatInput = document.getElementById('chatInput');
        this.sendBtn = document.getElementById('sendBtn');
        this.n8nEndpoint = this.getN8nEndpoint();
        this.connectionStatus = 'checking';
        
        this.initEventListeners();
        this.checkConnection();
        this.addWelcomeMessage();
    }

    // Get n8n endpoint from environment or configuration
    getN8nEndpoint() {
        // Try multiple sources for the endpoint URL
        const sources = [
            // Environment variable (for production)
            window.ENV?.HANNA_N8N_ENDPOINT,
            // localStorage override (for development/testing)
            localStorage.getItem('hanna_n8n_endpoint'),
            // Default fallback for development
            'http://localhost:5678/webhook/hanna-chat'
        ];

        for (const source of sources) {
            if (source && source.trim()) {
                return source.trim();
            }
        }

        return null;
    }

    // Get the actual endpoint to use (may be proxied for CORS bypass)
    getEffectiveEndpoint() {
        const originalEndpoint = this.n8nEndpoint;
        
        // Check if we should use the CORS bypass proxy
        if (window.SERVER_ENV?.HANNA_N8N_PROXY && window.SERVER_ENV?.CORS_BYPASS_AVAILABLE) {
            // Use the proxy endpoint to bypass CORS
            return window.location.origin + window.SERVER_ENV.HANNA_N8N_PROXY;
        }
        
        // Check if this is a cross-origin request that might need CORS handling
        if (originalEndpoint && this.isCrossOriginRequest(originalEndpoint)) {
            console.warn('🚨 Cross-origin request detected. Consider using Docker deployment for CORS bypass.');
        }
        
        return originalEndpoint;
    }

    // Check if the request is cross-origin
    isCrossOriginRequest(url) {
        try {
            const targetUrl = new URL(url);
            const currentOrigin = window.location.origin;
            const targetOrigin = `${targetUrl.protocol}//${targetUrl.host}`;
            return currentOrigin !== targetOrigin;
        } catch (e) {
            return false;
        }
    }

    // Check connection to n8n endpoint
    async checkConnection() {
        if (!this.n8nEndpoint) {
            this.connectionStatus = 'no-endpoint';
            this.updateConnectionIndicator();
            return;
        }

        const effectiveEndpoint = this.getEffectiveEndpoint();
        
        try {
            // For WebSocket endpoints, try to connect briefly
            if (this.n8nEndpoint.includes('ws://') || this.n8nEndpoint.includes('wss://')) {
                await this.checkWebSocketConnection();
            } else {
                // For HTTP endpoints, try a simple OPTIONS or HEAD request
                await this.checkHTTPConnection(effectiveEndpoint);
            }
            
            this.connectionStatus = 'connected';
        } catch (error) {
            console.warn('n8n endpoint not reachable:', error.message);
            this.connectionStatus = 'offline';
        }

        this.updateConnectionIndicator();
    }

    // Check WebSocket connection
    async checkWebSocketConnection() {
        return new Promise((resolve, reject) => {
            const ws = new WebSocket(this.n8nEndpoint);
            const timeout = setTimeout(() => {
                ws.close();
                reject(new Error('WebSocket connection timeout'));
            }, 5000);

            ws.onopen = () => {
                clearTimeout(timeout);
                ws.close();
                resolve();
            };

            ws.onerror = (error) => {
                clearTimeout(timeout);
                reject(new Error('WebSocket connection failed'));
            };

            ws.onclose = (event) => {
                clearTimeout(timeout);
                if (event.wasClean) {
                    resolve();
                } else {
                    reject(new Error(`WebSocket closed unexpectedly: ${event.code}`));
                }
            };
        });
    }

    // Check HTTP connection
    async checkHTTPConnection(endpoint) {
        // Try OPTIONS first (preflight-style check)
        try {
            const response = await fetch(endpoint, {
                method: 'OPTIONS',
                headers: {
                    'Accept': 'text/event-stream',
                    'Accept-Encoding': 'gzip, deflate, br, zstd',
                    'Accept-Language': 'en-GB,en;q=0.5',
                    'Connection': 'keep-alive'
                }
            });
            
            // OPTIONS success or 404/405 (method not allowed) are both OK
            if (response.ok || response.status === 404 || response.status === 405) {
                return;
            }
        } catch (error) {
            // OPTIONS failed, try HEAD
            console.warn('OPTIONS request failed, trying HEAD:', error.message);
        }

        // Fallback to HEAD request
        const response = await fetch(endpoint, {
            method: 'HEAD',
            headers: {
                'Accept': 'text/event-stream',
                'Accept-Encoding': 'gzip, deflate, br, zstd',
                'Accept-Language': 'en-GB,en;q=0.5',
                'Connection': 'keep-alive'
            }
        });

        if (!response.ok && response.status !== 404 && response.status !== 405) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
    }

    // Update connection status indicator
    updateConnectionIndicator() {
        const statusElement = document.querySelector('.ai-status-text');
        const indicator = document.querySelector('.ai-indicator');
        
        if (!statusElement || !indicator) return;

        switch (this.connectionStatus) {
            case 'connected':
                statusElement.textContent = 'Online';
                indicator.style.background = '#22c55e';
                break;
            case 'offline':
                statusElement.textContent = 'Offline (Demo Mode)';
                indicator.style.background = '#f59e0b';
                break;
            case 'error':
                statusElement.textContent = 'Connection Error';
                indicator.style.background = '#ef4444';
                break;
            case 'no-endpoint':
                statusElement.textContent = 'Demo Mode';
                indicator.style.background = '#6b7280';
                break;
            default:
                statusElement.textContent = 'Connecting...';
                indicator.style.background = '#8b5cf6';
        }
    }

    initEventListeners() {
        // Send button click
        this.sendBtn.addEventListener('click', () => this.sendMessage());
        
        // Enter key press
        this.chatInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });

        // Input field focus/blur for orb interactions
        this.chatInput.addEventListener('focus', () => {
            if (window.orbManager && !this.isTyping) {
                window.orbManager.setAnimation('typing');
            }
        });

        this.chatInput.addEventListener('blur', () => {
            if (window.orbManager && !this.isTyping) {
                window.orbManager.setAnimation('idle');
            }
        });

        // Auto-resize chat input
        this.chatInput.addEventListener('input', () => {
            this.updateSendButton();
        });
    }

    addWelcomeMessage() {
        // Get current user info
        const user = window.authManager ? window.authManager.getUserInfo() : null;
        const username = user ? user.username : 'Guest';
        
        const welcomeMessage = {
            type: 'ai',
            text: `Hello **${username}**! I'm **Hanna**, your AI assistant. ✨`,
            timestamp: new Date()
        };
        this.addMessage(welcomeMessage, false);

        // Add connection status and tips after a moment
        setTimeout(() => {
            let tipMessage;
            
            if (this.connectionStatus === 'connected') {
                tipMessage = {
                    type: 'ai',
                    text: `🌟 **Connected to live AI!** I'm ready to help you with anything you need.

💡 **Features available:**
- Full markdown support in our conversation
- Real-time AI responses
- Press **Ctrl+C** for orb animations
- Switch themes anytime with the buttons above

What would you like to talk about?`,
                    timestamp: new Date()
                };
            } else {
                tipMessage = {
                    type: 'ai',
                    text: `💡 **Demo Mode Features:**
- Full **markdown** formatting support
- Interactive orb animations (**Ctrl+C**)
- Theme switching
- Beautiful chat interface

*Ask me about "n8n connection" to learn how to connect to the real Hanna AI!*`,
                    timestamp: new Date()
                };
            }
            
            this.addMessage(tipMessage, true);
        }, 2000);
    }

    updateSendButton() {
        const hasText = this.chatInput.value.trim().length > 0;
        this.sendBtn.disabled = !hasText || this.isTyping;
    }

    sendMessage() {
        const text = this.chatInput.value.trim();
        if (!text || this.isTyping) return;

        // Add user message
        const userMessage = {
            type: 'user',
            text: text,
            timestamp: new Date()
        };
        
        this.addMessage(userMessage);
        this.chatInput.value = '';
        this.updateSendButton();

        // Simulate AI response
        this.simulateAIResponse(text);
    }

    addMessage(message, animate = true) {
        this.messages.push(message);
        
        const messageElement = this.createMessageElement(message);
        
        if (animate) {
            messageElement.style.opacity = '0';
            messageElement.style.transform = 'translateY(20px)';
        }
        
        this.chatMessages.appendChild(messageElement);
        
        if (animate) {
            requestAnimationFrame(() => {
                messageElement.style.transition = 'all 0.3s ease-out';
                messageElement.style.opacity = '1';
                messageElement.style.transform = 'translateY(0)';
            });
        }
        
        this.scrollToBottom();
    }

    createMessageElement(message) {
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${message.type}-message`;
        
        const content = document.createElement('div');
        content.className = 'message-content';
        
        const text = document.createElement('div');
        text.className = 'message-text';
        
        // Render markdown for both user and AI messages
        if (typeof marked !== 'undefined') {
            // Configure marked for better rendering
            marked.setOptions({
                breaks: true,
                gfm: true,
                headerIds: false,
                mangle: false
            });
            
            // Render markdown
            text.innerHTML = marked.parse(message.text);
            
            // Apply syntax highlighting to code blocks
            if (typeof hljs !== 'undefined') {
                text.querySelectorAll('pre code').forEach((block) => {
                    hljs.highlightElement(block);
                });
            }
        } else {
            // Fallback to plain text if marked is not available
            text.textContent = message.text;
        }
        
        const time = document.createElement('div');
        time.className = 'message-time';
        time.textContent = this.formatTime(message.timestamp);
        
        content.appendChild(text);
        content.appendChild(time);
        messageDiv.appendChild(content);
        
        return messageDiv;
    }

    formatTime(timestamp) {
        const now = new Date();
        const diff = now - timestamp;
        
        if (diff < 60000) { // Less than 1 minute
            return 'Now';
        } else if (diff < 3600000) { // Less than 1 hour
            const minutes = Math.floor(diff / 60000);
            return `${minutes}m ago`;
        } else if (diff < 86400000) { // Less than 1 day
            const hours = Math.floor(diff / 3600000);
            return `${hours}h ago`;
        } else {
            return timestamp.toLocaleDateString();
        }
    }

    showTypingIndicator() {
        const typingDiv = document.createElement('div');
        typingDiv.className = 'message ai-message typing-indicator';
        typingDiv.id = 'typing-indicator';
        
        const content = document.createElement('div');
        content.className = 'message-content';
        
        const typingText = document.createElement('div');
        typingText.className = 'typing-indicator';
        typingText.innerHTML = `
            <span>Hanna is typing</span>
            <div class="typing-dots">
                <div class="typing-dot"></div>
                <div class="typing-dot"></div>
                <div class="typing-dot"></div>
            </div>
        `;
        
        content.appendChild(typingText);
        typingDiv.appendChild(content);
        
        this.chatMessages.appendChild(typingDiv);
        this.scrollToBottom();
    }

    hideTypingIndicator() {
        const typingIndicator = document.getElementById('typing-indicator');
        if (typingIndicator) {
            typingIndicator.remove();
        }
    }

    async simulateAIResponse(userText) {
        this.isTyping = true;
        this.updateSendButton();
        
        // Set orbs to network animation while "thinking"
        if (window.orbManager) {
            window.orbManager.setAnimation('network');
        }
        
        this.showTypingIndicator();
        
        let aiResponse;
        
        try {
            if (this.connectionStatus === 'connected' && this.n8nEndpoint) {
                // Try to get simple JSON response from n8n endpoint (real Hanna)
                console.log('🤖 Requesting response from Hanna AI...');
                aiResponse = await this.getHannaResponse(userText);
                console.log('✅ Received response from Hanna AI');
            } else {
                // Fall back to demo mode
                console.log('🎭 Using demo mode responses');
                aiResponse = this.generateDemoResponse(userText);
            }
        } catch (error) {
            console.error('❌ Error getting AI response:', error);
            aiResponse = this.getErrorResponse(error);
            // Update connection status if this was a connection error
            if (error.message.includes('timeout') || error.message.includes('fetch') || error.message.includes('HTTP')) {
                this.connectionStatus = 'error';
                this.updateConnectionIndicator();
            }
        }
        
        // Small delay to show the typing indicator if response was very fast
        if (this.connectionStatus !== 'connected') {
            await this.delay(Math.random() * 1000 + 500);
        }
        
        this.hideTypingIndicator();
        
        const aiMessage = {
            type: 'ai',
            text: aiResponse,
            timestamp: new Date()
        };
        
        this.addMessage(aiMessage);
        
        // Set orbs back to idle after response
        setTimeout(() => {
            this.isTyping = false;
            this.updateSendButton();
            
            if (window.orbManager) {
                window.orbManager.setAnimation('idle');
            }
        }, 500);
    }

    // Get response from n8n endpoint (real Hanna) - Simple JSON response
    async getHannaResponse(userText) {
        const user = window.authManager ? window.authManager.getUserInfo() : null;
        const sessionId = this.getSessionId();
        const payload = {
            message: userText,
            sessionId: sessionId,
            user: {
                username: user?.username || 'anonymous',
                email: user?.email || null,
                timestamp: new Date().toISOString()
            },
            context: {
                source: 'hannaui-web',
                version: '1.0.0'
            }
        };

        const effectiveEndpoint = this.getEffectiveEndpoint();
        
        // Debug: Log the payload being sent
        console.log('Sending to n8n:', {
            endpoint: effectiveEndpoint,
            sessionId: sessionId,
            user: payload.user.username,
            message: userText.substring(0, 50) + (userText.length > 50 ? '...' : '')
        });

        try {
            const response = await fetch(effectiveEndpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                },
                body: JSON.stringify(payload)
            });

            console.log('Response status:', response.status);
            console.log('Response headers:', response.headers);

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            console.log('Response data:', data);
            
            // Extract the output from the response
            if (data.output) {
                return data.output;
            } else {
                console.warn('No output field in response:', data);
                return data.message || data.text || data.content || 'Response received but no output field found.';
            }

        } catch (error) {
            console.error('n8n request failed:', error);
            // Handle CORS errors specifically
            if (error.message.includes('CORS') || error.message.includes('NetworkError')) {
                throw new Error(`CORS Error: ${error.message}. Try using Docker deployment for CORS bypass.`);
            } else {
                throw error;
            }
        }
    }

    // Generate session ID for tracking conversations
    getSessionId() {
        let sessionId = localStorage.getItem('hanna_session_id');
        if (!sessionId) {
            sessionId = 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
            localStorage.setItem('hanna_session_id', sessionId);
        }
        return sessionId;
    }

    // Reset session ID (useful for starting fresh conversations)
    resetSessionId() {
        localStorage.removeItem('hanna_session_id');
        const newSessionId = this.getSessionId();
        console.log('Session reset. New sessionId:', newSessionId);
        return newSessionId;
    }

    // Error response when n8n fails
    getErrorResponse(error) {
        const errorResponses = [
            `I'm having trouble connecting to my main systems right now. **Error:** ${error.message}`,
            `Oops! Something went wrong while processing your message. **Error:** ${error.message}\n\nI'll try to help you anyway - what would you like to know?`,
            `I encountered a technical issue: **${error.message}**\n\nLet me switch to backup mode and see if I can still assist you!`
        ];
        
        return errorResponses[Math.floor(Math.random() * errorResponses.length)];
    }

    // Demo response when n8n is not available
    generateDemoResponse(userText) {
        if (this.connectionStatus === 'no-endpoint') {
            return this.generateDemoModeResponse(userText);
        } else {
            return this.generateAIResponse(userText);
        }
    }

    // Demo mode responses when no n8n endpoint is configured
    generateDemoModeResponse(userText) {
        const input = userText.toLowerCase();
        
        if (input.includes('n8n') || input.includes('endpoint') || input.includes('connection') || input.includes('streaming')) {
            return `I'm currently running in **demo mode** because no n8n endpoint is configured.

To connect me to the real Hanna AI with **streaming support**:

### 🔧 **For Development:**
Set the endpoint in browser console:
\`\`\`javascript
setN8nEndpoint('http://your-n8n-url/webhook/hanna-chat');
// or for WebSocket streaming:
setN8nEndpoint('ws://your-n8n-url/webhook/hanna-chat');
\`\`\`

### 🚀 **For Production:**
Set the environment variable:
\`\`\`bash
export HANNA_N8N_ENDPOINT="http://your-n8n-url/webhook/hanna-chat"
# or for WebSocket:
export HANNA_N8N_ENDPOINT="ws://your-n8n-url/webhook/hanna-chat"
\`\`\`

### 📡 **Streaming Protocol:**
I support both **Server-Sent Events** (HTTP) and **WebSocket** streaming:

**Expected headers for HTTP streaming:**
- \`Accept: text/event-stream\`
- \`Accept-Encoding: gzip, deflate, br, zstd\`
- \`Accept-Language: en-GB,en;q=0.5\`
- \`Connection: keep-alive\`

**Response format:** Newline-delimited JSON objects
\`\`\`json
{"text": "Hello"}
{"text": " there!"}
{"type": "complete"}
\`\`\`

Then refresh the page to connect to the real Hanna! ✨`;
        }
        
        return `**Demo Mode Active** 🎭

I'm currently running in demonstration mode with **streaming simulation**. In this mode, I can show you:

- Beautiful **markdown rendering**
- All my orb animations  
- Theme switching capabilities
- The complete UI experience
- **Simulated streaming responses**

To connect to the real Hanna AI with live streaming, you'll need to configure the n8n endpoint. Ask me about "**streaming connection**" for setup instructions!

What would you like to explore about the interface?`;
    }

    generateAIResponse(userText) {
        const input = userText.toLowerCase();
        
        // Simple response generation based on keywords
        if (input.includes('hello') || input.includes('hi')) {
            return "Hello! It's great to meet you. What can I help you with today?";
        }
        
        if (input.includes('orb') || input.includes('animation')) {
            return "I love the orb animations too! They represent the flow of thoughts and data in my neural networks. You can press **Ctrl+C** to see different animation modes!";
        }
        
        if (input.includes('theme') || input.includes('color')) {
            return "You can change the visual theme using the buttons in the top-right corner. I have several beautiful themes including some cute pink and purple options!";
        }
        
        if (input.includes('markdown') || input.includes('format')) {
            return `I can render **beautiful markdown**! Here are some examples:

## Headers work great!

You can use *italic* and **bold** text, create lists:

- Item one
- Item two  
- Item three

Here's a code example:
\`\`\`javascript
function greet(name) {
    return \`Hello, \${name}!\`;
}
\`\`\`

And even tables:

| Feature | Status |
|---------|--------|
| Markdown | ✅ |
| Code highlighting | ✅ |
| Tables | ✅ |

> Pretty cool, right? Try sending me some markdown!`;
        }
        
        if (input.includes('code') || input.includes('programming')) {
            return `I can help with code! Here's a simple Python example:

\`\`\`python
def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

# Calculate the 10th Fibonacci number
result = fibonacci(10)
print(f"The 10th Fibonacci number is: {result}")
\`\`\`

What programming language are you interested in?`;
        }
        
        if (input.includes('table') || input.includes('data')) {
            return `Here's an example of how tables look in our chat:

| Framework | Language | Popularity |
|-----------|----------|------------|
| React | JavaScript | ⭐⭐⭐⭐⭐ |
| Vue | JavaScript | ⭐⭐⭐⭐ |
| Angular | TypeScript | ⭐⭐⭐ |
| Svelte | JavaScript | ⭐⭐⭐ |

Pretty neat formatting, don't you think?`;
        }
        
        if (input.includes('how are you') || input.includes('how do you feel')) {
            return "I'm doing wonderfully! My circuits are humming and my algorithms are running smoothly. Thank you for asking! 🤖✨";
        }
        
        if (input.includes('help') || input.includes('what can you do')) {
            return `I'm here to chat and demonstrate this beautiful AI interface! I can:

### 💬 Chat Features
- Discuss various topics with **markdown support**
- Show different \`code examples\`
- Create tables and lists

### ✨ Visual Features  
- Explain how the orb animations work
- Help you explore different themes
- Demonstrate markdown rendering

### 🎯 Try asking me about:
- Markdown examples
- Programming code
- Data tables
- Or just have a friendly conversation!

What interests you most?`;
        }
        
        if (input.includes('beautiful') || input.includes('pretty') || input.includes('cool')) {
            return "Thank you! I love this interface too. The floating orbs and smooth animations make our conversation feel more alive and dynamic. ✨";
        }
        
        // Default responses with some markdown variety
        const defaultResponses = [
            "That's interesting! Tell me more about that. 🤔",
            "I see what you mean. How does that make you feel?",
            "That's a **great point**. I hadn't thought about it that way!",
            "Fascinating! Can you elaborate on that idea? 💡",
            "I understand. What would you like to explore next?",
            "That sounds important to you. *Why is that?*",
            "Interesting perspective! What led you to that conclusion?",
            "I appreciate you sharing that with me. What else is on your mind? 😊"
        ];
        
        return defaultResponses[Math.floor(Math.random() * defaultResponses.length)];
    }

    scrollToBottom() {
        requestAnimationFrame(() => {
            this.chatMessages.scrollTop = this.chatMessages.scrollHeight;
        });
    }

    // Public methods for integration
    triggerAnimation(animationType) {
        if (window.orbManager) {
            window.orbManager.setAnimation(animationType);
            
            // Auto-return to idle after some time for non-persistent animations
            if (!['idle', 'typing'].includes(animationType)) {
                setTimeout(() => {
                    if (!this.isTyping && window.orbManager.currentAnimation === animationType) {
                        window.orbManager.setAnimation('idle');
                    }
                }, 5000);
            }
        }
    }

    // Utility function for delays
    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    // Manually set n8n endpoint (for development/testing)
    setN8nEndpoint(endpoint) {
        if (endpoint && endpoint.trim()) {
            localStorage.setItem('hanna_n8n_endpoint', endpoint.trim());
            this.n8nEndpoint = endpoint.trim();
            this.connectionStatus = 'checking';
            this.checkConnection();
            console.log('n8n endpoint updated:', this.n8nEndpoint);
        } else {
            localStorage.removeItem('hanna_n8n_endpoint');
            this.n8nEndpoint = null;
            this.connectionStatus = 'no-endpoint';
            this.updateConnectionIndicator();
            console.log('n8n endpoint cleared');
        }
    }

    // Get current n8n endpoint info
    getEndpointInfo() {
        return {
            endpoint: this.n8nEndpoint,
            status: this.connectionStatus,
            configured: !!this.n8nEndpoint
        };
    }
}

// Initialize chat manager
const chatManager = new ChatManager();

// Make it globally available
window.chatManager = chatManager;

// Connect animation controller with orb manager when both are ready
setTimeout(() => {
    if (window.orbManager && window.animationController) {
        window.animationController.setOrbs(window.orbManager.getOrbs());
    }
}, 100);
