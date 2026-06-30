---
note: GENERATED. DO NOT EDIT.
title: rill project skip-partition
---
## rill project skip-partition

Skip partitions for a model

### Synopsis

Mark partitions as skipped so they are excluded from execution and from the model's error state. Skipped partitions remain skipped until they are explicitly triggered (e.g. via 'rill project refresh --partition').

```
rill project skip-partition [<project>] <model> [flags]
```

### Flags

```
      --project string      Project Name
      --path string         Project directory (default ".")
      --branch string       Target deployment by Git branch (default: primary deployment)
      --model string        Model Name
      --partition strings   Skip specific partitions by key
      --pending             Skip all pending partitions
      --errored             Skip all errored partitions
      --local               Target locally running Rill
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

