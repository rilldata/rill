---
title: "Dashboard Access Policies"
description: Granular, row-level security for dashboards
sidebar_label: "Dashboard Access Policies"
sidebar_position: 40
---

Rill supports granular access policies for dashboards. They allow the dashboard developer to configure dashboard-level, row-level and column-level restrictions based on user attributes such as email address and domain. Our goal with access to policies is to avoid dashboard sprawl by creating a single configuration of the dashboard that can then be sliced or restricted into multiple views via different policies. Using those access controls, a single dashboard can now serve dozens of teams and use cases to ensure consistent metric definitions and better dashboard findability.

Some of the typical use cases include:

- Restricting access to certain dashboards to admin users
- Limiting dashboards to relevent fields or metrics by team for ease of use (creating a lookup and filter by role) 
- Limiting access to sensitive dimensions or measures to users from a specific department
- Partner-filtered dashboards where external users can only see the subset of data that relates to them

:::note User Access vs. Access Policies
Access policies only apply to users who have been invited to access the project. They provide granular access control for your data, but are not the first layer of security for your project.
:::

## Configuration

You define access policies for dashboards under the `security` key in a dashboard's YAML file. The key properties are:

- **Dashboard-level access:** `access` – a boolean expression that determines if a user can or can't access the dashboard
- **Row-level access:** `row_filter` – a SQL expression that will be injected into the `WHERE` clause of all dashboard queries to restrict access to a subset of rows
- **Column-level access:** `include` or `exclude` – lists of boolean expressions that determine which dimension and measure names will be available to the user

![access](../../static/img/manage/security/access.png)

See the [Dashboard YAML](/reference/project-files/explore-dashboards) reference docs for all the available fields.

