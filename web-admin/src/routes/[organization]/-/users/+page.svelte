<script lang="ts">
  import { page } from "$app/stores";
  import type { V1OrganizationInvite } from "@rilldata/web-admin/client";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizationInvitesInfinite,
    createAdminServiceListOrganizationMemberUsersInfinite,
  } from "@rilldata/web-admin/client";
  import ChangeBillingContactDialog from "@rilldata/web-admin/features/billing/contact/ChangeBillingContactDialog.svelte";
  import { getOrganizationBillingContactUser } from "@rilldata/web-admin/features/billing/contact/selectors";
  import AddUsersDialog from "@rilldata/web-admin/features/organizations/users/AddUsersDialog.svelte";
  import ChangingBillingContactRoleDialog from "@rilldata/web-admin/features/organizations/users/ChangingBillingContactRoleDialog.svelte";
  import OrgUsersFilters from "@rilldata/web-admin/features/organizations/users/OrgUsersFilters.svelte";
  import OrgUsersTable from "@rilldata/web-admin/features/organizations/users/OrgUsersTable.svelte";
  import RemovingBillingContactDialog from "@rilldata/web-admin/features/organizations/users/RemovingBillingContactDialog.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { Plus } from "lucide-svelte";

  const PAGE_SIZE = 12;

  let userEmail = "";
  let userRole = "";
  let isSuperUser = false;

  let isAddUserDialogOpen = false;
  let isRemovingBillingContactDialogOpen = false;
  let isChangingBillingContactRoleDialogOpen = false;
  let isUpdateBillingContactDialogOpen = false;

  let searchText = "";
  let filterSelection: "all" | "members" | "guests" | "pending" = "all";

  let scrollToTopTrigger = null;
  $: {
    // Update trigger when filter selection changes to scroll to top
    scrollToTopTrigger = filterSelection;
  }

  $: organization = $page.params.organization;

  $: orgMemberUsersInfiniteQuery =
    createAdminServiceListOrganizationMemberUsersInfinite(
      organization,
      {
        pageSize: PAGE_SIZE,
      },
      {
        query: {
          getNextPageParam: (lastPage) => {
            if (lastPage.nextPageToken !== "") {
              return lastPage.nextPageToken;
            }
            return undefined;
          },
        },
      },
    );
  $: orgInvitesInfiniteQuery =
    createAdminServiceListOrganizationInvitesInfinite(
      organization,
      {
        pageSize: PAGE_SIZE,
      },
      {
        query: {
          getNextPageParam: (lastPage) => {
            if (lastPage.nextPageToken !== "") {
              return lastPage.nextPageToken;
            }
            return undefined;
          },
        },
      },
    );

  $: allOrgMemberUsersRows =
    $orgMemberUsersInfiniteQuery.data?.pages.flatMap(
      (page) => page.members ?? [],
    ) ?? [];
  $: allOrgInvitesRows =
    $orgInvitesInfiniteQuery.data?.pages.flatMap(
      (page) => page.invites ?? [],
    ) ?? [];

  function coerceInvitesToUsers(invites: V1OrganizationInvite[]) {
    return invites.map((invite) => ({
      ...invite,
      userEmail: invite.email,
      roleName: invite.roleName,
    }));
  }

  $: combinedRows = [
    ...allOrgMemberUsersRows,
    ...coerceInvitesToUsers(allOrgInvitesRows),
  ];

  $: currentUserRole = allOrgMemberUsersRows.find(
    (member) => member.userEmail === $currentUser.data?.user.email,
  )?.roleName;

  // Filter by role
  // Filter by search text
  $: filteredUsers = combinedRows
    .filter((user) => {
      const searchLower = searchText.toLowerCase();
      const matchesSearch =
        (user.userEmail?.toLowerCase() || "").includes(searchLower) ||
        ("userName" in user &&
          (user.userName?.toLowerCase() || "").includes(searchLower));

      let matchesRole = false;

      if (filterSelection === "all") {
        // All org users (members + guests + pending invites)
        matchesRole = true;
      } else if (filterSelection === "members") {
        // Only members (org admin, editor, viewer)
        matchesRole =
          !("invitedBy" in user) &&
          (user.roleName === OrgUserRoles.Admin ||
            user.roleName === OrgUserRoles.Editor ||
            user.roleName === OrgUserRoles.Viewer);
      } else if (filterSelection === "guests") {
        // Only guests
        matchesRole = user.roleName === OrgUserRoles.Guest;
      } else if (filterSelection === "pending") {
        // Only users with pending invites
        matchesRole = "invitedBy" in user;
      }

      return matchesSearch && matchesRole;
    })
    .sort((a, b) => {
      // Sort by current user first
      if (a.userEmail === $currentUser.data?.user.email) return -1;
      if (b.userEmail === $currentUser.data?.user.email) return 1;
      return 0;
    });

  const currentUser = createAdminServiceGetCurrentUser();
  $: billingContactUser = getOrganizationBillingContactUser(organization);
