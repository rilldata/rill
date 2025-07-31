---
title: Who Can See Your Data
description: Control who has access to view your metrics and data
sidebar_label: Data Access Control
sidebar_position: 10
---

Rill supports granular access policies for dashboards. They allow the dashboard developer to configure dashboard-level, row-level and column-level restrictions based on user attributes such as email address and domain. Our goal with access to policies is to avoid dashboard sprawl by creating a single configuration of the dashboard that can then be sliced or restricted into multiple views via different policies. Using those access controls, a single dashboard can now serve dozens of teams and use cases to ensure consistent metric definitions and better dashboard findability.

Typical use cases include:

- **Granting or Restricting Access** to data and as a result, dashboards
- **Hiding specific dimensions and measures** from specific groups of users, creating a tailored dashboard experience
- **Restricting Access to Internal users** of your organization, allowing specific dashboards to be viewed by internal users only
- **Partner-filtered Dashboards** where external users can only access the subset of their data
- **Embedded** use cases, passing custom attributes to Rill
- **Combination of all the above**


:::tip Assuming Access to Project is already given

Access Policies assume that the user already has access to the project in Rill Cloud. For more information on user management, see our [User Management](/manage/user-management) and [Project Management](/manage/project-management) for more information!

:::

## Configurations

There are three levels of considerations for access policies. 

- **Data Access** - `access`– a boolean expression that determines if a user can or can't access the dashboard
- **Row-level access:** `row_filter` – a SQL expression that will be injected into the WHERE clause of all dashboard queries to restrict access to a subset of rows
- **Column-level access**: `include` or `exclude` – lists of boolean expressions that determine which dimension and measure names will be available to the user

<img src='/img/manage/security/access.png' />

## Setting up Data Access

There are two locations that control data access in Rill.

1. Project level access.
2. Metrics View level access.

### Project Level Defaults

By default, when a user is granted access to your project, they have access to all metrics views. This is not always the desired behavior as some organizations will invite partner users to the Rill Cloud UI. In these instances, project level defaults that only give access to internal domain is required.


### Metrics View Specific



## User Attributes


## Testing Policies in Rill Developer


### Rill Cloud


### Embedded Dashboards