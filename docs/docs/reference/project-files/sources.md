---
title: Source YAML
sidebar_label: Source YAML
sidebar_position: 10
hide_table_of_contents: true
---

In your Rill project directory, create a `<source_name>.yaml` file in the `sources` directory containing a `type` and location (`uri` or `path`). Rill will automatically detect and ingest the source next time you run `rill start`.

:::tip Did you know?

Files that are *nested at any level* under your native `sources` directory will be assumed to be sources (unless **otherwise** specified by the `type` property).

:::

## Properties

**`type`** - Refers to the resource type and must be `source` _(required)_.

**`connector`**
 —  Refers to the connector type for the source _(required)_.
  - _`https`_ — public files accessible through the web via a http/https URL endpoint
  - _`s3`_ — a file available on amazon s3
    - **Note**: Rill also supports ingesting data from other storage providers that support S3 API. Refer to the `endpoint` property below.
  - _`gcs`_ — a file available on google cloud platform
  - `local_file` — a locally available file in a supported format (e.g. parquet, csv, etc.)
  - _`motherduck`_ - data stored in motherduck
  - _`athena`_ - a data store defined in Amazon Athena
  - _`redshift`_ - a data store in Amazon Redshift
  - _`postgres`_ - data stored in Postgres
  - _`sqlite`_ - data stored in SQLite
  - _`snowflake`_ - data stored in Snowflake
  - _`bigquery`_ - data stored in BigQuery
  - _`duckdb`_ - use the [embedded DuckDB](../olap-engines/duckdb.md) engine to submit a DuckDB-supported native [SELECT](https://duckdb.org/docs/sql/statements/select.html) query (should be used in conjunction with the `sql` property)

**`type`**
 — _Deprecated_ but preserves a legacy alias to `connector`. Can be used instead to specify the source connector, instead of the resource type (see above), **only** if the source YAML file belongs in the `<RILL_HOME>/sources/` directory (preserved primarily for backwards compatibility).

**`uri`**
 —  Refers to the URI of the remote connector you are using for the source. Rill also supports glob patterns as part of the URI for S3 and GCS _(required for type: http, s3, gcs)_.
  - `s3://your-org/bucket/file.parquet` —  the s3 URI of your file
  - `gs://your-org/bucket/file.parquet` —  the gsutil URI of your file
  - `https://data.example.org/path/to/file.parquet` —  the web address of your file

**`path`**
 — Refers to the _local path_ of the connector you are using for the source relative to your project's root directory _(required for type: file)_.
- `/path/to/file.csv` —  the path to your file

**`sql`**
 — Sets the SQL query to extract data from a SQL source: DuckDB/Motherduck/Athena/BigQuery/Postrgres/SQLite/Snowflake _(optional)_.

**`region`**
 — Sets the cloud region of the S3 bucket or Athena you want to connect to using the cloud region identifier (e.g. `us-east-1`). Only available for S3 and Athena _(optional)_.

**`endpoint`**
 — Overrides the S3 endpoint to connect to. This should **only** be used to connect to S3-compatible services, such as Cloudflare R2 or MinIO _(optional)_.

**`output_location`**
 — Sets the query output location and result files in Athena. Please note that Rill will remove the result files but setting a S3 file retention rule for the output location would make sure no orphaned files are left _(optional)_.

**`workgroup`**
 — Sets a workgroup for Athena connector. The workgroup is also used to determine an output location. A workgroup may override `output_location` if [Override client-side settings](https://docs.aws.amazon.com/athena/latest/ug/workgroups-settings-override.html) is turned on for the workgroup _(optional)_.  

**`project_id`**
 — Sets a project id to be used to run BigQuery [jobs](https://cloud.google.com/bigquery/docs/jobs-overview) _(required for type: bigquery)_.

**`glob.max_total_size`**
 — Applicable if the URI is a glob pattern. The max allowed total size (in bytes) of all objects matching the glob pattern _(optional)_.
  - Default value is _`10737418240 (10GB)`_

**`glob.max_objects_matched`**
 — Applicable if the URI is a glob pattern. The max allowed number of objects matching the glob pattern _(optional)_.
  - Default value is _`1,000`_

**`glob.max_objects_listed`**
 — Applicable if the URI is a glob pattern. The max number of objects to list and match against glob pattern, not inclusive of files already excluded by the glob prefix _(optional)_.
  - Default value is _`1,000,000`_

**`timeout`**
 — The maximum time to wait for souce ingestion _(optional)_.

**`refresh`** - Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying source data _(optional)_.
  - **`cron`** - a cron schedule expression, which should be encapsulated in single quotes, e.g. `'* * * * *'` _(optional)_
  - **`every`** - a Go duration string, such as `24h` ([docs](https://pkg.go.dev/time#ParseDuration)) _(optional)_

**`extract`** - Limits the data ingested from remote sources. Only available for S3 and GCS _(optional)_.
  - **`rows`** - limits the size of data fetched
    - **`strategy`** - strategy to fetch data (_head_ or _tail_)
    - **`size`** - size of data to be fetched (like `100MB`, `1GB`, etc). This is best-effort and <u>may</u> fetch more data than specified.
  - **`files`** - limits the total number of files to be fetched as per glob pattern
    - **`strategy`** - strategy to fetch files (_head_ or _tail_)
    - **`size`** -  number of files

:::tip A note on semantics
    - If both `rows` and `files` are specified, each file matching the `files` clause will be extracted according to the `rows` clause.
    - If only `rows` is specified, no limit on number of files is applied. For example, getting a 1 GB `head` extract will download as many files as necessary.
    - If only `files` is specified, each file will be fully ingested.
:::

**`db`**
 — Sets the database for motherduck connections and/or the path to the DuckDB/SQLite `db` file _(optional)_.
  - For DuckDB / SQLite, [if deploying to Rill Cloud](/deploy/existing-project), this `db` file will need to be accessible from the <u>root</u> directory of your project on Github.

**`database_url`**
 — Postgres connection string that should be used. Refer to Postgres [documentation](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING) for more details _(optional)_.
  - If not specified in the source YAML, the `connector.postgres.database_url` connection string will need to be set when [deploying the project to Rill Cloud](/build/credentials/#setting-credentials-for-a-rill-cloud-project).

**`duckdb`** – Specifies the raw parameters to inject into the DuckDB [`read_csv`](https://duckdb.org/docs/data/csv/overview.html), [`read_json`](https://duckdb.org/docs/data/json/overview.html) or [`read_parquet`](https://duckdb.org/docs/data/parquet/overview) statement that Rill generates internally. See the DuckDB [documentation](https://duckdb.org/docs/data/overview) for a full list of available parameters _(optional)_. 

Example usage:
```yaml
duckdb:
  header: True
  delim: "'|'"
  columns: "columns={'FlightDate': 'DATE', 'UniqueCarrier': 'VARCHAR', 'OriginCityName': 'VARCHAR', 'DestCityName': 'VARCHAR'}"
```

**`dsn`** - Used to set the Snowflake connection string. For more information, refer to our [Snowflake connector page](/reference/connectors/snowflake.md) and the official [Go Snowflake Driver](https://pkg.go.dev/github.com/snowflakedb/gosnowflake#hdr-Connection_String) documentation _(optional)_.
  - If not specified in the source YAML, the `connector.snowflake.dsn` connection string will need to be set when [deploying the project to Rill Cloud](/build/credentials/#setting-credentials-for-a-rill-cloud-project).
