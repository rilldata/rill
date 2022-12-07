# `runtime/queries/`

This package implements pre-defined analytical queries. Each query should adhere to the `runtime.Query` interface to enable efficient caching and cache invalidation of query results. 

## Adding a new query

Each query should be defined in a separate file and have its own test file containing at least one unit test and exactly one benchmark. The benchmark should be implemented against the `ad_bids` test project. See `column_topk.go` and `column_topk_test.go` for an example.

## Running benchmarks

From the repo root, you can benchmark all queries by running:
```bash
go test -bench=. -benchmem ./runtime/queries/...
```
