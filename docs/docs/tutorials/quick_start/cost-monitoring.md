---
title: "Walkthough the Demo Project"
sidebar_label: "Cost Monitoring Analytics Demo"
hide_table_of_contents: false
tags:
    - Tutorial
    - Quickstart
---

# Cost Monitoring Demo

Learn how to monitor and analyze cloud costs with Rill using the Cost Monitoring demo project. This guide shows you how to track spending across different services, projects, and time periods to optimize your cloud infrastructure costs.

## Step 1: Clone the Project

```bash
# Clone the Cost Monitoring demo
git clone https://github.com/rilldata/rill-examples.git
cd rill-examples/rill-cost-monitoring

# Start Rill Developer
rill start
```

Visit [http://localhost:9009](http://localhost:9009) to explore your cost monitoring dashboard.

## Step 2: Project Structure

The project is organized as follows:

```
rill-cost-monitoring/
├── rill.yaml                           # Project configuration
├── sources/                            # Data source definitions
│   ├── aws_costs.yaml                  # AWS cost data
│   ├── gcp_costs.yaml                  # GCP cost data
│   └── azure_costs.yaml                # Azure cost data
├── models/                             # SQL transformations
│   └── unified_costs_model.sql         # Combined cost data model
├── metrics/                            # Defined measures and dimensions
│   └── cost_metrics.yaml               # Cost analysis metrics
├── dashboards/                         # Dashboard configurations
│   ├── cost_overview.yaml              # Main cost dashboard
│   ├── cost_optimization.yaml          # Optimization insights
│   └── budget_tracking.yaml            # Budget vs actual
├── alerts/                             # Cost alert definitions
│   └── cost_spike_alert.yaml           # Cost spike detection
└── README.md                           # Project documentation
```

## Step 3: Data Sources

### AWS Costs Source

The AWS costs source tracks spending across AWS services:

```yaml
# sources/aws_costs.yaml
type: source
connector: "s3"
uri: "s3://your-cost-bucket/aws-costs/*.parquet"
```

**What this does:**
- Connects to your S3 bucket containing AWS cost data
- Fetches daily cost breakdowns by service, region, and account
- Tracks usage metrics and billing information
- Data includes service names, regions, usage types, and costs

### GCP Costs Source

The GCP costs source monitors Google Cloud spending:

```yaml
# sources/gcp_costs.yaml
type: source
connector: "gcs"
uri: "gs://your-cost-bucket/gcp-costs/*.parquet"
```

**What this does:**
- Connects to your GCS bucket with GCP billing data
- Tracks costs by project, service, and SKU
- Monitors resource usage and pricing
- Enables cross-project cost analysis

### Azure Costs Source

The Azure costs source tracks Microsoft Azure spending:

```yaml
# sources/azure_costs.yaml
type: source
connector: "azure"
uri: "https://your-storage-account.blob.core.windows.net/azure-costs/*.parquet"
```

**What this does:**
- Connects to Azure Blob Storage with cost data
- Tracks spending by subscription, resource group, and service
- Monitors usage patterns and billing cycles
- Enables Azure-specific cost optimization

## Step 4: Data Models

### Unified Cost Model

This model combines cost data from all cloud providers into a unified view:

```sql
-- Model SQL
-- Reference documentation: https://docs.rilldata.com/reference/project-files/models
-- @materialize: true

SELECT
    billing_date,
    cloud_provider,
    service_name,
    region,
    project_id,
    resource_group,
    usage_type,
    usage_quantity,
    unit_cost,
    total_cost,
    currency,
    account_id,
    subscription_id,
    tags,
    CASE 
        WHEN cloud_provider = 'AWS' THEN 'Amazon Web Services'
        WHEN cloud_provider = 'GCP' THEN 'Google Cloud Platform'
        WHEN cloud_provider = 'Azure' THEN 'Microsoft Azure'
        ELSE cloud_provider
    END AS provider_full_name,
    EXTRACT(YEAR FROM billing_date) AS billing_year,
    EXTRACT(MONTH FROM billing_date) AS billing_month,
    EXTRACT(DAY FROM billing_date) AS billing_day,
    DATE_TRUNC('month', billing_date) AS billing_month_start,
    CASE 
        WHEN total_cost > 1000 THEN 'High Cost'
        WHEN total_cost > 100 THEN 'Medium Cost'
        ELSE 'Low Cost'
    END AS cost_category
FROM (
    SELECT 
        billing_date,
        'AWS' AS cloud_provider,
        service_name,
        region,
        project_id,
        resource_group,
        usage_type,
        usage_quantity,
        unit_cost,
        total_cost,
        currency,
        account_id,
        subscription_id,
        tags
    FROM aws_costs_source
    
    UNION ALL
    
    SELECT 
        billing_date,
        'GCP' AS cloud_provider,
        service_name,
        region,
        project_id,
        resource_group,
        usage_type,
        usage_quantity,
        unit_cost,
        total_cost,
        currency,
        account_id,
        subscription_id,
        tags
    FROM gcp_costs_source
    
    UNION ALL
    
    SELECT 
        billing_date,
        'Azure' AS cloud_provider,
        service_name,
        region,
        project_id,
        resource_group,
        usage_type,
        usage_quantity,
        unit_cost,
        total_cost,
        currency,
        account_id,
        subscription_id,
        tags
    FROM azure_costs_source
)
```

## Step 5: Creating your Metrics View

Metrics in Rill define the measures and dimensions that power your cost monitoring dashboards:

```yaml
# metrics/cost_metrics.yaml
title: "Cloud Cost Monitoring Metrics"
description: "Key metrics for analyzing and optimizing cloud costs"

measures:
  - name: "total_cost"
    description: "Total cost across all cloud providers"
    expression: "SUM(total_cost)"
    
  - name: "daily_cost"
    description: "Average daily cost"
    expression: "AVG(total_cost)"
    
  - name: "monthly_cost"
    description: "Total monthly cost"
    expression: "SUM(total_cost)"
    
  - name: "cost_trend"
    description: "Cost trend over time"
    expression: "total_cost"
    
  - name: "cost_per_service"
    description: "Cost per service"
    expression: "SUM(total_cost)"
    
  - name: "cost_per_project"
    description: "Cost per project"
    expression: "SUM(total_cost)"
    
  - name: "cost_per_region"
    description: "Cost per region"
    expression: "SUM(total_cost)"
    
  - name: "usage_quantity"
    description: "Total usage quantity"
    expression: "SUM(usage_quantity)"
    
  - name: "avg_unit_cost"
    description: "Average unit cost"
    expression: "AVG(unit_cost)"
    
  - name: "cost_efficiency"
    description: "Cost efficiency ratio"
    expression: "total_cost / usage_quantity"

dimensions:
  - name: "billing_date"
    description: "Billing date"
    expression: "billing_date"
    
  - name: "cloud_provider"
    description: "Cloud provider (AWS, GCP, Azure)"
    expression: "cloud_provider"
    
  - name: "service_name"
    description: "Cloud service name"
    expression: "service_name"
    
  - name: "region"
    description: "Cloud region"
    expression: "region"
    
  - name: "project_id"
    description: "Project or account ID"
    expression: "project_id"
    
  - name: "resource_group"
    description: "Resource group (Azure) or similar"
    expression: "resource_group"
    
  - name: "usage_type"
    description: "Type of usage"
    expression: "usage_type"
    
  - name: "cost_category"
    description: "Cost category (High, Medium, Low)"
    expression: "cost_category"
    
  - name: "billing_month_start"
    description: "Start of billing month"
    expression: "billing_month_start"

time_grain: "day"
```

**What this metrics file does:**

- **Measures** define the key cost monitoring indicators:
  - `total_cost` - Overall spending across all providers
  - `daily_cost` - Average daily spending patterns
  - `cost_per_service` - Service-specific cost analysis
  - `cost_efficiency` - Cost per unit of usage

- **Dimensions** enable analysis across different segments:
  - `cloud_provider` - Multi-cloud cost comparison
  - `service_name` - Service-specific optimization
  - `project_id` - Project-level cost tracking
  - `region` - Geographic cost analysis

### Creating Advanced Cost Metrics

You can add more sophisticated cost analytics:

```yaml
# Additional measures for advanced cost analysis
measures:
  - name: "cost_variance"
    description: "Cost variance from previous period"
    expression: "total_cost - LAG(total_cost, 1) OVER (ORDER BY billing_date)"
    
  - name: "cost_growth_rate"
    description: "Month-over-month cost growth"
    expression: "(total_cost - LAG(total_cost, 30) OVER (ORDER BY billing_date)) / LAG(total_cost, 30) OVER (ORDER BY billing_date)"
    
  - name: "budget_utilization"
    description: "Percentage of budget used"
    expression: "total_cost / 10000"  # Assuming $10k monthly budget
    
  - name: "cost_per_user"
    description: "Cost per active user"
    expression: "total_cost / 100"  # Assuming 100 active users
```

## Step 6: Dashboard Exploration

### Main Cost Dashboard

The Cost Monitoring dashboard provides comprehensive spending insights:

#### **Cost Overview**
- **Daily cost trends** - Line chart showing spending over time
- **Cost by cloud provider** - Pie chart of spending distribution
- **Monthly cost comparison** - Bar chart comparing months

#### **Service Analysis**
- **Top spending services** - Services consuming the most budget
- **Service cost trends** - How service costs change over time
- **Cost per service** - Detailed service breakdown

#### **Project Insights**
- **Project cost allocation** - How costs are distributed across projects
- **Project cost trends** - Project spending patterns
- **Budget vs actual** - Project budget tracking

#### **Optimization Opportunities**
- **Cost anomalies** - Unusual spending patterns
- **Regional cost differences** - Geographic cost optimization
- **Usage efficiency** - Cost per unit of usage

### Interactive Features

The dashboard includes several interactive elements:

#### **Filters**
- **Date range selector** - Analyze specific time periods
- **Cloud provider filter** - Focus on specific providers
- **Service filter** - Analyze specific services
- **Project filter** - Project-specific analysis

#### **Drill-down Capabilities**
- Click on any chart to see detailed breakdowns
- Navigate from provider overview to specific services
- Explore individual project costs

