---
note: GENERATED. DO NOT EDIT.
title: rill project partitions skip
---
## rill project partitions skip

Skip partitions for a model

### Synopsis

Mark partitions as skipped so they are excluded from execution and from the model's error state. Skipped partitions remain skipped until they are explicitly triggered (e.g. via 'rill project refresh --partition').

```
rill project partitions skip [<project>] <model> [flags]
```

### Flags

```
      --project string      Project Name
      --path string         Project directory (default ".")
      --branch string       Target deployment by Git branch (default: primary deployment)
      --model string        Model Name
      --partition strings   Skip specific partitions by key
      --all                 Skip all pending partitions
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

* [rill project partitions](partitions.md)	 - List partitions for a model

