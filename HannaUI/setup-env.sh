#!/bin/bash

# HannaUI Environment Setup Script
# This script helps set up environment variables for the HannaUI web interface

# Default values
DEFAULT_PORT=8000
DEFAULT_N8N_ENDPOINT=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🌟 HannaUI Environment Setup${NC}"
echo "=============================="

# Check if n8n endpoint is provided via environment variable
if [ ! -z "$HANNA_N8N_ENDPOINT" ]; then
    echo -e "${GREEN}✅ n8n endpoint found in environment: $HANNA_N8N_ENDPOINT${NC}"
    N8N_ENDPOINT="$HANNA_N8N_ENDPOINT"
else
    echo -e "${YELLOW}⚠️  No HANNA_N8N_ENDPOINT environment variable found${NC}"
    echo "The chat interface will run in demo mode."
    echo ""
    echo "To connect to real Hanna AI, set the environment variable:"
    echo -e "${BLUE}export HANNA_N8N_ENDPOINT=\"http://your-n8n-url/webhook/hanna-chat\"${NC}"
    N8N_ENDPOINT=""
fi

# Create environment JavaScript file
echo "Creating environment configuration..."
cat > env-config.js << EOF
// Auto-generated environment configuration
// Generated on: $(date)

// Server-provided environment variables
window.SERVER_ENV = {
    HANNA_N8N_ENDPOINT: "${N8N_ENDPOINT}",
    BUILD_TIME: "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    VERSION: "1.0.0"
};

console.log('Server environment loaded:', window.SERVER_ENV);
EOF

# Update HTML files to include env-config.js
echo "Updating HTML files..."
for file in index.html login.html register.html; do
    if [ -f "$file" ]; then
        # Check if env-config.js is already included
        if ! grep -q "env-config.js" "$file"; then
            # Add env-config.js before env.js
            sed -i 's|<script src="env.js">|<script src="env-config.js"></script>\n    <script src="env.js">|g' "$file"
            echo -e "${GREEN}✅ Updated $file${NC}"
        else
            echo -e "${YELLOW}⚠️  $file already includes env-config.js${NC}"
        fi
    fi
done

echo ""
echo -e "${GREEN}🚀 Setup Complete!${NC}"
echo ""

if [ ! -z "$N8N_ENDPOINT" ]; then
    echo -e "${GREEN}✅ Configured with n8n endpoint: $N8N_ENDPOINT${NC}"
    echo "The chat interface will connect to the real Hanna AI."
else
    echo -e "${YELLOW}⚠️  Running in demo mode${NC}"
    echo "To enable real AI, set the environment variable and run this script again:"
    echo -e "${BLUE}export HANNA_N8N_ENDPOINT=\"http://your-n8n-url/webhook/hanna-chat\"${NC}"
    echo -e "${BLUE}./setup-env.sh${NC}"
fi

echo ""
echo "Start the server with:"
echo -e "${BLUE}python3 -m http.server ${DEFAULT_PORT}${NC}"
echo ""
echo "Then visit: http://localhost:${DEFAULT_PORT}"
echo ""

# Make the script executable
chmod +x "$0" 2>/dev/null || true
