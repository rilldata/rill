import {
  getAdminServiceGetBillingSubscriptionQueryKey,
  getAdminServiceListOrganizationBillingIssuesQueryKey,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export function invalidateBillingInfo(org: string) {
  return Promise.all([
    queryClient.refetchQueries(
      getAdminServiceGetBillingSubscriptionQueryKey(org),
    ),
    queryClient.refetchQueries(
      getAdminServiceListOrganizationBillingIssuesQueryKey(org),
    ),
  ]);
}
