---
title: "Building Rill Projects with AI"
sidebar_label: "Agentic Quickstart"
sidebar_position: 5
---

Rill projects are defined as YAML and SQL files, which makes them a natural fit for AI coding agents. This guide walks through using an AI agent like Claude Code or Cursor to build a Rill project from scratch.

## Prerequisites

- [Rill CLI installed](/developers/get-started/install)
- An AI coding agent: [Claude Code](https://docs.anthropic.com/en/docs/claude-code/overview), [Cursor](https://www.cursor.com/), or another MCP-compatible tool

## Step 1: Initialize a project with agent instructions

Run `rill init` to create a new project. The interactive setup will prompt you for a project name, OLAP engine, and agent instructions:

```bash
rill init
? Project name my-rill-project
? OLAP engine clickhouse
? Agent instructions all
```

This creates a project directory with:
- `rill.yaml` — project configuration
- `.claude/CLAUDE.md` and `.claude/skills/` — entry point and skills for Claude Code, with detailed instructions for each resource type (models, metrics views, dashboards, etc.)
- `.cursor/` — the same instructions as Cursor rules
- `AGENTS.md` and `.agents/skills/` — the same instructions in the tool-agnostic `AGENTS.md` format
- `.mcp.json` — connects your agent to Rill's local [MCP server](/guide/ai/mcp)

See the [Agent Skills documentation](/developers/agent-skills) for details on what each skill covers and the files generated for each tool.

:::tip Only using one AI agent?
Select a specific option in the "Agent instructions" prompt, or use the `--agent` flag to generate files for just that tool:

```bash
# Claude Code skills
rill init my-project --agent claude

# Cursor rules
rill init my-project --agent cursor

# Tool-agnostic AGENTS.md format
rill init my-project --agent agentsmd
```
:::

### Adding agent instructions to an existing project

If you already have a Rill project, run `rill init` with only the `--agent` flag to generate the agent instruction files:

```bash
rill init ./my-existing-project --agent all
```

The command only writes agent instruction files and MCP configuration — but note that it [overwrites existing agent files](/developers/agent-skills/install#existing-project) at the generated paths.

## Step 2: Start Rill in preview mode

Launch Rill Developer in [preview mode](/developers/build/getting-started#preview-mode) to get a clean, dashboard-only view while your AI agent handles the code:

```bash
rill start my-project --preview
```

This also starts a local MCP server at `http://localhost:9009/mcp`. If you generated agent instructions in Step 1, `.mcp.json` registers that server as `rill-developer`. In Claude Code, approve `rill-developer` when asked. That approval connects Claude Code to the server and its tools. For the Claude desktop app, add the Rill plugin and fully quit and restart the app, as described below.

The MCP server gives your AI agent access to:
- **Project status** — see which resources are healthy, errored, or pending
- **Table schemas** — inspect columns, types, and sample data
- **SQL queries** — run analytical queries against your OLAP engine
- **File operations** — read and write project files

### Add the Rill plugin in the Claude desktop app

In the Claude desktop app, add the [Rill plugin](https://claude.com/marketplace/plugins/rill). Open **Customize → Plugins**, find **Rill**, and select **Add**.

The plugin connects to Rill Developer at `http://localhost:9009/mcp`. That address uses the default HTTP port, `9009` (`rill start --port`, default `9009`). Keep Rill Developer on port 9009. Starting it with a different `--port` leaves the plugin pointed at 9009, so it cannot connect.

:::warning Fully quit Claude, then start it again
Claude loads plugins when the process starts. Closing the window leaves Claude running, so the Rill tools stay missing until you terminate Claude and open it again.

- **macOS:** Quit from the menu bar, or press **Cmd+Q**, then open Claude again.
- **Windows:** Quit from the system tray (right-click the Claude icon and choose **Quit**), or end the Claude process. Closing the window leaves Claude running in the tray.

After Claude restarts, with Rill Developer still running, the Rill tools are available.
:::

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

- **[Agent Skills](/developers/agent-skills)** — learn more about the skills, supported tools, and how to keep them updated
- **[Deploy to Rill Cloud](/developers/deploy/deploy-dashboard)** — share your dashboards with your team
- **[AI Chat](/guide/ai/ai-chat)** — ask questions about your data in natural language from Rill Cloud
- **[AI Configuration](/developers/build/ai-configuration)** — add `ai_instructions` to improve AI responses for your project
- **[Rill MCP Server](/guide/ai/mcp)** — connect Claude Desktop, ChatGPT, or other AI clients to Rill Cloud projects
