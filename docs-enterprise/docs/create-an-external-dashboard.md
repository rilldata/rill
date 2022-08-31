---
title: "Create External Dashboards"
slug: "create-an-external-dashboard"
excerpt: "Use a saved \"Parent\" Dashboard to create a \"Child\" based on predetermined filters"
hidden: false
createdAt: "2021-08-11T20:06:30.979Z"
updatedAt: "2022-07-13T07:13:26.466Z"
---
To create dashboards for a set of stakeholders, external users or other partners, start by creating a Parent dashboard view that contains all of the dimensions and metrics you wish to display. That parent dashboard can then be filtered by a set of criteria that then limits the data available to each user you wish to grant access. Further, that child dashboard can be embedded to be a specific application view within your product/portal for that set of users.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/cc0ee04-Child_Dash.png",
        "Child Dash.png",
        1560,
        808,
        "#d2d9dd"
      ],
      "sizing": "80"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "Create a Parent Dashboard"
}
[/block]
To create a parent dashboard, a small edit at the end of any dashboard configuration adds the ability to inherit the dashboard set-up along with prompt for filter criteria.

  * thirdPartySubsets = object to define the Child dash
  * datasource = datasource(s) from your current parent dashboard you want to make available to child dashboard
  * dimension = dimension (criteria) that is required to be inputted before creating the child dashboard 
[block:code]
{
  "codes": [
    {
      "code": "  \"thirdPartySubsets\": [\n    {\n      \"dataSource\": \"dash-metrics-datasource\",\n      \"dimension\": [\n        \"context.view.druidDataSource\"\n      ]\n    }\n  ]",
      "language": "json"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "Create a Child from a Parent Dashboard"
}
[/block]
In the [Admin view > Dashboards](https://docs.rilldata.com/docs/explore-admin), select your parent dashboard you wish to create a Child from
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/0c4c937-Parent_Dash.png",
        "Parent Dash.png",
        735,
        477,
        "#e9ecef"
      ],
      "sizing": "80"
    }
  ]
}
[/block]
Selecting Create a child dashboard will bring up the screen to create your dashboard. 

On that screen, you will be able to enter:

  * Dashboard name
  * Dashboard URL slug
  * Create a Security Policy (optional)
  * Add yourself to the policy (optional)
  * Enter required criteria for the Child to inherit and filter 
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/3806f1e-Child_Dash_Create.png",
        "Child Dash Create.png",
        950,
        1034,
        "#f7f9fa"
      ],
      "sizing": "80"
    }
  ]
}
[/block]

[block:callout]
{
  "type": "info",
  "title": "Add Additional Users to Child Dashboards",
  "body": "To add more users to a given dashboard, you can add the new dashboard to an existing security policy or [add users to the newly created policy](https://enterprise.rilldata.com/docs/admin-security)."
}
[/block]