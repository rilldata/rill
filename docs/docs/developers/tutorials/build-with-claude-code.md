---
title: "Build a Complete Project with Claude Code"
sidebar_label: "Build with Claude Code"
sidebar_position: 05
tags:
    - Tutorial
    - AI
    - Claude Code
---

# Build a Complete Rill Project with Claude Code

This tutorial walks through building a full Rill project — from raw data to polished dashboards — using Claude Code as your AI development agent. You'll learn how to use prompts effectively, connect MCP for live introspection, and iterate on your project without manually writing YAML or SQL.

## What you'll build

By the end of this tutorial you'll have:
- A connector and source model that loads data
- A derived model with transformations
- A metrics view with measures and dimensions
- An explore dashboard for drill-down analysis
- A canvas dashboard with charts and KPIs

## Prerequisites

| Tool | Install |
|------|---------|
| Rill CLI | `curl https://rill.sh \| sh` |
| Claude Code | `npm install -g @anthropic-ai/claude-code` |
| Anthropic API key | [console.anthropic.com](https://console.anthropic.com/) |

## Part 1: Project setup

### Create a new project

```bash
mkdir my-rill-project && cd my-rill-project
rill init
```

### Add Claude Code instructions

```bash
rill init --template claude
```

This generates:
- `.claude/CLAUDE.md` — main instructions covering Rill's architecture, resource types, and development workflow
- `.claude/skills/rill-*/SKILL.md` — resource-specific skills (connector, model, metrics_view, explore, canvas, theme, rillyaml) loaded on-demand

### Start Rill Developer

```bash
rill start
```

Keep this running in a separate terminal. Rill watches for file changes and reconciles resources automatically.

### Launch Claude Code with MCP

Connect Claude Code to your running Rill instance so it can introspect data, check project status, and validate changes:

```bash
claude mcp add --transport http rill http://localhost:9009/mcp
claude
```

## Part 2: Ingest data

### Prompt: Load data from a source

Tell Claude Code about your data. For this tutorial we'll use a public dataset:

> *"Create a source model that loads data from the parquet file at gs://rilldata-public/auction_data.parquet. Materialize it into DuckDB."*

Claude Code will:
1. Create a model YAML file (e.g., `models/auction_data.yaml`)
2. Configure the GCS source and DuckDB output
3. Rill will automatically reconcile and load the data

### Verify with MCP

Because Claude Code is connected via MCP, it can verify the model loaded correctly by checking `project_status` and querying the table schema. If there are errors, it will iterate automatically.

### Prompt: Explore the data shape

> *"What columns are in the auction_data table? Show me a few sample rows and the row count."*

Claude Code will use `query_sql` via MCP to inspect the data, giving you a feel for the schema before defining metrics.

## Part 3: Transform data

### Prompt: Create a derived model

> *"Create a derived model called auction_events that selects from auction_data. Parse the timestamp column as a proper timestamp, extract the top-level domain from the page URL, and filter out rows where bid_price is null or zero."*

Claude Code creates a transformation layer, keeping the source model clean and building business logic in a separate model.

## Part 4: Define metrics

### Prompt: Create a metrics view

> *"Create a metrics view on the auction_events model with:*
> - *Time series on the timestamp column*
> - *Measures: total impressions (count), total bid price (sum), average bid price, unique domains (count distinct)*
> - *Dimensions: domain, device_type, country, bid_type, top_level_domain"*

Claude Code generates a metrics view YAML with the specified measures and dimensions. Rill validates it immediately.

### Prompt: Refine the metrics

> *"Add a measure for win rate — the percentage of auctions where the bid was successful. Also add descriptions to all measures and dimensions so they show up in AI Chat and MCP."*

Adding descriptions makes the metrics view self-documenting and improves AI responses in [AI Chat](/guide/ai/ai-chat) and [MCP](/guide/ai/mcp).

## Part 5: Build dashboards

### Prompt: Create an explore dashboard

> *"Create an explore dashboard for the auction_events metrics view. Set the default time range to the last 7 days and show all measures by default."*

Open `http://localhost:9009` to see the explore dashboard. You can now slice and dice across all dimensions.

### Prompt: Create a canvas dashboard

> *"Create a canvas dashboard called auction_overview with:*
> - *A KPI row showing total impressions, average bid price, and unique domains*
> - *A line chart of total bid price over time*
> - *A bar chart of top 10 domains by impressions*
> - *A table showing bid type breakdown with all measures"*

Canvas dashboards give you a report-style overview alongside the drill-down explore dashboard.

## Part 6: Polish and deploy

### Prompt: Add a theme

> *"Create a dark theme with blue accent colors and apply it to all dashboards."*

### Prompt: Add AI instructions

> *"Add ai_instructions to rill.yaml explaining that this project tracks real-time bidding auction data. Route questions about bidding performance to the auction_events metrics view. Define that 'win rate' means the percentage of successful bids."*

This context helps [AI Chat](/guide/ai/ai-chat) and [MCP](/guide/ai/mcp) users get better answers.

### Deploy to Rill Cloud

When you're happy with the project:

```bash
rill deploy
```

See the [deployment guide](/developers/deploy/deploy-dashboard) for details on connecting to GitHub for continuous deploys.

## Tips for prompting effectively

- **Be specific about column names** — Claude Code can query the schema via MCP, but explicit names reduce back-and-forth
- **Build incrementally** — start with a few measures/dimensions, validate, then add more
- **Let MCP do the validation** — with MCP connected, Claude Code automatically checks for errors and fixes them
- **Use `rill validate`** — if you're not using MCP, tell Claude Code to run `rill validate` after each change
- **Reference existing resources** — say "create a metrics view *on the auction_events model*" rather than leaving it ambiguous

## What's next?

- **[AI Agents reference](/developers/build/ai-agents)** — full setup details for Claude Code and Cursor
- **[AI Configuration](/developers/build/ai-configuration)** — write effective `ai_instructions` for MCP and AI Chat
- **[Rill MCP Server](/guide/ai/mcp)** — connect Claude Desktop or ChatGPT to query your deployed dashboards
- **[Tutorials](/developers/tutorials/)** — more example projects and walkthroughs
