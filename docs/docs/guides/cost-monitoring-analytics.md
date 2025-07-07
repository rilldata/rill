---
title: "Cost Monitoring Analytics Demo"
sidebar_label: "Cost Monitoring Analytics Demo"
sidebar_position: 6
hide_table_of_contents: false
tags:
  - Tutorial
  - Quickstart
  - Example Project
  - AWS
  - Cost Management
---

# Cost Monitoring Analytics Demo

This guide walks you through the Cost Monitoring Analytics demo project, which showcases Rill's capabilities for analyzing AWS cost and usage data. You'll learn how to clone the project, understand its structure, and explore the cost analytics dashboard.

## Overview

The Cost Monitoring Analytics demo analyzes AWS cost and usage data, providing insights into:
- **Cost trends** - Daily, weekly, and monthly spending patterns
- **Service breakdown** - Which AWS services are driving costs
- **Resource utilization** - Understanding cost efficiency across resources
- **Budget tracking** - Monitoring spending against budgets and forecasts
- **Cost optimization** - Identifying opportunities for savings

## Step 1: Clone the Project

### Clone from GitHub

```bash
# Clone the Cost Monitoring Analytics project
git clone https://github.com/rilldata/rill-examples.git
cd rill-examples/rill-cost-monitoring
```

### Alternative: Use Rill CLI

```bash
# Clone directly using Rill CLI
rill project clone rill-cost-monitoring
cd rill-cost-monitoring
```

## Step 2: Project Structure

The project is organized as follows:

```
rill-cost-monitoring/
├── rill.yaml                           # Project configuration
├── sources/                            # Data source definitions
│   └── metrics_margin_monitoring.yaml  # AWS billing details
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

The source here is to our static public dataset

```yaml
# Visit https://docs.rilldata.com/ to learn more about Rill code artifacts.
type: source
connector: "https"
uri: "https://storage.googleapis.com/rilldata-public/metrics_margin_monitoring.parquet"
```

**What this does:**
- Connects to your AWS Cost and Usage Reports stored in S3
- Fetches detailed cost and usage data across all AWS services
- Includes resource-level cost attribution and tagging information
- Data includes service costs, usage types, and billing dimensions

### AWS Billing Data Source

The billing data source provides additional billing context:

```yaml
# sources/aws_billing_data.yaml
type: source
connector: "s3"
uri: "s3://your-billing-bucket/billing-reports/*/billing-*.parquet"
```

**What this does:**
- Provides billing period information and invoice details
- Includes account-level billing summaries
- Tracks payment methods and billing adjustments
- Enables month-over-month billing comparisons

## Step 4: Data Models

### Cost by Service Model

This model aggregates costs by AWS service and provides service-level insights:

```sql
-- models/cost_by_service.sql
-- Reference documentation: https://docs.rilldata.com/reference/project-files/models
-- @materialize: true

SELECT
    usage_date AS date,
    product_product_name AS service_name,
    CASE 
        WHEN product_product_name LIKE '%EC2%' THEN 'Compute'
        WHEN product_product_name LIKE '%S3%' THEN 'Storage'
        WHEN product_product_name LIKE '%RDS%' THEN 'Database'
        WHEN product_product_name LIKE '%Lambda%' THEN 'Serverless'
        ELSE 'Other'
    END AS service_category,
    line_item_usage_account_id AS account_id,
    product_region AS region,
    line_item_usage_type AS usage_type,
    line_item_operation AS operation,
    SUM(line_item_unblended_cost) AS unblended_cost,
    SUM(line_item_blended_cost) AS blended_cost,
    SUM(reservation_effective_cost) AS effective_cost,
    SUM(line_item_usage_amount) AS usage_amount,
    resource_tags_user_name AS resource_owner,
    resource_tags_user_environment AS environment,
    resource_tags_user_project AS project
FROM aws_cost_data
WHERE line_item_line_item_type != 'Tax'
GROUP BY 
    usage_date,
    product_product_name,
    service_category,
    line_item_usage_account_id,
    product_region,
    line_item_usage_type,
    line_item_operation,
    resource_tags_user_name,
    resource_tags_user_environment,
    resource_tags_user_project
```

### Daily Costs Model

This model provides daily cost aggregations for trend analysis:

```sql
-- models/daily_costs.sql
-- @materialize: true

