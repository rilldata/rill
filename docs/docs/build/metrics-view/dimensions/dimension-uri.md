---
title: "Clickable Dimension Links"
description: "Use lookup dimensions to enrich your metrics view data with reference information at query time"
sidebar_label: "Clickable Dimension Links"
sidebar_position: 50
---

You can make dimension values clickable by adding a `uri` parameter to your dimension configuration. This enables users to click directly on dimension values in the dashboard to navigate to external URLs, making your dashboards more interactive and useful for data exploration.



 <img src = '/img/build/dashboard/clickable-dimension.png' class='rounded-gif' />
<br />

The simplest set up is to set the `uri` parameter to `true` if your column is already a URL. 

```yaml
dimensions:
  - display_name: Company Url
    column: Company URL
    uri: true 
```

For more advanced use cases, you can use an expression to create the URI to click. There are two ways to do this depending if you want the value of the dimension to show the URL string or the original value.

### Keep the original value
The reason this stays as the value is that the `column` is defined as the column value and URI is dynamically generated.
```yaml
dimensions:
  - name: profile_url
    display_name: Bluesky Profile Link
    column: profile_id
    uri: CONCAT('https://bsky.app/profile/',profile_id)
```

### Show the URL string
In this case, we use expression to change the value of the displayed value and use this URL as the clickable link.
```yaml
  - name: profile_url
    display_name: Bluesky Profile Link
    expression: CONCAT('https://bsky.app/profile/',profile_id)
    uri: true
```