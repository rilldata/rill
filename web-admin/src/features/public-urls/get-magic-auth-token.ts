/**
 * TanStack Query helpers for custom `GetMagicAuthToken` functionality.
 *
 * Problem: Generated `GetCurrentMagicAuthToken` functions in `web-admin/src/client`
 * assume the token is already in the Authorization header.
 *
 * Solution: `GetMagicAuthToken` wraps `GetCurrentMagicAuthToken`, adding a `token`
 * parameter and setting it in the Authorization header explicitly.
 */
import type {
  RpcStatus,
  V1GetCurrentMagicAuthTokenResponse,
} from "@rilldata/web-admin/client";
import httpClient from "@rilldata/web-admin/client/http-client";
import {
  createQuery,
  type CreateQueryOptions,
  type CreateQueryResult,
  type QueryFunction,
  type QueryKey,
} from "@tanstack/svelte-query";

export const adminServiceGetMagicAuthToken = (
  token: string,
  signal?: AbortSignal,
) => {
  return httpClient<V1GetCurrentMagicAuthTokenResponse>({
    url: `/v1/magic-tokens/current`,
    method: "get",
    signal,
    // We pass the token via the Authorization header
    headers: {
      Authorization: `Bearer ${token}`,
    },
    // There's no need to send credentials (cookies) with this request
    withCredentials: false,
  });
};

export const getAdminServiceGetMagicAuthTokenQueryKey = (token: string) => [
  `/v1/magic-tokens/${token}`, // Ensures each token has its own entry in the QueryCache
];

export type AdminServiceGetMagicAuthTokenQueryResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceGetMagicAuthToken>>
>;
export type AdminServiceGetMagicAuthTokenQueryError = RpcStatus;

export const createAdminServiceGetMagicAuthToken = <
  TData = Awaited<ReturnType<typeof adminServiceGetMagicAuthToken>>,
  TError = RpcStatus,
>(
  token: string,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof adminServiceGetMagicAuthToken>>,
      TError,
      TData
    >;
  },
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getAdminServiceGetMagicAuthTokenQueryKey(token);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServiceGetMagicAuthToken>>
  > = ({ signal }) => adminServiceGetMagicAuthToken(token, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof adminServiceGetMagicAuthToken>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!token,
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};
