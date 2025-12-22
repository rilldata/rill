---
note: GENERATED. DO NOT EDIT.
title: rill usergroup grant-resource
---
## rill usergroup grant-resource

Grant a user group access to specific project resources (viewer scoped)

```
rill usergroup grant-resource [flags]
```

### Flags

```
      --group string           User group (required)
      --org string             Organization
      --project string         Project (required)
      --resource stringArray   Resource to grant in the format kind/name (repeat for multiple)
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

