import { defineConfig } from "orval";

export default defineConfig({
  api: {
    input: "../proto/gen/rill/admin/v1/admin.swagger.yaml",
    output: {
      workspace: "./src/client/",
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
          AdminService_GetDeploymentCredentials: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          AdminService_GenerateReportYAML: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          AdminService_GetBillingProjectCredentials: {
            query: {
              useQuery: true,
              useMutation: false,
            },
          },
          AdminService_ListMagicAuthTokens: {
            query: {
              useQuery: true,
              useInfinite: true,
              useInfiniteQueryParam: "pageToken",
            },
          },
          AdminService_ListOrganizationInvites: {
            query: {
              useQuery: true,
              useInfinite: true,
              useInfiniteQueryParam: "pageToken",
            },
          },
          AdminService_ListOrganizationMemberUsers: {
            query: {
              useQuery: true,
              useInfinite: true,
              useInfiniteQueryParam: "pageToken",
            },
          },
        },
      },
    },
  },
});
