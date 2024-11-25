import {
  adminServiceListOrganizationMemberUsers,
  getAdminServiceListOrganizationMemberUsersQueryKey,
  type AdminServiceListOrganizationMemberUsersParams,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import {
  createInfiniteQuery,
  type CreateInfiniteQueryOptions,
  type CreateInfiniteQueryResult,
  type QueryFunction,
  type QueryKey,
} from "@tanstack/svelte-query";

export const createAdminServiceListOrganizationMemberUsersInfiniteQuery = <
  TData = Awaited<ReturnType<typeof adminServiceListOrganizationMemberUsers>>,
  TError = RpcStatus,
>(
  organization: string,
  params?: AdminServiceListOrganizationMemberUsersParams,
  options?: {
    query?: CreateInfiniteQueryOptions<
      Awaited<ReturnType<typeof adminServiceListOrganizationMemberUsers>>,
      TError,
      TData
    >;
  },
): CreateInfiniteQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getAdminServiceListOrganizationMemberUsersQueryKey(organization, params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServiceListOrganizationMemberUsers>>
  > = ({ pageParam, signal }) =>
    adminServiceListOrganizationMemberUsers(
      organization,
      { ...params, pageToken: pageParam },
      signal,
    );

  const query = createInfiniteQuery<
    Awaited<ReturnType<typeof adminServiceListOrganizationMemberUsers>>,
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
    enabled: !!organization,
    ...queryOptions,
  }) as CreateInfiniteQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};
