import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import {
  adminServiceListOrganizations,
  getAdminServiceListOrganizationsQueryKey,
} from "@rilldata/web-admin/client";
import { redirect } from "@sveltejs/kit";

export async function load({ parent }) {
  await parent();

  const orgListResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceListOrganizationsQueryKey(),
    queryFn: () => adminServiceListOrganizations(),
    staleTime: Infinity,
  });
  if (orgListResp.organizations?.length) {
    throw redirect(307, "/-/welcome/organization/select");
  } else {
    throw redirect(307, "/-/welcome/organization/create");
  }
}
