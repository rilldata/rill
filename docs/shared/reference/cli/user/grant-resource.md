---
note: GENERATED. DO NOT EDIT.
title: rill user grant-resource
---
## rill user grant-resource

Grant a user access to specific project resources (viewer scoped)

```
rill user grant-resource [flags]
```

### Flags

```
      --email string           Email of the user (required)
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

* [rill user](user.md)	 - Manage users

