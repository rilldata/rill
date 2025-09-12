---
title: Monitoring Dashboards in Rill
sidebar_label: "Monitoring Dashboard in Rill"
hide_table_of_contents: false
sidebar_position: 50

tags:
    - Quickstart
    - Tutorial
---


## Overview

This guide demonstrates how to build powerful monitoring dashboards in Rill using a modern observability stack with Prometheus, Thanos, and ClickHouse. You'll learn to visualize infrastructure metrics, application performance, and business KPIs in real-time.

## Architecture Overview

```
Applications → Prometheus → Thanos → ClickHouse → Rill
     ↓           ↓           ↓         ↓        ↓
  Metrics    Collection   Long-term  Analytics  Interactive
Generation   & Rules     Storage    Database   Dashboards
```

## Prerequisites

This guide does not cover the details of setting up Prometheus and Thanos, and instead assumes that you already have this infrastructure running with your metrics being written to a readable location such as S3 or GCS. 

In this guide, we are writing our metrics directly to ClickHouse, which is being used as a [live connector in Rill](/connect/olap/clickhouse).

## Part 1: Connect your data to Rill

Whether you are writing directly to ClickHouse or any other analytical database, or have your data being written to cloud-based storage, you'll need to connect this to Rill.

In our case, we're using ClickHouse, so a live connector will suffice. If you are ingesting the data into Rill's analytical database, we recommend creating an [incremental model](/build/models) to ingest your stream of metrics.

### 1.1 Rill Live Connector 

This is a sample connector YAML to connect to ClickHouse. For more detailed information, see our [ClickHouse Docs](/connect/olap/clickhouse).

```yaml
# Connector YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/connectors
  
type: connector

driver: clickhouse
host: "clickhouse"
port: 9000
username: "default"
password: "{{ .env.connector.clickhouse.password }}"
```

### 1.2 Check tables are visible in Rill

Once you've configured the ClickHouse connector, you should see your Prometheus metrics tables appear in the left panel of Rill Developer. Look for tables like:

- `prometheus_metrics` - Raw metrics data


<img src='/img/guides/monitoring/clickhouse-tables.png' class='rounded-gif' />
<br />


### 1.3 Prepare the Data

Since the labels column seems to be nested, we'll need to unnest that in a model file in Rill.
Create a model by selecting the "..." next to the ClickHosue table.

<details> 
<summary> Click to reveal sample YAML </summary>

```yaml
connector: clickhouse
refresh:
  every: 10m
  run_in_dev: true
  
sql: > 
    WITH parsed_labels AS (
        SELECT
            timestamp,
            metric_name,
            value,
            -- Extract specific labels using string functions
            extract(labels, '''__name__'': ''([^'']*)''') as metric_type,
            extract(labels, '''cluster'': ''([^'']*)''') as cluster,
            extract(labels, '''criticality'': ''([^'']*)''') as criticality,
            extract(labels, '''environment'': ''([^'']*)''') as environment,
            extract(labels, '''hostname'': ''([^'']*)''') as hostname,
            extract(labels, '''instance'': ''([^'']*)''') as instance,
            extract(labels, '''instance_type'': ''([^'']*)''') as instance_type,
            extract(labels, '''job'': ''([^'']*)''') as job,
            extract(labels, '''monitoring_tier'': ''([^'']*)''') as monitoring_tier,
            extract(labels, '''os'': ''([^'']*)''') as os,
            extract(labels, '''quantile'': ''([^'']*)''') as quantile,
            extract(labels, '''region'': ''([^'']*)''') as region,
            extract(labels, '''replica'': ''([^'']*)''') as replica,
            extract(labels, '''service_owner'': ''([^'']*)''') as service_owner,
            -- New columns for Rill metrics
            extract(labels, '''otel_scope_name'': ''([^'']*)''') as otel_scope_name,
            extract(labels, '''otel_scope_version'': ''([^'']*)''') as otel_scope_version,
            extract(labels, '''db_system'': ''([^'']*)''') as db_system,
            extract(labels, '''instance_id'': ''([^'']*)''') as instance_id,
            extract(labels, '''method'': ''([^'']*)''') as method,
            extract(labels, '''status'': ''([^'']*)''') as status,
            extract(labels, '''le'': ''([^'']*)''') as le
        FROM
            default.prometheus_metrics
        WHERE instance = 'rill:10010'
    )
    SELECT
        timestamp,
        metric_name,
        value,
        metric_type,
        cluster,
        criticality,
        environment,
        hostname,
        instance,
        instance_type,
        job,
        monitoring_tier,
        os,
        quantile,
        region,
        replica,
        service_owner,
        -- New columns for Rill metrics
        otel_scope_name,
        otel_scope_version,
        db_system,
        instance_id,
        method,
        status,
        le,
        
    FROM
        parsed_labels
    ORDER BY
        timestamp DESC

output:
  connector: clickhouse
```