SELECT
    usage_date AS date,
    line_item_usage_account_id AS account_id,
    SUM(line_item_unblended_cost) AS daily_cost,
    SUM(line_item_usage_amount) AS daily_usage,
    COUNT(DISTINCT product_product_name) AS services_used,
    COUNT(DISTINCT product_region) AS regions_used,
    AVG(line_item_unblended_cost) AS avg_line_item_cost,
    -- Calculate month-to-date costs
    SUM(SUM(line_item_unblended_cost)) OVER (
        PARTITION BY DATE_TRUNC('month', usage_date), line_item_usage_account_id
        ORDER BY usage_date
        ROWS UNBOUNDED PRECEDING
    ) AS month_to_date_cost,
    -- Calculate running 7-day average
    AVG(SUM(line_item_unblended_cost)) OVER (
        PARTITION BY line_item_usage_account_id
        ORDER BY usage_date
        ROWS 6 PRECEDING
    ) AS seven_day_avg_cost
FROM aws_cost_data
WHERE line_item_line_item_type != 'Tax'
GROUP BY usage_date, line_item_usage_account_id
ORDER BY usage_date DESC
```

## Step 5: Creating your Metrics View

Metrics in Rill define the measures and dimensions that power your cost monitoring dashboards:

```yaml
# metrics/cost_metrics.yaml
title: "AWS Cost Monitoring Metrics"
description: "Key metrics for analyzing AWS costs and usage"

measures:
  - name: "total_cost"
    description: "Total unblended cost"
    expression: "SUM(unblended_cost)"
    format_preset: "currency_usd"
    
  - name: "effective_cost"
    description: "Total effective cost (including RI/SP benefits)"
    expression: "SUM(effective_cost)"
    format_preset: "currency_usd"
    
  - name: "average_daily_cost"
    description: "Average daily cost"
    expression: "AVG(daily_cost)"
    format_preset: "currency_usd"
    
  - name: "cost_per_service"
    description: "Average cost per service"
    expression: "total_cost / COUNT(DISTINCT service_name)"
    format_preset: "currency_usd"
    
  - name: "month_over_month_change"
    description: "Month-over-month cost change percentage"
    expression: "((SUM(CASE WHEN date >= DATE_TRUNC('month', CURRENT_DATE) THEN unblended_cost END) - SUM(CASE WHEN date >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month') AND date < DATE_TRUNC('month', CURRENT_DATE) THEN unblended_cost END)) / SUM(CASE WHEN date >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month') AND date < DATE_TRUNC('month', CURRENT_DATE) THEN unblended_cost END)) * 100"
    format_preset: "percentage_1"
    
  - name: "total_usage_amount"
    description: "Total usage amount"
    expression: "SUM(usage_amount)"
    
  - name: "cost_efficiency_ratio"
    description: "Ratio of effective cost to unblended cost"
    expression: "SUM(effective_cost) / SUM(unblended_cost)"
    format_preset: "percentage_2"
    
  - name: "unique_services"
    description: "Number of unique AWS services used"
    expression: "COUNT(DISTINCT service_name)"
    
  - name: "unique_regions"
    description: "Number of unique regions used"
    expression: "COUNT(DISTINCT region)"

dimensions:
  - name: "date"
    description: "Usage date"
    expression: "date"
    
  - name: "service_name"
    description: "AWS service name"
    expression: "service_name"
    
  - name: "service_category"
    description: "AWS service category"
    expression: "service_category"
    
  - name: "account_id"
    description: "AWS account ID"
    expression: "account_id"
    
  - name: "region"
    description: "AWS region"
    expression: "region"
    
  - name: "usage_type"
    description: "Usage type"
    expression: "usage_type"
    
  - name: "operation"
    description: "AWS operation"
    expression: "operation"
    
  - name: "environment"
    description: "Environment tag"
    expression: "environment"
    
  - name: "project"
    description: "Project tag"
    expression: "project"
    
  - name: "resource_owner"
    description: "Resource owner tag"
    expression: "resource_owner"

