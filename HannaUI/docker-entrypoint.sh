#!/bin/sh

# Docker entrypoint script for HannaUI
# This script injects environment variables into the web application

echo "🌟 Starting HannaUI..."
echo "================================"

# Default values
HANNA_N8N_ENDPOINT=${HANNA_N8N_ENDPOINT:-""}
HANNA_VERSION=${HANNA_VERSION:-"1.0.0"}
HANNA_BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Log configuration
echo "Environment Configuration:"
echo "- Version: $HANNA_VERSION"
echo "- Build Time: $HANNA_BUILD_TIME"

if [ -n "$HANNA_N8N_ENDPOINT" ]; then
    echo "- n8n Endpoint: $HANNA_N8N_ENDPOINT"
    echo "✅ Real AI mode enabled"
    
    # Configure nginx proxy for CORS bypass
    echo "🔧 Configuring CORS proxy..."
    
    # Extract host from URL for proxy_pass
    N8N_HOST=$(echo "$HANNA_N8N_ENDPOINT" | sed -E 's|^https?://([^/]+).*|\1|')
    N8N_PROTOCOL=$(echo "$HANNA_N8N_ENDPOINT" | sed -E 's|^(https?)://.*|\1|')
    N8N_PATH=$(echo "$HANNA_N8N_ENDPOINT" | sed -E 's|^https?://[^/]+(/.*)|\1|' | sed 's|^$|/|')
    
    # Update nginx configuration with actual n8n endpoint
    sed -i "s|\$n8n_upstream|${N8N_PROTOCOL}://${N8N_HOST}${N8N_PATH}|g" /etc/nginx/conf.d/default.conf
    sed -i "s|\$proxy_host|${N8N_HOST}|g" /etc/nginx/conf.d/default.conf
    
    echo "- Proxy configured: ${N8N_PROTOCOL}://${N8N_HOST}${N8N_PATH}"
    echo "- CORS bypass available at: /api/n8n/"
else
    echo "- n8n Endpoint: Not configured"
    echo "⚠️  Demo mode enabled"
    
    # Remove proxy configuration if no endpoint
    sed -i 's|proxy_pass \$n8n_upstream;|return 404 "n8n endpoint not configured";|g' /etc/nginx/conf.d/default.conf
fi

# Create environment configuration JavaScript file
cat > /usr/share/nginx/html/env-config.js << EOF
// Auto-generated environment configuration
// Generated on: $HANNA_BUILD_TIME

// Server-provided environment variables
window.SERVER_ENV = {
    HANNA_N8N_ENDPOINT: "${HANNA_N8N_ENDPOINT}",
    HANNA_N8N_PROXY: $([ -n "$HANNA_N8N_ENDPOINT" ] && echo '"/api/n8n/"' || echo 'null'),
    BUILD_TIME: "${HANNA_BUILD_TIME}",
    VERSION: "${HANNA_VERSION}",
    MODE: $([ -n "$HANNA_N8N_ENDPOINT" ] && echo '"production"' || echo '"demo"'),
    CORS_BYPASS_AVAILABLE: $([ -n "$HANNA_N8N_ENDPOINT" ] && echo 'true' || echo 'false')
};

console.log('🐳 Docker environment loaded:', window.SERVER_ENV);
if (window.SERVER_ENV.HANNA_N8N_PROXY) {
    console.log('🔧 CORS bypass proxy available at:', window.SERVER_ENV.HANNA_N8N_PROXY);
}
EOF

# Update HTML files to include env-config.js
echo "Configuring web files..."

for file in index.html login.html register.html; do
    filepath="/usr/share/nginx/html/$file"
    if [ -f "$filepath" ]; then
        # Check if env-config.js is already included
        if ! grep -q "env-config.js" "$filepath"; then
            # Add env-config.js before env.js using sed
            sed -i 's|<script src="env.js">|<script src="env-config.js"></script>\n    <script src="env.js">|g' "$filepath"
            echo "✅ Updated $file"
        fi
    fi
done

echo "🚀 HannaUI configured successfully!"
echo "================================"

# Execute the main command (nginx)
exec "$@"
