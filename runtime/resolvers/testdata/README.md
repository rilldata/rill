# `runtime/resolvers/testdata`

## Introduction

This directory contains files describing test cases for resolvers. They provide a concise way to initialize a Rill project with connectors, and then run a series of resolvers against it and testing that they produce an expected output.

Each file in this directory is executed against a freshly initialized runtime instance and when possible a freshly initialized connector (with the exception of readonly external connectors that are pre-populated with test data).

## Test file format

See `runtime/resolvers/resolvers_test.go` for details about the test file YAML syntax.

## Running a test file

Example commands for running the tests:
```bash
# Run all resolver tests
go test -run ^TestResolvers$ ./runtime/resolvers

# Run tests for one file
go test -run ^TestResolvers/metrics_sql_duckdb$ ./runtime/resolvers

# Run one test case in a test file
go test -run ^TestResolvers/metrics_sql_duckdb/simple$ ./runtime/resolvers
```

## Updating the expected output

To avoid manually entering expected values, you can run tests with the `-update` flag to overwrite the YAML files with the resolver output.

WARNING: When using this feature, carefully check that the output is correct before committing it.

Example commands:
```bash
# Update the expected output in one test file
go test -run ^TestResolvers/metrics_sql_duckdb$ ./runtime/resolvers -update
```
