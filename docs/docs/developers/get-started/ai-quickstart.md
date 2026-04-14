---
title: "Building Rill Projects with AI"
sidebar_label: "Agentic Quickstart"
sidebar_position: 15
---

Rill projects are defined as YAML and SQL files, which makes them a natural fit for AI coding agents. This guide walks through using an AI agent like Claude Code or Cursor to build a Rill project from scratch.

## Prerequisites

- [Rill CLI installed](/developers/get-started/install)
- An AI coding agent: [Claude Code](https://docs.anthropic.com/en/docs/claude-code/overview), [Cursor](https://www.cursor.com/), or another MCP-compatible tool

## Step 1: Initialize a project with agent instructions

Run `rill init` to create a new project. By default, Rill generates Claude Code instruction files that teach your AI agent how Rill projects work:

```bash
rill init my-project
```

This creates a project directory with:
- `rill.yaml` — project configuration
- `.claude/CLAUDE.md` — entry point that tells Claude Code to load Rill development skills
- `.claude/skills/` — detailed instructions for each resource type (models, metrics views, dashboards, etc.)
- `.mcp.json` — connects Claude Code to Rill's local [MCP server](/guide/ai/mcp)

:::tip Using a different AI agent?
Use the `--agent` flag to generate instructions for your tool of choice:

```bash
# Cursor rules
rill init my-project --agent cursor

# Tool-agnostic AGENTS.md format
rill init my-project --agent agentsmd

# All formats at once
rill init my-project --agent all
```
:::

### Adding agent instructions to an existing project

If you already have a Rill project, you can add agent instructions without affecting your existing files:

```bash
rill init ./my-existing-project --agent claude
```

## Step 2: Start Rill in preview mode

Launch Rill Developer in [preview mode](/developers/build/getting-started#preview-mode) to get a clean, dashboard-only view while your AI agent handles the code:

```bash
rill start my-project --preview
```

This also starts a local MCP server at `http://localhost:9009/mcp`. If you generated agent instructions in Step 1, your AI agent will connect to this server automatically via the `.mcp.json` config — no additional setup required.

The MCP server gives your AI agent access to:
- **Project status** — see which resources are healthy, errored, or pending
- **Table schemas** — inspect columns, types, and sample data
- **SQL queries** — run analytical queries against your OLAP engine
- **File operations** — read and write project files

## Step 3: Build with your AI agent

With Rill running, open your AI agent in the project directory and start building. Here are some examples of what you can ask:

### Connect a data source
> "Connect to the parquet file at `gs://rilldata-public/auction_data.parquet`"

The agent will create a source YAML file and Rill will automatically ingest the data.

### Create models
> "Create a model that cleans the auction data — filter out null bids and add a `bid_bucket` column that groups bids into $0-1, $1-5, $5-10, and $10+ ranges"

### Define metrics
> "Create a metrics view on the auction model with measures for total bids, average bid price, and win rate, broken down by dimensions like domain, device type, and bid bucket"

### Build dashboards
> "Create an explore dashboard for the auction metrics view"

> "Create a canvas dashboard with KPI cards for total bids and win rate, a time series chart, and a breakdown table by domain"

### Iterate
The agent has full context on Rill's resource types and YAML schemas. It can fix errors, refactor models, add new measures, and restructure your project — just describe what you want.

:::tip Check project status
If something isn't working, ask your agent to check the project status. The MCP connection lets it see parse errors, reconciliation failures, and resource health directly.
:::

## Next steps

- **[Deploy to Rill Cloud](/developers/deploy/deploy-dashboard)** — share your dashboards with your team
- **[AI Chat](/guide/ai/ai-chat)** — ask questions about your data in natural language from Rill Cloud
- **[AI Configuration](/developers/build/ai-configuration)** — add `ai_instructions` to improve AI responses for your project
- **[Rill MCP Server](/guide/ai/mcp)** — connect Claude Desktop, ChatGPT, or other AI clients to Rill Cloud projects
