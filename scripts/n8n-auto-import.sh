#!/bin/bash
set -e

echo "🔧 Auto-importing n8n workflows and credentials..."

# Wait for n8n to be ready
echo "⏳ Waiting for n8n to be ready..."
timeout=60
counter=0
while ! wget --quiet --tries=1 --spider "http://n8n:5678/healthz" >/dev/null 2>&1; do
    if [ $counter -ge $timeout ]; then
        echo "❌ Timeout waiting for n8n"
        exit 1
    fi
    sleep 2
    counter=$((counter + 2))
done

echo "✅ n8n is ready"

# Check if already imported
SETUP_MARKER="/home/node/.n8n/.hanna-import-complete"
if [ -f "$SETUP_MARKER" ]; then
    echo "✅ Import already completed"
    exit 0
fi

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
        # Activate the workflow
        echo "🔄 Activating workflow..."
        n8n update:workflow --id=1 --active=true || echo "⚠️ Failed to activate workflow"
    else
        echo "⚠️ Workflow import failed"
    fi
else
    echo "⚠️ No workflows file found"
fi

# Create completion marker
touch "$SETUP_MARKER"
echo "🎉 Import completed successfully!"