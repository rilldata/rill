---
title: Get Started with Rill
slug: /
sidebar_label: Home
sidebar_position: 00
---

import FeatureList from '@site/src/components/FeatureList';

Rill Developer users encompass the data engineers, analytics engineers, BI developers, and platform teams — all building data pipelines, defining metrics, and delivering interactive dashboards with Rill. In addition to Rill Developer for local project development, these docs cover deploying to Rill Cloud, embedding dashboards into your own applications, and integrating with external tools and APIs.

Looking to **explore dashboards and data** that your team has already set up? Head over to the [User Guide](/guide).

## Install

Install Rill using the command below, then launch your first project to explore core features hands-on. For more installation methods, see our [complete installation instructions](/developers/get-started/install).

```bash
curl https://rill.sh | sh
rill start my-rill-project
```

New to Rill? Follow the [Quickstart](/developers/get-started/quickstart) to build your first dashboard with a public dataset, or browse our [tutorials](/developers/tutorials/).

## Start Developing

Rill Developer is a local application that makes it easy to build end-to-end analytics pipelines. Connect to data sources, write SQL or YAML models for last-mile transformations, define a metrics layer with measures and dimensions, and preview interactive dashboards — all from your local machine. See the full [Build](/developers/build) docs for details.

<FeatureList items={[
  { name: "Connectors", description: "Connect to S3, GCS, BigQuery, Snowflake, ClickHouse, and more", link: "/developers/build/connectors" },
  { name: "Models", description: "Transform and prepare data with SQL or YAML models", link: "/developers/build/models" },
  { name: "Metrics Views", description: "Define measures, dimensions, and time series for dashboards", link: "/developers/build/metrics-view" },
  { name: "Dashboards", description: "Build Explore and Canvas dashboards", link: "/developers/build/dashboards" },
  { name: "Custom APIs", description: "Create API endpoints to retrieve data outside of Rill", link: "/developers/build/custom-apis" },
]} />

Looking for hands-on examples? Browse our [tutorials and example projects](/developers/tutorials/) for guided walkthroughs you can clone and explore.

## Deploy to Rill Cloud

Rill Developer is great for building and testing locally, but once you're ready to share your work, deploy your project to Rill Cloud. Rill Cloud is a fully managed service where your team can explore dashboards, ask questions with AI Chat, set up alerts, and schedule reports — no local setup required for consumers.

<FeatureList items={[
  { name: "Local vs Cloud", description: "Understand the differences between local and cloud", link: "/developers/deploy/cloud-vs-developer" },
  { name: "Deploy to Rill Cloud", description: "Publish your project to Rill Cloud", link: "/developers/deploy/deploy-dashboard" },
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

<div className="tile-icon-grid">
<TileIcon
  header="Connect Data Sources"
  content="Connect to your data sources and start ingesting data into Rill for analysis."
  link="/developers/build/connectors"
  icon={<img src="/img/get-started/connect.svg" alt="Connect" style={{ width: 24, height: 24 }} />}
/>
<TileIcon
  header="Prepare Your Data"
  content="Transform and prepare your data with Rill's powerful ETL capabilities."
  link="/developers/build/models"
  icon={<img src="/img/get-started/model.svg" alt="Model" style={{ width: 24, height: 24 }} />}
/>
<TileIcon
  header="Create a Metrics Layer"
  content="Build a metrics layer to define key business metrics and KPIs."
  link="/developers/build/metrics-view"
  icon={<img src="/img/get-started/metrics.svg" alt="Metrics" style={{ width: 24, height: 24 }} />}
/>
<TileIcon
  header="Embed a Dashboard"
  content="Embed Rill dashboards into your applications and workflows."
  link="/developers/embed/dashboards"
  icon={<img src="/img/get-started/embed.svg" alt="Embed" style={{ width: 24, height: 24 }} />}
/>
## Join the Community

Rill is open source — browse the code on [GitHub](https://github.com/rilldata/rill) or join the community on [Discord](https://discord.gg/2ubRfjC7Rh). If you're a customer, reach out directly via your dedicated Slack channel or contact support.
