---
title: Source YAML
sidebar_label: Source YAML
sidebar_position: 10
---

In your Rill project directory, create a `<source_name>.yaml` file in the `sources` directory containing a `type` and location (`uri` or `path`). Rill will automatically detect and ingest the source next time you run `rill start`.

## Properties

**`type`**
 —  the type of connector you are using for the source _(required)_. Possible values include:
  - _`https`_ — public files available on the web.
  - _`s3`_ — a file available on amazon s3. 
    - **Note** : Rill also supports ingesting data from other storage providers that support S3 API. Refer to the `endpoint` property below.
  - _`gcs`_ — a file available on google cloud platform.
  - _`local_file`_ — a locally available file.

**`uri`**
 —  the URI of the remote connector you are using for the source _(required for type: http, s3, gcs)_. Rill also supports glob patterns as part of the URI for S3 and GCS.
  - _`s3://your-org/bucket/file.parquet`_ —  the s3 URI of your file
  - _`gs://your-org/bucket/file.parquet`_ —  the gsutil URI of your file
  - _`https://data.example.org/path/to/file.parquet`_ —  the web address of your file

**`path`**
 — the _local path_ of the connector you are using for the source relative to your project's root directory.   _(required for type: file)_
- _`/path/to/file.csv`_ —  the path to your file

**`region`**
 — Optionally sets the cloud region of the bucket you want to connect to. Only available for S3.
  - _`us-east-1`_ —  the cloud region identifer

**`endpoint`**
 — Optionally overrides the S3 endpoint to connect to. This should only be used to connect to S3-compatible services, such as Cloudflare R2 or MinIO.

**`glob.max_total_size`**
 — Applicable if the URI is a glob pattern. The max allowed total size (in bytes) of all objects matching the glob pattern.
  - default value is _`10737418240 (10GB)`_

**`glob.max_objects_matched`**
 — Applicable if the URI is a glob pattern. The max allowed number of objects matching the glob pattern.
  - default value is _`1,000`_

**`glob.max_objects_listed`**
 — Appplicable if the URI is a glob pattern. The max number of objects to list and match against glob pattern (excluding files excluded by the glob prefix).
  - default value is _`1,000,000`_

**`timeout`**
 — The maximum time to wait for souce ingestion.

**`hive_partitioning`**
 — If set to true, hive style partitioning is transformed into column values in the data source on ingestion.
 - _`true`_ by default

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
