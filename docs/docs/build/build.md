---
title: Building Your Data Analytics Pipeline in Rill 
sidebar_label: Build
sidebar_position: 00
---

import TileIcon from '@site/src/components/TileIcon';

<div className="tile-icon-grid">
    <h2>Modeling</h2><br/>
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
    <div className="tile-icon-grid">
    <h2>Define Measures and Dimensions</h2><br/>
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
    <div className="tile-icon-grid">
    <h2>Build Visualizations</h2><br/>
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
    <div className="tile-icon-grid">
    <h2>Project Settings</h2><br/>
    <TileIcon
    header="Rill Project Defaults"
    content="Need to set project defaults on access, env variables, and time settings?"
    link="/build/rill-project-file"
    />

</div>


Rill provides a comprehensive platform for building end-to-end data analytics solutions. Once you've connected to your source data, or connected to your own OLAP database, you can start building with Rill.

## Prepare and Transform Your Data
<img src = '/img/build/advanced-models/advanced-model.png' class='rounded-gif' />
<br />

If your data is already prepared and clean, you can skip straight to [creating a metrics view](/build/metrics-view). However, we understand that real-world data often requires some last-minute transformations and joins. You might need to join lookup tables to enrich your data, apply business logic, or perform data cleaning operations.

This is where Rill's powerful modeling capabilities come into play. You can create models using either [SQL](/build/models) or [YAML](/build/advanced-models) configurations, depending on your business needs and technical preferences. These models serve as the foundation for your analytics, ensuring data consistency and business logic implementation.

:::tip Need incremental refreshes?

If you need to incrementally refresh, partition, or are using your own OLAP database, take a look at our [advanced model docs](/build/advanced-models) to get started.
:::

## Define Your Metrics Layer
<img src = '/img/build/advanced-models/advanced-model.png' class='rounded-gif' />
<br />

Once your data is properly prepared, you can create a metrics view either manually or with the help of Rill's AI assistant. This is the crucial layer in Rill that allows you, as the designer, to decide what metrics are important and what they mean to your business.

The metrics layer transforms your raw data into meaningful business KPIs and measures. You can define aggregations, calculations, and business rules that turn data into actionable insights. This layer serves as the semantic model that business users will interact with.

## Visualize and Explore Your Data

<img src = '/img/build/advanced-models/advanced-model.png' class='rounded-gif' />
<br />
Finally, once you've created your metrics, you can visualize your data in multiple ways:

- **Slice-and-Dice Explore Dashboards**: Interactive exploration dashboards that allow users to drill down, filter, and analyze data dynamically
- **Canvas Dashboards**: Traditional dashboard layouts using Vega-Lite for custom visualizations

Using Vega-Lite, Rill allows users to build comprehensive dashboards that reference multiple metrics views, providing a holistic view of your business. You can directly navigate to and explore data from the canvas, creating a seamless experience from high-level dashboards to detailed analysis.

