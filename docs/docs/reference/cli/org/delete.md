---
note: GENERATED. DO NOT EDIT.
title: rill org delete
---
## rill org delete

Delete organization

### Synopsis

Delete an organization and all its associated projects.
This operation cannot be undone. Use --force to skip confirmation.

```
rill org delete [<org-name>] [flags]
```

### Examples

```
  rill org delete myorg
  rill org delete myorg --force
```

### Flags

```
      --force   Delete forcefully, skips the confirmation
```

### Global flags

```
      --api-token string   Token for authenticating with the cloud API
      --format string      Output format (options: "human", "json", "csv") (default "human")
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters (default true)
```

### SEE ALSO

* [rill org](org.md)	 - Manage organisations

