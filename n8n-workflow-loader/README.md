# N8N Workflow Loader

This sidecar container automatically loads workflows into n8n when the Docker Compose stack starts up.

## How it works

1. **Startup**: The container waits for n8n to be ready by checking the `/healthz` endpoint
2. **Discovery**: It scans the `/workflows` directory for `.json` workflow files
3. **Import**: For each workflow file, it:
   - Checks if a workflow with the same name already exists
   - If not, creates the workflow using the n8n REST API
   - Activates the newly created workflow
4. **Cleanup**: After processing all workflows, the container exits

## Configuration

Environment variables:
- `N8N_HOST`: The n8n instance URL (default: `http://n8n:5678`)
- `WEBHOOK_TOKEN`: Token used for webhook authentication (default: `secret123`)

## Adding new workflows

1. Place your workflow JSON files in the `workflows/` directory
2. Ensure each workflow has a unique `name` field
3. Remove the `id` field from the workflow JSON (the loader handles this automatically)
4. Restart the Docker Compose stack to trigger the loader

## Workflow format

Workflows should be in standard n8n export format:

```json
{
  "name": "My Workflow",
  "nodes": [...],
  "connections": {...},
  "active": true,
  "settings": {},
  "versionId": "1"
}
```

## Default workflows

The loader comes with:
- **Hanna Echo Bot**: A simple echo bot that responds to IRC messages through the Hanna bot integration

## Integration with Hanna Bot

The loader automatically configures the `TRIGGER_CONFIG` environment variable for the Hanna bot container using the webhook ID from the loaded workflows, enabling seamless integration between n8n workflows and IRC events.