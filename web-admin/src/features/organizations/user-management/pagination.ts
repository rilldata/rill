type PaginatedQuery = {
  hasNextPage: boolean;
  isFetching: boolean;
  isSuccess: boolean;
  fetchNextPage: () => Promise<unknown>;
};

// Only invitations need client-side filtering: ListOrganizationInvites has no
// search or role parameters. Members use the API's searchPattern and role filters.
// Continue fetching invitations even when none of the loaded rows match.
export function loadNextInvitePageForFilter(
  query: PaginatedQuery,
  hasActiveFilters: boolean,
) {
  if (
    hasActiveFilters &&
    query.hasNextPage &&
    query.isSuccess &&
    !query.isFetching
  ) {
    void query.fetchNextPage();
  }
}
