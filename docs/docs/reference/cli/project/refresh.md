---
note: GENERATED. DO NOT EDIT.
title: rill project refresh
---
## rill project refresh

Refresh one or more resources

```
rill project refresh [<project-name>] [flags]
```

### Flags

```
      --project string         Project name
      --path string            Project directory (default ".")
      --local                  Target locally running Rill
      --all                    Refresh all resources except alerts and reports (default)
      --full                   Fully reload the targeted models (use with --all or --model)
      --model strings          Refresh a model
      --partition strings      Refresh a model partition (must set --model)
      --errored-partitions     Refresh all model partitions with errors (must set --model)
      --source strings         Refresh a source
      --metrics-view strings   Refresh a metrics view
      --alert strings          Refresh an alert
      --report strings         Refresh a report
      --connector strings      Re-validate a connector
      --parser                 Refresh the parser (forces a pull from Github)
```

### Global flags

```
      --api-token string   Token for authenticating with the cloud API
      --format string      Output format (options: "human", "json", "csv") (default "human")
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters (default true)
      --org string         Organization Name
```

### SEE ALSO

* [rill project](project.md)	 - Manage projects

