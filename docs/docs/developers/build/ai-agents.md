---
title: Develop with AI Agents
description: Use rill init to set up Claude Code or Cursor for AI-assisted Rill project development
sidebar_label: AI Agents (Claude & Cursor)
sidebar_position: 05
---

# Develop Rill Projects with AI Agents

Rill ships built-in instructions and skills that teach AI coding agents how to build Rill projects. Running a single `rill init` command scaffolds everything your agent needs — resource schemas, best practices, and tool-specific conventions — so you can go from a blank project to a working dashboard using natural language.

## Quick start

```bash
# 1. Create a new Rill project (if you don't have one yet)
rill init

# 2. Add Claude Code instructions
rill init --template claude

# Or add Cursor rules instead
rill init --template cursor
```

Then open the project in your agent of choice and start prompting.

## Claude Code

[Claude Code](https://docs.anthropic.com/en/docs/claude-code) is Anthropic's CLI agent for software development. The `claude` template generates a set of files that give Claude Code deep knowledge of Rill's resource types, YAML schemas, and development workflow.

### Setup

1. **Install Claude Code** if you haven't already:
   ```bash
   npm install -g @anthropic-ai/claude-code
   ```

2. **Initialize instructions** in your Rill project directory:
   ```bash
   rill init --template claude
   ```

   This creates:
   | Path | Purpose |
   |------|---------|
   | `.claude/CLAUDE.md` | Main instructions — project structure, resource types, development workflow |
   | `.claude/skills/rill-connector/SKILL.md` | How to create and configure connectors |
   | `.claude/skills/rill-model/SKILL.md` | How to create models (SQL + YAML) |
   | `.claude/skills/rill-metrics-view/SKILL.md` | How to define metrics views with dimensions and measures |
   | `.claude/skills/rill-explore/SKILL.md` | How to configure explore dashboards |
   | `.claude/skills/rill-canvas/SKILL.md` | How to build canvas dashboards |
   | `.claude/skills/rill-theme/SKILL.md` | How to customize themes |
   | `.claude/skills/rill-rillyaml/SKILL.md` | How to configure `rill.yaml` |

3. **Start Claude Code** from your project root:
   ```bash
   claude
   ```

   Claude Code automatically reads `.claude/CLAUDE.md` on startup. Skills are loaded on-demand when invoked, keeping the context lean.

### How it works

- **`CLAUDE.md`** is always loaded and gives Claude Code the full picture of Rill's architecture — resource types, the project DAG, execution model, and development best practices.
- **Skills** contain detailed instructions and JSON schemas for each resource type. Claude Code loads them only when relevant (e.g., when you ask it to create a metrics view, it pulls in the `rill-metrics-view` skill).
- Instructions are kept in sync with your version of Rill. Run `rill init --template claude --force` to update them after upgrading.

### Example prompts

Once set up, you can ask Claude Code things like:

- *"Connect to my Postgres database and create a model that loads the `orders` table"*
- *"Create a metrics view on the `events` model with measures for total events and unique users, and dimensions for country and device type"*
- *"Add an explore dashboard for my sales metrics view with a dark theme"*
- *"Build a canvas dashboard that shows revenue trends, top products, and regional breakdown"*

### Connect MCP for live development

:::tip Recommended
Connecting MCP transforms Claude Code from a file editor into a full development agent that can introspect your data, validate changes, and fix errors automatically.
:::

Rill's MCP server gives Claude Code access to live tools during development:

| Tool | What it does |
|------|-------------|
| `project_status` | Check resource status and errors |
| `query_sql` | Run SQL queries against your OLAP database |
| `list_tables` / `get_table` | Discover tables and inspect schemas |
| `query_metrics_view` | Query metrics views to validate results |

**Connect to a locally running Rill instance:**
```bash
claude mcp add --transport http rill http://localhost:9009/mcp
```

**Connect to a Rill Cloud project:**
```bash
claude mcp add --transport http rill https://api.rilldata.com/v1/orgs/{org}/projects/{project}/runtime/mcp
```

With MCP connected, Claude Code can:
- Discover what tables exist in your database before writing models
- Check for reconcile errors after creating resources and fix them automatically
- Profile data (row counts, column cardinality, date ranges) to inform metrics view design
- Validate that dashboards render correctly by querying the metrics view

See the [Rill MCP Server](/guide/ai/mcp) guide for full setup instructions including authentication.

## Cursor

[Cursor](https://cursor.com) is an AI-powered code editor. The `cursor` template generates rule files that teach Cursor's AI assistant how to work with Rill projects.

### Setup

1. **Initialize rules** in your Rill project directory:
   ```bash
   rill init --template cursor
   ```

   This creates rule files under `.cursor/rules/`:
   | Path | Purpose |
   |------|---------|
   | `.cursor/rules/development.mdc` | Main instructions (always applied) |
   | `.cursor/rules/resources/connector.mdc` | Connector development rules |
   | `.cursor/rules/resources/model.mdc` | Model development rules |
   | `.cursor/rules/resources/metrics_view.mdc` | Metrics view development rules |
   | `.cursor/rules/resources/explore.mdc` | Explore dashboard rules |
   | `.cursor/rules/resources/canvas.mdc` | Canvas dashboard rules |
   | `.cursor/rules/resources/theme.mdc` | Theme rules |
   | `.cursor/rules/resources/rillyaml.mdc` | `rill.yaml` configuration rules |

2. **Open the project** in Cursor. The `development.mdc` rule is marked `alwaysApply: true` so Cursor loads it automatically. Other rules are loaded on-demand based on context.

## Other AI editors

The instructions that `rill init` generates are derived from the same source — Rill's embedded instruction set. If your preferred editor supports a similar rule or instruction format, you can adapt the generated files. The Claude Code skills (`.claude/skills/`) and Cursor rules (`.cursor/rules/`) contain the same content in different formats.

## Updating instructions

When you upgrade Rill, the built-in instructions may have been updated with new resource types or improved guidance. To refresh the generated files:

```bash
# Overwrite existing Claude Code instructions
rill init --template claude --force

# Overwrite existing Cursor rules
rill init --template cursor --force
```

:::caution
Using `--force` will overwrite any custom edits you've made to the generated files. If you've added project-specific instructions to `CLAUDE.md` or the rule files, back them up first.
:::
