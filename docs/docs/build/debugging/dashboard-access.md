---
title: "Debugging Dashboard Access"
description: "Troubleshooting dashboard access and data control issues in Rill"
sidebar_label: "Debugging Dashboard Access"
sidebar_position: 00
---

# Debugging Dashboard Access

Dashboard access and data control are fundamental aspects of Rill, offering multiple configuration options for securing your dashboards. This complexity can sometimes lead to access issues that require troubleshooting.

This guide provides essential troubleshooting steps and solutions to quickly resolve dashboard access problems and restore dashboard functionality.

:::note 

This guide assumes you have already set up and understand [dashboard and data access policies](/build/metrics-view/security). 

:::

## Where to Start

There are three main types of data access issues you might encounter in Rill:

1. [**403 Permission Denied**](#troubleshooting-general-access-issues): Dashboard not shown in the project
2. [**Failed to Load Dashboard**](/build/debugging/dashboard-access#troubleshooting-row-access-filters): "This dashboard currently has no data to display. This may be due to access permissions."
3. [**Canvas Dashboard Component Failed to Load**](#canvas-dashboards)

The troubleshooting approach depends on which type of error you're experiencing.

:::tip Not an Admin or Editor?

If you don't have admin or editor access, or access to the underlying GitHub repository, you'll need to contact your internal team to resolve the issue.

:::

## Where Are Policies Defined?

In Rill, policies are defined at two levels:
- **Project defaults** in `rill.yaml`
- **Object-level policies** in individual resource files

For detailed information about policy configuration, refer to our [data access policy documentation](/build/metrics-view/security#creating-access-policies).

Once you've identified where these policies are defined, you can begin troubleshooting the specific issue.


## Secutiry Policies Visualized

```mermaid
flowchart TD
    subgraph MV_side[Metrics View Side]
        A1{MV defined?}
        A2[Use MV policy]
        A3{PDMV defined?}
        A4[Use PDMV policy]
        A5[Use open-to-org default]
    end

    subgraph DB_side[Dashboard Side]
        B1{DB defined?}
        B2[Use DB policy]
        B3{PDD defined?}
        B4[Use PDD policy]
        B5[Use open-to-org default]
    end

    A1 -- Yes --> A2
    A1 -- No --> A3
    A3 -- Yes --> A4
    A3 -- No --> A5

    B1 -- Yes --> B2
    B1 -- No --> B3
    B3 -- Yes --> B4
    B3 -- No --> B5

    %% Final Access Combination
    A2 --> LHS[MV-effective]
    A4 --> LHS
    A5 --> LHS

    B2 --> RHS[DB-effective]
    B4 --> RHS
    B5 --> RHS

    LHS --> AND_OP{{AND}}
    RHS --> AND_OP

    AND_OP --> FINAL[MV-effective AND DB-effective]
    
    %% Style definitions
    classDef mv fill:#cce5ff,stroke:#004085,stroke-width:1px;
    classDef db fill:#d4edda,stroke:#155724,stroke-width:1px;
    classDef andnode fill:#fff,stroke:#856404,stroke-width:2px;
    classDef final fill:#e2d6f3,stroke:#4b0082,stroke-width:2px,font-weight:bold;

    %% Assign styles
    class A1,A2,A3,A4,A5,LHS mv;
    class B1,B2,B3,B4,B5,RHS db;
    class AND_OP andnode;
    class FINAL final;

    %% Subgraph (cluster) styles
    style MV_side fill:#fff,stroke:#004085,stroke-width:2px;
    style DB_side fill:#fff,stroke:#155724,stroke-width:2px;
```

**Abbreviations:**
- **MV**: Metrics View Policy
- **PDMV**: Project Default Metrics View Policy  
- **DB**: Dashboard Policy
- **PDD**: Project Default Dashboard Policy

For a full table of possible combinations, see [tables of examples](/build/debugging/dashboard-access#table-of-examples).

## Troubleshooting General _Access_ Issues

**How it's surfaced**: The dashboard doesn't appear in the project view, or a component isn't showing in the metrics view.

Don't forget the behavior for security policies:

```bash
Project Defaults < (Metrics View YAML AND Dashboard YAML)
```

**Check these levels in order:**

1. **Project defaults**: Are there project-level access policies? [Yes/No]
2. **Metrics view policies**: Are there object-level metrics view policies? [Yes/No] (This overrides project defaults)
3. **Dashboard policies**: Are there object-level dashboard policies? [Yes/No] (This overrides project defaults)
   
:::tip No policies in object level?
Even if you have no policies defined in the metrics view, any policies defined in the project level will get added dynamically. Don't forget to add the project defaults in your calculations! See our [table of examples](#table-of-examples) for all the possible combinations.

:::

**Example policy structure:**

```sql
('{{ .user.admin }}' OR '{{ .user.domain }}' == 'example.com') AND ('{{ has "partners" .user.groups }}')
```

This policy would only allow:
- Admins who are part of the "partners" group, OR
- Users with "example.com" domain who are part of the "partners" group

:::tip Checking User Groups

You can verify your user group membership through:
- **UI**: https://ui.rilldata.com/your_org/-/users/groups
- **CLI**:
```bash
rill usergroup list
rill user list --group group_name
```
:::

### Canvas Dashboards

Canvas dashboards reference multiple metrics views. If you're missing components from specific metrics views, check the `access` parameter for each individual metrics view.

<img src="/img/build/debugging/canvas-metrics-false.png" class="rounded-gif" alt="Canvas metrics access example" /> <br/>


## Troubleshooting _Row Filters_ Issues

**How it's surfaced**: The data displayed in the dashboard is not being filtered properly.

Don't forget the behavior for security policies:

```bash
Project Defaults < (Metrics View YAML)
```

**Check these levels in order:**

1. **Project defaults**: Are there project-level access policies? [Yes/No] 
2. **Metrics view policies**: Are there object-level metrics view policies? [Yes/No] (This overrides project defaults)

Note that dashboards don't apply here since `row_filter` is set on metrics views only.

:::tip No policies in object level?
Even if you have no policies defined in the metrics view, any policies defined in the project level will get added dynamically. Don't forget to add the project defaults in your calculations! See our [table of examples](#table-of-examples) for all the possible combinations.

:::



## Table of Examples

<div class="rounded-gif">

| Project Default Metrics View                  | Project Default Dashboard                         | Metrics View                                                                       | Dashboard                                                                         | Resulting                                                                                                                  |
| --------------------------------------------- | ------------------------------------------------- | ---------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| _not defined_                                 | _not defined_                                     | _not defined_                                                                      | _not defined_                                                                     | **All dashboards open** to org users                                                                                       |
| _not defined_                                 | _not defined_                                     | _not defined_                                                                      | `access: '{{ has "partners" .user.groups }}'`                                     | Dashboard accessible to users in **usergroup**, "partners"                                                                 |
| _not defined_                                 | _not defined_                                     | `access: '{{ has "partners" .user.groups }}'`                                      | _not defined_                                                                     | Dashboard accessible to users in **usergroup**, "partners"                                                                 |
| _not defined_                                 | _not defined_                                     | `access: '{{ has "partners" .user.groups }}'`                                      | `access: "'{{ .user.domain }}' == 'example.com'"`                                 | Dashboard accessible to users in **usergroup**, "partners" AND **domain** ending in "example.com"                          |
| _not defined_                                 | `access: false`                                   | _not defined_                                                                      | _not defined_                                                                     | No Dashboards are accessible _(From Project Default)_                                                                      |
| _not defined_                                 | ~~`access: false`~~                               | _not defined_                                                                      | `access: "'{{ .user.domain }}' == 'example.com'"`(**overwrites** project default) | Dashboard accessible to user **domain** ending in "example.com"                                                            |
| _not defined_                                 | `access: '{{ has "partners" .user.groups }}'`     | `access: "'{{ .user.domain }}' == 'example.com'"`                                  | _not defined_                                                                     | Dashboard accessible to users in **usergroup**, "partners" _(From Project Default)_ AND **domain** ending in "example.com" |
| _not defined_                                 | ~~`access: '{{ has "partners" .user.groups }}'`~~ | `access: "'{{ .user.domain }}' == 'domain.com'"`                                   | `access: '{{ has "external" .user.groups }}'` (**overwrites** project default)    | Dashboard accessible to users in **usergroup**, "external" AND **domain** ending in "domain.com"                           |
| `access: false`                               | _not defined_                                     | _not defined_                                                                      | _not defined_                                                                     | No Dashboards are accessible _(From Project Default)_                                                                      |
| `access: '{{ has "partners" .user.groups }}'` | _not defined_                                     | _not defined_                                                                      | `access: "'{{ .user.domain }}' == 'example.com'"`                                 | Dashboard accessible to users in **usergroup**, "partners" _(From Project Default)_ AND **domain** ending in "example.com" |
| ~~`access: false`~~                           | _not defined_                                     | `access: "'{{ .user.domain }}' == 'example.com'"` (**overwrites** project default) | _not defined_                                                                     | Dashboard accessible to user **domain** ending in "example.com"                                                            |
| ~~`access: false`~~                           | _not defined_                                     | `access: "'{{ .user.domain }}' == 'example.com'"` (**overwrites** project default) | `access: true`                                                                    | Dashboard accessible to user **domain** ending in "example.com"                                                            |
| `access: false`                               | `access: false`                                   | _not defined_                                                                      | _not defined_                                                                     | No Dashboards are accessible _(From Project Default)_                                                                      |
| `access: false`                               | ~~`access: false`~~                               | _not defined_                                                                      | `access: "'{{ .user.domain }}' == 'example.com'"` (**overwrite** project default) | No Dashboards are accessible _(From Project Default)_                                                                      |
| ~~`access: false`~~                           | `access: true`                                    | `access: "'{{ .user.domain }}' == 'example.com'"` (**overwrites** project default) | _not defined_                                                                     | Dashboard accessible to user **domain** ending in "example.com"                                                            |
| ~~`access: false`~~                           | ~~`access: false`~~                               | `access: "'{{ .user.domain }}' == 'example.com'"` (**overwrites** project default) | `access: '{{ has "partners" .user.groups }}'`(**overwrites** project default)     | Dashboard accessible to users in **usergroup**, "partners" AND **domain** ending in "example.com"                          |
</div>