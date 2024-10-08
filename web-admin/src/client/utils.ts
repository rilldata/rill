import type { RpcStatus } from "@rilldata/web-admin/client/gen/index.schemas";
import type { Query } from "@tanstack/svelte-query";
import type { AxiosError } from "axios";
import { derived, type Readable } from "svelte/store";

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

export function mergedQueryStatusStatus(
  queriesOrMutations: Readable<{
    isLoading: boolean;
    isError: boolean;
    error?: any;
  }>[],
) {
  return derived(queriesOrMutations, (queriesOrMutations) => {
    const isLoading = queriesOrMutations.some((q) => q.isLoading);
    const isError = queriesOrMutations.some((q) => q.isError);
    const errors = queriesOrMutations
      .map((q) => q.error)
      .filter(Boolean) as AxiosError<RpcStatus>[];
    return {
      isLoading,
      isError,
      errors: errors.map((e) => e.response?.data?.message ?? e.message),
    };
  });
}
