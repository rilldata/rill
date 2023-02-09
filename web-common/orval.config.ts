import { defineConfig } from "orval";

export default defineConfig({
  api: {
    input: "../proto/gen/rill/runtime/v1/runtime.swagger.yaml",
    output: {
      workspace: "./src/runtime-client/",
      target: "gen/index.ts",
      client: "svelte-query",
      mode: "tags-split",
      mock: false,
      prettier: true,
      override: {
        mutator: {
          path: "http-client.ts", // Relative to workspace path set above
          name: "httpClient",
        },
        // Override queries and mutations here
        operations: {
          // Turn MetricsViewMeta into a query even though it's a POST request
          QueryService_MetricsViewMeta: {
            query: {
              useQuery: true,
            },
          },
          QueryService_ColumnRollupInterval: {
            query: {
              useQuery: true,
            },
          },
          QueryService_ColumnTopK: {
            query: {
              useQuery: true,
            },
          },
          QueryService_ColumnTimeSeries: {
            query: {
              useQuery: true,
            },
          },
          QueryService_TableColumns: {
            query: {
              useQuery: true,
            },
          },
          QueryService_MetricsViewTotals: {
            query: {
              useQuery: true,
            },
          },
          QueryService_MetricsViewTimeSeries: {
            query: {
              useQuery: true,
            },
          },
          QueryService_MetricsViewToplist: {
            query: {
              useQuery: true,
            },
          },
        },
      },
    },
  },
});
