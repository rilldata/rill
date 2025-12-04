---
title: "Clickable Dimension Links"
description: "Make dimension values clickable by adding URI parameters to create interactive links in your dashboards"
sidebar_label: "Clickable Dimension Links"
sidebar_position: 50
---

You can make dimension values clickable by adding a `uri` parameter to your dimension configuration. This enables users to click directly on dimension values in the dashboard to navigate to external URLs, making your dashboards more interactive and useful for data exploration.

<img src='/img/build/dashboard/clickable-dimension.png' class='rounded-gif' />
<br />

## Simple Setup: Column Already Contains URLs

If your column values are already URLs, simply add `uri: true` to the dimension:

```yaml
dimensions:
  - display_name: Company URL
    column: company_url
    uri: true 
```

## Advanced Setup: Dynamic URL Generation

For more advanced use cases, you can dynamically create URLs using expressions. The dimension displays the generated URL and uses it as the clickable link:

```yaml
dimensions:
  - name: profile_url
    display_name: Bluesky Profile Link
    expression: CONCAT('https://bsky.app/profile/', profile_id)
    uri: true
```