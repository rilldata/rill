# Benchmark concurrent profiling on spending.parquet

(Currently only works against binaries built on [#1405](https://github.com/rilldata/rill-developer/pull/1405))

1. Start Rill:
```
./rill start --no-open --db "stage.db?rill_pool_size=1"
```

2. Run benchmark:
```
go run ./benchmark.go
```
