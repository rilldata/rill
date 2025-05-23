---
title: "Connecting to Rill MCP Server"
description: How to connect to Rill MCP and query your deployment
sidebar_label: "Rill MCP Server"
sidebar_position: 1
---

![ ]( img here)

## What is Rill MCP Server?

The Rill Model Context Protocol (MCP) server exposes Rill's most essential APIs to LLMs. It's designed primarily for data analysts, not data engineers, and focuses on consuming Rill metric views—not creating them.

The server can also generate VegaLite-backed visualizations from natural language prompts and data tables. This feature uses OpenAI to translate prompts into chart specs. However:

Not all MCP clients yet support the Image datatype.
Some clients may offer their own visualization systems you may prefer.

## Installation


### Pre-requisites

### Create a Rill service token
```
# Install the Rill CLI if you haven't already
curl https://rill.sh | sh

# Create a service token
rill service create mcp-server
```

### (Optional) Create an OpenAI API key
Generate one at https://platform.openai.com/api-keys if you want to enable chart generation.

### Configure Claude Desktop
:::warning
As of 2025-05-08, it's recommended to use Claude Desktop as your MCP client. Cursor currently struggles to compose complex structured payloads (see issue), and Windsurf often thinks it can find the answer in your current code project. Other MCP clients have not yet been tested.
:::

Edit your claude_desktop_config.json file:
Replace /path/to/docker below with the path from step 1. Without the full path to Docker, Claude may fail to find it.
```
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
Restart Claude Desktop for the changes to take effect.

### Troubleshooting:
If Claude Desktop cannot connect to the MCP server:

In Claude Desktop, click on Developer → Open MCP Log File and check the logs for any errors.
Double-check the Docker path is correct.
Ensure all required environment variables are set.
Make sure Docker is running.

## Using MCP Server in Claude

![ ]( img here)