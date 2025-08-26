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
- [**Performance**](/build/models/performance) - Optimize model performance and data refresh strategies
- [**Additional Model Templating**](/build/models/templating) - Separate your production and development settings 


### Define Measures and Dimensions

- [**Define Measures and Dimensions**](/build/metrics-view) - Define measures and dimensions 
- [**Advanced Measures**](/build/metrics-view/advanced-expressions) - Aggregate further in your metrics view
- [**Customize Metrics View Settings**](/build/metrics-view/customize) - Set the smallest selectable grain, day_of_week, month_of_year, etc.
  
<!-- - [Define Data Access](#security) - Define access and row access policies -->

### Build Visualizations

- [**Slice-and-Dice Dashboards**](/build/dashboards) - Explore and find insights in your data
- [**Traditional Visualizations**](/build/canvas) - Visualize your data with various chart types
<!-- - [**Define Dashboard Access**](/build/dashboards#define-dashboard-access) - Set a SQL boolean query that defines access to dashboard-->

### Build Integrations with External Applications
- [**Custom API**](/build/custom-apis) - Create custom API endpoints to retreive data outside of Rill


### Project Settings

- [**rill.yaml**](/build/rill-project-file) - Configure your Rill project defaults
- [**Structure your Project**](/build/structure) - Structure folder architecture in Rill
- [**Debugging Rill Developer**](/build/debugging/trace-viewer) - Troubleshoot dashboard access, trace your reconciled resources, and understand project logs  