See the [Examples](#examples) below for how to set up each type of configuration.

## User attributes

When developing access policies, you can leverage a fixed set of user attributes to resolve access at runtime. The attributes are:

- `.user.email` – the current user's email address, for example `john.doe@example.com` (string)
- `.user.domain` – the domain of the current user's email address, for example `example.com` (string)
- `.user.name` - the current user's name, for example `John Doe` (string)
- `.user.admin` – a boolean value indicating whether the current user is an org or project admin, for example `true` (bool)
- `.user.groups` - a list of user groups the user belongs to in the project's org (list of strings), e.g. `["marketing","sales","finance"]`

Note: Rill requires users to confirm their email address before letting them interact with the platform so a user cannot fake an email address or email domain.

If you require additional user attributes to enforce access policies, see the [example for custom attributes below](#advanced-example-custom-attributes) for more details.

## Templating and expression evaluation

When a user loads a dashboard, the policies are resolved in two phases:

1. The templating engine first replaces expressions like `{{ .user.domain }}` with actual values ([Templating reference](../deploy/templating.md))
2. The resulting expression is then evaluated contextually:
  - The `access` and `if` values are evaluated as SQL expressions and resolved to a `true` or `false` value
  - The `row_filter` value is injected into the `WHERE` clause of the SQL queries used to render the dashboard

## Testing your policies

### In Rill Developer

In development (on `localhost`), you can test your policies by adding "mock users" to your project and viewing the dashboard as one of them.

In your project's `rill.yaml` file, add a `mock_users` section. Each mock user must have an `email` attribute, and can optionally have `name` and `admin` attributes. For example:
```yaml
# rill.yaml
mock_users:
- email: john@yourcompany.com
  name: John Doe
  admin: true
- email: jane@partnercompany.com
  groups:
    - partners
- email: anon@unknown.com
```

On the dashboard page (provided you've added a policy) you'll see a "View as" button in the top right corner. Click this button and select one of your mock users. You'll see the dashboard as that user would see it.

### In Rill Cloud (Admin only)

In case you want to test what your users are seeing in Rill Cloud after deploying, you can find this in the dropdown of your account. You will see the actual users in the dropdown of this list, not the mock users defined in the rill.yaml file. 

![rc_view_as](/img/manage/access-policies/rill-cloud-view-as.png)

## Examples

### Restrict dashboard access to users matching specific criteria

Let's say you want to restrict dashboard access to admin users or users whose email domain is `example.com`. Add the following clause to your dashboard's YAML:
```yaml
security:
  access: "{{ .user.admin }} OR '{{ .user.domain }}' == 'example.com'"
```

:::note DEFAULT SECURITY IS FALSE
If the `security` section is defined and `access` is not, then `access` will default to `false`, meaning that it won't be accessible to anyone and users will need to invited individually.
:::

### Restrict dashboard access to specific user groups

Group membership can be utilized to specify which users have access to a specific dashboard (using the templating function `has`). For example:
```yaml
security:
  access: '{{ has "partners" .user.groups }}'
```


### Show only data from the user's own domain

You can limit the data available to the dashboard by applying a filter on the underlying data. Assuming the dashboard's underlying model has a `domain` column, adding the following clause to the dashboard's YAML will only show dimension and measure values for the current user's email domain:

```yaml
security:
  access: true
  row_filter: "domain = '{{ .user.domain }}'"
```

:::note FILTERS SHOULD BE VALID SQL
The `filter` value needs to be valid SQL syntax for a `WHERE` clause. It will be injected into every SQL query used to render the dashboard.
:::

### Conditionally hide a dashboard dimension or measure

You can include or exclude dimensions and measures based on a boolean expression. For example, to exclude a dimension named `ssn` and `id` for users whose email domain is not `example.com`:

```yaml
security:
  access: true
  exclude:
    - if: "'{{ .user.domain }}' != 'example.com'"
      names: 
        - ssn
        - id
```

Alternatively, you can explicitly define the dimensions and measures to include using the `include` key. It uses the same syntax as `exclude` and automatically excludes all names not explicitly defined in the list. See the [Dashboard YAML](/reference/project-files/explore-dashboards) reference for details.

### Use wildcards to select all dimensions and measures

When defining inclusion policies, you can easily and automatically select all columns by using `names: '*'` as a wildcard. For example:
```yaml
security:
  access: true
  include:
    - if: true
      names:
        - ssn
        - id
    - if: "{{ .user.admin }}"
      names: '*'
```

Note that the `'*'` must be quoted (using single or double quotes), and **must** be provided as a scalar value, not as an entry in a list.

### Filter queries based on the user's groups

You can directly inject the groups that a user belongs to into the row filter itself, such as:
```yaml 
security:
  access: true
  row_filter: "groups IN ('{{ .user.groups | join \"', '\" }}')"
```

### Hide dimensions or measures for members of a certain group

You can check group membership using the templating function `has`. For example:
```yaml
security:
  access: true
  exclude:
    - if: '{{ has "partners" .user.groups }}'
      names:
        - cost
        - profit
```


### Advanced Example: Mapping Dimensions and Attributes

For some use cases, the built-in user attributes do not provide sufficient context to correctly restrict access. For example, a dashboard for a multi-tenant SaaS application might have a `tenant_id` column and external users should only be able to see data for the tenant they belong to.

To support this, ingest a separate data [source](/build/connect) containing mappings of user email addresses to tenant IDs and reference it in the row-level filter. This can be a locally created csv file or any hosted data source.

For example, a locally created `mappings.csv` file in the `data` directory of our Rill project with the following contents:
```csv
email,tenant_id
john.doe@example.com,1
jane.doe@example.com,2
```

This requires to be ingested as a source in Rill like any other data source:
```yaml
# sources/mappings.yaml
type: local_file
path: data/mappings.csv
```
(In practice, you would probably ingest the data from a regularly updated export in S3 with a source refresh.)

We can now refer to the mappings data using a SQL sub-query as follows:
```yaml
security:
  access: true
  row_filter: "tenant_id IN (SELECT tenant_id FROM mappings WHERE email = '{{ .user.email }}')"
```


### Advanced Example: Custom attributes (Embed Dashboards)

Another use case for row access policies is to ensure that your embed dashboards is providing a specific view for your end users. During the [embed dashboard request](/integrate/embedding), you can pass custom attributes (other than the ones provided OOB) that maps directly to a value within your Rill explore dashboard.

```yaml
row_filter: >
      dimension_1 = '{{ .user.custom_variable_1 }}' AND
      dimension_2 = '{{ .user.custom_variable_2 }}' 
```

In order to test the view of your embed dashboard, you can add the same custom variables to [your mock users](#testing-your-policies) as seen below:
```yaml
- email: embed@rilldata.com
  name: embed
  custom_variable_1: Value_1
  custom_variable_2: Value_2
  ```