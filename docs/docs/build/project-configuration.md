---
title: "Project Configuration"
description: "Complete guide to configuring your Rill project with rill.yaml and other project files"
sidebar_label: "Project Configuration"
sidebar_position: 50
---

# Project Configuration

This guide covers all aspects of configuring your Rill project, from basic settings to advanced security and testing configurations.
## OLAP Connector

When you add an OLAP connector to your project, Rill automatically updates the `olap_connector` field in `rill.yaml` with the new connector name (e.g., `clickhouse`, `druid`). 

If you create multiple connectors of the same type, Rill will number them sequentially. For example, if you create three ClickHouse connectors, they will be named:
- `clickhouse` (first connector)
- `clickhouse_2` (second connector)  
- `clickhouse_3` (third connector)

The most recently created connector becomes the default and is referenced in `rill.yaml`.

```yaml
olap_connector: clickhouse
```

The default OLAP connector is used as the default `output` for all of your models unless otherwise specified.



## Model Defaults

Configure default behaviors for all models in your project. These settings control refresh schedules and materialization behavior that apply across all models unless overridden in individual model files.

### Model Refresh Schedule

Set up your project's model refresh schedule. You can override this in the model's YAML file if needed.

```yaml
models:
    refresh:
        cron: '0 * * * *'
```

