---
title: Quickstart & Examples
sidebar_label: Quickstart & Examples
sidebar_position: 10
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Quickstart

Run the following from the CLI to start Rill. Select a project from the UI or add your own data.

```bash
rill start my-rill-project
```

![img](/img/tutorials/101/new-rill-project.png)


## Example Projects Repository

We have created several example projects that highlight common use cases for Rill. 

The monorepo for these examples can be found at: https://github.com/rilldata/rill-examples/

Each example project includes a ReadMe with more details on:

- Source data in the dataset
- Dimension and metric definitions
- Example dashboard analyses

Current projects include:

- [App Engagement](https://github.com/rilldata/rill-examples/tree/main/rill-app-engagement): a conversion dataset used by marketers, mobile developers or product teams to analyze funnel steps
- [Cost Monitoring](https://github.com/rilldata/rill-examples/tree/main/rill-cost-monitoring): based off of Rill's own internal dashboards, cloud infrastructure data (compute, storage, pipeline statistics, etc.) merged with customer data to analyze bottlenecks and look for efficiencies
- [GitHub Analytics](https://github.com/rilldata/rill-examples/tree/main/rill-github-analytics): analyze GitHub activity to understand what parts of your codebase are most active, analyze contributor productivity, and evaluate the intersections between commits and files
- [Programmatic Ads/OpenRTB](https://github.com/rilldata/rill-examples/tree/main/rill-openrtb-prog-ads): bidstream data for programmtic advertisers to optimize pricing strategies, look for inventory opportunities, and improve campaign performance
- [311 Operations](https://github.com/rilldata/rill-examples/tree/main/rill-311-ops): a live datastream of 311 call centers from various locations in the US for example operational analytics 
- [Final Tutorial Project](https://github.com/rilldata/rill-examples/tree/main/my-rill-tutorial): A finalized version of the tutorial project with many working examples, a good place to reference any newer features, updated regularly.

## Installing Examples

You can install `rill` using our installation script:

```
curl https://rill.sh | sh
```

To run an example (in this case our Programmatic/OpenRTB dataset):
```
git clone https://github.com/rilldata/rill-examples.git
cd rill-examples/rill-openrtb-prog-ads
rill start
```

Rill will build your project from data sources to dashboard and then launch in a new browser window.


## We want to hear from you

You can file an issue [on GitHub](https://github.com/rilldata/rill/issues/new/choose) or reach us in our [Discord channel](https://discord.gg/DJ5qcsxE2m). If you want to contact Rill Support, please see our [Contact](contact.md#contacting-support) page for additional options.
