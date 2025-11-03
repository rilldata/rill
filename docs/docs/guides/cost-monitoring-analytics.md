---
title: "Cost Monitoring Analytics Demo"
sidebar_label: "Cost Monitoring Analytics Demo"
sidebar_position: 20
hide_table_of_contents: false
tags:
  - Tutorial
  - Quickstart
  - Example Project
  - Cost Management
  - Margin Analysis
  - Business Intelligence
---

# Cost Monitoring Analytics Demo

This guide walks you through the Cost Monitoring Analytics demo project, which showcases Rill's capabilities for analyzing cost and usage data. You'll learn how to clone the project, understand its structure, and explore the dashboard.

## Overview

This dataset is modeled after a similar dashboard we use internally at Rill to both identify opportunities to improve our cloud infrastructure operations and to manage customer implementations. Typical users include engineering, customer success, and finance. In this example, we've combined cloud services, other hosting costs, and revenue metrics.

The Cost Monitoring Analytics demo analyzes operational costs and revenue data, providing insights into:
- **Margin trends** – Daily, weekly, and monthly profitability patterns
- **Customer profitability** – Which customers are driving the highest margins
- **Cost efficiency** – Understanding operational cost effectiveness across services
- **Revenue optimization** – Monitoring revenue performance against operational costs
- **Business intelligence** – Identifying opportunities for improved profitability

## Step 1: Clone the Project

### Clone from GitHub

```bash
# Clone the Cost Monitoring Analytics project
git clone https://github.com/rilldata/rill-examples.git
cd rill-examples/rill-cost-monitoring
```


## Step 2: Project Structure

The project is organized as follows:

```
rill-cost-monitoring/
├── rill.yaml                           # Project configuration
├── sources/                            # Data source definitions
│   └── metrics_margin_monitoring.yaml  # Margin source dataset
├── models/                             # SQL transformations
│   └── metrics_margin_model.sql        # Cost trend analysis
├── metrics/                            # Defined measures and dimensions
│   └── metrics_margin_metrics.yaml     # Cost monitoring metrics
├── dashboards/                         # Dashboard configurations
│   ├── margin_scorecard.yaml           # Canvas dashboard
│   └── metrics_margin_explore.yaml     # Explore Dashboard
└── README.md                           # Project documentation
```

## Step 3: Data Sources

The source connects to our static public dataset containing operational cost and revenue data. You can modify this [source](/build/connectors) to point to your own data export.

```yaml
# Visit https://docs.rilldata.com/ to learn more about Rill code artifacts.
type: source
connector: "https"
uri: "https://storage.googleapis.com/rilldata-public/metrics_margin_monitoring.parquet"
```

**What this data contains:**
- **Cost data** – Operational expenses broken down by component, environment, and pipeline
- **Revenue data** – Customer revenue associated with different billing plans and SKUs
- **Customer information** – Company names and associated billing plans
- **Operational metadata** – Location, environment, application, and pipeline details
- **Time series data** – Daily granular data for trend analysis

**What this source does:**
- Connects to our public GCS bucket (if modified, will need to be verified via [credentials](/build/connectors/credentials))
- Ingests the data into Rill's OLAP Engine (DuckDB)
- Provides the foundation for margin analysis and business intelligence



## Step 4: Data Models
:::tip Modeling
In our example, we've already processed the data, but if you need to do some last-mile ETL in Rill, this is possible via a [model](/build/models).
:::


## Step 5: Creating your Metrics View

Metrics in Rill define the measures and dimensions that power your margin monitoring dashboards:

```yaml
# Metrics view YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics-views
# This file was generated using AI.

version: 1
type: metrics_view

display_name: Metrics Margin Model KPIs
model: metrics_margin_model
timeseries: __time
smallest_time_grain: "day"

dimensions:
  - name: customer
    display_name: Customer
    column: company
    description: "The name of the customer"
  - name: plan_name
    display_name: Plan Name
    column: plan_name
    description: "The name of the billing plan"
  - name: location
    display_name: "Cost by Region"
    column: "location"
    description: "The region incurring costs"
  - name: component
    display_name: Cost by Component
    column: component
    description: "The component generating costs"
  - name: app_name
    display_name: "Cost by App Name"
    column: "app_name"
    description: "The app generating costs"
  - name: sku_description
    display_name: "Cost by SKU"
    column: "sku_description"
    description: "The sku description for costs"
  - name: pipeline
    display_name: "Cost by Data Pipeline"
    column: "pipeline"
    description: "The pipeline incurring costs"
  - name: environment
    display_name: "Cost by Environment"
    column: "environment"
    description: "The environment incurring costs"
    
measures:
  - display_name: "Total Cost"
    expression: "SUM(cost)"
    name: total_cost
    description: "The sum of cost"
    format_preset: currency_usd
  - display_name: "Total Revenue"
    expression: SUM(revenue)
    name: total_revenue
    description: The sum of revenue
    format_preset: currency_usd
  - display_name: "Net Revenue"
    expression: "SUM(revenue) - SUM(cost)"
    name: net_revenue
    description: "The sum of revenue minus the sum of cost"
    format_preset: currency_usd
  - display_name: "Gross Margin %"
    expression: "(SUM(revenue) - SUM(cost))/SUM(revenue)"
    name: gross_margin_percent
    description: "Net revenue divided by sum of revenue"
    format_preset: percentage
  - display_name: "Unique Customers"
    expression: "COUNT(DISTINCT company)"
    name: unique_customers
    description: "The count of unique companies"
    format_preset: humanize
```

**What this metrics file does:**

- **Measures** define the key business calculations you want to perform:
  - `total_cost` – Sum of all operational costs
  - `total_revenue` – Sum of all revenue generated
  - `net_revenue` – Revenue minus costs (profit calculation)
  - `gross_margin_percent` – Profitability percentage calculation
  - `unique_customers` – Count of distinct customers

- **Dimensions** define how you can slice and dice the margin data:
  - `customer` – Analysis by individual customer/company
  - `plan_name` – Breakdown by billing plan types
  - `location` – Regional cost and revenue analysis
  - `component` – Analysis by system components
  - `app_name` – Application-specific margin analysis
  - `sku_description` – Product SKU-level insights
  - `pipeline` – Data pipeline cost attribution
  - `environment` – Environment-based analysis (dev/staging/prod)

## Step 6: Dashboard Exploration

#### **Features**
- **Margin Overview** – High-level profitability trends and key metrics
- **Customer Analysis** – Detailed margin analysis by customer
- **Product Breakdown** – Profitability by plan types and SKUs
- **Regional Analysis** – Cost and revenue distribution across locations
- **Component Analysis** – Margin breakdown by system components
- **Environment Tracking** – Cost allocation across environments

#### **Selectors**
- **Date range selector** – Analyze specific time periods
- **Time Comparison Toggle** – Compare with previous periods
- **Customer filter** – Focus on specific customers
- **Plan filter** – Analyze specific billing plans

#### **Filters**
- **Region filter** – Analyze specific regions
- **Component filter** – Focus on specific system components
- **Margin threshold filter** – Focus on high/low margin segments

### Let's answer some key margin questions!

**Which customers are driving the highest margins?**
<img src='/img/tutorials/quickstart/customer-margins-1.png' class='rounded-gif'/>
<br />


**Which billing plans have the best profitability?**
<img src='/img/tutorials/quickstart/customer-margins-2.png' class='rounded-gif'/>
<br />

These are just some of the insights that you can find within your explore dashboard, but you'll find more hidden gems in your data as you continue to use Rill. Please let us know if you have any other questions!
