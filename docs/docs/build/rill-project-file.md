---
title: "Rill Project File"
description: For documenting required migrations
sidebar_label: "Rill Project File"
sidebar_position: 00
---

The `rill.yaml` file is automatically generated when you start any project in Rill and serves as the central configuration hub for your entire project. While often overlooked, this powerful file enables you to set project-wide defaults, configure environment variables, define connector settings, create test users, and establish security policies across all your metrics views and dashboards.

Let's walk through the key capabilities of `rill.yaml`, including setting up [`mock_users`](/build/metrics-view/security#testing-policies-in-rill-developer) to test row access policies, configuring default security settings for your metrics views and explore dashboards, establishing refresh schedules, defining MCP `ai_instructions`, and setting the default OLAP connector and more.

Here is an example YAML that uses many of our features.
```yaml
compiler: rillv1

display_name: Rill Project Dev

# The project's default OLAP connector.
# Learn more: https://docs.rilldata.com/reference/olap-engines
olap_connector: duckdb

#Project Defaults
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
        
ai_instructions: |
  Greet the user and remind them that not all AI answers should be used to make decisions without checking the underlying dashboard first.
  When you run queries with rill, you will include corresponding Rill Explore URLs in your answer. All URLs should start with the BASE_URL, which is defined below. 
  The full URL should include the time range (tr) used in the report, the timezone (tz), and any measures or dimensions that are relevant to the report. See the examples below.
  # Example
  URL for an explore with multiple metrics and dimensions
  ## Description
  A link to an online dashboard from Rill. Contains all selected metrics in the report, all dimensions used in the report, and up to 1-3 additional dimensions. Time range includes the range used as the focus of the report, plus a comparison period for enriched visualization. It is in markdown format, and has a link that describes the purpose of the link.
  ## Format 
  Markdown
  ## Link
  [https://ui.rilldata.com/demo/rill-openrtb-prog-ads/explore/bids_explore?tr=2025-05-17T23%3A00%3A00.000Z%2C2025-05-19T23%3A00%3A00.000Z&tz=Europe%2FLondon&compare_tr=rill-PP&measures=overall_spend%2Ctotal_bids%2Cwin_rate%2Cvideo_completes%2Cavg_bid_floor&dims=advertiser_name%2Csites_domain%2Capp_site_name%2Cdevice_type%2Ccreative_type%2Cpub_name](Explore change in advertising bids due to composition of advertisers)

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

For a list of all supported settings, see our [project YAML reference page](/reference/project-files/rill-yaml).
##  OLAP Connector
When adding an OLAP connector to your project, this will automatically populate with the new OLAP connector. e.g., `ClickHouse`, `Druid` If you create multiple connectors, we will append "_#" to the file name and use this as the default connector.

The default OLAP connector is used as the default `output` for all of your models unless otherwise specified.



## Project Defaults

### Model Refresh Schedule
Set up your project's model refresh schedule. You can override this in the model's YAML file if needed.
```yaml
models:
    refresh:
        cron: '0 * * * *'
```
### Metrics Views Time Modifiers

Set the default time modifiers such as `first_day_of_week` or `smallest_time_grain` as seen below. For more parameters, see [metrics view referene page](/reference/project-files/metrics-views).

```yaml
metrics_views:
    first_day_of_week: 1
    smallest_time_grain: month
```

### Metrics Views Security Policy
By default, Rill is open to access (to your organization users), unless otherwise defined. To add project-level access to the Rill project, you can add a default metrics view security policy in the `rill.yaml` file. Like a metrics_view, you can define the security as shown below. For more information, read our [data access documentation](/build/metrics-view/security#examples).

```yaml
metrics_views:
  security:
    access: {boolean expression}
    row_filter: {SQL expression}
```


:::tip Security Policy Rules

Rill YAML settings < (Metrics View YAML AND Dashboard YAML)

For detailed guide on security policies, review our [data access policies](/build/metrics-view/security) doc.
:::


### Explore Security Policy
Similar to metrics views, you can set [security for an explore dashboard](/build/dashboards/#define-dashboard-access). (Note that only `access` can be set at the dashboard level.)

```yaml
explores:
  security:
    access: {boolean expression}
```

### Dashboard Defaults

You are also able to set the `defaults` parameter in the explore dashboard to define your default time range, as well as the available `time_zones` and `time_ranges` in an Explore dashboard.
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


:::tip Why dont I see the YAML view?

In Rill Cloud, we save a user's last state on the explore dashboard. Therefore, your users will not see the defined view above but the view they last left on. 

Rill YAML settings < Explore Dashboard YAML < Bookmarks in Rill Cloud < User Last State
:::

### Differentiating dev and prod environments

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



## `ai_instructions`
Use the `ai_instructions` field to provide information that is **unique to your project**. This helps the agent deliver more relevant and actionable insights tailored to your specific needs.

**What to include:**
- Guidance on which metrics views are most important or should be prioritized for your project.
- Any custom business logic, definitions, or terminology unique to your data or organization.
- Preferences for aggregations, filters, or dimensions that are especially relevant to your use case.

**Example:**
```yaml
ai_instructions: |
  Focus on the `ad_performance` and `revenue_overview` metrics views, as these are most critical for our business users.
  When possible, highlight trends by region and product category.
  Use our internal terminology: "campaign" refers to a single ad initiative, and "placement" refers to a specific ad slot.
```

:::note 
For metric-level specific instructions, `ai_instructions` can also be applied there. 
:::

## Test Access Policies in Rill Developer
Access to your environment is a crucial step in creating your project in Rill Developer. By doing so, you can confidently push your dashboard changes to Rill Cloud. This is done via the `mock_users` in the project file. You can create pseudo-users with specific domains, or admin and non-admin users or user groups, to ensure that access is correct. 

Let's assume that the following are applied to the metrics view.

```yaml
security:
    access: "{{ .user.admin }} OR '{{ .user.domain }}' == 'example.com'"
    row_filter: "domain = '{{ .user.domain }}'"
```

In order to test both access to the dashboard, as well as the row filter, you can create the following in the project YAML.

```yaml
mock_users:
  - email: royendo@rilldata.com
    admin: true
  - email: your_email@domain.com
    groups:
      - tutorial-admin
  - email: your_email2@another_domain.com
```

### Custom Attributes
other type of dashboard that you'll want to test in Rill Developer, mainly because there is no way to pass custom variables in Rill Cloud to ensure that access and data are being presented correctly. To do this, you'll need to add the following to your `mock_users`:
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

If you are interested in testing our upcoming features and experimental functionality, you can enable feature flags in your `rill.yaml` file. These flags allow you to access beta features and provide early feedback on new capabilities before they become generally available.

To enable feature flags, add the `features` section to your `rill.yaml`:


```yaml
features:
  - cloudDataViewer
```