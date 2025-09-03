---
title: Build Rill Projects with Your Favorite IDE
description: Use VS Code, IntelliJ, or any IDE to create and edit Rill projects with real-time feedback
sidebar_label: External IDE Integration
sidebar_position: 00
---


## Use Any IDE for Rill Development

Rill projects are just files and folders that you can edit with any code editor or IDE. Whether you prefer VS Code, IntelliJ, Vim, or any other editor, you can create and modify Rill projects directly from your favorite development environment.

## How It Works

BI-as-code is a modern approach to business intelligence that treats analytics assets as code, bringing the same benefits of version control, collaboration, and automation that software development teams have enjoyed for years. With Rill, you can define your entire analytics stack—from data models to dashboards—using code, while still maintaining the flexibility to make UI-based adjustments when needed.

### How It Works in Rill

Rill implements BI-as-code through a combination of:

1. **SQL-based Definitions**: Define your models via SQL to connect to your various sources
2. **YAML Configuration**: Configure your metrics views, dashboards, and project settings via YAML
3. **Git Integration**: Version control your analytics assets
4. **CLI Tools**: Deploy and manage your analytics stack from the command line
   
<div style={{ textAlign: 'center' }}>
  <img src="/img/concepts/metrics-view/metrics-view-components.png" style={{ width: '100%', borderRadius: '15px', padding: '20px' }} />
</div>


For a deeper dive into BI-as-code and its benefits, check out our blog post: [What is BI-as-code?](https://www.rilldata.com/blog/bi-as-code-and-the-new-era-of-genbi)


You can edit these files in any IDE, and Rill will automatically detect changes and provide real-time feedback.

<img src='https://cdn.rilldata.com/docs/release-notes/36_hot_reload.gif' class='rounded-gif' />
<br />



## Using AI Agents to Help Build Rill Projects

You can also use IDEs with integrated AI agents like Cursor or WindSurf to accelerate your Rill development workflow.

:::warning AI Agent Limitations

**Important**: AI agents can hallucinate parameters, functions, and syntax that don't exist in Rill or your underlying database engine. Always verify:
- Generated SQL syntax against your database's capabilities
- YAML configuration parameters against Rill's documentation
- Function names and parameters for accuracy
- Generated code in a development environment before production use

:::

### Best Practices When Using AI Agents

When working with AI agents, start with clear prompts that specify your Rill version, database engine, and project context. Always review generated code thoroughly rather than copying and pasting without understanding the logic. Test changes incrementally by validating small modifications before implementing larger ones. Cross-reference all parameters and syntax with official Rill documentation to ensure accuracy. Remember to treat AI-generated code as a starting template to refine rather than final production code.