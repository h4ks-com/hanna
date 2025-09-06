#!/bin/bash
set -e

echo "🔧 Importing Hanna workflows and credentials into n8n..."

# Wait for n8n to be ready
echo "⏳ Waiting for n8n to be ready..."
timeout=60
counter=0
while ! curl -f -s "http://localhost:5678/healthz" >/dev/null 2>&1; do
    if [ $counter -ge $timeout ]; then
        echo "❌ Timeout waiting for n8n"
        exit 1
    fi
    sleep 2
    counter=$((counter + 2))
done

echo "✅ n8n is ready"

# Check if setup files exist
SETUP_DIR="/home/node/.n8n/hanna-setup"
if [ ! -d "$SETUP_DIR" ]; then
    echo "❌ Setup directory not found: $SETUP_DIR"
    exit 1
fi

# Import credentials
if [ -f "$SETUP_DIR/credentials.json" ]; then
    echo "🔑 Importing credentials..."
    if n8n import:credentials --input="$SETUP_DIR/credentials.json"; then
        echo "✅ Credentials imported"
    else
        echo "⚠️ Credential import failed"
    fi
else
    echo "⚠️ No credentials file found"
fi

# Import workflows
if [ -f "$SETUP_DIR/workflows.json" ]; then
    echo "⚙️ Importing workflows..."
    if n8n import:workflow --input="$SETUP_DIR/workflows.json"; then
        echo "✅ Workflows imported"
    else
        echo "⚠️ Workflow import failed"
    fi
else
    echo "⚠️ No workflows file found"
fi

echo "🎉 Import completed!"