</script>

<div class="flex flex-col w-full">
  {#if $orgMemberUsersInfiniteQuery.isLoading || $orgInvitesInfiniteQuery.isLoading}
    <DelayedSpinner
      isLoading={$orgMemberUsersInfiniteQuery.isLoading ||
        $orgInvitesInfiniteQuery.isLoading}
      size="1rem"
    />
  {:else if $orgMemberUsersInfiniteQuery.isError || $orgInvitesInfiniteQuery.isError}
    <div class="text-red-500">
      Error loading organization members: {$orgMemberUsersInfiniteQuery.error ??
        $orgInvitesInfiniteQuery.error}
    </div>
  {:else if $orgMemberUsersInfiniteQuery.isSuccess && $orgInvitesInfiniteQuery.isSuccess}
    <div class="flex flex-col">
      <div class="flex flex-row gap-x-4">
        <Search
          placeholder="Search"
          bind:value={searchText}
          large
          autofocus={false}
          showBorderOnFocus={false}
        />
        <OrgUsersFilters bind:filterSelection />
        <Button
          type="primary"
          large
          onClick={() => (isAddUserDialogOpen = true)}
        >
          <Plus size="16px" />
          <span>Add users</span>
        </Button>
      </div>
      <div class="mt-6">
        <OrgUsersTable
          data={filteredUsers}
          usersQuery={$orgMemberUsersInfiniteQuery}
          invitesQuery={$orgInvitesInfiniteQuery}
          currentUserEmail={$currentUser.data?.user.email}
          {currentUserRole}
          billingContact={$billingContactUser?.email}
          {scrollToTopTrigger}
          onAttemptRemoveBillingContactUser={() =>
            (isRemovingBillingContactDialogOpen = true)}
          onAttemptChangeBillingContactUserRole={() =>
            (isChangingBillingContactRoleDialogOpen = true)}
        />
      </div>
      {#if filteredUsers.length > 0}
        <div class="px-2 py-3">
          <span class="font-medium text-sm text-gray-500">
            {filteredUsers.length} total user{filteredUsers.length === 1
              ? ""
              : "s"}
          </span>
        </div>
      {/if}
    </div>
  {/if}
</div>

<AddUsersDialog
  bind:open={isAddUserDialogOpen}
  email={userEmail}
  role={userRole}
  {isSuperUser}
/>

<RemovingBillingContactDialog
  bind:open={isRemovingBillingContactDialogOpen}
  onChange={() => (isUpdateBillingContactDialogOpen = true)}
/>

<ChangingBillingContactRoleDialog
  bind:open={isChangingBillingContactRoleDialogOpen}
  onChange={() => (isUpdateBillingContactDialogOpen = true)}
/>

<ChangeBillingContactDialog
  bind:open={isUpdateBillingContactDialogOpen}
  {organization}
  currentBillingContact={$billingContactUser?.email}
/>
