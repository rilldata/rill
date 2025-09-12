---
title: Building Your Data Analytics Pipeline in Rill
sidebar_label: Build
sidebar_position: 0
---

Rill provides a comprehensive platform for building end-to-end data analytics solutions. Once you've connected to your source data or connected to your own OLAP database, you can start building with Rill. Don't forget to check out the [reference documentation!](/reference/project-files)

### Modeling and Last Mile ETL
- [**Models Overview**](/build/models) - Overview of models available in Rill
- [**Getting Started with Models**](/build/models/models-101) - What are models and important topics to remember as you create your own.
- [**Differences between SQL and YAML Models**](/build/models/model-differences) - Compare SQL and YAML models and learn when to use 
- [**SQL Models**](/build/models/sql-models) - Transform and prepare your data
- [**Data Refresh**](/build/models/data-refresh) - Schedule your data refresh  
- [**YAML Models**](/build/models/model-differences#yaml-models) - Incremental ingestion, data partitions, and more
- [**Additional Model Templating**](/build/models/templating) - Separate your production and development settings 
- [**Performance Optimization**](/build/models/performance) - Optimize model performance and data refresh strategies



### Define Measures and Dimensions
- [**What are Metrics Views?**](/build/metrics-view/what-are-metrics-views) - Learn about Metrics Views!
- [**Time Series**](/build/metrics-view/time-series) - The backbone of Rill Dashboards, time
- [**Dimensions**](/build/metrics-view/dimensions) - Aggregate your dimensions further in your metrics view
- [**Measures**](/build/metrics-view/measures) - Aggregate your measures further in your metrics view
- [**Define Data Access**](/build/metrics-view/security) - Define access and row access policies

### Build Visualizations

- [**Slice-and-Dice Dashboards**](/build/dashboards) - Explore and find insights in your data
- [**Traditional Visualizations**](/build/canvas) - Visualize your data with various chart types
- [**Define Dashboard Access**](/build/dashboards#define-dashboard-access) - Set a SQL boolean query that defines access to dashboard

### Build Integrations with External Applications
- [**Custom API**](/build/custom-apis) - Create custom API endpoints to retrieve data outside of Rill


### Project Settings

- [**rill.yaml**](/build/rill-project-file) - Configure your Rill project defaults
- [**Structure your Project**](/build/structure) - Structure folder architecture in Rill
- [**Use your favorite IDE**](/build/ide) - Utilize your favorite IDE to build Rill projects
- [**Debugging Rill Developer**](/build/debugging/trace-viewer) - Troubleshoot dashboard access, trace your reconciled resources, and understand project logs  