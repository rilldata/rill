---
title: Project execution
description: How Rill parses and executes project resources
sidebar_label: Project execution
sidebar_position: 10
---

# Project Execution

When you save a file in a Rill project, it flows through a pipeline of components:

```
  Files  →  Watcher  →  Parser  →  Catalog  →  Controller  →  Reconciler
                           ↓           ↓                          ↓
                     Parse errors   DAG order              Reconcile errors
```

1. **Watcher**: Detects file changes in the project directory.
2. **Parser**: Converts files into internal resource definitions and organizes them into a dependency graph (DAG).
3. **Catalog**: Persistent store holding the declared "spec" and current "state" of each resource.
4. **Controller**: Watches the catalog and triggers reconciliation for changed resources in DAG order.
5. **Reconciler**: Executes actions to make each resource's current state match its declared spec.

## Resource Lifecycle

Each resource type has its own reconciler containing domain-specific logic. For example, the `model` reconciler runs SQL queries and tracks ingestion state, while the `metrics_view` reconciler validates dimensions and measures.

Some reconcilers are always cheap; others can trigger slow or costly operations. The `model` reconciler is notably expensive when it:
- Re-ingests data from an external database
- Re-creates tables with complex SQL transformations

Reconcilers are designed to run fast when no action is needed—if a model already exists, its reconciler returns quickly. This allows Rill to trigger reconciliation liberally. Reconcilers run when:
- A parent resource in the DAG finishes reconciling
- The runtime restarts
- `rill.yaml` or environment variables change
- A scheduled time is reached (to honor `cron` refresh expressions)

## Internal Resources

Some resources exist without corresponding files:
- `project_parser`: Global resource that handles file parsing and stores parse errors.
- `alert` and `report`: Can be created via the Rill Cloud UI instead of files.
- `component`: Created internally as part of a `canvas` resource.
- `explore`: Can optionally be created as part of a `metrics_view` resource.
