---
note: GENERATED. DO NOT EDIT.
title: rill project deploy
---
## rill project deploy

Deploy project to Rill Cloud by uploading the project files

```
rill project deploy [<path>] [flags]
```

### Flags

```
      --path string           Path to project repository (default: current directory) (default ".")
      --subpath string        Relative path to project in the repository (for monorepos)
      --org string            Org to deploy project in
      --project string        Project name (default: Git repo name)
      --description string    Project description
      --public                Make dashboards publicly accessible
      --provisioner string    Project provisioner
      --prod-version string   Rill version (default: the latest release version) (default "latest")
      --prod-branch string    Git branch to deploy from (default: the default Git branch)
```

### Global flags

```
      --api-token string   Token for authenticating with the cloud API
      --format string      Output format (options: "human", "json", "csv") (default "human")
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters (default true)
```

### SEE ALSO

* [rill project](project.md)	 - Manage projects

