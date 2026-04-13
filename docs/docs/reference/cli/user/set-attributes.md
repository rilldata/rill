---
note: GENERATED. DO NOT EDIT.
title: rill user set-attributes
---
## rill user set-attributes

Set custom attributes for a user

```
rill user set-attributes [flags]
```

### Flags

```
      --attribute stringToString   Attributes in key=value format (--attribute app=foo --attribute dept=bar) (default [])
      --email string               Email of the user (required)
      --force                      Skip confirmation prompt when overwriting existing attributes
      --json string                Attributes as JSON object (--json '{"app":"foo","dept":"bar"}')
      --org string                 Organization
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

