---
note: GENERATED. DO NOT EDIT.
title: rill service create
---
## rill service create

Create service

```
rill service create <service-name> [flags]
```

### Flags

```
      --attributes string     JSON object of key-value pairs for service attributes
      --org-role string       Organization role to assign to the service (admin, editor, viewer, guest)
      --project string        Project to assign the role to (required if project-role is set)
      --project-role string   Project role to assign to the service (admin, editor, viewer)
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

* [rill service](service.md)	 - Manage service accounts

