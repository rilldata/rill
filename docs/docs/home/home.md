---
title: Ready for Fast Dashboards with Rill?
slug: /
sidebar_label: Home
sidebar_position: 00
hide_table_of_contents: true
---

## Install 
Install Rill using the command below, then launch your first project to explore core features hands-on.

```bash
curl https://rill.sh | sh
rill start my-rill-project
```

<!-- <img src = 'https://storage.googleapis.com/prod-cdn.rilldata.com/docs/rill_hero.gif' class='rounded-gif' />
<br /> -->

##  Examples

Explore our [example repository](https://github.com/rilldata/rill-examples/) to kickstart your project or see them in action in our [demo project](https://ui.rilldata.com/demo).

<div className="tile-icon-grid">
<TileIcon
  header="Programmatic Ads / OpenRTB"
  content="Bidstream data for programmatic advertisers to optimize pricing strategies and improve campaign performance."
  link="https://ui.rilldata.com/demo/rill-openrtb-prog-ads/explore/auction_explore"
  linkLabel="Explore Demo"
  target="_blank"
  rel="noopener noreferrer"
  githubLink="https://github.com/rilldata/rill-examples/tree/main/rill-openrtb-prog-ads"
  walkthroughLink="/guides/openrtb-analytics"
/>
<TileIcon
  header="GitHub Analytics"
  content="Analyze GitHub activity to understand codebase activity, contributor productivity, and commit patterns."
  link="https://ui.rilldata.com/demo/rill-github-analytics/explore/mcp_servers_commits_explore?tr=rill-YTD&grain=week"
  linkLabel="Explore Demo"
  target="_blank"
  rel="noopener noreferrer"
  githubLink="https://github.com/rilldata/rill-examples/tree/main/rill-github-analytics"
  walkthroughLink="/guides/github-analytics"
/>
<TileIcon
  header="Cost Monitoring"
  content="Cloud infrastructure data merged with customer data to analyze bottlenecks and find efficiencies."
  link="https://ui.rilldata.com/demo/rill-cost-monitoring/explore/metrics_margin_explore"
  linkLabel="Explore Demo"
  target="_blank"
  rel="noopener noreferrer"
  githubLink="https://github.com/rilldata/rill-examples/tree/main/rill-cost-monitoring"
  walkthroughLink="/guides/cost-monitoring-analytics"
/>
<TileIcon
  header="Tutorial Project"
  content="A comprehensive tutorial project with working examples — perfect reference for newer features."
  link="https://ui.rilldata.com/demo/my-rill-tutorial/explore/advanced_explore?tr=P30D"
  linkLabel="Explore Demo"
  target="_blank"
  rel="noopener noreferrer"
  githubLink="https://github.com/rilldata/rill-examples/tree/main/my-rill-tutorial"
  walkthroughLink="/guides/tutorial/rill-basics/launch"
/>
<TileIcon
  header="App Engagement"
  content="Conversion dataset for marketers and product teams to analyze funnel steps and user behavior."
  link="https://ui.rilldata.com/demo/rill-app-engagement/explore/mobile_events_explore"
  linkLabel="Explore Demo"
  target="_blank"
  rel="noopener noreferrer"
  githubLink="https://github.com/rilldata/rill-examples/tree/main/rill-app-engagement"
/>
<TileIcon
  header="Podcasting"
  content="Podcast analytics for creators and platforms to track engagement and episode performance."
  link="https://ui.rilldata.com/demo/sample-podcast-project/canvas/amplify"
  linkLabel="Explore Demo"
  target="_blank"
  rel="noopener noreferrer"
  githubLink="https://github.com/rilldata/rill-examples/tree/main/sample-podcast-project"
/>
</div>


Clone the repository and launch any example project to get started:

```bash
git clone https://github.com/rilldata/rill-examples.git
cd rill-examples/rill-openrtb-prog-ads
curl https://rill.sh | sh rill start
```

## Quickstart

Take a look at our [Start Guide](get-started/get-started.md) for a quick guide to get started with your own data! We use a public GCS dataset but you can follow along with your own data. Take a look at our [connectors docs](/reference/connectors/) for the exact steps to connect to your data.

Or, if you're looking for more guides, check out our [Guides](/guides/) section!



## Release Notes

Curious about what's new? Check out our latest and greatest updates in the [release notes!](https://docs.rilldata.com/notes)



## Next Steps
import TileIcon from '@site/src/components/TileIcon';

<div className="tile-icon-grid">
<TileIcon
  header="Connect Sources"
  content="Connect to your data sources and start ingesting data into Rill for analysis."
  link="/reference/connectors/"
/>
<TileIcon
  header="Last Mile ETL"
  content="Transform and prepare your data with Rill's powerful ETL capabilities."
  link="/build/models/"
/>
<TileIcon
  header="Create Metrics Layer"
  content="Build a metrics layer to define key business metrics and KPIs."
  link="/build/metrics-view/"
/>
<TileIcon
  header="Explore Data"
  content="Use Rill's interactive data exploration tools to discover insights."
  link="/explore/dashboard-101"
/>
<TileIcon
  header="Embed Dashboard"
  content="Embed Rill dashboards into your applications and workflows."
  link="/integrate/embedding/"
/>
<TileIcon
  header="Manage Users"
  content="Set up user management and access controls for your Rill projects."
  link="/manage/user-management/"
/>
<TileIcon
  header="Deploy to Cloud"
  content="Deploy your Rill project to production and share with your team."
  link="/deploy/deploy-dashboard/"
/>
<TileIcon
  header="See Demo Project"
  content="Explore our demo projects to see Rill in action with real data."
  link="https://ui.rilldata.com/demo"
  target="_blank"
  rel="noopener noreferrer"
/>
</div>