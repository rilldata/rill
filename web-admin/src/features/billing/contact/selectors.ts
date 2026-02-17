import {
  createAdminServiceGetCurrentUser,
  createAdminServiceGetOrganization,
  type V1OrganizationMemberUser,
  type V1User,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { derived, type Readable } from "svelte/store";
import { getOrgAdminMembers } from "@rilldata/web-admin/features/organizations/user-management/selectors.ts";

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
      getOrgAdminMembers(organization),
    ],
    ([orgResp, currentUser, orgAdminsResp], set) => {
      if (
        orgResp.data?.organization?.billingEmail ===
        currentUser.data?.user?.email
      ) {
        set(currentUser.data?.user);
        return;
      }

      let adminUser: V1OrganizationMemberUser | null = null;
      orgAdminsResp.data.pages.forEach((p) => {
        const user = p.members.find(
          (m) => m.userEmail === orgResp.data?.organization?.billingEmail,
        );
        if (user) adminUser = user;
      });
      if (!adminUser) {
        set(undefined);
        return;
      }

      set({
        id: adminUser.userId,
        email: adminUser.userEmail,
        displayName: adminUser.userName,
      } satisfies V1User);
    },
  );
}
