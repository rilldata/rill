---
note: GENERATED. DO NOT EDIT.
title: rill usergroup revoke-resource
---
## rill usergroup revoke-resource

Remove resource-level access previously granted to a user group

```
rill usergroup revoke-resource [flags]
```

### Flags

```
      --group string           User group (required)
      --org string             Organization
      --project string         Project (required)
      --resource stringArray   Resource to revoke in the format kind/name (repeat for multiple)
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

