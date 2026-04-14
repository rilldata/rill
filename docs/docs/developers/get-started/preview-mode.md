---
title: "Preview Mode"
sidebar_label: "Preview Mode"
sidebar_position: 20
---

Preview mode starts Rill Developer with a dashboard-only interface — no file editor, no code. It's designed for sharing a local Rill instance with stakeholders who don't need the development environment.

```bash
rill start my-project --preview
```

## What's available in preview mode

| Available | Hidden |
|---|---|
| Explore dashboards | File editor |
| Canvas dashboards | Connector configuration |
| AI Chat | Resource graph |
| Project status | |

Preview mode also sets the application to **read-only**, so dashboards cannot be modified through the UI.

## Switching modes

You can switch between Developer and Preview modes from within the app using the mode toggle in the top-left header. You don't need to restart Rill to change modes.

- Navigating to a developer route (like the file editor) automatically switches to Developer mode
- Navigating to a dashboard route from preview keeps you in Preview mode

## When to use preview mode

- **Demos and presentations** — show dashboards without exposing project internals
- **Stakeholder access** — let non-technical users explore dashboards locally without the development UI
- **Pre-deploy review** — see how dashboards look in a production-like view before deploying to [Rill Cloud](/developers/deploy/deploy-dashboard)

:::tip
For a full team deployment with authentication and access controls, [deploy to Rill Cloud](/developers/deploy/deploy-dashboard).
:::
