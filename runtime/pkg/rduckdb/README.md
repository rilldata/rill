# rduckdb

## Motivation
1. As an embedded database, DuckDB does not inherently provide the same isolation for ETL and serving workloads that other OLAP databases offer.
2. We have observed significant degradation in query performance during data ingestion.
3. In a Kubernetes environment, it is recommended to use local disks instead of network disks, necessitating separate local disk backups.

## Features
1. Utilizes separate DuckDB handles for reading and writing, each with distinct CPU and memory resources.
2. Automatically backs up writes to GCS in real-time.
3. Automatically restores from backups when starting with an empty local disk.

## Examples
1. Refer to `examples/main.go` for a usage example.

## Future Work
1. Enable writes and reads to be executed on separate machines.
2. Limit read operations to specific tables to support ephemeral tables (intermediate tables required only for writes).
