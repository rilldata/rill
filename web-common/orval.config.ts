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
              useMutation: false,
            },
          },
          QueryService_ColumnRollupInterval: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          QueryService_ColumnTopK: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          QueryService_ColumnTimeSeries: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          QueryService_TableColumns: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          QueryService_MetricsViewAggregation: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          QueryService_MetricsViewTotals: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          QueryService_MetricsViewTimeSeries: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          QueryService_MetricsViewToplist: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          QueryService_MetricsViewComparison: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          QueryService_MetricsViewRows: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          QueryService_MetricsViewTimeRange: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          QueryService_MetricsViewSearch: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          QueryService_MetricsViewTimeRanges: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          QueryService_MetricsViewAnnotations: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          QueryService_ResolveComponent: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          QueryService_ResolveCanvas: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          RuntimeService_IssueDevJWT: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          RuntimeService_GetModelPartitions: {
            query: {
              useInfinite: true,
              useInfiniteQueryParam: "pageToken",
            },
          },
        },
      },
    },
  },
});
