---
title: "Rill Agent Skills"
description: Teach AI coding agents like Claude Code and Cursor how to build Rill projects
sidebar_label: "Overview"
sidebar_position: 0
---

Rill ships built-in agent skills that teach AI coding agents how to develop Rill projects. Because Rill projects are defined as YAML and SQL files, they are a natural fit for agentic development: the skills give your agent detailed knowledge of Rill's project structure, resource types, and development workflow, so it can build connectors, models, metrics views, and dashboards on your behalf.

A single command generates everything your agent needs:

```bash
rill init --agent all
```

The skills work with Claude Code, Cursor, and any agent that supports the `AGENTS.md` convention, such as OpenAI Codex, Gemini CLI, and GitHub Copilot. See the [Installation guide](/developers/agent-skills/install) for setup instructions for each tool.

:::note Agent skills vs. the Rill MCP Server
Agent skills and the [Rill MCP Server](/guide/ai/mcp) are complementary. The skills teach your agent *how* to build a Rill project — resource types, YAML schemas, and best practices. The MCP server gives it *live access* to your running project — resource status, table schemas, query results, and file operations. `rill init --agent` sets up both: it generates the skill files and an MCP configuration that connects your agent to Rill Developer's local MCP server.
:::

## What you can do

With agent skills installed, your AI agent can:

- Connect to data sources like S3, GCS, BigQuery, and Snowflake
- Create and refactor SQL and YAML models, including incremental and partitioned models
- Define metrics views with measures, dimensions, and access policies
- Build explore and canvas dashboards
- Configure project-wide settings in `rill.yaml` and custom themes
- Inspect resource status, debug parse errors, and fix reconciliation failures
- Run analytical queries against your metrics views and summarize the results

## Supported tools

| Tool | Format | Generated files |
| --- | --- | --- |
| [Claude Code](/developers/agent-skills/install#claude-code) | Claude skills | `.claude/CLAUDE.md`, `.claude/skills/`, `.mcp.json` |
| [Cursor](/developers/agent-skills/install#cursor) | Cursor rules | `.cursor/rules/`, `.cursor/mcp.json` |
| [Other agents](/developers/agent-skills/install#agentsmd-compatible-agents) (Codex, Gemini CLI, Copilot, ...) | `AGENTS.md` | `AGENTS.md`, `.agents/skills/`, `.mcp.json` |

## Example prompts

Once installed, the skills activate automatically based on your request. Ask naturally:

```text
"Connect to the parquet file at gs://rilldata-public/auction_data.parquet"
"Create a model that cleans the auction data and buckets bids into price ranges"
"Create a metrics view with measures for total bids, average bid price, and win rate"
"Build an explore dashboard for the auction metrics view"
"Add a canvas dashboard with KPI cards and a time series chart"
"Why is my events model erroring?"
"What were my top domains by win rate last week?"
```

## Related

- [Agentic Quickstart](/developers/get-started/ai-quickstart) — step-by-step guide to building a project with an AI agent
- [Installing agent skills](/developers/agent-skills/install)
- [Agent skills reference](/developers/agent-skills/skills)
- [AI Configuration](/developers/build/ai-configuration) — add `ai_instructions` to improve AI responses for your project
- [Rill MCP Server](/guide/ai/mcp) — connect AI assistants to Rill Cloud projects for data analysis
