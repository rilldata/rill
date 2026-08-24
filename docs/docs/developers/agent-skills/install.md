---
title: "Installing Agent Skills"
description: Install Rill agent skills for Claude Code, Cursor, and AGENTS.md-compatible agents
sidebar_label: "Installation"
sidebar_position: 10
---

Rill's agent skills are bundled with the Rill CLI and generated directly into your project with `rill init --agent`. This page covers installation for each supported tool.

## Requirements

- [Rill CLI installed](/developers/get-started/install)
- An AI coding agent: [Claude Code](https://docs.anthropic.com/en/docs/claude-code/overview), [Cursor](https://www.cursor.com/), or another agent that supports the `AGENTS.md` convention

## Install with the CLI

### New project

Run `rill init` interactively and select an option at the "Agent instructions" prompt:

```bash
rill init
? Project name my-rill-project
? OLAP engine duckdb
? Agent instructions all
```

Or pass the `--agent` flag directly:

```bash
rill init my-project --agent all
```

The `all` option generates the skills in every supported format, so the project works with any agent out of the box. To generate files for a single tool only, pass `claude`, `cursor`, or `agentsmd` instead.

### Existing project

To add agent skills to an existing Rill project without affecting your other files, run `rill init` with only the `--agent` flag:

```bash
rill init ./my-existing-project --agent all
```

### Options

The `--agent` flag accepts the following values:

| Option | Description |
| --- | --- |
| `claude` | Claude Code skills in `.claude/` (default) |
| `cursor` | Cursor rules in `.cursor/` |
| `agentsmd` | Tool-agnostic `AGENTS.md` format |
| `all` | All of the above |
| `none` | No agent instructions |

:::tip Commit the generated files
The generated files are plain text and designed to be committed to version control. Check them in so everyone on your team — and every agent — works from the same instructions.
:::

## Claude Code

The `all` and `claude` options generate:

- **`.claude/CLAUDE.md`** — entry point that tells Claude Code to load the Rill development skills before making changes
- **`.claude/skills/rill-*/SKILL.md`** — one skill per topic: the development workflow, data analysis, and each resource type (see the [skills reference](/developers/agent-skills/skills))
- **`.mcp.json`** — connects Claude Code to Rill Developer's local MCP server

Open Claude Code in the project directory. When prompted, trust the project's MCP configuration so Claude Code can connect to the Rill MCP server. The skills activate automatically based on your requests.

## Cursor

The `all` and `cursor` options generate:

- **`.cursor/rules/AGENTS.mdc`** — an always-applied rule that points Cursor at the Rill development instructions
- **`.cursor/rules/*.mdc`** — rules for the development workflow, data analysis, and each resource type, applied automatically based on their descriptions
- **`.cursor/mcp.json`** — connects Cursor to Rill Developer's local MCP server

Open the project in Cursor and enable the `rill-developer` MCP server when prompted.

## AGENTS.md-compatible agents

This format works with agents that follow the [AGENTS.md](https://agents.md) convention, including OpenAI Codex, Gemini CLI, and GitHub Copilot. The `all` and `agentsmd` options generate:

- **`AGENTS.md`** — entry point at the project root
- **`.agents/skills/rill-*/SKILL.md`** — one skill per topic, in the same format as the Claude Code skills
- **`.mcp.json`** — MCP server configuration for agents that support it

## Connect the MCP server

The generated MCP configuration points at Rill Developer's local MCP server:

```json
{
  "mcpServers": {
    "rill-developer": {
      "type": "http",
      "url": "http://localhost:9009/mcp"
    }
  }
}
```

The server becomes available when you start Rill Developer:

```bash
rill start my-project
```

Through the MCP server, your agent can check resource status, inspect table schemas, run SQL and metrics queries, and read and write project files with immediate feedback on parse and reconcile errors.

## Next steps

- [Agentic Quickstart](/developers/get-started/ai-quickstart) — build a project end-to-end with an AI agent
- [Agent skills reference](/developers/agent-skills/skills) — what each skill covers
- [Updating agent skills](/developers/agent-skills/update) — keep the skills up to date with new Rill releases
