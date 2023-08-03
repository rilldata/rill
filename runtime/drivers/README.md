Each driver in the package can implement one or more of following interfaces:

Following interfaces are system interfaces and is present for all instances. Depending on cloud/local these can be shared with all instances or private to an instance as well. Check `runtime.Options.GlobalDrivers` and `runtime.Options.LocalDrivers`.
- **OLAPStore** for storing data and running analytical queries
- **CatalogStore** for storing sources, models and metrics views, including metadata like last refresh time
- **RepoStore** for storing code artifacts (this is essentially a file system shim)
- **RegistryStore** for tracking instances (DSNs for OLAPs and repos, instance variables etc)

Following interfaces are only available as source connectors. These are always instance specific connectors.
- **ObjectStore** for downloading files from remote object stores like s3,gcs etc
- **SQLStore** for runnning arbitrary SQL queries against DataWarehouses like bigquery. Caution: Not to be confused with postgres, duckdb etc.
- **FileStore** stores path for local files.

Special interfaces. Also instance specific.
- **Transporter** for transfering data from one infra to other. 
