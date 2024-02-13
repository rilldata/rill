---
note: GENERATED. DO NOT EDIT.
title: rill project logs
---
## rill project logs

Show project logs

```
rill project logs [<project-name>] [flags]
```

### Flags

```
      --project string   Project Name
      --path string      Project directory (default ".")
  -f, --follow           Follow logs
  -t, --tail int         Number of lines to show from the end of the logs, use -1 for all logs (default -1)
      --level string     Minimum log level to show (DEBUG, INFO, WARN, ERROR, FATAL) (default "INFO")
```

### Global flags

```
      --api-token string   Token for authenticating with the admin API
      --format string      Output format (options: "human", "json", "csv") (default "human")
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters (default true)
      --org string         Organization Name
```

### SEE ALSO

* [rill project](project.md)	 - Manage projects

