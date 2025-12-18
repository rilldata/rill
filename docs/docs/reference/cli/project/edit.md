---
note: GENERATED. DO NOT EDIT.
title: rill project edit
---
## rill project edit

Edit the project details

```
rill project edit [<project-name>] [flags]
```

### Flags

```
      --project string         Project Name
      --description string     Project Description
      --prod-branch string     Production branch name
      --public                 Make dashboards publicly accessible
      --path string            Project directory (default ".")
      --remote-url string      Github remote URL
      --subpath string         Relative path to project in the repository (for monorepos)
      --provisioner string     Project provisioner (default: current provisioner)
      --prod-ttl-seconds int   Time-to-live in seconds for production deployment (0 means no expiration)
      --prod-version string    Specify the Rill version for production deployment (default: current version)
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

* [rill project](project.md)	 - Manage projects

