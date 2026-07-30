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
  return getOrganizationBillingUser(organization, "billingEmail");
}

export function getOrganizationBillingPortalAdminUser(
  organization: string,
): Readable<V1User | undefined> {
  return getOrganizationBillingUser(organization, "billingPortalAdmin");
}

function getOrganizationBillingUser(
  organization: string,
  field: "billingEmail" | "billingPortalAdmin",
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
      const email = orgResp.data?.organization?.[field];
      if (!email) {
        set(undefined);
        return;
      }

      if (email === currentUser.data?.user?.email) {
        set(currentUser.data?.user);
        return;
      }

      // The org admins query fails for users without member read access
      // (e.g. a billing portal admin), in which case we fall back to the raw email below.
      const adminUser: V1OrganizationMemberUser | undefined =
        orgAdminsResp.data?.pages
          ?.flatMap((p) => p.members ?? [])
          ?.find((m) => m.userEmail === email);

      if (adminUser) {
        set({
          id: adminUser.userId,
          email: adminUser.userEmail,
          displayName: adminUser.userName,
          photoUrl: adminUser.userPhotoUrl,
        } satisfies V1User);
        return;
      }

      // The email does not belong to an org admin (or the members list is unavailable);
      // show the raw email.
      set({ email } satisfies V1User);
    },
  );
}
