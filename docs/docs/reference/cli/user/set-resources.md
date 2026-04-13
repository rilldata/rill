---
note: GENERATED. DO NOT EDIT.
title: rill user set-resources
---
## rill user set-resources

Set a user's project resources and restriction flag (overwrites existing list)

```
rill user set-resources [flags]
```

### Flags

```
      --canvas stringArray    Canvas Resource to set (repeat for multiple)
      --email string          Email of the user (required)
      --explore stringArray   Explore Resource to set (repeat for multiple)
      --org string            Organization
      --project string        Project (required)
      --restrict-resources    Whether to restrict the user to the provided resources (defaults to true when resources are provided)
```

### Global flags

```
      --api-token string   Token for authenticating with the cloud API
      --format string      Output format (options: "human", "json", "csv") (default "human")
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters (default true)
```

### SEE ALSO

* [rill user](user.md)	 - Manage users

