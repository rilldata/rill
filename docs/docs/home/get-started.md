---
title: Quickstart & Examples
sidebar_label: Quickstart & Examples
sidebar_position: 10
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

Rill's BI-as-code approach allows you to build and manage dashboards in a codeful way locally + integrate your analytics into other Git processes.
This tutorial is a step by step guide that will help you install Rill, ingest sources, model data and deploy a dashboard online.

:::tip Example projects
If you're looking for a place to get started, check out one of our [example projects](#installing-examples) which cover a variety of sources and use cases. 
:::


## Install Rill

Install `rill` on Mac or Linux ([WSL](https://learn.microsoft.com/en-us/windows/wsl/install) for Windows users) using the installation script:

```
curl https://rill.sh | sh
```

:::tip sharing dashboards in Rill cloud? clone your git repo first
If you plan to share your dashboards, it is helpful to start by creating a repo in Git. Go to https://github.com/new to create a new repo. Then, run the [Rill install script](#install) in your cloned location locally to make deployment easier. 

More details on deploying Rill via Git in our [Deploy section](../deploy/existing-project.md).
:::

## Create your project

Start a new Rill project:

```
rill start my-rill-project
```

The Rill web app runs locally at `http://localhost:9009` and will create code files in the `my-rill-project` directory.

## Load and transform data

On the welcome screen, initialize an example project or load up Rill with your own data (use local files, cloud storage and/or database connections)

Use the app to transform your data (data models) and quickly generate visualizations (dashboards).

## Deploy your dashboard

Once complete, you can deploy any Rill project with a dashboard to an authenticated hosted endpoint:

1. Create a new Github repository on [https://github.com/new](https://github.com/new) and push your `my-rill-project` directory to it
2. Setup continuous deployment from Github to Rill Cloud:
    ```
    cd my-rill-project
    rill deploy
    ```

## Share your dashboard

Projects on Rill Cloud are private by default. To invite others to explore your project, run:
```
rill user add
```

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

You can file an issue [on GitHub](https://github.com/rilldata/rill/issues/new/choose) or reach us in our [Discord channel](https://bit.ly/3unvA05).
