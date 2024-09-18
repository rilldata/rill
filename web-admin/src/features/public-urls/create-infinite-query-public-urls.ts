import {
  adminServiceListMagicAuthTokens,
  getAdminServiceListMagicAuthTokensQueryKey,
  type AdminServiceListMagicAuthTokensParams,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import {
  createInfiniteQuery,
  type CreateInfiniteQueryOptions,
  type CreateInfiniteQueryResult,
  type QueryFunction,
  type QueryKey,
} from "@tanstack/svelte-query";

// Create an infinite query for listing magic auth tokens
// Support `nextPageToken` pagination
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
  > = ({ pageParam, signal }) =>
    adminServiceListMagicAuthTokens(
      organization,
      project,
      { ...params, pageToken: pageParam },
      signal,
    );

  const query = createInfiniteQuery<
    Awaited<ReturnType<typeof adminServiceListMagicAuthTokens>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    getNextPageParam: (lastPage) => {
      if (!lastPage.nextPageToken || lastPage.nextPageToken === "") {
        return undefined;
      }
      return lastPage.nextPageToken;
    },
    enabled: !!(organization && project),
    ...queryOptions,
  }) as CreateInfiniteQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};
