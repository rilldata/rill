---
note: GENERATED. DO NOT EDIT.
title: rill project deployments create
---
## rill project deployments create

Create a deployment for a specific branch

```
rill project deployments create [<project>] <branch> [flags]
```

### Flags

```
      --editable             Make the deployment editable (changes are persisted back to git repo)
      --environment string   Optional environment to create for (options: dev, prod) (default "dev")
      --path string          Project directory (default ".")
      --project string       Project name
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

* [rill project deployments](deployments.md)	 - Manage project deployments

