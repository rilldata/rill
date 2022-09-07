---
title: "Create External Dashboards"
slug: "create-an-external-dashboard"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text='Use a saved "Parent" Dashboard to create a "Child" based on predetermined filters' />

To create dashboards for a set of stakeholders, external users or other partners, start by creating a Parent dashboard view that contains all of the dimensions and metrics you wish to display. That parent dashboard can then be filtered by a set of criteria that then limits the data available to each user you wish to grant access. Further, that child dashboard can be embedded to be a specific application view within your product/portal for that set of users.
![](https://images.contentful.com/ve6smfzbifwz/2LCl2vzrcK3mguqOmEqBdD/35756cfa9dcfe9e91f402ec31d6d6879/cc0ee04-Child_Dash.png)

## Create a Parent Dashboard

To create a parent dashboard, a small edit at the end of any dashboard configuration adds the ability to inherit the dashboard set-up along with prompt for filter criteria.
  * thirdPartySubsets = object to define the Child dash
  * datasource = datasource(s) from your current parent dashboard you want to make available to child dashboard
  * dimension = dimension (criteria) that is required to be inputted before creating the child dashboard 

```json
  "thirdPartySubsets": [
    {
      "dataSource": "dash-metrics-datasource",
      "dimension": [
        "context.view.druidDataSource"
      ]
    }
  ]
```

## Create a Child from a Parent Dashboard

In the [Admin view > Dashboards](/explore-admin), select your parent dashboard you wish to create a Child from
![](https://images.contentful.com/ve6smfzbifwz/3aT5HhfECIpwT1ij9Ygzc0/432cac5f7bcc505e84d0434b188ceb77/0c4c937-Parent_Dash.png)
Selecting Create a child dashboard will bring up the screen to create your dashboard. 

On that screen, you will be able to enter:

  * Dashboard name
  * Dashboard URL slug
  * Create a Security Policy (optional)
  * Add yourself to the policy (optional)
  * Enter required criteria for the Child to inherit and filter 
![](https://images.contentful.com/ve6smfzbifwz/31Ik5EfYRsOGOUgvQocxS7/0ec38bff22826990dc89189fa4d06f67/3806f1e-Child_Dash_Create.png)

:::info Add Additional Users to Child Dashboards
To add more users to a given dashboard, you can add the new dashboard to an existing security policy or [add users to the newly created policy](/admin-security).
:::