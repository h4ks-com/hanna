#!/bin/bash

set -e

N8N_HOST=${N8N_HOST:-"http://n8n:5678"}
WORKFLOWS_DIR="/workflows"
WEBHOOK_TOKEN=${WEBHOOK_TOKEN:-"secret123"}
N8N_OWNER_EMAIL=${N8N_OWNER_EMAIL:-"admin@hanna-bot.local"}
N8N_OWNER_PASSWORD=${N8N_OWNER_PASSWORD:-"hanna123!"}
COOKIE_FILE="/tmp/n8n_cookies.txt"

echo "N8N Workflow Loader starting..."
echo "N8N Host: $N8N_HOST"
echo "Webhook Token: $WEBHOOK_TOKEN"

# Wait for n8n to be ready
echo "Waiting for n8n to be available..."
until curl -f "$N8N_HOST/healthz" >/dev/null 2>&1; do
    echo "Waiting for n8n health endpoint..."
    sleep 5
done

echo "n8n health endpoint is ready, now checking API..."

# Also check if the REST API is available
until curl -s "$N8N_HOST/rest/workflows" >/dev/null 2>&1; do
    echo "Waiting for n8n REST API..."
    sleep 3
done

echo "n8n is fully ready!"

# Use n8n CLI instead of REST API for better reliability
# We'll use docker exec to run CLI commands in the n8n container

# Function to run n8n CLI commands
n8n_cli() {
    docker exec hanna-n8n-1 n8n "$@"
}

# Function to check if a workflow exists by name
workflow_exists() {
    local workflow_name="$1"
    # Use n8n CLI to list workflows and filter for actual workflow lines  
    # The format is: timestamp | [32minfo[39m | [32mworkflow_id|workflow_name[39m {"file":"workflow.js"}
    local workflow_output=$(n8n_cli list:workflow 2>&1 | grep "|.*${workflow_name}.*\[39m")
    
    if [ -n "$workflow_output" ]; then
        echo "exists"
    else
        echo ""
    fi
}

# Function to create or update a workflow
import_workflow() {
    local workflow_file="$1"
    local workflow_name=$(jq -r '.name // "Unknown"' "$workflow_file" 2>/dev/null || echo "Unknown")
    
    echo "Processing workflow: $workflow_name"
    
    # Check if workflow already exists
    if [ "$(workflow_exists "$workflow_name")" = "exists" ]; then
        echo "✓ Workflow '$workflow_name' already exists"
        
        # Get the existing workflow ID and ensure it's activated
        local existing_id=$(n8n_cli list:workflow 2>&1 | grep "|.*${workflow_name}.*\[39m" | head -1 | sed 's/.*\[32m\([^|]*\)|.*/\1/')
        if [ -n "$existing_id" ]; then
            echo "Ensuring workflow '$workflow_name' is activated (ID: $existing_id)..."
            # Use REST API to activate the workflow
            if curl -s -X PATCH "$N8N_HOST/rest/workflows/$existing_id" \
                -H "Content-Type: application/json" \
                -d '{"active": true}' | grep -q '"active":true'; then
                echo "✓ Workflow '$workflow_name' is now activated"
            else
                echo "⚠ Failed to activate existing workflow '$workflow_name'"
            fi
        fi
        
        return 0
    fi
    
    echo "Creating workflow '$workflow_name' using n8n CLI..."
    
    # Copy workflow file to n8n container
    docker cp "$workflow_file" hanna-n8n-1:/tmp/workflow.json
    
    # Import the workflow using n8n CLI
    if n8n_cli import:workflow --input=/tmp/workflow.json; then
        echo "✓ Successfully imported workflow '$workflow_name'"
        
        # Get the workflow ID that was just imported  
        # Extract ID from pattern: [32mworkflow_id|workflow_name[39m
        local imported_id=$(n8n_cli list:workflow 2>&1 | grep "|.*${workflow_name}.*\[39m" | head -1 | sed 's/.*\[32m\([^|]*\)|.*/\1/')
        
        if [ -n "$imported_id" ]; then
            echo "Activating workflow '$workflow_name' (ID: $imported_id)..."
            # Use REST API to activate the workflow since CLI doesn't work while n8n is running
            if curl -s -X PATCH "$N8N_HOST/rest/workflows/$imported_id" \
                -H "Content-Type: application/json" \
                -d '{"active": true}' | grep -q '"active":true'; then
                echo "✓ Successfully activated workflow '$workflow_name'"
            else
                echo "⚠ Workflow imported but activation failed"
            fi
        else
            echo "⚠ Workflow imported but could not get ID for activation"
        fi
        
        # Clean up temp file (ignore permission errors)
        docker exec hanna-n8n-1 rm -f /tmp/workflow.json 2>/dev/null || true
        
        return 0
    else
        echo "✗ Failed to import workflow '$workflow_name'"
        # Clean up temp file even on failure (ignore permission errors)
        docker exec hanna-n8n-1 rm -f /tmp/workflow.json 2>/dev/null || true
        return 1
    fi
}

# Process all workflow files
if [ -d "$WORKFLOWS_DIR" ]; then
    echo "Looking for workflow files in $WORKFLOWS_DIR..."
    
    for workflow_file in "$WORKFLOWS_DIR"/*.json; do
        if [ -f "$workflow_file" ]; then
            echo "Found workflow file: $(basename "$workflow_file")"
            import_workflow "$workflow_file"
        fi
    done
else
    echo "No workflows directory found at $WORKFLOWS_DIR"
fi

echo "Workflow loading completed!"

# Keep container running for a short time to ensure everything is processed
sleep 10