#!/bin/sh
echo "🔧 Setting up n8n data directory..."

# Skip if already setup
if [ -f "/data/.hanna-setup-complete" ]; then
  echo "✅ Setup already completed"
  exit 0
fi

# Create directories
mkdir -p /data/hanna-setup/credentials /data/hanna-setup/workflows

# Process credential templates with environment variables
echo "📋 Processing credential templates..."
if [ -d "/templates/credentials" ] && [ "$(ls -A /templates/credentials/*.json 2>/dev/null)" ]; then
    for credential_file in /templates/credentials/*.json; do
        filename=$(basename "$credential_file")
        echo "   Processing: $filename"
        envsubst < "$credential_file" > "/data/hanna-setup/credentials/$filename"
    done
else
    echo "   No credential templates found"
fi

# Copy workflow templates (no environment substitution needed)
echo "📋 Processing workflow templates..."
if [ -d "/templates/workflows" ] && [ "$(ls -A /templates/workflows/*.json 2>/dev/null)" ]; then
    for workflow_file in /templates/workflows/*.json; do
        filename=$(basename "$workflow_file")
        echo "   Processing: $filename"
        cp "$workflow_file" "/data/hanna-setup/workflows/$filename"
    done
else
    echo "   No workflow templates found"
fi

# Create marker
echo "Setup completed at $(date)" > /data/.hanna-setup-complete
echo "✅ Setup files ready in /data/hanna-setup/"