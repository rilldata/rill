<script lang="ts">
  import { page } from "$app/stores";
  import type { V1OrganizationInvite } from "@rilldata/web-admin/client";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import ChangeBillingContactDialog from "@rilldata/web-admin/features/billing/contact/ChangeBillingContactDialog.svelte";
  import { getOrganizationBillingContactUser } from "@rilldata/web-admin/features/billing/contact/selectors";
  import AddUsersDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/AddUsersDialog.svelte";
  import ChangingBillingContactRoleDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/ChangingBillingContactRoleDialog.svelte";
  import EditUserGroupDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/EditUserGroupDialog.svelte";
  import OrgUsersFilters from "@rilldata/web-admin/features/organizations/user-management/OrgUsersFilters.svelte";
  import OrgUsersTable from "@rilldata/web-admin/features/organizations/user-management/table/users/OrgUsersTable.svelte";
  import RemovingBillingContactDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/RemovingBillingContactDialog.svelte";
  import {
    getOrgUserInvites,
    getOrgUserMembers,
  } from "@rilldata/web-admin/features/organizations/user-management/selectors.ts";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { Plus } from "lucide-svelte";
  import type { PageData } from "./$types";

  export let data: PageData;
  $: ({ organizationPermissions } = data);

  let userEmail = "";
  let userRole = "";
  let isSuperUser = false;

  let isAddUserDialogOpen = false;
  let isRemovingBillingContactDialogOpen = false;
  let isChangingBillingContactRoleDialogOpen = false;
  let isUpdateBillingContactDialogOpen = false;
  let isEditUserGroupDialogOpen = false;
  let editingUserGroupName = "";

  let searchText = "";
  let filterSelection: "all" | "members" | "guests" | "pending" = "all";

  let scrollToTopTrigger = null;
  $: {
    // Update trigger when filter selection changes to scroll to top
    scrollToTopTrigger = filterSelection;
  }

  $: organization = $page.params.organization;

  $: orgMemberUsersInfiniteQuery = getOrgUserMembers({
    organization,
    guestOnly: false,
  });
  $: orgInvitesInfiniteQuery = getOrgUserInvites(organization);

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

  // Filter by role
  // Filter by search text
  $: filteredUsers = combinedRows
    .filter((user) => {
      if (user.roleName === OrgUserRoles.Guest) return false;

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
          {organization}
          data={filteredUsers}
          usersQuery={$orgMemberUsersInfiniteQuery}
          invitesQuery={$orgInvitesInfiniteQuery}
          currentUserEmail={$currentUser.data?.user.email}
          {organizationPermissions}
          billingContact={$billingContactUser?.email}
          {scrollToTopTrigger}
          guestOnly={false}
          onAttemptRemoveBillingContactUser={() =>
            (isRemovingBillingContactDialogOpen = true)}
          onAttemptChangeBillingContactUserRole={() =>
            (isChangingBillingContactRoleDialogOpen = true)}
          onEditUserGroup={(groupName) => {
            editingUserGroupName = groupName;
            isEditUserGroupDialogOpen = true;
          }}
          onConvertToMember={() => {}}
        />
      </div>
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

{#if editingUserGroupName}
  <EditUserGroupDialog
    bind:open={isEditUserGroupDialogOpen}
    groupName={editingUserGroupName}
    currentUserEmail={$currentUser.data?.user.email}
  />
{/if}
