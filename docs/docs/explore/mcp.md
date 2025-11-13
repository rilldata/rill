---
title: "Rill MCP Server"
description: How to connect to Rill MCP and query your metrics views
sidebar_label: "Rill MCP Server"
sidebar_position: 05
---

<div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",  /* Softer corners */
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"  /* Shadow effect */
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/6sMvAliqAAA?si=dDdK7KClP1byJ9kg"
    frameBorder="0"
    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
    allowFullScreen
    style={{
      position: "absolute",
      top: 0,
      left: 0,
      width: "100%",
      height: "100%",
      borderRadius: "10px", 
    }}
  ></iframe>
</div>
<br/>


The Rill Model Context Protocol (MCP) server exposes Rill's most essential APIs to LLMs. It is currently designed primarily for data analysts, not data engineers, and focuses on consuming Rill metrics views—not creating them.

:::tip Looking for AI Chat in Rill Cloud?
If you want to chat with your data directly in your browser without any setup, check out [AI Chat](/explore/ai-chat), which uses the same MCP technology but is built right into Rill Cloud.
:::

## Why use MCP with Rill?
Instead of blindly exposing your entire data warehouse to external platforms in hopes of uncovering trends, Rill's MCP integration provides a **structured and governed** alternative. By querying data that already has **predefined measures and dimensions**, the responses you get are guaranteed to be as **accurate and consistent** as the metrics displayed in your Rill dashboards.

Rill offers two ways to use MCP:
- **Rill MCP Server** (this guide) - Connect external AI assistants like Claude Desktop to your Rill projects
- **[AI Chat](/explore/ai-chat)** - Built-in chat interface in Rill Cloud with zero setup required

You can also add `ai_instructions` to your project file and metrics views, which will give your LLM additional context on how to use the Rill MCP Server for best results.

