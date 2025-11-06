---
title: Building Your Data Analytics Pipeline in Rill
sidebar_label: Build
sidebar_position: 0
---

Rill provides a comprehensive platform for building end-to-end data analytics solutions. Once you've connected to your source data or connected to your own OLAP database, you can start building with Rill. Don't forget to check out the [reference documentation!](/reference/project-files)

### What is a Rill project?
- [**Getting Started with your Rill Project**](/developers/build/getting-started) - Understand Rill project structure and configuration

### Connect to your Data
- [**Connectors Overview**](/developers/build/connectors) - Learn about connecting Rill to data sources and OLAP engines
- [**OLAP Engines**](/developers/build/connectors/olap) - Configure DuckDB, ClickHouse, Druid, or Pinot for analytics
- [**Data Sources**](/developers/build/connectors/data-source) - Connect to S3, GCS, BigQuery, Snowflake, and more
- [**Configure Local Credentials**](/developers/build/connectors/credentials) - Set up credentials for local development
- [**Dev/Prod Connectors**](/developers/build/connectors/templating) - Separate your production and development data sources 
  
### Modeling and Last Mile ETL
- [**Models Overview**](/developers/build/models) - Overview of models available in Rill
- [**Getting Started with Models**](/developers/build/models/models-101) - What are models and important topics to remember as you create your own.
- [**Differences between SQL and YAML Models**](/developers/build/models/model-differences) - Compare SQL and YAML models and learn when to use 
- [**SQL Models**](/developers/build/models/sql-models) - Transform and prepare your data
- [**Data Refresh**](/developers/build/models/data-refresh) - Schedule your data refresh  
- [**YAML Models**](/developers/build/models/model-differences#yaml-models) - Incremental ingestion, data partitions, and more
- [**Additional Model Templating**](/developers/build/models/templating) - Separate your production and development settings 
- [**Performance Optimization**](/developers/build/models/performance) - Optimize model performance and data refresh strategies

### Define Measures and Dimensions
- [**What are Metrics Views?**](/developers/build/metrics-view/what-are-metrics-views) - Learn about Metrics Views!
- [**Underlying Model/Table**](/developers/build/metrics-view/underlying-model) - Powering your metrics view with a `model` or `table`
- [**Time Series**](/developers/build/metrics-view/time-series) - The backbone of Rill Explore Dashboards, time
- [**Dimensions**](/developers/build/metrics-view/dimensions) - Expand your dimensions capabilities further in your metrics view
- [**Measures**](/developers/build/metrics-view/measures) - Aggregate your measures further in your metrics view
- [**Define Data Access**](/developers/build/metrics-view/security) - Define access and row access policies

### Build Visualizations
- [**Dashboards Types**](/developers/build/dashboards/dashboards-101) - What are the visualization options in Rill, and how are they different?
- [**Slice-and-Dice Dashboards**](/developers/build/dashboards/explore) - Explore and find insights in your data
- [**Traditional Visualizations**](/developers/build/dashboards/canvas) - Visualize your data with various chart types
- [**Canvas Components**](/developers/build/dashboards/canvas-widgets) - See all of our supported components!
- [**Define Dashboard Access**](/developers/build/dashboards/customization#define-dashboard-access) - Set a SQL boolean query that defines access to dashboard

### Build Integrations with External Applications
- [**Custom API**](/developers/build/custom-apis) - Create custom API endpoints to retrieve data outside of Rill

### Project Settings
- [**Project Configuration**](/developers/build/project-configuration) - Configure your Rill Project and set default behavior
- [**Structure your Project**](/developers/build/structure) - Structure folder architecture in Rill
- [**Use your favorite IDE**](/developers/build/ide) - Utilize your favorite IDE to build Rill projects
- [**Debugging Rill Developer**](/developers/build/debugging/trace-viewer) - Troubleshoot dashboard access, trace your reconciled resources, and understand project logs  