import {
  adminServiceListOrganizationInvites,
  getAdminServiceListOrganizationInvitesQueryKey,
  type AdminServiceListOrganizationInvitesParams,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import {
  createInfiniteQuery,
  type CreateInfiniteQueryOptions,
  type CreateInfiniteQueryResult,
  type QueryFunction,
  type QueryKey,
} from "@tanstack/svelte-query";

export const createAdminServiceListOrganizationInvitesInfiniteQuery = <
  TData = Awaited<ReturnType<typeof adminServiceListOrganizationInvites>>,
  TError = RpcStatus,
>(
  organization: string,
  params?: AdminServiceListOrganizationInvitesParams,
  options?: {
    query?: CreateInfiniteQueryOptions<
      Awaited<ReturnType<typeof adminServiceListOrganizationInvites>>,
      TError,
      TData
    >;
  },
): CreateInfiniteQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getAdminServiceListOrganizationInvitesQueryKey(organization, params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServiceListOrganizationInvites>>
  > = ({ pageParam, signal }) =>
    adminServiceListOrganizationInvites(
      organization,
      { ...params, pageToken: pageParam },
      signal,
    );

  const query = createInfiniteQuery<
    Awaited<ReturnType<typeof adminServiceListOrganizationInvites>>,
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
