---
title: Build a Project with Claude Code
sidebar_label: Claude Code Quickstart
sidebar_position: 11
---

# Build a Rill Project with Claude Code

This quickstart shows how to create a complete Rill project — from data source to dashboard — using Claude Code as your AI coding agent. Instead of manually writing YAML and SQL, you'll describe what you want in natural language and let Claude Code build it.

## Prerequisites

- **Rill CLI** installed ([installation guide](/developers/get-started/install))
- **Claude Code** installed (`npm install -g @anthropic-ai/claude-code`)
- An [Anthropic API key](https://console.anthropic.com/) or Claude Max subscription

## Step 1: Create a project and add Claude Code instructions

```bash
# Create a new Rill project
rill init

# Add Claude Code instructions and skills
rill init --template claude
```

This scaffolds `.claude/CLAUDE.md` and a set of skills under `.claude/skills/` that teach Claude Code how to work with every Rill resource type.

## Step 2: Start Rill and Claude Code

Open two terminals:

**Terminal 1** — start Rill Developer so Claude Code can validate changes in real time:
```bash
rill start
```

**Terminal 2** — launch Claude Code in the same project directory:
```bash
claude
```

## Step 3: Connect Claude Code to Rill via MCP

For the best experience, connect Claude Code to your running Rill instance. This gives it access to live tools like `project_status`, `query_sql`, and `list_tables`.

In Claude Code, run:
```
/mcp add --transport http rill http://localhost:9009/mcp
```

Or from the terminal before launching Claude Code:
```bash
claude mcp add --transport http rill http://localhost:9009/mcp
```

## Step 4: Build your project with prompts

Now you can describe what you want and Claude Code will create the files. Here's an example flow using a public dataset:

### Connect to data

> *"Create a model that loads data from gs://rilldata-public/auction_data.parquet"*

Claude Code will create a model YAML file that reads the Parquet file and materializes it as a DuckDB table.

### Create a metrics view

> *"Create a metrics view on the auction_data model. Add measures for total impressions, total bid price, and average bid price. Add dimensions for domain, device type, country, and timestamp as the time series."*

Claude Code will create a metrics view YAML file with the measures and dimensions you described.

### Generate a dashboard

> *"Create an explore dashboard for the auction data metrics view"*

Claude Code will create an explore resource. Open `http://localhost:9009` to see your dashboard in action.

### Iterate

> *"Add a canvas dashboard with a line chart showing bid price over time, a bar chart of top 10 domains by impressions, and a KPI card showing total impressions"*

Keep prompting to refine your project. Claude Code will iterate on errors automatically when connected via MCP.

## Step 5: Explore your dashboard

Open your browser to `http://localhost:9009`. You'll see the explore dashboard and any canvas dashboards Claude Code created. From here you can:

- Slice and dice data across dimensions
- Filter by time ranges
- Drill down into specific segments
- Compare metrics across categories

## What's next?

- **[AI Agents guide](/developers/build/ai-agents)** — full reference for Claude Code and Cursor setup
- **[Build with Claude Code tutorial](/developers/tutorials/build-with-claude-code)** — a more in-depth walkthrough
- **[Rill MCP Server](/guide/ai/mcp)** — connect Claude Desktop or ChatGPT to deployed Rill projects
- **[Deploy to Rill Cloud](/developers/deploy/deploy-dashboard)** — share your project with your team
