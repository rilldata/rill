# `runtime/blob`

Package blob provides a way to download a batch of files ingested from remote sources like s3/gcs using google's go cdk (https://pkg.go.dev/gocloud.dev) as per user's glob pattern.

How many files are downloaded and how much data from a file is downloaded is controlled by `runtimev1.Source_ExtractPolicy`
It also has support for ingesting partial files for some formats like parquet, unzipped csv/txt/tsv files.

It uses a `planner` to implement strategies for downloading.

A planner has a `container` which keeps track of files to be downloaded and a `rowplanner` which plans how much data per file needs to be downloaded.

For partial parquet file ingestion it uses `apache arrow for go` : https://github.com/apache/arrow/tree/master/go