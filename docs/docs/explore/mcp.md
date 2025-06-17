---
title: "Rill MCP Server"
description: How to connect to Rill MCP and query your metrics views
sidebar_label: "Rill MCP Server"
sidebar_position: 1
---

The Rill Model Context Protocol (MCP) server exposes Rill's most essential APIs to LLMs. It's currently designed primarily for data analysts, not data engineers, and focuses on consuming Rill metric views—not creating them.

## Why use MCP with Rill?
Instead of blindly exposing your entire data warehouse to external platforms in hopes of uncovering trends, Rill's MCP server provides a **structured and governed** alternative. By querying data that already has **pre-defined measures and dimensions**, the responses you get are guaranteed to be as **accurate and consistent** as the metrics displayed in your Rill dashboards.

This ensures **trustworthy, governed analytics** while empowering users to **self-serve answers** to everyday business questions—without delays caused by email chains or ticket requests. The result: greater team productivity, clearer data ownership, and faster, more confident decision-making across your organization.

## Installation

### Prerequisites

To use the Rill MCP server, you'll need:

- An **MCP client** (we recommend [Claude Desktop](https://claude.ai/download), but you can use any compatible client. [Why?](#configure-claude-desktop))
- A **running Rill project** (locally or hosted on Rill Cloud)
- **Node.js**, which can be downloaded from [nodejs.org](https://nodejs.org/en)

### Create a Rill Personal Access Token (if your project is on Rill Cloud)
You can navigate to the AI tab in your project to retrieve both the JSON and create a Rill personal access token.
<img src='/img/explore/mcp/project-ai.png' class='rounded-gif'/>
<br />

Alternatively, if you want to create the token via the CLI:

```bash
# Install the Rill CLI if you haven't already
curl https://rill.sh | sh

# Create a token
rill token issue
```

### Configure Claude Desktop
:::warning
As of 2025/05/08, it's recommended to use Claude Desktop as your MCP client. Cursor currently struggles to compose complex structured payloads (see [issue](https://forum.cursor.com/t/issue-with-mcp-server-and-pydantic-model-object-as-tool-parameter-in-cursor/77110/5)), and Windsurf often thinks it can find the answer in your current code project. Other MCP clients have not yet been tested.
:::

Edit your `claude_desktop_config.json` file. 
By default, the JSON is found in the following directories:

- macOS: `/Users/{USER}/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `C:\Users\{USER}\AppData\Roaming\Claude\claude_desktop_config.json`

### config.json
Depending on which Rill instance you are trying to connect to (locally running Rill Developer, public Rill project on Rill Cloud, or private Rill project on Rill Cloud (default)), the configuration will vary. For Rill Cloud deployed projects, you can navigate to the AI page to retrieve the `config.json`.

__*Private Rill Project on Rill Cloud*__

Replace `org` and `project` with the ID of your organization and project.

```json
{
    "mcpServers": {
        "rill": {
            "command": "npx",
            "args": [
                "mcp-remote",
                "https://api.rilldata.com/v1/organizations/{org}/projects/{project}/runtime/mcp/sse",
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

__*Public Rill Project on Rill Cloud*__

See [our demo page](https://ui.rilldata.com/demo) for public projects to test.

```json
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

```json
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

:::tip Restart Claude!
Restart Claude Desktop for any changes to your JSON to take effect.
:::

### Troubleshooting
If Claude Desktop cannot connect to the MCP server, check that Rill is running (locally) or that you are able to connect to your [Rill project](https://ui.rilldata.com) from your browser. If your project is private, check that the token is valid via the CLI or create a new one in the UI and edit the `config.json` file.

If you're still experiencing issues, check the logs in Claude Desktop.
Click on Developer → Open MCP Log File and check the logs for any errors.


## Using Rill MCP Server in Claude

<img src='/img/explore/mcp/mcp-main.gif' class='rounded-gif'/>
<br />

### Supported Actions

- __*List metrics views*__ – Use `list_metrics_views` to discover available metrics views in the project.
- __*Get metrics view spec*__ – Use `get_metrics_view` to fetch a metrics view's specification. This is important to understand all the dimensions and measures in a metrics view.
- __*Query the time range*__ – Use `query_metrics_view_time_range` to obtain the available time range for a metrics view. This is important to understand what time range the data spans.
- __*Query the metrics*__ – Use `query_metrics_view` to run queries to get aggregated results.


### Usage Examples

- __*What are my available metrics views?*__ – Provides a list of existing metrics views in your project.
- __*What are the available metrics and dimensions in XXX metrics view?*__ – Provides the list of pre-defined measures and dimensions that have been defined in the `metrics_view.yaml`.
- __*What is the available time range of my data?*__ – Provides an overview of your data's time range, earliest/latest date, etc.
- __*Provide me a TIME_RANGE broken by TIME_GRAIN analysis of XXX measure.*__ – Builds on the previous queries to give you an actual analysis of your data in dimensions and measures based on the time range provided. 


### Using Claude Styles

You can use [Styles](https://www.anthropic.com/news/styles) to automatically provide Claude with additional context whenever you are talking to your data. For instance, by showing Rill how to use text characters to make simple data visualizations, you can get "fast vis" inline with the response, avoiding the delays that can happen when Claude decides to build an entire webpage as part of its answer.

<img src='/img/explore/mcp/claude-desktop-custom-style.png' class='rounded-gif'/>
<br />

You can also teach Claude how to build a URL that links back to the Rill explore. This makes it easy to go from having a conversation in Claude chat, to exploring the data using Rill's dedicated data exploration interface.

<img src='/img/explore/mcp/claude-desktop-explore-link.png' class='rounded-gif'/>
<br />

In order to use a custom style, you need to update our [template](https://github.com/rilldata/rill-claude-styles/blob/main/DataExplorer.md) with the links and explore names for your project. 

1. Edit the BASE_URL to match your project in Rill Cloud
2. Create an METRICS_EXPLORE_SLUG for each Explore in the project
3. Use Rill to create an entire URL for each example
4. Copy that URL into the Link: using the format below
5. Repeat for each example:
    - Multiple metrics and dimensions
    - Individual metric
    - Individual dimension

Full instructions and the style template are included in the [rill-claude-styles](https://github.com/rilldata/rill-claude-styles) repository.


Using all the above concepts, you can ask the Rill MCP server questions like:
- What are my *week-on-week* __increases or decreases in sales__ of `XYZ service`?
- During the *current year*, do I have any __outliers in website views__? What might this correlate to?
- In the *previous quarter*, compared to the current ongoing quarter, what are the __trends for customer access__?


## Conclusion
While [Explore dashboards](./dashboard-101/dashboard-101.md) are a great way to slice and dice to find insights, sometimes you just need a quick, overall summary of your data via a text conversation. MCP servers are the easiest way to do so! Since Rill MCP is built on top of your existing metrics, you can be confident that the returned data will be correct.


## Need help?
[Contact our team](/contact) if you have any questions, comments, or concerns!