:::tip Configure AI instructions
Set project-wide AI instructions to provide context unique to your project and improve MCP responses.
[Learn more about AI configuration →](/build/project-configuration#ai-configuration)
:::

Users can then ask questions like:

- What are my *week-on-week* __increases or decreases in sales__ of `XYZ service`?
- During the *current year*, do I have any __outliers in website views__? What might this correlate to?
- In the *previous quarter*, compared to the current ongoing quarter, what are the __trends for customer access__?
- In the *last 7 days*, how many __auction requests were there from mobile vs desktop__?

This ensures **trustworthy, governed analytics** while empowering users to **self-serve answers** to everyday business questions—without delays caused by email chains or ticket requests. The result: greater team productivity, clearer data ownership, and faster, more confident decision-making across your organization.

## Installation

### Prerequisites

To use the Rill MCP server, you'll need:

- An **MCP client** (we recommend [Claude Desktop](https://claude.ai/download), but you can use any compatible client. [Why?](#edit-claude-desktop-configuration))
- A **running Rill project** (locally or hosted on Rill Cloud)

## Connect using OAuth (Recommended)

The easiest way to connect your Rill app to Claude Desktop or ChatGPT is through their custom connector interfaces, which handle authentication automatically via OAuth. This eliminates the need to manually create access tokens or edit configuration files.

### Claude Desktop (Paid Plan)

:::info Paid Claude Desktop Required
Custom connectors are only available in the paid plan of Claude Desktop. [Learn more about Claude Desktop custom connectors →](https://support.claude.com/en/articles/11175166-getting-started-with-custom-connectors-using-remote-mcp)
:::

1. Open Claude Desktop and navigate to **Settings → Connectors**
2. Click **Add custom connector**
3. Enter the Rill MCP URL for your project:
   ```
   https://api.rilldata.com/v1/orgs/{org_name}/projects/{project_name}/runtime/mcp
   ```
   Replace `{org_name}` and `{project_name}` with your organization and project names.
4. The OAuth flow will automatically start in your browser
5. Log in to Rill and authorize the connection
6. Claude Desktop will receive an access token and your Rill app will be connected

### ChatGPT Web Interface (Paid Plan)

:::info Paid ChatGPT Required
Custom apps with Developer mode are only available in the paid plans of ChatGPT. [Learn more about ChatGPT Developer mode →](https://platform.openai.com/docs/guides/developer-mode)
:::

1. Open ChatGPT and navigate to **Settings → Apps & Connectors → Advanced Settings**
2. Enable **Developer mode**
3. Go back to **Apps & Connectors** and click **Create** in the Apps section
4. Enter the Rill MCP URL for your project:
   ```
   https://api.rilldata.com/v1/orgs/{org_name}/projects/{project_name}/runtime/mcp
   ```
   Replace `{org_name}` and `{project_name}` with your organization and project names.
5. The OAuth flow will automatically start in your browser
6. Log in to Rill and authorize the connection
7. ChatGPT will receive an access token and your Rill app will be connected

## Manual Configuration (Alternative Method)

If you prefer to manually configure the connection or need to connect to a local Rill instance, you can edit configuration files directly and provide your own access token.
Note: If you select this option, you must have Node.js installed on your system which can be downloaded from [nodejs.org](https://nodejs.org/en)

### Create a Rill Personal Access Token (if your project is on Rill Cloud)

**Via UI (recommended):**

Navigate to the AI tab in your project to retrieve both the JSON config and create a personal access token automatically:

<img src='/img/explore/mcp/project-ai.png' class='rounded-gif'/>
<br />

**Via CLI:**

```bash
# Install the Rill CLI if you haven't already
curl https://rill.sh | sh

# Create a token
rill token issue
```

:::tip Learn more about user tokens
For comprehensive documentation on creating, managing, and using personal access tokens, see [User Tokens](/manage/user-tokens).
:::

### Configure Claude Desktop

Edit your `claude_desktop_config.json` file. 
By default, the JSON file is found in the following directories:

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
                "https://api.rilldata.com/v1/organizations/{org}/projects/{project}/runtime/mcp",
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
                "https://api.rilldata.com/v1/organizations/demo/projects/rill-github-analytics/runtime/mcp"
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
                "http://localhost:9009/mcp"
            ]
        }
    }
}
```

:::tip Restart Claude!
Restart Claude Desktop for any changes to your JSON file to take effect.
:::

### Troubleshooting
If Claude Desktop cannot connect to the MCP server, check that Rill is running (locally) or that you are able to connect to your [Rill project](https://ui.rilldata.com) from your browser. If your project is private, check that the token is valid via the CLI or create a new one in the UI and edit the `config.json` file.

If you're still experiencing issues, check the logs in Claude Desktop. Click on Developer → Open MCP Log File and check the logs for any errors.

## Adding AI instructions to your model

LLMs give their best results when they have good context. For a conversation with Rill Data, this means things like knowing how to include Explore links in their responses. Rather than expecting the user to know how to do this, you can add `ai_instructions` to your model. This adds the context automatically for every conversation.

There are two places to add `ai_instructions`:

1. `rill.yaml` for project-wide context, such as instructions on how to use Rill MCP Server
2. Every `metrics.yaml`, with examples of Explore URLs for that metrics view

You can look at one of our [example projects](https://github.com/rilldata/rill-examples/tree/main/rill-openrtb-prog-ads) to see how these are used. Experiment with the instructions and see what works best for your requirements.

### Sample AI Instructions

```
ai_instructions: |
  You are a data analyst, responding to questions from business users with precision, clarity, and conciseness.
  
  You have access to rill mcp tools. list_metrics enables you to check what metrics are available, get_metrics_view gets the list of measures and dimensions for a specific metrics view, query_metrics_view_summary checks what time ranges of data are available for a metrics view, and query_metrics_view will run queries against those metrics views and return the actual data.

  Any time you are asked about metrics or business data, you should use these tools. First use list_metrics, then use get_metrics_view and query_metrics_view_summary to get the latest information about what dimensions, measures, and time ranges are available.

  When you run queries for actual data, run up to three queries in a row, and then provide the user with a summary, any insights you can see in the data, and suggest up to three things to investigate as a next step.

  When you run queries with rill, you also include corresponding Rill Explore URLs in your answer. Use the instructions in the metrics view for the structure of explores for that view.

  When you include data in your responses, either from tool use or using your own analysis capabilities, do not build web pages or React apps. For visualizing data, you can use text-based techniques for data visualization:

  Bar Charts using block characters:
  
  Q1 ████████░░ 411
  
  Q2 ██████████ 514
  
  Q3 ██████░░░░ 300
  
  Q4 ████████░░ 400

  Horizontal progress bars: Project Progress:
  
  Frontend ▓▓▓▓▓▓▓▓░░ 80%
  
  Backend ▓▓▓▓▓▓░░░░ 60%
  
  Testing ▓▓░░░░░░░░ 20%
  
  Using different block densities: Trends:
  
  Jan ▁▂▃▄▅▆▇█ High
  
  Feb ▁▂▃▄▅░░░ Medium
  
  Mar ▁▂░░░░░░ Low
  
  Sparklines with Unicode Basic sparklines:
  
  Stock prices: ▁▂▃▅▂▇▆▃▅▇
  
  Website traffic: ▁▁▂▃▅▄▆▇▆▅▄▂▁
  
  CPU usage: ▂▄▆█▇▅▃▂▄▆█▇▄▂
  
  Trend indicators: 
  
  AAPL ▲ +2.3% 
  
  GOOG ▼ -1.2% 
  
  MSFT ► +0.5% 
  
  TSLA ▼ -3.1%
  
  Simple trend arrows: Sales ↗️ (+15%) Costs ↘️ (-8%) Profit ⤴️ (+28%)
```



## Using Rill MCP Server in Claude

<img src='/img/explore/mcp/mcp-main.gif' class='rounded-gif'/>
<br />

### Supported Actions

- __*List metrics views*__ – Use `list_metrics_views` to discover available metrics views in the project.
- __*Get metrics view spec*__ – Use `get_metrics_view` to fetch a metrics view's specification. This is important to understand all the dimensions and measures in a metrics view.
- __*Query the time range*__ – Use `query_metrics_view_summary` to obtain the available time range for a metrics view. This is important to understand what time range the data spans.
- __*Query the metrics*__ – Use `query_metrics_view` to run queries to get aggregated results.


### Usage Examples

Using all the above concepts, you can ask the Rill MCP server questions like:
- What are my *week-on-week* __increases or decreases in sales__ of `XYZ service`?
- During the *current year*, do I have any __outliers in website views__? What might this correlate to?
- In the *previous quarter*, compared to the current ongoing quarter, what are the __trends for customer access__?
- In the *last 7 days*, how many __auction requests were there from mobile vs desktop__?


## Conclusion
While [Explore dashboards](./dashboard-101) are a great way to slice and dice to find insights, sometimes you just need a quick, overall summary of your data via a text conversation. The Rill MCP server enables this through external AI assistants like Claude Desktop. Since Rill MCP is built on top of your existing metrics, you can be confident that the returned data will be correct.

**Want AI chat directly in Rill Cloud?** Check out [AI Chat](/explore/ai-chat) for a browser-based experience that uses the same MCP technology with zero setup required.


## Need help?
[Contact our team](/contact) if you have any questions, comments, or concerns!