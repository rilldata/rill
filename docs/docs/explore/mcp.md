---
title: "Rill MCP Server"
description: How to connect to Rill MCP and query your deployment
sidebar_label: "Rill MCP Server"
sidebar_position: 1
---
The Rill Model Context Protocol (MCP) server exposes Rill's most essential APIs to LLMs. It's designed primarily for data analysts, not data engineers, and focuses on consuming Rill metric views—not creating them.

The server can also generate VegaLite-backed visualizations from natural language prompts and data tables. This feature uses OpenAI to translate prompts into chart specs. However:

Not all MCP clients yet support the Image datatype.
Some clients may offer their own visualization systems you may prefer.

## Why use MCP?


## Installation

### Prerequisites
In order to use Rill MCP server, you'll need a MCP Client (Recommendation: Claude, see [below](#configure-claude-desktop), but you can use whatever existing client that you might already be using) and a running Rill project (locally or on Rill Cloud).

Before you can use Rill MCP server, you'll also need to install Node.js. You can download it from [nodejs.org](https://nodejs.org/en)

### Create a Rill service token (if your project is on Rill Cloud)
```
# Install the Rill CLI if you haven't already
curl https://rill.sh | sh

# Create a service token
rill service create mcp-server
```

### Configure Claude Desktop
:::warning
As of 2025/05/08, it's recommended to use Claude Desktop as your MCP client. Cursor currently struggles to compose complex structured payloads (see issue), and Windsurf often thinks it can find the answer in your current code project. Other MCP clients have not yet been tested.
:::

Edit your claude_desktop_config.json file. 
By default, the JSON is found in the following directories: 


- MacOS: `/Users/{USER}/Library/Application Support/Claude/claude_desktop_config.json`

- Windows: `C:\Users\{USER}\AppData\Roaming\Claude\claude_desktop_config.json`

### Platforms
Depending on which Rill instance you are trygin to connect to (locally running Rill Developer, Public Rill project on Rill Cloud, Private Rill Project on Rill Cloud (default)) the configuration will vary.


__*Private Rill Project on Rill Cloud*__
```
{
    "mcpServers": {
        "rill": {
            "command": "npx",
            "args": [
                "mcp-remote",
                "https://api.rilldata.com/v1/organizations/demo/projects/rill-openrtb-prog-ads/runtime/mcp/sse",
                "--header",
                "Authorization:${AUTH_HEADER}"
            ],
            "env": {
                "AUTH_HEADER": "Bearer <Rill access token>"
            }
        }
    }
}
```

__*Public Rill project on Rill Cloud*__

```
{
    "mcpServers": {
        "rill": {
            "command": "npx",
            "args": [
                "mcp-remote",
                "https://api.rilldata.com/v1/organizations/demo/projects/rill-github-analytics/runtime/mcp/sse"
            ]
        }
    }
}
```

__*Locally Running Rill Developer*__

```
{
    "mcpServers": {
        "rill": {
            "command": "npx",
            "args": [
                "mcp-remote",
                "http://localhost:9009/mcp/sse"
            ]
        }
    }
}
```


Restart Claude Desktop for the changes to take effect.

### Troubleshooting:
If Claude Desktop cannot connect to the MCP server:

In Claude Desktop, click on Developer → Open MCP Log File and check the logs for any errors.
Double-check the Docker path is correct.
Ensure all required environment variables are set in the `env` parameter.
Make sure Docker is running.

## Using MCP Server in Claude

<img src ='/img/explore/mcp/mcp-main.gif' class='rounded-gif'/>
<br />

### Supported Actions

- __*Action1*__ -
- __*Action2*__ -
- __*Action3*__ -
- __*Action4*__ -
- __*Action5*__ -

### Usage Examples

__*Example Interactions*__
- __*Example1*__ -
- __*Example2*__ -
- __*Example3*__ -
- __*Example4*__ -


## Conclusion


## Resources


## Need help? 
Contact our team through our [various forms of communication](/contact) such as email, Chat, and on Discord! 