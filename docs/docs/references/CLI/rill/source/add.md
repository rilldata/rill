## rill source add

Add a local file source

### Synopsis

Add a local file source. Supported file types include .parquet, .csv, .tsv.

```
rill source add <file> [flags]
```

### Flags

```
      --name string        Source name (defaults to file name)
  -f, --force              Overwrite the source if it already exists
      --db string          Database DSN (default "stage.db")
      --db-driver string   Database driver (default "duckdb")
      --delimiter string   CSV delimiter override (defaults to autodetect)
      --verbose            Sets the log level to debug
```

### Global flags

```
  -h, --help   Print usage
```

### SEE ALSO

* [rill source](source.md)	 - Create or drop a source

