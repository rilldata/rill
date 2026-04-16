import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import {
  adminServiceListOrganizations,
  getAdminServiceListOrganizationsQueryKey,
} from "@rilldata/web-admin/client";
import { redirect } from "@sveltejs/kit";

export const load = async () => {
  const orgListResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceListOrganizationsQueryKey(),
    queryFn: () => adminServiceListOrganizations(),
    staleTime: Infinity,
  });
  // Safeguard to ensure we dont show create org page when user already has orgs.
  // Adding it here to avoid checking it in every page that might redirect here.
  // As of now this can only happen when user enters `/-/welcome/*` pages directly.
  if (orgListResp.organizations?.length) {
    throw redirect(307, "/");
  }
};
