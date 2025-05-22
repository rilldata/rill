# rill-mcp-server

The Rill [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server exposes Rill's most essential APIs to LLMs. It's designed primarily for **data analysts**, not data engineers, and focuses on **consuming** Rill metric views—not creating them.

The server can also generate [VegaLite](https://vega.github.io/vega-lite/)-backed visualizations from natural language prompts and data tables. This feature uses OpenAI to translate prompts into chart specs. However:
- Not all MCP clients yet support the `Image` datatype.
- Some clients may offer their own visualization systems you may prefer.

## Installation

### Option 1: Using Docker (Recommended)

You'll need:
- Docker installed
- A Rill service token
- (Optional) An OpenAI API key for prompt-based visualizations

#### 1. Create a Rill service token

```bash
# Install the Rill CLI if you haven't already
curl https://rill.sh | sh

# Create a service token
rill service create mcp-server
```

#### 2. (Optional) Create an OpenAI API key

Generate one at https://platform.openai.com/api-keys if you want to enable chart generation.

#### 3. Configure Claude Desktop

As of 2025-05-08, it's recommended to use [Claude Desktop](https://claude.ai/download) as your MCP client. Cursor currently struggles to compose complex structured payloads ([see issue](https://forum.cursor.com/t/issue-with-mcp-server-and-pydantic-model-object-as-tool-parameter-in-cursor/77110/4?u=ericpgreen)), and Windsurf often thinks it can find the answer in your current code project. Other MCP clients have not yet been tested.

**To configure Claude Desktop:**

1. **Find your Docker executable path:**  
   Run `which docker` in your terminal and copy the output (e.g., `/Users/<your-username>/.docker/bin/docker`).

2. **Edit your `claude_desktop_config.json` file:**  
   Replace `/path/to/docker` below with the path from step 1. Without the full path to Docker, Claude may fail to find it.

   ```json
   {
     "mcpServers": {
       "rill": {
         "command": "/path/to/docker", // Use the full path from `which docker`
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

3. **Restart Claude Desktop** for the changes to take effect.

#### 4. Run with Docker

```bash
docker run --rm -i \
  -e RILL_ORGANIZATION_NAME="your-org-name" \
  -e RILL_PROJECT_NAME="your-project-name" \
  -e RILL_SERVICE_TOKEN="your-rill-service-token" \
  -e OPENAI_API_KEY="your-openai-api-key" \
  rilldata/rill-mcp-server
```

### Option 2: Using Go Directly

You'll need:
- Go 1.21 or later
- A Rill service token
- (Optional) An OpenAI API key for prompt-based visualizations

#### 1. Set up environment variables

```bash
export RILL_ORGANIZATION_NAME="your-org-name"
export RILL_PROJECT_NAME="your-project-name"
export RILL_SERVICE_TOKEN="your-rill-service-token"
export OPENAI_API_KEY="your-openai-api-key"  # Optional
```

#### 2. Run the server

```bash
go run ./cmd/mcp-server/main.go
```

## Usage

The MCP server exposes the following tools:

1. **List Metrics Views**
```json
{
  "name": "list_metrics_views",
  "arguments": {}
}
```

2. **Get Metrics View Spec**
```json
{
  "name": "get_metrics_view_spec",
  "arguments": {
    "name": "your_metrics_view_name"
  }
}
```

3. **Get Time Range Summary**
```json
{
  "name": "get_metrics_view_time_range_summary",
  "arguments": {
    "metrics_view": "your_metrics_view_name"
  }
}
```

4. **Get Metrics View Aggregation**
```json
{
  "name": "get_metrics_view_aggregation",
  "arguments": {
    "metrics_view": "your_metrics_view_name",
    "measures": [{"name": "total_revenue"}],
    "dimensions": [
      {"name": "transaction_timestamp", "time_grain": "TIME_GRAIN_MONTH"},
      {"name": "country"}
    ],
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "sort": [
      {"name": "transaction_timestamp"},
      {"name": "total_revenue", "desc": true}
    ]
  }
}
```

5. **Generate Chart** (if visualization is enabled)
```json
{
  "name": "generate_chart",
  "arguments": {
    "data": {
      // Your data here
    },
    "prompt": "Create a line chart showing revenue over time by country"
  }
}
```

## Development

### Building the Docker Image

```bash
# Build the image
docker build -t rilldata/rill-mcp-server:latest .

# Authenticate with Docker Hub
docker login -u rilldataops  # Use a personal access token from https://app.docker.com/settings/personal-access-tokens

# Push the image
docker push rilldata/rill-mcp-server:latest
```

### Building the Go Binary

```bash
# Build the binary
go build -o mcp-server ./cmd/mcp-server

# Run the binary
./mcp-server
```

## Troubleshooting

If you encounter issues:

1. **Docker Issues**:
   - Ensure Docker is running
   - Check Docker logs: `docker logs <container_id>`
   - Verify environment variables are set correctly

2. **Go Issues**:
   - Check Go version: `go version`
   - Update dependencies: `go mod tidy`
   - Verify environment variables: `env | grep RILL_`

3. **API Issues**:
   - Verify your Rill service token is valid
   - Check your organization and project names
   - Ensure your OpenAI API key is valid (if using visualization)

4. **Client Issues**:
   - Check client logs for connection errors
   - Verify the client supports all required features
   - Ensure the client is configured correctly

5. **Claude Desktop Issues**:
   - In Claude Desktop, click on **Developer → Open MCP Log File** and check the logs for any errors
   - Double-check the Docker path is correct
   - Ensure all required environment variables are set
   - Make sure Docker is running

