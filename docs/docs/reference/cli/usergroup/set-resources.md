---
note: GENERATED. DO NOT EDIT.
title: rill usergroup set-resources
---
## rill usergroup set-resources

Set a user group's project resources and restriction flag (overwrites existing list)

```
rill usergroup set-resources [flags]
```

### Flags

```
      --canvas stringArray    Canvas resource to restrict to (repeat for multiple)
      --explore stringArray   Explore resource to restrict to (repeat for multiple)
      --group string          User group (required)
      --org string            Organization
      --project string        Project (required)
      --restrict-resources    Whether to restrict the group to the provided resources (defaults to true when resources are provided)
```

### Global flags

```
      --api-token string   Token for authenticating with the cloud API
      --format string      Output format (options: "human", "json", "csv") (default "human")
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters (default true)
```

### SEE ALSO

* [rill usergroup](usergroup.md)	 - Manage user groups

