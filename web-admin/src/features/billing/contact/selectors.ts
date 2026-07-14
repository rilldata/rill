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

      const billingEmail = orgResp.data?.organization?.billingEmail;

      const adminUser: V1OrganizationMemberUser | undefined = billingEmail
        ? orgAdminsResp.data?.pages
            ?.flatMap((p) => p.members ?? [])
            ?.find((m) => m.userEmail === billingEmail)
        : undefined;

      if (!adminUser) {
        set(undefined);
        return;
      }

      set({
        id: adminUser.userId,
        email: adminUser.userEmail,
        displayName: adminUser.userName,
        photoUrl: adminUser.userPhotoUrl,
      } satisfies V1User);
    },
  );
}
