import type {
  AdminServiceListMagicAuthTokensParams,
  RpcStatus,
  V1ListMagicAuthTokensResponse,
} from "@rilldata/web-admin/client";
import httpClient from "@rilldata/web-admin/client/http-client";
import {
  createInfiniteQuery,
  type CreateInfiniteQueryOptions,
  type CreateInfiniteQueryResult,
  type QueryFunction,
  type QueryKey,
} from "@tanstack/svelte-query";

export const adminServiceListMagicAuthTokens = (
  organization: string,
  project: string,
  params?: AdminServiceListMagicAuthTokensParams,
  signal?: AbortSignal,
) => {
  return httpClient<V1ListMagicAuthTokensResponse>({
    url: `/v1/organizations/${organization}/projects/${project}/tokens/magic`,
    method: "get",
    params,
    signal,
  });
};

export const getAdminServiceListMagicAuthTokensQueryKey = (
  organization: string,
  project: string,
  params?: AdminServiceListMagicAuthTokensParams,
) => [
  `/v1/organizations/${organization}/projects/${project}/tokens/magic`,
  ...(params ? [params] : []),
];

export type AdminServiceListMagicAuthTokensQueryResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceListMagicAuthTokens>>
>;
export type AdminServiceListMagicAuthTokensQueryError = RpcStatus;

export const createAdminServiceListMagicAuthTokensInfiniteQuery = <
  TData = Awaited<ReturnType<typeof adminServiceListMagicAuthTokens>>,
  TError = RpcStatus,
>(
  organization: string,
  project: string,
  params?: AdminServiceListMagicAuthTokensParams,
  options?: {
    query?: CreateInfiniteQueryOptions<
      Awaited<ReturnType<typeof adminServiceListMagicAuthTokens>>,
      TError,
      TData
    >;
  },
): CreateInfiniteQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getAdminServiceListMagicAuthTokensQueryKey(organization, project, params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServiceListMagicAuthTokens>>
  > = ({ signal }) =>
    adminServiceListMagicAuthTokens(organization, project, params, signal);

  const query = createInfiniteQuery<
    Awaited<ReturnType<typeof adminServiceListMagicAuthTokens>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    getNextPageParam: (lastPage) => lastPage.nextPageToken ?? undefined,
    enabled: !!(organization && project),
    ...queryOptions,
  }) as CreateInfiniteQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};
