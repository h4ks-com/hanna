#!/bin/sh
echo "🔧 Setting up n8n data directory..."

# Skip if already setup
if [ -f "/data/.hanna-setup-complete" ]; then
  echo "✅ Setup already completed"
  exit 0
fi

# Create directories
mkdir -p /data/hanna-setup

# Process templates with environment variables
envsubst < /templates/credentials.json > /data/hanna-setup/credentials.json
cp /templates/workflows.json /data/hanna-setup/workflows.json

# Create marker
echo "Setup completed at $(date)" > /data/.hanna-setup-complete
echo "✅ Setup files ready in /data/hanna-setup/"