---
title: Ready for Fast Dashboards with Rill?
slug: /
sidebar_label: Home
sidebar_position: 00
hide_table_of_contents: true
---

import TileIcon from '@site/src/components/TileIcon';

## Install 
Install Rill using the command below, then launch your first project to explore core features hands-on.

```bash
curl https://rill.sh | sh
rill start my-rill-project
```

<!-- <img src = 'https://storage.googleapis.com/prod-cdn.rilldata.com/docs/rill_hero.gif' class='rounded-gif' />
<br /> -->

##  Examples

Browse our [example repository](https://github.com/rilldata/rill-examples/) to find a project that suites your needs or see them in action in our demo project by hitting [Live Demo →](https://ui.rilldata.com/demo). Some projects have a walkthrough, too! 

```bash
git clone https://github.com/rilldata/rill-examples.git
cd rill-examples/rill-openrtb-prog-ads
curl https://rill.sh | sh
rill start
```


  **Programmatic Ads/OpenRTB**: bidstream data for programmatic advertisers to optimize pricing strategies, look for inventory opportunities, and improve campaign performance
  
  [GitHub →](https://github.com/rilldata/rill-examples/tree/main/rill-openrtb-prog-ads) | [Walkthrough →](/guides/openrtb-analytics) | [Live Demo →](https://ui.rilldata.com/demo/rill-openrtb-prog-ads)


  **Cost Monitoring**: based on Rill's own internal dashboards, cloud infrastructure data (compute, storage, pipeline statistics, etc.) merged with customer data to analyze bottlenecks and look for efficiencies

  [GitHub →](https://github.com/rilldata/rill-examples/tree/main/rill-cost-monitoring) | [Walkthrough →](/guides/cost-monitoring-analytics) | [Live Demo →](https://ui.rilldata.com/demo/rill-cost-monitoring)


  **GitHub Analytics**: analyze GitHub activity to understand what parts of your codebase are most active, analyze contributor productivity, and evaluate the intersections between commits and files

  [GitHub →](https://github.com/rilldata/rill-examples/tree/main/rill-github-analytics) | [Walkthrough →](/guides/github-analytics) | [Live Demo →](https://ui.rilldata.com/demo/rill-github-analytics)


  **Final Tutorial Project**: A finalized version of the tutorial project with many working examples. It's a good place to reference any newer features and is updated regularly

  [GitHub →](https://github.com/rilldata/rill-examples/tree/main/my-rill-tutorial) | [Walkthrough →](/guides/rill-basics/launch) | [Live Demo →](https://ui.rilldata.com/demo/my-rill-tutorial)
  

  **App Engagement**: a conversion dataset used by marketers, mobile developers, or product teams to analyze funnel steps

  [GitHub →](https://github.com/rilldata/rill-examples/tree/main/rill-app-engagement) | [Live Demo →](https://ui.rilldata.com/demo/rill-app-engagement)



## Quickstart

Take a look at our [Quick start](get-started/get-started.md) for a quick guide to get started with your own data! We use a public GCS dataset but you can follow along with your own data. Take a look at our [connectors docs](/reference/connectors/) for the exact steps to connect to your data.

Or, if you're looking for more guides, check out our [Guides](/guides/) section!



## Release Notes

Curious about what's new? Check out our latest and greatest updates in the [release notes!](https://docs.rilldata.com/notes)



## Next Steps

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