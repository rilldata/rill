---
title: Core Concepts
sidebar_label: Core Concepts
sidebar_position: 10
---

Rill is built around a few key concepts that work together to create a complete analytics workflow:

## Core Concepts Overview

1. **[OLAP Engine](#what-is-olap)** - Connect to your data sources and import data into Rill's embedded analytics engine or connect to your OLAP engine via a live connector
2. **[BI-As-Code](/get-started/why-rill#bi-as-code)** - Create Rill objects (models, metrics views, dashboards) with code components (YAML) to prepare your data for visualizations.
3. **[Rill Cloud vs Rill Developer](#what-is-rill-cloud-and-rill-developer)** - Deploy and share dashboards with your team for data-driven decision making

## What is OLAP?

OLAP (or Online Analytical Processing) is a computational approach designed to enable rapid, multidimensional analysis of large volumes of data. With OLAP, data is typically organized into cubes instead of traditional two-dimensional tables, which can facilitate complex queries and data analysis in a way that is significantly more efficient and user-friendly for analytical tasks. In particular, OLAP databases can be especially well suited for BI use cases that require deep, multidimensional analysis or real-time / user-facing analytics and applications. Additionally, many modern OLAP databases are optimized to ingest large volumes of data, execute low-latency queries with high throughput, and process billions of rows quickly with an emphasis on speed and efficiency in data retrieval. 

Unlike traditional relational databases or data warehouses that are optimized for transaction processing (with a focus on CRUD operations), OLAP databases are designed for query speed and complex analysis. Rather than storing data in a row-oriented manner, optimizing for transactional efficiency and operational queries, most OLAP databases are columnar and use pre-aggregated multidimensional cubes to speed up analytical queries. This allows a broad range of ad hoc queries and analysis to be performed without needing predefined schemas that are tailored to specific queries, and it's this flexibility that enables the highly interactive slice-and-dice exploration of data that powers Rill dashboards. This paradigm allows OLAP to be particularly well-suited for organizations and teams that want to dive deep into and understand their data to support decision-making processes, where speed and flexibility in the actual data analysis are important. 

See our dedicated [Connect Docs](/build/connectors) for more information on [OLAP engines](/build/connectors/olap) and available [data source connectors](/build/connectors/data-source).

### Live Connectors in Rill

Rill supports [live connectors](/build/connectors/olap) that allow you to connect directly to existing tables in external OLAP engines like ClickHouse, Druid, and Pinot. Instead of importing data into Rill, you can build metrics views and dashboards on top of your existing tables, keeping your data where it already lives.

This approach is ideal when:
- Your data is already in a production OLAP database
- You need to query large datasets that are managed elsewhere
- Your organization has existing data infrastructure you want to leverage

With live connectors, you define your [metrics views](/build/metrics-view) that reference external tables, and Rill queries them directly to power your dashboards and visualizations.

:::info Want to see OLAP in action?

Check [here](https://www.rilldata.com/case-studies) to see examples of use cases that can be powered by OLAP.

:::

## What is a Metrics Layer?

A metrics layer is a `centralized framework` used to define and organize **key metrics** for your organization. Having a centralized layer allows an organization to easily manage and reuse calculations across various reports, dashboards, and data tools. As Rill continues to grow, we decided to separate the metrics layer from the dashboard configuration.

:::tip
Starting from version 0.50, the operation of creating a dashboard via AI will create a metrics view and dashboard separately in their own respective folders and navigate you to a preview of your dashboard. If you find that some metrics need to be modified, you will need to navigate to your [metrics/model_name_metrics.yaml](/build/metrics-view) file. 


Assuming that you have the '*' (select all) in your dashboard configurations, any changes will automatically be reflected on your [explore dashboard](/build/dashboards).
:::


Within Rill, we refer to metrics layers as a metrics view. It's a single view or file that contains all of your measures and dimensions that will be used to display the data in various ways. The metrics view also contains some configuration settings that are required to ensure that the data being displayed is as accurate as you need it to be. 


<img src = '/img/tutorials/rill-basics/new-viz-editor.png' class='rounded-gif' />
<br />

:::tip
It is possible to develop the metrics layer in a traditional BI-as-code manner as well as via the UI. To switch between the two, select the toggle in the top right corner.
:::

For more information on Metrics, see our [metrics view docs](/build/metrics-view)!

## What is Rill Cloud and Rill Developer?

Rill offers two unique but complementary experiences within our broader product suite, **Rill Cloud** and **Rill Developer**.

As the name suggests, Rill _Developer_ is designed with the developer in mind, where project development actually occurs. Rill Developer is meant for the primary developers of project assets and dashboards, allowing them to import, wrangle, iterate on, and explore the data before presenting it for broader consumption by the team. Rill Developer is meant to run on your local machine - see here for some [recommendations and best practices](/build/models/performance) - but it is a simple process to [deploy a project](/deploy/deploy-dashboard) once ready to Rill Cloud.


Rill Cloud, on the other hand, is designed for our dashboard consumers and allows broader team members to easily collaborate. Once the developer has deployed the dashboard onto Rill Cloud, these users will be able to utilize the dashboards to interact with their data, set alerts / bookmarks, investigate nuances / anomalies, or otherwise perform everyday tasks for their business needs at Rill speed.

For more information on deploying to Rill Cloud, see our [Deploy section](/deploy).