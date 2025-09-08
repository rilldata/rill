---
title: "Clickable Dimension Links"
description: "Use lookup dimensions to enrich your metrics view data with reference information at query time"
sidebar_label: "Clickable Dimension Links"
sidebar_position: 50
---

You can make dimension values clickable by adding a `uri` parameter to your dimension configuration. This enables users to click directly on dimension values in the dashboard to navigate to external URLs, making your dashboards more interactive and useful for data exploration.

 <img src = '/img/build/dashboard/clickable-dimension.png' class='rounded-gif' />
<br />

```yaml
dimensions:
  - display_name: Company Url
    column: Company URL
    uri: true # if already set to the URL, also accepts SQL expressions
```

For more advanced use cases, you can use an `expression` to dynamically generate URLs based on your data.

```yaml
dimensions:
  - display_name: Bluesky Profile Link
    expression: profile_id
    uri: CONCAT('https://bsky.app/profile/',profile_id)
```
