# rill-mcp-server

The Rill [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server exposes Rill's most essential APIs to LLMs. It's designed primarily for **data analysts**, not data engineers, and focuses on **consuming** Rill metric views—not creating them.

The server can also generate [VegaLite](https://vega.github.io/vega-lite/)-backed visualizations from natural language prompts and data tables. This feature uses OpenAI to translate prompts into chart specs. However:
- Not all MCP clients yet support the `Image` datatype.
- Some clients may offer their own visualization systems you may prefer.

## Installation

This MCP server runs via [Docker](https://www.docker.com). You'll need:

- Docker installed
- A Rill service token
- (Optional) An OpenAI API key for prompt-based visualizations

### 1. Create a Rill service token

```bash
# Install the Rill CLI if you haven't already
curl https://rill.sh | sh

# Create a service token
rill service create mcp-server
```

### 2. (Optional) Create an OpenAI API key

Generate one at https://platform.openai.com/api-keys if you want to enable chart generation.

### 3. Configure your MCP client

Create a `mcp.json` file for your client (e.g. Claude Desktop, Cursor, Windsurf):

```json
{
  "mcpServers": {
    "rill": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-e", "RILL_ORGANIZATION_NAME",
        "-e", "RILL_PROJECT_NAME",
        "-e", "RILL_SERVICE_TOKEN",
        "-e", "OPENAI_API_KEY",
        "rilldata/rill-mcp-server"
      ],
      "env": {
        "RILL_ORGANIZATION_NAME": "your-org-name",
        "RILL_PROJECT_NAME": "your-project-name",
        "RILL_SERVICE_TOKEN": "your-rill-service-token",
        "OPENAI_API_KEY": "your-openai-api-key" // Optional
      }
    }
  }
}
```

Claude Desktop notes:
- Restart the app after editing `mcp.json`.
- You may need to provide a full path to docker in the command field (e.g. `/opt/homebrew/bin/docker` on macOS with Homebrew).


## Usage
TODO



## Development

To build and push the Docker image:

```bash
# Build the image
docker build -t rilldata/rill-mcp-server:latest .

# Authenticate with Docker Hub
docker login -u rilldataops  # Use a personal access token from https://app.docker.com/settings/personal-access-tokens

# Push the image
docker push rilldata/rill-mcp-server:latest
```