time_grain: "day"
```

**What this metrics file does:**

- **Measures** define the cost calculations you want to perform:
  - `total_cost` - Sum of all unblended costs
  - `effective_cost` - Cost including Reserved Instance and Savings Plan benefits
  - `average_daily_cost` - Daily cost averages for trend analysis
  - `month_over_month_change` - Growth rate calculations
  - `cost_efficiency_ratio` - Savings from RI/SP usage

- **Dimensions** define how you can slice and dice the cost data:
  - `date` - Time-based cost analysis
  - `service_name` - Per-service cost breakdown
  - `account_id` - Multi-account cost analysis
  - `region` - Regional cost distribution
  - `environment` - Cost by environment (dev/staging/prod)

- **Time grain** sets the default time aggregation to daily

### Creating Custom Cost Metrics

You can add more sophisticated cost metrics:

```yaml
# Additional measures for advanced cost analysis
measures:
  - name: "cost_per_region"
    description: "Average cost per region"
    expression: "total_cost / unique_regions"
    format_preset: "currency_usd"
    
  - name: "weekly_cost_trend"
    description: "Week-over-week cost change"
    expression: "((SUM(CASE WHEN date >= DATE_TRUNC('week', CURRENT_DATE) THEN unblended_cost END) - SUM(CASE WHEN date >= DATE_TRUNC('week', CURRENT_DATE - INTERVAL '1 week') AND date < DATE_TRUNC('week', CURRENT_DATE) THEN unblended_cost END)) / SUM(CASE WHEN date >= DATE_TRUNC('week', CURRENT_DATE - INTERVAL '1 week') AND date < DATE_TRUNC('week', CURRENT_DATE) THEN unblended_cost END)) * 100"
    format_preset: "percentage_1"
    
  - name: "top_service_cost_share"
    description: "Percentage of total cost from top service"
    expression: "MAX(SUM(unblended_cost)) / SUM(unblended_cost) * 100"
    format_preset: "percentage_1"
```

## Step 6: Dashboard Exploration

#### **Features**
- **Cost Overview** - High-level cost trends and key metrics
- **Service Breakdown** - Detailed analysis by AWS service
- **Regional Analysis** - Cost distribution across regions
- **Tag-based Analysis** - Cost allocation by projects, environments, and owners
- **Budget Tracking** - Monitor spending against budgets

#### **Selectors**
- **Date range selector** - Analyze specific time periods
- **Time Comparison Toggle** - Compare with previous periods
- **Account filter** - Focus on specific AWS accounts
- **Service filter** - Analyze specific AWS services

#### **Filters**
- **Environment filter** - Focus on dev, staging, or production
- **Region filter** - Analyze specific AWS regions
- **Cost threshold filter** - Focus on high-cost resources

### Let's answer some key cost questions!

**Which AWS services are driving the highest costs?**
<img src='/img/tutorials/cost-monitoring/service-breakdown.png' class='rounded-gif'/>
<br />

**How are costs trending over time? Are there any unusual spikes?**
<img src='/img/tutorials/cost-monitoring/cost-trends.png' class='rounded-gif'/>
<br />

**Which regions and accounts have the highest cost concentration?**
<img src='/img/tutorials/cost-monitoring/regional-analysis.png' class='rounded-gif'/>
<br />

**How effectively are we using Reserved Instances and Savings Plans?**
<img src='/img/tutorials/cost-monitoring/savings-analysis.png' class='rounded-gif'/>
<br />

## Cost Optimization Insights

The dashboard helps identify several cost optimization opportunities:

### 1. **Service Optimization**
- Identify services with high costs but low utilization
- Compare costs across similar services to find alternatives
- Track cost per unit of usage to identify inefficiencies

### 2. **Regional Cost Analysis**
- Identify regions with unexpectedly high costs
- Compare regional pricing for similar workloads
- Optimize data transfer costs between regions

### 3. **Reserved Instance and Savings Plan Optimization**
- Track RI/SP utilization and coverage
- Identify opportunities for additional commitments
- Monitor effective cost vs. on-demand costs

### 4. **Tag-based Cost Allocation**
- Identify untagged resources contributing to costs
- Allocate costs to projects and teams accurately
- Track cost accountability across the organization

## Getting Started with Your Own Data

To use this dashboard with your own AWS cost data:

1. **Set up AWS Cost and Usage Reports**:
   - Enable detailed billing reports in your AWS account
   - Configure reports to be delivered to an S3 bucket
   - Ensure reports include resource IDs and tags

2. **Update source configurations**:
   - Modify `sources/aws_cost_data.yaml` to point to your S3 bucket
   - Update any credential configurations as needed

3. **Customize metrics and dimensions**:
   - Add your organization's specific tags as dimensions
   - Create custom measures for your business metrics
   - Adjust time grains based on your reporting needs

4. **Set up alerts and monitoring**:
   - Create alerts for cost thresholds
   - Set up anomaly detection for unusual spending patterns
   - Configure scheduled reports for stakeholders

These are just some of the insights you can discover with your cost monitoring dashboard. The combination of detailed AWS cost data and Rill's powerful analytics capabilities provides a comprehensive view of your cloud spending patterns and optimization opportunities.

For more advanced cost optimization strategies, consider integrating additional data sources such as:
- AWS CloudWatch metrics for utilization data
- AWS Trusted Advisor recommendations
- Third-party cost optimization tools
- Business metrics to correlate costs with business outcomes

Please let us know if you have any questions about implementing cost monitoring analytics with Rill!
