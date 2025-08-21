---
title: "Clickable Dimension Links"
description: "Use lookup dimensions to enrich your metrics view data with reference information at query time"
sidebar_label: "Clickable Dimension Links"
sidebar_position: 50
---

Adding an additional parmater to the dimension allows you to click directly on a link in the dimension leaderboard to navigate to the displayed value.

 <img src = '/img/build/dashboard/clickable-dimension.png' class='rounded-gif' />
<br />

```yaml
dimensions:
  - label: Company Url
    column: Company URL
    uri: true #if already set to the URL, also accepts SQL expressions
```

A bit more complex example is using an `expression` and create the URL dynamically.

```yaml
dimensions:
  - label: Company Url
    expression: Company URL
    uri: true #if already set to the URL, also accepts SQL expressions
```
