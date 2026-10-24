<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { page } from "$app/stores";
  import type { V1OrganizationMemberUser } from "@rilldata/web-admin/client";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { getOrganizationBillingContactUser } from "@rilldata/web-admin/features/billing/contact/selectors";
  import AddUsersDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/AddUsersDialog.svelte";
  import AddGuestsDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/AddGuestsDialog.svelte";
  import ConvertGuestToMemberDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/ConvertGuestToMemberDialog.svelte";
  import EditUserGroupDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/EditUserGroupDialog.svelte";
  import ManageUserGroupsDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/ManageUserGroupsDialog.svelte";
  import OrgUsersFilters from "@rilldata/web-admin/features/organizations/user-management/OrgUsersFilters.svelte";
  import OrgUsersTable from "@rilldata/web-admin/features/organizations/user-management/table/users/OrgUsersTable.svelte";
  import {
    getOrgUserInvites,
    getOrgUserMembers,
    type OrgUserMemberFilters,
  } from "@rilldata/web-admin/features/organizations/user-management/selectors.ts";
  import {
    coerceInvitesToUsers,
    type OrgUserRow,
  } from "@rilldata/web-admin/features/organizations/user-management/utils.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import { Search } from "@rilldata/web-common/components/search";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import { onDestroy } from "svelte";
  import { writable } from "svelte/store";
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
  let isManageGroupsDialogOpen = false;
  let manageGroupsUser: {
    email: string;
    userId: string;
    pendingAcceptance: boolean;
    usergroups: string[];
  } | null = null;
  let convertGuestUser: V1OrganizationMemberUser | undefined = undefined;
  let convertGuestDialogOpen = false;

  let searchText = "";
  let debouncedSearchText = "";
  const updateSearch = debounce((value: string) => {
    debouncedSearchText = value;
  }, 250);
  $: updateSearch(searchText);
  onDestroy(updateSearch.cancel);

  const memberFilters = writable<OrgUserMemberFilters>({
    organization: "",
    guestOnly: true,
  });
  let filterSelection: "all" | "members" | "guests" | "pending" = "all";

  let scrollToTopTrigger: unknown = null;
  $: {
    // Scroll to the top when the search or filters change.
    scrollToTopTrigger = { searchText, filterSelection };
  }

  $: organization = $page.params.organization;

  $: memberFilters.set({
    organization,
    guestOnly: true,
    searchText: debouncedSearchText,
    enabled: filterSelection !== "pending",
  });
  const orgMemberUsersInfiniteQuery = getOrgUserMembers(memberFilters);
  $: orgInvitesInfiniteQuery = getOrgUserInvites(organization);

  $: allOrgMemberUsersRows =
    $orgMemberUsersInfiniteQuery.data?.pages.flatMap(
      (page) => page.members ?? [],
    ) ?? [];
  $: allOrgInvitesRows =
    $orgInvitesInfiniteQuery.data?.pages.flatMap(
      (page) => page.invites ?? [],
    ) ?? [];

  $: combinedRows = [
    ...(allOrgMemberUsersRows as OrgUserRow[]),
    ...coerceInvitesToUsers(allOrgInvitesRows),
  ];

  // Members are searched by the API. Only invitations need local text matching.
  $: filteredUsers = combinedRows.filter((user) => {
    const searchLower = debouncedSearchText.toLowerCase();
    const matchesSearch =
      !user.pendingAcceptance ||
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
        !user.pendingAcceptance &&
        (user.roleName === OrgUserRoles.Admin ||
          user.roleName === OrgUserRoles.Editor ||
          user.roleName === OrgUserRoles.Viewer);
    } else if (filterSelection === "pending") {
      // Only users with pending invites
      matchesRole = !!user.pendingAcceptance;
    }

    return matchesSearch && matchesRole;
  });

  const currentUser = createAdminServiceGetCurrentUser();
  $: billingContactUser = getOrganizationBillingContactUser(organization);
</script>

<div class="flex flex-col w-full">
  {#if $orgMemberUsersInfiniteQuery.isError || $orgInvitesInfiniteQuery.isError}
    <div class="text-red-500">
      {m.users_error_loading_members()}
      {$orgMemberUsersInfiniteQuery.error ?? $orgInvitesInfiniteQuery.error}
    </div>
  {/if}
  <div class="flex flex-col">
    <div class="flex flex-row flex-wrap gap-x-4 gap-y-2 sm:flex-nowrap sm:h-9">
      <Search
        bind:value={searchText}
        large
        autofocus={false}
        showBorderOnFocus={false}
      />
      <OrgUsersFilters
        bind:filterSelection
        showMembers={false}
        showRoleFilter={false}
      />
      <Button
        type="primary"
        large
        onClick={() => (isAddGuestsDialogOpen = true)}
      >
        <Plus size="16px" />
        <span>{m.users_add_guest()}</span>
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
        isSearchPending={searchText !== debouncedSearchText}
        showMembers={filterSelection !== "pending"}
        showInvites={filterSelection !== "members"}
        hasActiveFilters={!!searchText || filterSelection !== "all"}
        guestOnly
        onAttemptRemoveBillingContactUser={() => {}}
        onAttemptChangeBillingContactUserRole={() => {}}
        onEditUserGroup={(groupName) => {
          editingUserGroupName = groupName;
          isEditUserGroupDialogOpen = true;
        }}
        onManageGroups={(user: OrgUserRow) => {
          manageGroupsUser = {
            email: user.userEmail ?? "",
            userId: user.userId ?? "",
            pendingAcceptance: !!user.pendingAcceptance,
            usergroups: user.usergroups ?? [],
          };
          isManageGroupsDialogOpen = true;
        }}
        onConvertToMember={(user) => {
          convertGuestUser = user;
          convertGuestDialogOpen = true;
        }}
      />
    </div>
  </div>
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

{#if manageGroupsUser}
  <ManageUserGroupsDialog
    bind:open={isManageGroupsDialogOpen}
    {organization}
    email={manageGroupsUser.email}
    userId={manageGroupsUser.userId}
    pendingAcceptance={manageGroupsUser.pendingAcceptance}
    currentGroups={manageGroupsUser.usergroups}
  />
{/if}

<ConvertGuestToMemberDialog
  bind:open={convertGuestDialogOpen}
  user={convertGuestUser}
/>
