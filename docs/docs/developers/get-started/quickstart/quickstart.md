---
title: Getting Started with Rill Developer 
sidebar_label: Quickstart
sidebar_position: 10
---
import Video from '@site/src/components/Video';

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

:::tip Rill's Default Engine
This guide assumes you'll be using Rill's default embedded engine, DuckDB. If you're looking to set up Rill with ClickHouse, check out our [ClickHouse Guide](/guides/rill-clickhouse)!
:::

## What is Rill Developer?

Rill Developer is your comprehensive toolkit for transforming raw data into actionable insights. It's designed to streamline the entire analytics workflow—from data ingestion to dashboard creation—all within a single, intuitive interface.

**Key capabilities:**
- **Connect to any data source** - Support for databases, cloud storage, APIs, and more
- **Transform data with ease** - Built-in ETL capabilities for data cleaning and preparation
- **Build metrics layer** - Define business KPIs and calculations
- **Create interactive dashboards** - AI-powered insights and visualizations
- **Deploy to production** - Seamlessly publish and share with your team

## Quick Start: From Zero to Dashboard in Minutes

### Step 1: Install and Launch Rill

Get started with just two commands:

```bash
curl https://rill.sh | sh
rill start my-rill-project
```

This will:
- Download and install Rill Developer
- Create a new project directory
- Launch the Rill interface in your browser

<img src = '/img/tutorials/rill-basics/new-rill-project.png' class='rounded-gif' />

<br/>
:::note Rill Developer vs Rill Cloud
Rill Developer is your local development environment where you build and test your analytics projects. Once ready, you can deploy to Rill Cloud for team collaboration and production use. For more details, see our [Developer vs Cloud comparison](/get-started/concepts/cloud-vs-developer).
:::

### Step 2: Connect Your Data

Rill supports a wide range of data sources through our [connector library](/build/connectors). For this tutorial, we'll use a sample dataset hosted on Google Cloud Storage. Select Add Data in the dropdown, GCS, and use the following dataset:
```
gs://rilldata-public/auction_data.parquet
```

**What happens when you connect data:**
- Rill automatically detects your data schema
- Provides a preview of the first 150 rows
- Analyzes data types, value ranges, and patterns
- Creates a foundation for your analytics

<div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/SWiEy0JgOgU?si=9rAdpgf5zqjZZ56K"
    frameBorder="0"
    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
    allowFullScreen
    style={{
      position: "absolute",
      top: 0,
      left: 0,
      width: "100%",
      height: "100%",
      borderRadius: "10px", 
    }}
  ></iframe>
</div>

### Step 3: Create Your First Dashboard

Rill's AI-powered dashboard creation makes it easy to go from data to insights in seconds. The system automatically generates relevant visualizations and suggests key metrics based on your data.

Select the model's side menu and select "Generate dashboard with AI". Watch as Rill configures your metrics view and creates an Explore dashboard. Once finished, you can navigate the different features of our [Explore dashboard](/explore).

<div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/_e_IbPKbA6k?si=Jm4jUT5aszL0oNJl"
    frameBorder="0"
    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
    allowFullScreen
    style={{
      position: "absolute",
      top: 0,
      left: 0,
      width: "100%",
      height: "100%",
      borderRadius: "10px", 
    }}
  ></iframe>
</div>

### Step 4: Explore and Analyze

Once your dashboard is created, you can:

- **Navigate different views** - Pivot tables, time-series analysis, and multi-measure charts
- **Drill down into data** - Click on any element to explore deeper insights
- **Customize visualizations** - Adjust charts, filters, and layouts
- **Export insights** - Share findings with your team



## What else can Rill do?

This quick start covered the basics, but Rill offers much more:

- **[Data Modeling](/build/models)** - Transform and prepare your data with SQL
- **[Metrics Layer](/build/metrics-view)** - Define business KPIs and calculations
- **[Deployment](/deploy/deploy-dashboard)** - Share your dashboards with your team

Ready to build something more complex? Check out our [comprehensive tutorial](/guides/rill-basics/launch) for a complete walkthrough of Rill's advanced features.