import type { Query } from "@tanstack/svelte-query";

export function isAdminServerQuery(query: Query): boolean {
  const [apiPath] = query.queryKey as string[];
  const adminApiEndpoints = [
    "/v1/deployments",
    "/v1/github",
    "/v1/organizations",
    "/v1/projects",
    "/v1/services",
    "/v1/superuser",
    "/v1/telemetry",
    "/v1/tokens",
    "/v1/users",
  ];

  return adminApiEndpoints.some((endpoint) => apiPath.startsWith(endpoint));
}
