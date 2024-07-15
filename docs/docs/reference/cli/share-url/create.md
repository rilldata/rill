---
note: GENERATED. DO NOT EDIT.
title: rill share-url create
---
## rill share-url create

Create a shareable URL

```
rill share-url create [<project-name>] <metrics view> [flags]
```

### Flags

```
      --project string    Project name
      --path string       Project directory (default ".")
      --ttl-minutes int   Duration until the token expires (use 0 for no expiry)
      --filter string     Limit access to the provided filter (json)
      --fields strings    Limit access to the provided fields
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

* [rill share-url](share-url.md)	 - Manage shareable URLs

