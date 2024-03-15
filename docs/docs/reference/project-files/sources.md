---
title: Source YAML
sidebar_label: Source YAML
sidebar_position: 10
hide_table_of_contents: true
---

In your Rill project directory, create a `<source_name>.yaml` file in the `sources` directory containing a `type` and location (`uri` or `path`). Rill will automatically detect and ingest the source next time you run `rill start`.

## Properties

**`connector`**
 —  the type of connector you are using for the source _(required)_. Possible values include:
  - _`https`_ — public files available on the web.
  - _`s3`_ — a file available on amazon s3. 
    - **Note** : Rill also supports ingesting data from other storage providers that support S3 API. Refer to the `endpoint` property below.
  - _`gcs`_ — a file available on google cloud platform.
  - `local_file` — a locally available file.
  - _`motherduck`_ - data stored in motherduck
  - _`athena`_ - a data store defined in Amazon Athena
  - _`redshift`_ - a data store in Amazon Redshift
  - _`postgres`_ - data stored in Postgres
  - _`sqlite`_ - data stored in SQLite
  - _`snowflake`_ - data stored in Snowflake
  - _`bigquery`_ - data stored in BigQuery
  - _`duckdb`_ - use the [embedded DuckDB](../olap-engines/duckdb.md) engine to submit a DuckDB-supported native [SELECT](https://duckdb.org/docs/sql/statements/select.html) query (should be used in conjunction with the `sql` property)

**`type`**
 —  deprecated but preserves a legacy alias to `connector`.

**`uri`**
 —  the URI of the remote connector you are using for the source _(required for type: http, s3, gcs)_. Rill also supports glob patterns as part of the URI for S3 and GCS.
  - `s3://your-org/bucket/file.parquet` —  the s3 URI of your file
  - `gs://your-org/bucket/file.parquet` —  the gsutil URI of your file
  - `https://data.example.org/path/to/file.parquet` —  the web address of your file

**`path`**
 — the _local path_ of the connector you are using for the source relative to your project's root directory.   _(required for type: file)_
- `/path/to/file.csv` —  the path to your file

**`sql`**
- Optionally sets the SQL query to extract data from a SQL source (DuckDB/Motherduck/Athena/BigQuery/Postrgres/SQLite/Snowflake) 

**`region`**
 — Optionally sets the cloud region of the bucket or Athena you want to connect to. Only available for S3 and Athena.
  - `us-east-1` —  the cloud region identifier

**`endpoint`**
 — Optionally overrides the S3 endpoint to connect to. This should only be used to connect to S3-compatible services, such as Cloudflare R2 or MinIO.

**`output_location`**
- Optionally sets the query output location and result files in Athena (Rill removes the result files but an S3 file retention rule for the output location would make sure no orphaned files are left)

**`workgroup`**
- Optionally sets a workgroup for Athena connector. The workgroup is also used to determine an output location. A workgroup may override `output_location` if [Override client-side settings](https://docs.aws.amazon.com/athena/latest/ug/workgroups-settings-override.html) is turned on for the workgroup.  

**`project_id`**
- Sets a project id to be used to run BigQuery [jobs](https://cloud.google.com/bigquery/docs/jobs-overview) (mandatory for BiqQuery connection)

**`glob.max_total_size`**
 — Applicable if the URI is a glob pattern. The max allowed total size (in bytes) of all objects matching the glob pattern.
  - default value is _`10737418240 (10GB)`_

**`glob.max_objects_matched`**
 — Applicable if the URI is a glob pattern. The max allowed number of objects matching the glob pattern.
  - default value is _`1,000`_

**`glob.max_objects_listed`**
 — Applicable if the URI is a glob pattern. The max number of objects to list and match against glob pattern (excluding files excluded by the glob prefix).
  - default value is _`1,000,000`_

**`timeout`**
 — The maximum time to wait for souce ingestion.

**`refresh`** - Optionally specify a schedule after which Rill should re-ingest the source
  - **`cron`** - a cron schedule expression, which should be encapsulated in single quotes e.g. `'* * * * *'` (optional)
  - **`every`** - a Go duration string, such as `24h` ([docs](https://pkg.go.dev/time#ParseDuration)) (optional)

**`extract`** - Optionally limit the data ingested from remote sources (S3/GCS only)
  - **`rows`** - limits the size of data fetched
    - **`strategy`** - strategy to fetch data (**head** or **tail**)
    - **`size`** - size of data to be fetched (like `100MB`, `1GB`, etc). This is best-effort and may fetch more data than specified.
  - **`files`** - limits the total number of files to be fetched as per glob pattern
    - **`strategy`** - strategy to fetch files (**head** or **tail**)
    - **`size`** -  number of files
  - Semantics
    - If both `rows` and `files` are specified, each file matching the `files` clause will be extracted according to the `rows` clause.
    - If only `rows` is specified, no limit on number of files is applied. For example, getting a 1 GB `head` extract will download as many files as necessary.
    - If only `files` is specified, each file will be fully ingested.

**`db`**
 — Optionally set database for motherduck connector or path to SQLite db file.

**`database_url`**
 — Postgres connection string. Refer Postgres [docs](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING) for format.  

**`duckdb`** – Optionally specify raw parameters to inject into the DuckDB [`read_csv`](https://duckdb.org/docs/data/csv/overview.html), [`read_json`](https://duckdb.org/docs/data/json/overview.html) or [`read_parquet`](https://duckdb.org/docs/data/parquet/overview) statement that Rill generates internally. See the DuckDB [docs](https://duckdb.org/docs/data/overview) for a full list of available parameters. Example usage:
```yaml
duckdb:
  header: True
  delim: "'|'"
  columns: "columns={'FlightDate': 'DATE', 'UniqueCarrier': 'VARCHAR', 'OriginCityName': 'VARCHAR', 'DestCityName': 'VARCHAR'}"
```

**`dsn`** - Optionally sets the Snowflake connection string. For more information, refer to our [Snowflake connector page](/reference/connectors/snowflake.md) and the official [Go Snowflake Driver](https://pkg.go.dev/github.com/snowflakedb/gosnowflake#hdr-Connection_String) documentation for the correct syntax to use.
