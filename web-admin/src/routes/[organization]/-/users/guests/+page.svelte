<script lang="ts">
  import { page } from "$app/stores";
  import type {
    V1OrganizationInvite,
    V1OrganizationMemberUser,
  } from "@rilldata/web-admin/client";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { getOrganizationBillingContactUser } from "@rilldata/web-admin/features/billing/contact/selectors";
  import AddUsersDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/AddUsersDialog.svelte";
  import AddGuestsDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/AddGuestsDialog.svelte";
  import ConvertGuestToMemberDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/ConvertGuestToMemberDialog.svelte";
  import EditUserGroupDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/EditUserGroupDialog.svelte";
  import OrgUsersFilters from "@rilldata/web-admin/features/organizations/user-management/OrgUsersFilters.svelte";
  import OrgUsersTable from "@rilldata/web-admin/features/organizations/user-management/table/users/OrgUsersTable.svelte";
  import {
    getOrgUserInvites,
    getOrgUserMembers,
  } from "@rilldata/web-admin/features/organizations/user-management/selectors.ts";
  import { Button } from "@rilldata/web-common/components/button";
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
  let isAddGuestsDialogOpen = false;
  let isEditUserGroupDialogOpen = false;
  let editingUserGroupName = "";
  let convertGuestUser: V1OrganizationMemberUser | undefined = undefined;
  let convertGuestDialogOpen = false;

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
    guestOnly: true,
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
  $: filteredUsers = combinedRows.filter((user) => {
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
      <div class="flex flex-row gap-x-4 h-9">
        <Search
          placeholder="Search"
          bind:value={searchText}
          large
          autofocus={false}
          showBorderOnFocus={false}
        />
        <OrgUsersFilters bind:filterSelection showMembers={false} />
        <Button
          type="primary"
          large
          onClick={() => (isAddGuestsDialogOpen = true)}
        >
          <Plus size="16px" />
          <span>Add guest</span>
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
          guestOnly
          onAttemptRemoveBillingContactUser={() => {}}
          onAttemptChangeBillingContactUserRole={() => {}}
          onEditUserGroup={(groupName) => {
            editingUserGroupName = groupName;
            isEditUserGroupDialogOpen = true;
          }}
          onConvertToMember={(user) => {
            convertGuestUser = user;
            convertGuestDialogOpen = true;
          }}
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

<AddGuestsDialog bind:open={isAddGuestsDialogOpen} />

{#if editingUserGroupName}
  <EditUserGroupDialog
    bind:open={isEditUserGroupDialogOpen}
    groupName={editingUserGroupName}
    currentUserEmail={$currentUser.data?.user.email}
  />
{/if}

<ConvertGuestToMemberDialog
  bind:open={convertGuestDialogOpen}
  user={convertGuestUser}
/>
