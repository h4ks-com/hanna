import { ChatRequest, ChatResponse, ChatEndpointInfo } from '@/types';
import { clientEnv } from './env';

export class ChatService {
  private static endpoint = clientEnv.HANNA_N8N_ENDPOINT;

  // Get n8n endpoint for chat
  static getEndpoint(): string {
    // Try multiple sources for the endpoint URL
    const sources = [
      // Environment variable
      clientEnv.HANNA_N8N_ENDPOINT,
      // localStorage override (for development/testing)
      typeof window !== 'undefined' ? localStorage.getItem('hanna_n8n_endpoint') : null,
      // Default fallback for development
      'http://localhost:5678/webhook/hanna-chat'
    ];

    for (const source of sources) {
      if (source && source.trim()) {
        return source.trim();
      }
    }

    return '';
  }

  // Set endpoint (for development/testing)
  static setEndpoint(endpoint: string): void {
    if (typeof window === 'undefined') return;
    
    if (endpoint && endpoint.trim()) {
      localStorage.setItem('hanna_n8n_endpoint', endpoint.trim());
      this.endpoint = endpoint.trim();
    } else {
      localStorage.removeItem('hanna_n8n_endpoint');
      this.endpoint = '';
    }
  }

  // Check if the request is cross-origin
  static isCrossOriginRequest(url: string): boolean {
    if (typeof window === 'undefined') return false;
    
    try {
      const targetUrl = new URL(url);
      const currentOrigin = window.location.origin;
      const targetOrigin = `${targetUrl.protocol}//${targetUrl.host}`;
      return currentOrigin !== targetOrigin;
    } catch (e) {
      return false;
    }
  }

  // Check WebSocket connection
  static async checkWebSocketConnection(endpoint: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const ws = new WebSocket(endpoint);
      const timeout = setTimeout(() => {
        ws.close();
        reject(new Error('WebSocket connection timeout'));
      }, 5000);

      ws.onopen = () => {
        clearTimeout(timeout);
        ws.close();
        resolve();
      };

      ws.onerror = () => {
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
  static async checkHTTPConnection(endpoint: string): Promise<void> {
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
      
      if (response.ok || response.status === 404 || response.status === 405) {
        return;
      }
    } catch (error) {
      console.warn('OPTIONS request failed, trying HEAD:', (error as Error).message);
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

  // Check connection status
  static async checkConnection(): Promise<ChatEndpointInfo> {
    const endpoint = this.getEndpoint();
    
    if (!endpoint) {
      return {
        endpoint: null,
        status: 'no-endpoint',
        configured: false
      };
    }

    try {
      if (endpoint.includes('ws://') || endpoint.includes('wss://')) {
        await this.checkWebSocketConnection(endpoint);
      } else {
        await this.checkHTTPConnection(endpoint);
      }
      
      return {
        endpoint,
        status: 'connected',
        configured: true
      };
    } catch (error) {
      console.warn('n8n endpoint not reachable:', (error as Error).message);
      return {
        endpoint,
        status: 'offline',
        configured: true
      };
    }
  }

  // Send message to n8n (using API route proxy in production)
  static async sendMessage(request: ChatRequest): Promise<string> {
    const endpoint = this.getEndpoint();
    
    if (!endpoint) {
      throw new Error('No n8n endpoint configured');
    }

    // In development, call n8n directly
    // In production with Docker, use API route proxy
    const targetUrl = endpoint.includes('localhost') || endpoint.includes('127.0.0.1') 
      ? endpoint 
      : '/api/chat';

    try {
      const response = await fetch(targetUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        body: JSON.stringify(request)
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data: ChatResponse = await response.json();
      
      // Extract the output from the response
      if (data.output) {
        return data.output;
      } else {
        console.warn('No output field in response:', data);
        return data.message || data.text || data.content || 'Response received but no output field found.';
      }

    } catch (error) {
      console.error('Chat request failed:', error);
      // Handle CORS errors specifically
      if ((error as Error).message.includes('CORS') || (error as Error).message.includes('NetworkError')) {
        throw new Error(`CORS Error: ${(error as Error).message}. Try using Docker deployment for CORS bypass.`);
      } else {
        throw error;
      }
    }
  }

  // Generate demo response for offline mode
  static generateDemoResponse(userText: string): string {
    const input = userText.toLowerCase();
    
    if (input.includes('n8n') || input.includes('endpoint') || input.includes('connection')) {
      return `I'm currently running in **demo mode** because no n8n endpoint is configured.

To connect me to the real Hanna AI:

### 🔧 **For Development:**
Set the endpoint in browser console:
\`\`\`javascript
// For HTTP:
localStorage.setItem('hanna_n8n_endpoint', 'http://your-n8n-url/webhook/hanna-chat');
// For WebSocket:
localStorage.setItem('hanna_n8n_endpoint', 'ws://your-n8n-url/webhook/hanna-chat');
\`\`\`

### 🚀 **For Production:**
Set the environment variable:
\`\`\`bash
export NEXT_PUBLIC_HANNA_N8N_ENDPOINT="http://your-n8n-url/webhook/hanna-chat"
\`\`\`

Then refresh the page to connect! ✨`;
    }
    
    if (input.includes('hello') || input.includes('hi')) {
      return "Hello! I'm running in **demo mode**. What can I help you explore today?";
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

Pretty cool, right? 🚀`;
    }
    
    // Default responses
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
}