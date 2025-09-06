// Environment configuration for HannaUI
// This file handles environment variables and configuration

class EnvironmentConfig {
    constructor() {
        this.loadEnvironment();
    }

    // Load environment variables from various sources
    loadEnvironment() {
        // Initialize global ENV object if it doesn't exist
        if (!window.ENV) {
            window.ENV = {};
        }

        // Try to load from server-provided environment (if available)
        this.loadFromServerEnv();
        
        // Load from URL parameters (for quick testing)
        this.loadFromUrlParams();
        
        // Load from localStorage (for persistent overrides)
        this.loadFromLocalStorage();
    }

    // Load from server-provided environment variables
    loadFromServerEnv() {
        // This would be populated by a server-side script in production
        // For now, we'll check if it was set by any server-side process
        if (typeof SERVER_ENV !== 'undefined') {
            Object.assign(window.ENV, SERVER_ENV);
        }
    }

    // Load from URL parameters (for testing)
    loadFromUrlParams() {
        const urlParams = new URLSearchParams(window.location.search);
        
        // Check for n8n endpoint in URL
        const n8nEndpoint = urlParams.get('n8n_endpoint') || urlParams.get('hanna_endpoint');
        if (n8nEndpoint) {
            window.ENV.HANNA_N8N_ENDPOINT = decodeURIComponent(n8nEndpoint);
            console.log('n8n endpoint loaded from URL:', window.ENV.HANNA_N8N_ENDPOINT);
        }

        // Check for WebSocket endpoint
        const wsEndpoint = urlParams.get('ws_endpoint') || urlParams.get('websocket_endpoint');
        if (wsEndpoint) {
            window.ENV.HANNA_N8N_ENDPOINT = decodeURIComponent(wsEndpoint);
            console.log('WebSocket endpoint loaded from URL:', window.ENV.HANNA_N8N_ENDPOINT);
        }
    }

    // Load from localStorage (persistent overrides)
    loadFromLocalStorage() {
        const storedEndpoint = localStorage.getItem('hanna_n8n_endpoint');
        if (storedEndpoint) {
            window.ENV.HANNA_N8N_ENDPOINT = storedEndpoint;
        }
    }

    // Set environment variable
    set(key, value) {
        window.ENV[key] = value;
        
        // Persist certain values to localStorage
        if (key === 'HANNA_N8N_ENDPOINT') {
            if (value) {
                localStorage.setItem('hanna_n8n_endpoint', value);
            } else {
                localStorage.removeItem('hanna_n8n_endpoint');
            }
        }
    }

    // Get environment variable
    get(key, defaultValue = null) {
        return window.ENV[key] || defaultValue;
    }

    // Get all environment variables
    getAll() {
        return { ...window.ENV };
    }

    // Clear environment variable
    clear(key) {
        delete window.ENV[key];
        
        if (key === 'HANNA_N8N_ENDPOINT') {
            localStorage.removeItem('hanna_n8n_endpoint');
        }
    }
}

// Initialize environment configuration
const envConfig = new EnvironmentConfig();

// Make it globally available
window.envConfig = envConfig;

// Helper functions for easy access
window.setN8nEndpoint = function(endpoint) {
    envConfig.set('HANNA_N8N_ENDPOINT', endpoint);
    if (window.chatManager) {
        window.chatManager.setN8nEndpoint(endpoint);
    }
    const protocol = endpoint?.includes('ws://') || endpoint?.includes('wss://') ? 'WebSocket' : 'HTTP';
    console.log(`n8n endpoint set (${protocol}):`, endpoint);
};

window.setWebSocketEndpoint = function(endpoint) {
    // Convenience function for WebSocket endpoints
    if (endpoint && !endpoint.startsWith('ws://') && !endpoint.startsWith('wss://')) {
        endpoint = 'ws://' + endpoint.replace(/^https?:\/\//, '');
    }
    window.setN8nEndpoint(endpoint);
};

window.getN8nEndpoint = function() {
    return envConfig.get('HANNA_N8N_ENDPOINT');
};

window.showEndpointInfo = function() {
    const info = window.chatManager ? window.chatManager.getEndpointInfo() : { endpoint: 'Chat manager not loaded' };
    const isWebSocket = info.endpoint?.includes('ws://') || info.endpoint?.includes('wss://');
    console.log('Current n8n endpoint configuration:', {
        ...info,
        protocol: isWebSocket ? 'WebSocket' : 'HTTP/SSE',
        streaming: true
    });
    return info;
};

// Log current configuration
console.log('HannaUI Environment loaded. Available commands:');
console.log('- setN8nEndpoint("http://your-n8n-url/webhook/hanna-chat")');
console.log('- setWebSocketEndpoint("ws://your-n8n-url/webhook/hanna-chat")');
console.log('- getN8nEndpoint()');
console.log('- showEndpointInfo()');
console.log('');
console.log('🌊 Streaming Support:');
console.log('  • HTTP: Server-Sent Events with newline-delimited JSON');
console.log('  • WebSocket: Real-time bidirectional streaming');
console.log('  • Headers: Accept: text/event-stream, Connection: keep-alive');

if (window.ENV.HANNA_N8N_ENDPOINT) {
    const isWebSocket = window.ENV.HANNA_N8N_ENDPOINT.includes('ws://') || window.ENV.HANNA_N8N_ENDPOINT.includes('wss://');
    console.log(`n8n endpoint configured (${isWebSocket ? 'WebSocket' : 'HTTP'}):`, window.ENV.HANNA_N8N_ENDPOINT);
} else {
    console.log('No n8n endpoint configured - running in demo mode');
}
