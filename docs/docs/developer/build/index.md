---
title: Building Your Data Analytics Pipeline in Rill
sidebar_label: Build
sidebar_position: 0
---

Rill provides a comprehensive platform for building end-to-end data analytics solutions. Once you've connected to your source data or connected to your own OLAP database, you can start building with Rill. Don't forget to check out the [reference documentation!](/reference/project-files)

### What is a Rill project?
- [**Getting Started with your Rill Project**](/developer/build/getting-started) - Understand Rill project structure and configuration

### Connect to your Data
- [**Connectors Overview**](/developer/build/connectors) - Learn about connecting Rill to data sources and OLAP engines
- [**OLAP Engines**](/developer/build/connectors/olap) - Configure DuckDB, ClickHouse, Druid, or Pinot for analytics
- [**Data Sources**](/developer/build/connectors/data-source) - Connect to S3, GCS, BigQuery, Snowflake, and more
- [**Configure Local Credentials**](/developer/build/connectors/credentials) - Set up credentials for local development
- [**Dev/Prod Connectors**](/developer/build/connectors/templating) - Separate your production and development data sources 
  
### Modeling and Last Mile ETL
- [**Models Overview**](/developer/build/models) - Overview of models available in Rill
- [**Getting Started with Models**](/developer/build/models/models-101) - What are models and important topics to remember as you create your own.
- [**Differences between SQL and YAML Models**](/developer/build/models/model-differences) - Compare SQL and YAML models and learn when to use 
- [**SQL Models**](/developer/build/models/sql-models) - Transform and prepare your data
- [**Data Refresh**](/developer/build/models/data-refresh) - Schedule your data refresh  
- [**YAML Models**](/developer/build/models/model-differences#yaml-models) - Incremental ingestion, data partitions, and more
- [**Additional Model Templating**](/developer/build/models/templating) - Separate your production and development settings 
- [**Performance Optimization**](/developer/build/models/performance) - Optimize model performance and data refresh strategies

### Define Measures and Dimensions
- [**What are Metrics Views?**](/developer/build/metrics-view/what-are-metrics-views) - Learn about Metrics Views!
- [**Underlying Model/Table**](/developer/build/metrics-view/underlying-model) - Powering your metrics view with a `model` or `table`
- [**Time Series**](/developer/build/metrics-view/time-series) - The backbone of Rill Explore Dashboards, time
- [**Dimensions**](/developer/build/metrics-view/dimensions) - Expand your dimensions capabilities further in your metrics view
- [**Measures**](/developer/build/metrics-view/measures) - Aggregate your measures further in your metrics view
- [**Define Data Access**](/developer/build/metrics-view/security) - Define access and row access policies

### Build Visualizations
- [**Dashboards Types**](/developer/build/dashboards/dashboards-101) - What are the visualization options in Rill, and how are they different?
- [**Slice-and-Dice Dashboards**](/developer/build/dashboards/explore) - Explore and find insights in your data
- [**Traditional Visualizations**](/developer/build/dashboards/canvas) - Visualize your data with various chart types
- [**Canvas Components**](/developer/build/dashboards/canvas-widgets) - See all of our supported components!
- [**Define Dashboard Access**](/developer/build/dashboards/customization#define-dashboard-access) - Set a SQL boolean query that defines access to dashboard

### Build Integrations with External Applications
- [**Custom API**](/developer/build/custom-apis) - Create custom API endpoints to retrieve data outside of Rill

### Project Settings
- [**Project Configuration**](/developer/build/project-configuration) - Configure your Rill Project and set default behavior
- [**Structure your Project**](/developer/build/structure) - Structure folder architecture in Rill
- [**Use your favorite IDE**](/developer/build/ide) - Utilize your favorite IDE to build Rill projects
- [**Debugging Rill Developer**](/developer/build/debugging/trace-viewer) - Troubleshoot dashboard access, trace your reconciled resources, and understand project logs  