For more information on configuring model refreshes, see the [models documentation](/build/models) and [model refresh reference](/reference/project-files/models#refresh).

### Model Materialization

By default, models in Rill are created as views rather than materialized tables. While views enable a dynamic keystroke-by-keystroke modeling experience, materializing models as tables can significantly improve performance for complex queries and large datasets.

You can set a default materialization behavior for all models in your project:

```yaml
models:
  materialize: true
```

This is particularly useful in production environments where performance is prioritized over the interactive modeling experience. Individual models can still override this setting using `-- @materialize: false` in their SQL file, or `materialize: false` in the YAML file.

:::tip When to materialize
- **Recommended:** Final models that power dashboards should be materialized for better performance
- **Consider:** Complex intermediate models with large datasets
- **Avoid:** Models with cross joins or when you need real-time keystroke-by-keystroke updates

For more details, see our [model performance guide](/build/models/performance#materialization).
:::

## Metrics Views Defaults

Configure default settings for all metrics views in your project. These settings control time-related behavior and security policies that apply across all metrics views unless overridden in individual metrics view files. For comprehensive information on metrics views, see the [metrics view documentation](/build/metrics-view).


### Metrics Views Time Modifiers

Set default time modifiers for all metrics views, such as `first_day_of_week` or `smallest_time_grain` as shown below. For more parameters, see the [metrics view reference page](/reference/project-files/metrics-views).

```yaml
metrics_views:
    first_day_of_week: 1
    smallest_time_grain: month
```

### Metrics Views Security Policy

By default, Rill is open to access (to your organization users), unless otherwise defined. To add project-level access to the Rill project, you can add a default metrics view security policy in the `rill.yaml` file. Like in a metrics view file, you can define the security as shown below. For more information, read our [data access documentation](/build/metrics-view/security#examples).

```yaml
metrics_views:
  security:
    access: '{{ has "partners" .user.groups }}'
    row_filter: "domain = '{{ .user.domain }}'"
```

:::tip Security Policy Rules

Rill YAML settings < (Metrics View YAML AND Dashboard YAML)

For detailed guide on security policies, review our [data access policies](/build/metrics-view/security) doc.
:::


## Dashboard Defaults

Rill supports two types of dashboards: **Explores** (metrics-focused dashboards) and **Canvases** (custom visualization dashboards). You can set default configurations for each type. Learn more about creating and customizing dashboards in our [dashboard documentation](/build/dashboards).

### Dashboard Security Policy

Similar to metrics views, you can set [security for a dashboard](/build/dashboards/customization#define-dashboard-access). (Note that only `access` can be set at the dashboard level.)

```yaml
explores:
  security:
    access: "'{{ .user.domain }}' == 'example.com'"
canvases:
  security:
    access:  '{{ has "dev" .user.groups }}'
```

### Explore Defaults

You are also able to set the `defaults` parameter in the explore dashboard to define your default time range, as well as the available `time_zones` and `time_ranges` in an Explore dashboard.

:::note Time Range Format
Time ranges use [ISO 8601 duration format](https://en.wikipedia.org/wiki/ISO_8601#Durations). Common examples:
- `PT24H` = 24 hours
- `P7D` = 7 days  
- `P3M` = 3 months
- `P24M` = 24 months
:::

```yaml
explores:
    defaults:
        time_range: P3M
    time_zones:
        - UTC
    time_ranges:
        - PT24H
        - P7D
        - P14D
        - P3M
```

### Canvas Defaults

Similarly, you can configure defaults for canvas dashboards:

```yaml
canvases:
    defaults:
        time_range: P7D
    time_zones:
        - UTC
    time_ranges:
        - PT24H
        - P7D
        - P14D
        - P3M
```

:::tip Understanding Dashboard Default Behavior

When setting dashboard defaults, keep in mind that Rill Cloud saves each user's last state on dashboards. This means users will see the view they last used rather than the defaults configured above.

**Settings precedence (lowest to highest):**  
Rill YAML settings < Dashboard YAML < Bookmarks in Rill Cloud < User Last State

Defaults are most useful for new users accessing a dashboard for the first time or when viewing in Rill Developer.
:::

## Environment Configuration

### Differentiating Dev and Prod Environments

Rill comes with default `dev` and `prod` properties defined, corresponding to Rill Developer and Rill Cloud, unless otherwise specified in the `rill start --environment (dev/prod)` command for Rill Developer. You can use these keys to set environment-specific YAML overrides or SQL logic.

For example, the following `rill.yaml` file explicitly sets the default materialization setting for models to `false` in development and `true` in production:

```yaml
dev:
  models:
    materialize: false

prod:
  models:
    materialize: true
```

:::note Specifying a custom environment

When using Rill Developer, instead of defaulting to `dev`, you can run your project in production mode using the following command:

```bash
rill start --environment prod
```

:::

## Variable Management

Variables in Rill enable dynamic templating throughout your project files. They can be used in SQL queries, YAML configurations, and security policies to make your project more flexible and maintainable.

### Setting Variables

Variables are defined in your `rill.yaml` file using the `env` key. This allows you to set variables that will be available in your Rill Cloud deployments while maintaining the ability to use different values locally during development.

```yaml
env:
  numeric_var: 10
  string_var: "string_value"
  api_endpoint: "https://api.example.com"
```

Once defined, these variables can be referenced in your project files using template syntax: `{{ .env.variable_name }}`

**Example usage in a model:**
```sql
SELECT *
FROM users
WHERE status = '{{ .env.string_var }}'
  AND score > {{ .env.numeric_var }}
```

## AI Configuration

### `ai_instructions`

Use the `ai_instructions` field to provide information that is **unique to your project**. This helps the AI agent deliver more relevant and actionable insights tailored to your specific needs. For more information on AI features in Rill, see our [AI documentation](/explore/mcp#adding-ai-instructions-to-your-model).

**What to include:**
- Guidance on which metrics views are most important or should be prioritized for your project
- Any custom business logic, definitions, or terminology unique to your data or organization
- Preferences for aggregations, filters, or dimensions that are especially relevant to your use case
- Specific business context that helps the AI understand your domain

**Examples:**

*E-commerce project:*
```yaml
ai_instructions: |
  Focus on the `ad_performance` and `revenue_overview` metrics views, as these are most critical for our business users.
  When possible, highlight trends by region and product category.
  Use our internal terminology: "campaign" refers to a single ad initiative, and "placement" refers to a specific ad slot.
  Always include conversion rates when discussing revenue metrics.
```

*SaaS analytics project:*
```yaml
ai_instructions: |
  Prioritize user engagement metrics over raw user counts.
  Our key business metrics are monthly recurring revenue (MRR) and customer lifetime value (CLV).
  Segment analysis by customer tier (Enterprise, Pro, Basic) is essential.
  Churn analysis should focus on the 30-day and 90-day windows.
```

:::note 
For metric-level specific instructions, `ai_instructions` can also be applied there. 
:::

## Testing Security

### Test Access Policies in Rill Developer

Testing access policies in your local environment is a crucial step before deploying to Rill Cloud. This is done via the `mock_users` in the project file. You can create pseudo-users with specific domains, admin/non-admin roles, or user groups to ensure that access policies work as intended. For comprehensive information on security policies, see our [data access policies documentation](/build/metrics-view/security). 

Let's assume that the following security policy is applied to the metrics view.

```yaml
security:
    access: "{{ .user.admin }} OR '{{ .user.domain }}' == 'rilldata.com'"
    row_filter: "region = '{{ .user.region }}'"
```

In order to test both access to the dashboard, as well as the row filter, you can create the following in the project YAML.

```yaml
mock_users:
  - email: royendo@rilldata.com
    admin: true
    region: us-west
  - email: your_email@domain.com
    groups:
      - tutorial-admin
    region: us-east
  - email: your_email2@another_domain.com
    region: europe
```

See our embedded example, [here](https://rill-embedding-example.netlify.app/rowaccesspolicy/basic).

### Custom Attributes

Embedded dashboards allow passing custom attributes (variables) from your application to control access and filtering. These attributes are set when generating the embed JWT token in your application code. For more information on embedding dashboards, see our [embedding documentation](/integrate/embedding).

To test embedded dashboards locally with custom attributes, add them to `mock_users`:

```yaml
- email: embed@rilldata.com
  name: embed
  custom_variable_1: Value_1 #this is passed at embed creation
  custom_variable_2: Value_2 #this is passed at embed creation
```

See our [Custom Attributes Embedded Dashboard](https://rill-embedding-example.netlify.app/rowaccesspolicy/custom) live!

Let's assume a similar setup to the above example. Within the metrics view, we define:

```yaml
security:
  access: true
  row_filter: >
    app_site_name = '{{ .user.app_site_name }}' AND
      pub_name = '{{ .user.pub_name }}'
```

Then within the application we are passing

```yaml
app_site_name='Sling'
pub_name='MobilityWare'
```

You can create a test mock user to ensure that this dashboard is working as designed with the following:

```yaml
- email: embed@rilldata.com
  name: embed
  app_site_name: 'Sling' 
  pub_name: 'MobilityWare'
```

<img src = '/img/tutorials/admin/custom-attribute-mock-user.png' class='rounded-gif' />
<br />

## Feature Flags

:::warning Experimental Features
Feature flags enable experimental functionality that may be unstable or change before general availability. Use with caution in production environments.
:::

If you are interested in testing our upcoming features and experimental functionality, you can enable feature flags in your `rill.yaml` file. These flags allow you to access beta features and provide early feedback on new capabilities before they become generally available.

To enable feature flags, add the `features` section to your `rill.yaml`:

```yaml
features:
  - cloudDataViewer
```


**Available feature flags:**
- `cloudDataViewer`: Enables the cloud data viewer interface for exploring data directly in the browser (default: `false`)
- `dimensionSearch`: Enables advanced dimension search functionality (default: `false`)
- `twoTieredNavigation`: Enables two-tiered navigation interface (default: `false`)
- `rillTime`: Enables Rill-specific time functionality (default: `false`)
- `hidePublicUrl`: Hides public URL sharing options (default: `false`)
- `exportHeader`: Enables export header functionality (default: `false`)
- `alerts`: Enables alerting features (default: `true`)
- `reports`: Enables reporting functionality (default: `true`)
- `darkMode`: Enables dark mode interface (default: `true`)
- `chat`: Enables chat functionality (default: `true`)
- `dashboardChat`: Enables chat features within dashboards (default: `false`)

**How to enable**: Add the `features` section to your `rill.yaml`

**Reporting issues**: If you encounter issues with feature flags, please [contact support](/contact) with details about the flag and behavior.

For a complete list of available feature flags and their current status, see our [feature flags reference](https://github.com/rilldata/rill/blob/main/web-common/src/features/feature-flags.ts#L36) in the codebase.

## Complete Example

Here is an example YAML that uses many of our features.

```yaml
compiler: rillv1

display_name: Rill Project Dev

# The project's default OLAP connector.
# Learn more: https://docs.rilldata.com/reference/olap-engines
olap_connector: duckdb

# Project variables for templating
env:
  numeric_var: 10
  string_var: "string_value"

# Project Defaults
models:
    refresh:
        cron: '0 * * * *'
        run_in_dev: true
metrics_views:
    first_day_of_week: 1
    smallest_time_grain: month
explores:
    defaults:
        time_range: P24M
    time_zones:
        - UTC
        - America/Los_Angeles
        - America/New_York
        - Europe/London
        - Europe/Paris
        - Asia/Tokyo
        - Australia/Sydney
    time_ranges:
        - PT24H
        - P6M
        - P12M
canvases:
  defaults:
      time_range: P24M
  time_zones:
      - UTC
      - America/Los_Angeles
      - America/New_York
      - Europe/London
      - Europe/Paris
      - Asia/Tokyo
  time_ranges:
      - PT24H
      - P7D
      - P14D
      - P30D
      - P3M
      - P6M
      - P12M

# These are example mock users to test your security policies.
# Learn more: https://docs.rilldata.com/manage/security
mock_users:
  - email: john@yourcompany.com
  - email: jane@partnercompany.com
  - email: your_email@domain.com
    groups:
      - tutorial-admin
  - email: embed@rilldata.com
    name: embed
    custom_variable_1: Value_1 #this is passed at embed creation
    custom_variable_2: Value_2 #this is passed at embed creation

features:
  - cloudDataViewer
```
