import { getAdminServiceListProjectsForOrganizationQueryOptions } from "@rilldata/web-admin/client";

const PAGE_SIZE = 1000;

export function listProjectsForOrgQueryOptions(org: string) {
  return getAdminServiceListProjectsForOrganizationQueryOptions(
    org,
    { pageSize: PAGE_SIZE },
    { query: { refetchOnMount: true, staleTime: Infinity } },
  );
}