</details>

## Part 2: Create the Metrics View With AI

Navigate to the metrics section and click "Generate metrics via AI" to automatically create a metrics view from your Prometheus data.

<img src='/img/guides/monitoring/ai-generate-metrics.png' class='rounded-gif' />
<br />

The AI will analyze your data structure and suggest relevant measures and dimensions such as:

**Measures:**
- `total metrics value` - Seems pretty generic but you'll see that in the explore, you're able to easily slice and dice.

**Dimensions:**
- `job` - Service name
- `instance` - Server instance
- `method` - HTTP method
- `status_code` - HTTP status
- ...

## Part 3: Creating Explore Dashboard in Rill

Click "Create Explore dashboard" to generate an interactive dashboard from your metrics view. This will create a comprehensive monitoring dashboard with:


<img src='/img/guides/monitoring/explore-dashboard.png' class='rounded-gif' />
<br />

### 3.1 Rill Project Structure

Current project structure should look like this:
```
monitoring-dashboard/
├── rill.yaml
├── models/
│   └── prometheus_expanded.yaml
├── metrics/
│   └── prometheus_expanded_metrics.yaml
└── dashboards/
    └── explore_dashboard.yaml
```

## Part 4: Explore the Dashboard and find Insights

### 4.1 Key Monitoring Questions

Your monitoring dashboard can answer critical questions like:

**Infrastructure Health:**
- Which services are consuming the most CPU and memory?
- Are there any memory leaks or resource spikes?
- How is disk usage trending across instances?

**Application Performance:**
- What's the average response time for each API endpoint?
- Which services have the highest error rates?
- Are there any performance regressions over time?

**Business Impact:**
- How do system performance issues correlate with user experience?
- Which incidents had the biggest business impact?
- What's the cost of downtime vs. infrastructure investment?

### 4.2 Dashboard Features

**Interactive Filtering:**
- Filter by time range to focus on specific incidents
- Drill down by service, instance, or environment
- Compare different time periods side-by-side

**Real-time Updates:**
- Live data refresh every 30 seconds
- Automatic scaling based on data volume
- Alert thresholds for critical metrics

**Export and Sharing:**
- Share dashboard links with team members
- Embed dashboards in other tools



## Conclusion

This guide demonstrates how Rill transforms raw infrastructure metrics into actionable insights through:

- **Real-time Performance**: Sub-second query responses on billions of metrics
- **Interactive Analytics**: Drill-down capabilities and dynamic filtering  
- **Scalable Architecture**: Handles enterprise-scale metric volumes
- **Business Context**: Correlates technical metrics with business outcomes

The combination of Prometheus (collection), Thanos (storage), ClickHouse (analytics), and Rill (visualization) provides a complete observability platform that scales from startups to enterprise.

<img src='/img/guides/monitoring/prometheus-explore.png' class='rounded-gif' />
<br />

## Resources

- [Rill Documentation](https://docs.rilldata.com)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)  
- [Thanos Architecture Guide](https://thanos.io/tip/thanos/design.md)