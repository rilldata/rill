## rill deploy

Deploy project to Rill Cloud

```
rill deploy [flags]
```

### Flags

```
      --path string             Project directory (default ".")
      --org string              Org to deploy project (default: default org)
      --prod-slots int          Slots to allocate for production deployments (default 2)
      --description string      Project description
      --region string           Deployment region
      --prod-db-driver string   Database driver (default "duckdb")
      --prod-db-dsn string      Database driver configuration
      --public                  Make dashboards publicly accessible
      --prod-branch string      Git branch to deploy from (default: the default Git branch)
      --project string          Project name (default: Git repo name)
      --remote string           Remote name (defaults: first github remote)
      --api-token string        Token for authenticating with the admin API
      --api-url string          Base URL for the admin API (default "https://admin.rilldata.io")
```

### Global flags

```
  -h, --help          Print usage
      --interactive   Prompt for missing required parameters (default true)
```

### SEE ALSO

* [rill](rill.md)	 - Rill CLI

