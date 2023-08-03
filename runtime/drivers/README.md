Each driver in the package can implement one or more of following interfaces:

- **OLAPStore** for storing data and running analytical queries
- **CatalogStore** for storing sources, models and metrics views, including metadata like last refresh time
- **RepoStore** for storing code artifacts (this is essentially a file system shim)
- **RegistryStore** for tracking instances (DSNs for OLAPs and repos, instance variables etc)
- **ObjectStore** for downloading files from remote object stores like s3,gcs etc
- **SQLStore** for runnning arbitrary SQL queries against DataWarehouses like bigquery. Caution: Not to be confused with postgres, duckdb etc.
- **Transporter** for transfering data from one infra to other. 
