/**
 * This file implements a variant of the `GetProject` client code that authenticates using a bearer token.
 *
 * Modifications from the original Orval-generated code in `/web-admin/src/client/gen/admin-service/admin-service.ts` include:
 * - `queryFn`: Authentication via `Authorization: Bearer ${token}` header, replacing cookie-based authentication.
 * - `queryKey`: Incorporation of the `token` to differentiate cache entries.
 * - `queryOptions`: Conditional enabling of the query based on the presence of `token`.
 */

import type {
  AdminServiceGetProjectParams,
  RpcStatus,
  V1GetProjectResponse,
} from "@rilldata/web-admin/client";
import httpClient from "@rilldata/web-admin/client/http-client";
import {
  createQuery,
  type CreateQueryOptions,
  type CreateQueryResult,
  type QueryFunction,
  type QueryKey,
} from "@tanstack/svelte-query";

export const adminServiceGetProjectWithBearerToken = (
  organizationName: string,
  name: string,
  token: string,
  params?: AdminServiceGetProjectParams,
  signal?: AbortSignal,
) => {
  return httpClient<V1GetProjectResponse>({
    url: `/v1/orgs/${organizationName}/projects/${name}`,
    method: "get",
    params,
    signal,
    // We use the bearer token to authenticate the request
    headers: {
      Authorization: `Bearer ${token}`,
    },
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
  });
};

export const getAdminServiceGetProjectWithBearerTokenQueryKey = (
  organizationName: string,
  name: string,
  token: string,
  params?: AdminServiceGetProjectParams,
) => [
  `/v1/orgs/${organizationName}/projects/${name}`,
  `token/${token}`, // Ensures each token has its own entry in the QueryCache
  ...(params ? [params] : []),
];

export type AdminServiceGetProjectWithBearerTokenQueryResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceGetProjectWithBearerToken>>
>;
export type AdminServiceGetProjectWithBearerTokenQueryError = RpcStatus;

export const createAdminServiceGetProjectWithBearerToken = <
  TData = Awaited<ReturnType<typeof adminServiceGetProjectWithBearerToken>>,
  TError = RpcStatus,
>(
  organizationName: string,
  name: string,
  token: string,
  params?: AdminServiceGetProjectParams,
  options?: {
    query?: Partial<
      CreateQueryOptions<
        Awaited<ReturnType<typeof adminServiceGetProjectWithBearerToken>>,
        TError,
        TData
      >
    >;
  },
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  // We enforce this query key
  const queryKey = getAdminServiceGetProjectWithBearerTokenQueryKey(
    organizationName,
    name,
    token,
    params,
  );

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServiceGetProjectWithBearerToken>>
  > = ({ signal }) =>
    adminServiceGetProjectWithBearerToken(
      organizationName,
      name,
      token,
      params,
      signal,
    );

  const query = createQuery<
    Awaited<ReturnType<typeof adminServiceGetProjectWithBearerToken>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(organizationName && name && token), // A token must be provided
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};
