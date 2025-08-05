---
title: Building Your Data Analytics Pipeline in Rill 
sidebar_label: Build
sidebar_position: 00
---

import TileIcon from '@site/src/components/TileIcon';


# Building Your Data Analytics Pipeline in Rill

Rill provides a comprehensive platform for building end-to-end data analytics solutions. Once you've connected to your source data, or connected to your own OLAP database, you can start building with Rill.

### Modeling and Last Mile ETL
- [SQL Models](/build/models) - Transform and prepare your data
- [Data Refresh](/build/models/data-refresh) - Schedule your data refresh  
- [Additional Model Templating](/build/models/environments) - Separate your prod and dev settings 
- [Advanced Modeling](/build/advanced-models) - Incremental ingestion, data partitions and more

### Define Measures and Dimensions
- [Define Measures and Dimensions](/build/metrics-view) - Define measures and dimensions 
- [Advanced Measures](/build/metrics-view/advanced-expressions) - Aggregate further in your metrics view
<!-- - [Define Data Access](#asd) - Define access and row access policies -->

### Build Visualizations

- [Slice-and-Dice Dashboard](/build/dashboards) - Explore and find insights in your data
- [Traditional Visualizations](/build/canvas) - Visualize your data with various chart types
  
### Project Settings
- [rill.yaml](/build/rill-project-file) - Configure your Rill project defaults
- [Structure your Project](/build/structure) - Structure folder architecture in Rill
- [Debugging Rill Developer](/build/debugging/trace-viewer) - Trace your reconciled resources to troubleshoot issues
  

<!-- 
## Modeling
<div className="tile-icon-grid">
    <TileIcon
    header="Prepare Your Data"
    content="Transform and prepare your data with Rill's powerful ETL capabilities."
    link="/build/models/"
    />
    <TileIcon
    header="Advanced Model Features"
    content="Need incremental refreshes or using ClickHouse Modeling? Click here!"
    link="/build/advanced-models"
    />
    </div>
## Define Measures and Dimensions
    <div className="tile-icon-grid">
    <TileIcon
    header="Create a Metrics Layer"
    content="Build a metrics layer to define key business metrics and KPIs."
    link="/build/metrics-view"
    />
    <TileIcon
    header="Advanced Measures"
    content="Build a metrics layer to define key business metrics and KPIs."
    link="/build/metrics-view"
    />
   </div>
## Build Visualizations
    <div className="tile-icon-grid">
    <TileIcon
    header="Explore Your Data"
    content="Use Rill's interactive data exploration tools to discover insights."
    link="/build/dashboards"
    />
    <TileIcon
    header="Canvas Your Data"
    content="Create your traditional Dashboard, referencing mutliple metric views."
    link="/build/canvas"
    />
        </div>
## Project Settings
    <div className="tile-icon-grid">
    <TileIcon
    header="Rill Project Defaults"
    content="Need to set project defaults on access, env variables, and time settings?"
    link="/build/rill-project-file"
    />
    <TileIcon
    header="Structure your Project"
    content="Need to set project defaults on access, env variables, and time settings?"
    link="/build/rill-project-file"
    />
</div>
 -->
