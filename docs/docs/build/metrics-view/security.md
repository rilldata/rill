---
title: Who Can Access Your Data
description: Control who can view your metrics and data
sidebar_label: Data Access Control
sidebar_position: 20
---

Rill supports **granular access policies** that let you control:

- **Who can access your data**
- **What rows they can see**
- **Which dimensions and measures are visible**

Policies are based on user attributes such as **email address**, **domain**, or **custom attributes**.  This avoids dashboard sprawl — instead of creating multiple dashboards for each audience, you can build _**one dashboard**_ and tailor it for many teams and use cases.


## How Does It Work?

Access policies are defined in the **metrics view** and/or **[dashboard YAML](/build/dashboards/customization#define-dashboard-access)**.  
There are three types of rules:

**General Access:** (`access`) A boolean expression deciding if a user can access the metrics view
```yaml
security:
  access: "{{ .user.admin }} OR '{{ .user.domain }}' == 'example.com'"
```

:::info Dashboard access

`access` can be set on both the dashboard YAML and metrics view YAML and policies are combined using logical AND operations. If no policies are defined on the dashboard, they are derived from the metrics view. For most set-ups, setting the access on the metrics view is sufficient.

:::

**Row-level access** (`row_filter`) – a SQL expression that will be injected into the WHERE clause of all dashboard queries to restrict access to a subset of rows
```yaml
security:
  row_filter: region = '{{ .user.region }}'
```

**Column-level access** (`include` or `exclude`) – lists of boolean expressions that determine which dimension and measure names will be available to the user
```yaml
security:
  exclude:
    - if: "'{{ .user.domain }}' != 'example.com'"
      names:
        - ssn
        - id
```

When a user loads a dashboard, the policies are resolved in two phases:
  1. The templating engine first replaces expressions like `{{ .user.domain }}` with actual values ([Templating reference](/build/connectors/templating))
  2. The resulting expression is then evaluated contextually:
     - The `access` and `if` values are evaluated as SQL expressions and resolved to a `true` or `false` value
     - The `row_filter` value is injected into the `WHERE` clause of the SQL queries used to render the dashboard


:::info What about MCP, and APIs?

Metrics views limit data access for all requests, including MCP integrations and custom APIs. When creating a token or copying from the AI tab, the user's attributes (such as email, domain, groups, and custom attributes) are automatically included in the request context. This ensures that the same security policies that apply to dashboard users also apply to programmatic access, maintaining consistent data governance across all access methods.

For more details, see [Service Tokens](/manage/service-tokens).
:::
Typical use cases include:

- [**Granting or Restricting Access**](#restrict-data-access-to-users-matching-specific-criteria) to data and, as a result, dashboards
- [**Hiding specific dimensions and measures**](#conditionally-hide-a-dashboard-dimension-or-measure) from specific groups of users, creating a tailored dashboard experience
- [**Restricting Access to Internal users**](#hide-dimensions-or-measures-for-members-of-a-certain-group) of your organization, allowing specific dashboards to be viewed by internal users only
- [**Partner-filtered Dashboards**](#show-only-data-from-the-users-own-domain) where external users can only access the subset of their data
- [**Embedded**](#advanced-example-custom-attributes-embed-dashboards) use cases, passing custom attributes to Rill

:::tip Project Access Required

Access Policies assume that the user already has access to the project in Rill Cloud. For more information on user management, see our [User Management](/manage/user-management) and [Project Management](/manage/project-management) documentation.

:::

## Creating Access Policies

There are two locations that control data access in Rill.

### Project Level Defaults

By default, when a user is granted access to your project, they have access to all metrics views and, if there is [no dashboard policy](/build/dashboards/customization#define-dashboard-access), all dashboards. While this is the default behavior, it can be easily changed in the project's `rill.yaml`. This will lock down all metrics views and block all users who are not Rill Administrators or do not have 'example.com' as their domain.

:::tip Set project-wide security defaults
Configure default security policies for all metrics views and dashboards in your project.
[Learn more about security defaults →](/build/project-configuration#metrics-views-security-policy)
:::

```yaml
metrics_views:
  security:
    access: "{{ .user.admin }} OR '{{ .user.domain }}' == 'example.com'"
    row_filter: "partner_id IN (SELECT id FROM mapping WHERE partner_domain = '{{ .user.domain }}') OR '{{ .user.domain }}' = 'example.com'"
    exclude:
      - if: "'{{ .user.domain }}' != 'example.com'"
        names: 
          - ssn
          - id
```

### Object Specific Policies

You can define policies directly in a specific metrics view or dashboard YAML to override the project-level defaults.

```yaml
security:
  access: '{{ has "partners" .user.groups }}'
  row_filter: "domain = '{{ .user.domain }}'"
  exclude:
    - if: "'{{ .user.domain }}' != 'example.com'"
      names: 
        - ssn
        - id
```

:::tip Access Policy Behavior

When combining access policies from project defaults and object-specific policies, remember that the object level policies will overwrite the project level ones. Dashboard and metrics view policies are combined using logical AND operations.
 

:::

## Dashboard Access

Dashboards also have an `access` key that can add additional security to the metrics view. Both [explore](/build/dashboards/customization#define-dashboard-access) and [canvas](/build/dashboards/customization#define-dashboard-access) dashboards can set the following:

```yaml
security:
  access: "'{{ .user.domain }}' == 'example.com'"
```

This will logically AND with your metrics view's access so ensure that a user who needs access to the dashboard passes **both** conditions.

:::tip complicated set-ups

Access Policies can get quite complicated as your use case grows and having to navigate multiple files to figure out why a user is able to or unable to access certain dashboards. 

A few recommendations:
1. Only change project level access if absolutely necessary. (They get overwritten by object level security)
2. Dashboard access can be derived from the metrics view, only add extra policies on the dashboard if absolutely necessary as this gets combined with the metrics view using logical AND operations anyway.
3. Solve project access issues higher up in the [user](/manage/user-management) / [usergroup](/manage/usergroup-management) settings, and keep default project security rules.

:::

## User Attributes
- `.user.email` – the current user's email address, for example john.doe@example.com (string)
- `.user.domain` – the domain of the current user's email address, for example example.com (string)
- `.user.name` - the current user's name, for example John Doe (string)
- `.user.admin`– a boolean value indicating whether the current user is an org or project admin, for example true (bool) 
- `.user.groups` - a list of user groups the user belongs to in the project's org (list of strings), e.g. ["marketing","sales","finance"]
- `.user.attribute` - where `attribute` is a custom variable that you can pass via an embedded dashboard from your application

Note: Rill requires users to confirm their email address before letting them interact with the platform, so a user cannot fake an email address or email domain.



## Testing Policies in Rill Developer

In development (on `localhost`), you can test your policies by adding "mock users" to your project and viewing the dashboard as one of them.

:::tip Test policies in Rill Developer
Use `mock_users` in rill.yaml to test your security policies before deploying.
[Learn more about testing security →](/build/project-configuration#testing-security)
:::

In your project's `rill.yaml` file, add a `mock_users` section. Each mock user must have an `email` attribute and can optionally have `name` and `admin` attributes. For example:
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

On the dashboard page (provided you've added a policy), you'll see a "View as" button in the top right corner. Click this button and select one of your mock users. You'll see the dashboard as that user would see it.

### Rill Cloud
If you want to test what your users are seeing in Rill Cloud after deploying, you can find this in the dropdown of your account. You will see the actual users in the dropdown of this list, not the mock users defined in the rill.yaml file.

<img src = '/img/manage/access-policies/rill-cloud-view-as.png' class='rounded-gif' />
<br />

### Embedded Dashboards

When [requesting an embedded dashboard from Rill](/integrate/embedding) from your frontend, you can pass the `attributes` parameter with custom names to ensure that the resulting dashboard displays the correct information.

For more information, see [our embedding docs](/integrate/embedding#backend-build-an-iframe-url).


## Examples

### Restrict data access to users matching specific criteria

Let's say you want to restrict dashboard access to admin users or users whose email domain is `example.com`. Add the following clause to your metrics view's YAML:
```yaml
security:
  access: "{{ .user.admin }} OR '{{ .user.domain }}' == 'example.com'"
```

:::note DEFAULT SECURITY IS FALSE
If the `security` section is defined and `access` is not, then `access` will default to `false`, meaning that it won't be accessible to anyone and users will need to be invited individually.
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
The `row_filter` value needs to be valid SQL syntax for a `WHERE` clause. It will be injected into every SQL query used to render the dashboard.
:::

### Conditionally hide a dashboard dimension or measure

You can include or exclude dimensions and measures based on a boolean expression. For example, to exclude dimensions named `ssn` and `id` for users whose email domain is not `example.com`:

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

Note that the `'*'` must be quoted (using single or double quotes) and **must** be provided as a scalar value, not as an entry in a list.

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

For some use cases, the built-in user attributes do not provide sufficient context to correctly restrict access. For example, a dashboard for a multi-tenant SaaS application might have a `tenant_id` column, and external users should only be able to see data for the tenant they belong to.

To support this, ingest a separate data [source](/build/connectors) containing mappings of user email addresses to tenant IDs and reference it in the row-level filter. This can be a locally created CSV file or any hosted data source.

For example, a locally created `mappings.csv` file in the `data` directory of your Rill project with the following contents:
```csv
email,tenant_id
john.doe@example.com,1
jane.doe@example.com,2
```

This needs to be ingested as a source in Rill like any other data source:
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

Another use case for row access policies is to ensure that your embedded dashboard provides a specific view for your end users. During the [embed dashboard request](/integrate/embedding), you can pass custom attributes (other than the ones provided out-of-the-box) that map directly to a value within your Rill explore dashboard.

```yaml
security:
  access: true
  row_filter: >
        dimension_1 = '{{ .user.custom_variable_1 }}' AND
        dimension_2 = '{{ .user.custom_variable_2 }}' 
```

In order to test the view of your embedded dashboard, you can add the same custom variables to [your mock users](#testing-policies-in-rill-developer) as seen below:
```yaml
- email: embed@rilldata.com
  name: embed
  custom_variable_1: Value_1
  custom_variable_2: Value_2
```


### Advanced Example: Access to Dashboard in Rill and Embedded

While not common, there are use cases where a dashboard is used both in the Rill Cloud UI and as an embedded dashboard. In this case, passing a similar user attribute could suffice, but if you need to pass a custom attribute, you'll need to add an extra layer of logic to your dashboard.

```yaml
security:
  access: "{{ .user.admin }} OR '{{ .user.domain }}' == 'example.com' {{- if .user.custom_variable_1 }} OR true {{- end }}"
```