---
title: Get Started with Rill
slug: /
sidebar_label: Home
sidebar_position: 00
---

import FeatureList from '@site/src/components/FeatureList';

Rill developer users encompass the data engineers, analytics engineers, BI developers, and platform teams — all building data pipelines, defining metrics, and delivering interactive dashboards with Rill. In addition to Rill Developer for local project development, these docs cover deploying to Rill Cloud, embedding dashboards into your own applications, and integrating with external tools and APIs.

Looking to **explore dashboards and data** that your team has already set up? Head over to the [User Guide](/guide).

## Install

Install Rill using the command below, then launch your first project to explore core features hands-on. For more installation methods, see our [complete installation instructions](/developers/get-started/install).

```bash
curl https://rill.sh | sh
rill start my-rill-project
```

New to Rill? Follow the [Quickstart](/developers/get-started/quickstart) to build your first dashboard with a public dataset, or browse our [tutorials](/developers/tutorials/).

## Build with AI Agents

Use **Claude Code**, **Cursor**, or other AI coding agents to build Rill projects using natural language. Run `rill init --template claude` to scaffold instructions and skills that teach your agent Rill's resource types, YAML schemas, and development workflow — then start prompting.

```bash
# Install Rill and create a project
curl https://rill.sh | sh
rill init

# Add Claude Code instructions and start building
rill init --template claude
claude
```

Connect to your running Rill instance via MCP for live data introspection and validation:

```bash
claude mcp add --transport http rill http://localhost:9009/mcp
```

See the [AI Agents guide](/developers/build/ai-agents) for full setup, or try the [Claude Code quickstart](/developers/get-started/claude-code-quickstart) to build your first project with AI.

## Start Developing

Rill Developer is a local application that makes it easy to build end-to-end analytics pipelines. Connect to data sources, write SQL or YAML models for last-mile transformations, define a metrics layer with measures and dimensions, and preview interactive dashboards — all from your local machine. See the full [Build](/developers/build) docs for details.

### Core Concepts

<FeatureList items={[
  { name: "Connectors", description: "Connect to S3, GCS, BigQuery, Snowflake, ClickHouse, and more", link: "/developers/build/connectors" },
  { name: "Models", description: "Transform and prepare data with SQL or YAML models", link: "/developers/build/models" },
  { name: "Metrics Views", description: "Define measures, dimensions, and time series for dashboards", link: "/developers/build/metrics-view" },
  { name: "Dashboards", description: "Build Explore and Canvas dashboards", link: "/developers/build/dashboards" },
  { name: "Custom APIs", description: "Create API endpoints to retrieve data outside of Rill", link: "/developers/build/custom-apis" },
  { name: "AI Agents", description: "Set up Claude Code or Cursor to build Rill projects with AI", link: "/developers/build/ai-agents" },
  { name: "Tutorials", description: "Guided walkthroughs and example projects to clone and explore", link: "/developers/tutorials/" },
]} />

## Deploy to Rill Cloud

Rill Developer is great for building and testing locally, but once you're ready to share your work, deploy your project to Rill Cloud. Rill Cloud is a fully managed service where your team can explore dashboards, ask questions with AI Chat, set up alerts, and schedule reports — no local setup required for consumers.

<FeatureList items={[
  { name: "Cloud vs Developer", description: "Understand the differences between local and cloud", link: "/developers/deploy/cloud-vs-developer" },
  { name: "Deploy a Dashboard", description: "Publish your project to Rill Cloud", link: "/developers/deploy/deploy-dashboard" },
  { name: "Credentials", description: "Configure deployment credentials", link: "/developers/deploy/deploy-credentials" },
]} />

## Integrate into your Application

Rill is designed to fit into your existing stack. Embed interactive dashboards directly into your web applications using iframes, programmatically control them with the Embed API, or build custom API endpoints to pull Rill data into external tools and workflows. See the full [Integrate](/developers/integrate) docs for details.

<FeatureList items={[
  { name: "Embed Dashboards", description: "Embed Rill dashboards into your applications using iframes", link: "/developers/integrate/embedding" },
  { name: "Embed API", description: "Programmatically control embedded dashboards", link: "/developers/integrate/embed-iframe-api" },
  { name: "Custom APIs", description: "Integrate custom APIs in external applications", link: "/developers/integrate/custom-api" },
  { name: "URL Parameters", description: "Reference for all dashboard URL parameters", link: "/developers/integrate/url-parameters" },
]} />

## Join the Community

Rill is open source and has an active community. Browse the source code, file issues, or request features on GitHub. Get help and connect with other users on Discord. If you're a customer, reach out directly via your dedicated Slack channel or contact support.

<FeatureList items={[
  { name: "GitHub", description: "Source code, issues, and feature requests", link: "https://github.com/rilldata/rill" },
  { name: "Discord", description: "Community support and discussion", link: "https://discord.gg/2ubRfjC7Rh" }
]} />
