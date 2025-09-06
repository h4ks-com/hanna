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

# Check if setup directories exist
SETUP_DIR="/home/node/.n8n/hanna-setup"
if [ ! -d "$SETUP_DIR" ]; then
    echo "❌ Setup directory not found: $SETUP_DIR"
    exit 1
fi

# Import credentials
if [ -d "$SETUP_DIR/credentials" ] && [ "$(ls -A $SETUP_DIR/credentials/*.json 2>/dev/null)" ]; then
    echo "🔑 Importing credentials..."
    for credential_file in $SETUP_DIR/credentials/*.json; do
        filename=$(basename "$credential_file")
        echo "   Importing: $filename"
        if n8n import:credentials --input="$credential_file"; then
            echo "   ✅ $filename imported"
        else
            echo "   ⚠️ Failed to import $filename"
        fi
    done
else
    echo "⚠️ No credential files found"
fi

# Import workflows
if [ -d "$SETUP_DIR/workflows" ] && [ "$(ls -A $SETUP_DIR/workflows/*.json 2>/dev/null)" ]; then
    echo "⚙️ Importing workflows..."
    for workflow_file in $SETUP_DIR/workflows/*.json; do
        filename=$(basename "$workflow_file")
        echo "   Importing: $filename"
        if n8n import:workflow --input="$workflow_file"; then
            echo "   ✅ $filename imported"
        else
            echo "   ⚠️ Failed to import $filename"
        fi
    done
    echo "ℹ️  Please activate workflows manually in the n8n web interface"
else
    echo "⚠️ No workflow files found"
fi

# Create completion marker
touch "$SETUP_MARKER"
echo "🎉 Import completed successfully!"