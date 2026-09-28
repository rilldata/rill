---
note: GENERATED. DO NOT EDIT.
title: rill user add
---
## rill user add

Add user to a project, organization or group

### Synopsis

Add user to a project, organization or group.

When --group is given together with --project, or with --role and no --project, the user is added to (or invited to)
the project or organization and its user groups in a single step. A user who has not signed up yet joins the groups when
they accept the invitation. When only --group is given, the user is added to the groups directly.

```
rill user add [flags]
```

### Flags

```
      --attribute stringToString   Custom attributes in key=value format (--attribute app=foo --attribute dept=bar) (default [])
      --canvas stringArray         Canvas resource to restrict to (repeat for multiple)
      --email string               Email of the user
      --explore stringArray        Explore resource to restrict to (repeat for multiple)
      --group stringArray          User group to add the user to (repeat for multiple)
      --json string                Custom attributes as JSON object (--json '{"app":"foo","dept":"bar"}')
      --org string                 Organization
      --project string             Project
      --restrict-resources         Restrict the user to the provided resources (defaults to true when resources are provided)
      --role string                Role of the user (options: admin, editor, viewer, guest)
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

