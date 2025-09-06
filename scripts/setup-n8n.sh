#!/bin/bash
set -e

echo "🔧 Setting up n8n with workflows and credentials..."

# Environment variables with defaults
HANNA_BOT_API_TOKEN=${HANNA_BOT_API_TOKEN:-"your-api-token-here"}
WEBHOOK_TOKEN=${WEBHOOK_TOKEN:-"secret123"}

# Directories
SETUP_DIR="/home/node/.n8n/hanna-setup"
N8N_DATA_DIR="/home/node/.n8n"
SETUP_MARKER="$N8N_DATA_DIR/.hanna-setup-complete"

# Check if setup is already done
if [ -f "$SETUP_MARKER" ]; then
    echo "✅ n8n already set up with Hanna workflows and credentials"
    exit 0
fi

# Wait for n8n database to be initialized
echo "⏳ Waiting for n8n database to be ready..."
timeout=60
counter=0
while [ ! -f "$N8N_DATA_DIR/database.sqlite" ] && [ $counter -lt $timeout ]; do
    sleep 1
    counter=$((counter + 1))
done

if [ $counter -eq $timeout ]; then
    echo "❌ Timeout waiting for n8n database"
    exit 1
fi

echo "📦 n8n database found, proceeding with setup..."

# Process credentials template with environment variables
echo "🔑 Setting up credentials..."
envsubst < "$SETUP_DIR/credentials.json" > "$SETUP_DIR/credentials-processed.json"

# Process workflows template with environment variables  
echo "⚙️ Setting up workflows..."
envsubst < "$SETUP_DIR/workflows.json" > "$SETUP_DIR/workflows-processed.json"

# Import credentials
echo "📥 Importing credentials..."
if n8n import:credentials --input="$SETUP_DIR/credentials-processed.json"; then
    echo "✅ Credentials imported successfully"
else
    echo "❌ Failed to import credentials"
    exit 1
fi

# Import workflows
echo "📥 Importing workflows..."
if n8n import:workflow --input="$SETUP_DIR/workflows-processed.json"; then
    echo "✅ Workflows imported successfully"
else
    echo "❌ Failed to import workflows"
    exit 1
fi

# Create setup marker
echo "✅ Creating setup completion marker..."
touch "$SETUP_MARKER"

echo "🎉 n8n setup completed successfully!"
echo "   - Hanna Bot credentials configured"
echo "   - Echo Bot workflow imported and activated"
echo "   - Webhook endpoint: /webhook/a3297d90-6c5f-40c1-bdf1-c26e6483b450/webhook"

# Clean up processed files
rm -f "$SETUP_DIR/credentials-processed.json" "$SETUP_DIR/workflows-processed.json"