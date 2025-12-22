---
title: Ready for Fast Dashboards with Rill?
slug: /
sidebar_label: Home
sidebar_position: 00
---

import TileIcon from '@site/src/components/TileIcon';

## Install 
Install Rill using the command below, then launch your first project to explore core features hands-on. For more installation methods, see our [complete installation instructions docs](/get-started/install). 

```bash
curl https://rill.sh | sh
rill start my-rill-project
```

## Quickstart

Take a look at our [Quick start](/get-started/quickstart) for a quick guide to get started with your own data! We use a public GCS dataset but you can follow along with your own data. Take a look at our [connectors docs](/build/connectors) for the exact steps to connect to your data.

Or, if you're looking for more guides, check out our [Guides](/guides) section!

## Examples

Browse our [repository of examples](https://github.com/rilldata/rill-examples) to find a project that suits your needs or see them in action in our demo project by hitting [Live Demo →](https://ui.rilldata.com/demo). Some projects have a walkthrough, too! 

```bash
git clone https://github.com/rilldata/rill-examples.git
cd rill-examples/rill-openrtb-prog-ads #swap this to the project that interests you!
curl https://rill.sh | sh
rill start
```

### Programmatic Ads/OpenRTB

Bidstream data for programmatic advertisers to optimize pricing strategies, look for inventory opportunities, and improve campaign performance.

- <a href="https://github.com/rilldata/rill-examples/tree/main/rill-openrtb-prog-ads">GitHub →</a><br />
- <a href="/guides/openrtb-analytics">Walkthrough →</a><br />
- <a href="https://ui.rilldata.com/demo/rill-openrtb-prog-ads">Live Demo →</a> 

### Cost Monitoring

Based on Rill's own internal dashboards, cloud infrastructure data (compute, storage, pipeline statistics, etc.) merged with customer data to analyze bottlenecks and look for efficiencies.

- <a href="https://github.com/rilldata/rill-examples/tree/main/rill-cost-monitoring">GitHub →</a><br />
- <a href="/guides/cost-monitoring-analytics">Walkthrough →</a><br />
- <a href="https://ui.rilldata.com/demo/rill-cost-monitoring">Live Demo →</a>

### GitHub Analytics

Analyze GitHub activity to understand what parts of your codebase are most active, analyze contributor productivity, and evaluate the intersections between commits and files.

- <a href="https://github.com/rilldata/rill-examples/tree/main/rill-github-analytics">GitHub →</a><br />
- <a href="/guides/github-analytics">Walkthrough →</a><br />
- <a href="https://ui.rilldata.com/demo/rill-github-analytics">Live Demo →</a>

### App Engagement

A conversion dataset used by marketers, mobile developers, or product teams to analyze funnel steps.

- <a href="https://github.com/rilldata/rill-examples/tree/main/rill-app-engagement">GitHub →</a><br />
- <a href="https://ui.rilldata.com/demo/rill-app-engagement">Live Demo →</a>

### Kitchen-sink example

A compilation of projects that deep dives into Rill's features using ClickHouse's GitHub commit information.

- <a href="https://github.com/rilldata/rill-examples/tree/main/my-rill-tutorial">GitHub →</a><br />
- <a href="/guides/rill-basics/launch">Walkthrough →</a><br />
- <a href="https://ui.rilldata.com/demo/my-rill-tutorial">Live Demo →</a>

## Explore Rill's Capabilities

<div className="tile-icon-grid">
<TileIcon
  header="Connect Data Sources"
  content="Connect to your data sources and start ingesting data into Rill for analysis."
  link="/connect"
  icon={<img src="/img/get-started/connect.svg" alt="Connect" style={{ width: 24, height: 24 }} />}
/>
<TileIcon
  header="Prepare Your Data"
  content="Transform and prepare your data with Rill's powerful ETL capabilities."
  link="/build/models"
  icon={<img src="/img/get-started/model.svg" alt="Model" style={{ width: 24, height: 24 }} />}
/>
<TileIcon
  header="Create a Metrics Layer"
  content="Build a metrics layer to define key business metrics and KPIs."
  link="/build/metrics-view"
  icon={<img src="/img/get-started/metrics.svg" alt="Metrics" style={{ width: 24, height: 24 }} />}
/>
<TileIcon
  header="Explore Your Data"
  content="Use Rill's interactive data exploration tools to discover insights."
  link="/explore/dashboard-101"
  icon={<img src="/img/get-started/explore.svg" alt="Explore" style={{ width: 24, height: 24 }} />}
/>
<TileIcon
  header="Embed a Dashboard"
  content="Embed Rill dashboards into your applications and workflows."
  link="/integrate/embedding"
  icon={<img src="/img/get-started/embed.svg" alt="Embed" style={{ width: 24, height: 24 }} />}
/>
<TileIcon
  header="Release Notes"
  content="Curious about what's new?"
  link="/notes"
  icon={<img src="/img/get-started/notification.svg" alt="Notification" style={{ width: 24, height: 24 }} />}
/>

</div>
