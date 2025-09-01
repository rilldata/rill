import {
  createAdminServiceGetCurrentUser,
  createAdminServiceGetOrganization,
  createAdminServiceGetUser,
  type V1User,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { derived, type Readable } from "svelte/store";

export function getOrganizationBillingContactUser(
  organization: string,
): Readable<V1User | undefined> {
  return derived(
    [
      createAdminServiceGetOrganization(
        organization,
        undefined,
        undefined,
        queryClient,
      ),
      createAdminServiceGetCurrentUser(undefined, queryClient),
    ],
    ([orgResp, currentUser], set) => {
      if (
        orgResp.data?.organization?.billingEmail ===
        currentUser.data?.user?.email
      ) {
        set(currentUser.data?.user);
        return;
      }

      return createAdminServiceGetUser(
        {
          email: orgResp.data?.organization?.billingEmail,
        },
        undefined,
        queryClient,
      ).subscribe((u) => set(u.data?.user));
    },
  );
}
