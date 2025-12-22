import type { RpcStatus } from "@rilldata/web-admin/client/gen/index.schemas";
import type { CreateBaseMutationResult, Query } from "@tanstack/svelte-query";
import type { AxiosError } from "axios";
import { derived, type Readable } from "svelte/store";

export function isAdminServerQuery(query: Query): boolean {
  const [apiPath] = query.queryKey as string[];
  const adminApiEndpoints = [
    "/v1/deployments",
    "/v1/github",
    "/v1/orgs",
    "/v1/projects",
    "/v1/services",
    "/v1/superuser",
    "/v1/telemetry",
    "/v1/tokens",
    "/v1/users",
    "/v1/billing",
  ];

  return adminApiEndpoints.some((endpoint) => apiPath.startsWith(endpoint));
}

const OrgUsageAPIRegex = /v1\/instances\/.*\/api\/usage-meter/;
export function isOrgUsageQuery(query: Query): boolean {
  const [apiPath] = query.queryKey as string[];
  return (
    apiPath === "/v1/billing/metrics-project-credentials" ||
    OrgUsageAPIRegex.test(apiPath)
  );
}

export function mergedQueryStatus(
  queriesOrMutations: Readable<{
    isPending: boolean;
    isError: boolean;
    error?: any;
  }>[],
) {
  return derived(queriesOrMutations, (queriesOrMutations) => {
    const isLoading = queriesOrMutations
      // access 'isLoading' of all queries. this seems to be necessary to get the correct status.
      // TODO: figure out why this is the case.
      .map((q) => q.isPending)
      .some((loading) => loading);
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

export function getErrorForMutation<T>(
  mutation: CreateBaseMutationResult<T, RpcStatus>,
) {
  return (
    (mutation.error as AxiosError<RpcStatus>)?.response?.data?.message ??
    mutation.error?.message
  );
}
