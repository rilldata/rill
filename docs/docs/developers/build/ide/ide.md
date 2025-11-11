---
title: Build Rill Projects with Your Favorite IDE
description: Use VS Code, IntelliJ, or any IDE to create and edit Rill projects with real-time feedback
sidebar_label: External IDE Integration
sidebar_position: 00
---


## Use Any IDE for Rill Development

Rill projects are just files and folders that you can edit with any code editor or IDE. Whether you prefer VS Code, IntelliJ, Vim, or any other editor, you can create and modify Rill projects directly from your favorite development environment.

### How It Works

Rill projects consist of:
- **SQL files** (`.sql`) for models and metrics views
- **YAML files** (`.yml`) for project configuration
- **Data files** in various formats

You can edit these files in any IDE, and Rill will automatically detect changes and provide real-time feedback.

<img src='https://cdn.rilldata.com/docs/release-notes/36_hot_reload.gif' class='rounded-gif' />
<br />

## Using AI Agents to Help Build Rill Projects

You can also use IDEs with integrated AI agents like Cursor or WindSurf to accelerate your Rill development workflow.


### Best Practices When Using AI Agents

When working with AI agents, start with clear prompts that specify your Rill version, database engine, and project context. Always review generated code thoroughly rather than copying and pasting without understanding the logic. Test changes incrementally by validating small modifications before implementing larger ones. Cross-reference all parameters and syntax with official Rill documentation to ensure accuracy. Remember to treat AI-generated code as a starting template to refine rather than final production code.

:::warning AI Agent Limitations

**Important**: AI agents can hallucinate parameters, functions, and syntax that don't exist in Rill or your underlying database engine. Always verify:
- Generated SQL syntax against your database's capabilities
- YAML configuration parameters against Rill's documentation
- Function names and parameters for accuracy
- Generated code in a development environment before production use

:::