# `runtime/metricsview/`

This package contains the core logic for Rill's metrics layer (semantic layer). Namely, that includes validating metrics views and generating SQL queries for querying dimensions and measures defined in a metrics view.

Some other code files relevant to metrics views that are not found in this package:

- `proto/rill/runtime/v1/resources.proto`: defines the spec for a metrics view
- `runtime/compilers/rillv1/parse_metrics_view.go`: contains the logic for parsing a metrics view spec from a YAML file
