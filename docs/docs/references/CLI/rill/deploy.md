## rill deploy

Deploy project to Rill Cloud

```
rill deploy [flags]
```

### Flags

```
      --prod-slots int          Slots to allocate for production deployments (default 2)
      --description string      Project description
      --region string           Deployment region
      --prod-db-driver string   Database driver (default "duckdb")
      --prod-db-dsn string      Database driver configuration
      --public                  Make dashboards publicly accessible
      --prod-branch string      Git branch to deploy from (default: the default Git branch)
      --name string             Project name (default: Git repo name)
      --api-token string        Token for authenticating with the admin API
      --api-url string          Base URL for the admin API (default "https://admin.rilldata.io")
```

### Global flags

```
  -h, --help   Print usage
```

### SEE ALSO

* [rill](rill.md)	 - Rill CLI

