---
title: rill project reconcile (deprecated)
---
## rill project reconcile

Send trigger to deployment

```
rill project reconcile [<project-name>] [flags]
```

:::warning Deprecation Notice 

This is a **legacy** command that's been preserved for backwards compatibility. *Starting in Rill v0.37.0 and newer*, we recommend using the [rill project refresh](refresh.md) and [rill project reset](reset.md) commands to refresh and reset/redeploy your projects respectively.

:::

### Flags

```
      --project string           Project name
      --path string              Project directory (default ".")
      --refresh                  Refresh all sources
      --refresh-source strings   Refresh specific source(s)
      --reset                    Reset and redeploy the project from scratch
```

### Global flags

```
      --api-token string   Token for authenticating with the admin API
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters (default true)
      --org string         Organization Name
```

### SEE ALSO

* [rill project](project.md)	       - Manage projects
* [rill project refresh](refresh.md) - Refresh your project
