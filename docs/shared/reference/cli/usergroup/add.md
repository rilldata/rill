---
note: GENERATED. DO NOT EDIT.
title: rill usergroup add
---
## rill usergroup add

Add a group to a project or organization

```
rill usergroup add [flags]
```

### Flags

```
      --canvas stringArray    Canvas resource to restrict to (repeat for multiple)
      --explore stringArray   Explore resource to restrict to (repeat for multiple)
      --group string          User group
      --org string            Organization
      --project string        Project
      --restrict-resources    Restrict the user group to provided resources (defaults to true when resources are provided)
      --role string           Role of the user group (options: admin, editor, viewer)
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

