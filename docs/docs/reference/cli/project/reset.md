---
note: GENERATED. DO NOT EDIT.
title: rill project reset
---
## rill project reset

Re-deploy project

### Synopsis

Create a new deployment for the project (and tear down the current one)

```
rill project reset [<project-name>] [flags]
```

### Flags

```
      --project string   Project name
      --path string      Project directory (default ".")
      --force            Force reset even if project is already deployed
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

