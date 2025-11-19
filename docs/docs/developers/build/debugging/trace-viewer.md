---
title: "Trace Viewer in Rill Developer"
description: Alter dashboard look and feel
sidebar_label: "Trace Viewer in Rill Developer"
sidebar_position: 00
---

Rill Developer provides a built-in trace viewer, enabling users to visually inspect operations performed when reconciling resources or fetching data for dashboards. This helps in diagnosing performance and operational issues.

<img src = '/img/build/debugging/trace-viewer-overview.png' class='rounded-gif' />

## How to Use the Trace Viewer

### Step 1: Start Rill Developer with Debug Mode

Launch Rill Developer with the debug flag enabled:

```bash
rill start --debug
```

### Step 2: Access the Trace Viewer

Open your web browser and navigate to:

```
http://localhost:9009/traces
```

### Step 3: Visualize Traces

There are two ways to visualize traces:

- **By Trace ID:** Enter a specific Trace ID to inspect a particular operation.
- **By Resource Reconciliation:** Browse operations associated with reconciling specific resources.

For example, to view the operations performed during the reconciliation of the `bids` model, enter `bids` in the resource name field.

## Understanding the Trace Graph

- **Horizontal Bars:** Each horizontal bar represents a distinct operation performed by Rill.
- **Bar Length:** The length of each bar indicates the duration of the operation.
- **Operation Details:** Click on any bar to view detailed tags and metadata in the pane on the right.
- **Nested Operations:** Operations may trigger other sub-operations, displayed as nested bars beneath the parent operation.
- **Parallel Operations:** Bars displayed side by side indicate operations executed concurrently.

## Retrieving the Trace ID for Dashboard Data Fetch

To find the Trace ID associated with fetching data for a dashboard:

1. **Open Browser Developer Tools:** Use your browser's developer tools (usually `F12` or `Cmd+Option+I`).
2. **Inspect API Requests:** Select the API call you're interested in from the network tab.
3. **Find `X-trace-id`:** Check the response headers to locate the `X-trace-id`.
4. **Use Trace ID:** Enter this Trace ID into the trace viewer to inspect the operation details.

<img src = '/img/build/debugging/capture-trace-id.png' class='rounded-gif' />

## Inner Workings

- The Trace Viewer UI uses OpenTelemetry (OTEL) traces.
- Traces are captured in JSON format and stored locally.
- The UI retrieves and filters traces based on query parameters and trace type.
