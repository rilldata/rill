<script lang="ts">
  import { page } from "$app/stores";
  import type { V1OrganizationInvite } from "@rilldata/web-admin/client";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizationInvitesInfinite,
    createAdminServiceListOrganizationMemberUsersInfinite,
  } from "@rilldata/web-admin/client";
  import { getOrganizationBillingContactUser } from "@rilldata/web-admin/features/billing/contact/selectors";
  import AddUsersDialog from "@rilldata/web-admin/features/organizations/users/AddUsersDialog.svelte";
  import ConvertGuestToMemberDialog from "@rilldata/web-admin/features/organizations/users/ConvertGuestToMemberDialog.svelte";
  import EditUserGroupDialog from "@rilldata/web-admin/features/organizations/users/EditUserGroupDialog.svelte";
  import OrgUsersTable from "@rilldata/web-admin/features/organizations/users/OrgUsersTable.svelte";
  import ShareProjectDialog from "@rilldata/web-admin/features/projects/user-management/ShareProjectDialog.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import type { PageData } from "./$types";

  const PAGE_SIZE = 12;

  export let data: PageData;
  $: ({ organizationPermissions } = data);

  let userEmail = "";
  let userRole = "";
  let isSuperUser = false;

  let isAddUserDialogOpen = false;
  let isEditUserGroupDialogOpen = false;
  let editingUserGroupName = "";
  let isShareProjectDialogOpen = false;
  let sharingProject = "";
  let convertGuestEmail = "";
  let convertGuestDialogOpen = false;

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
        role: OrgUserRoles.Guest,
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
  $: filteredUsers = combinedRows.filter((user) => {
    const searchLower = searchText.toLowerCase();
    const matchesSearch =
      (user.userEmail?.toLowerCase() || "").includes(searchLower) ||
      ("userName" in user &&
        (user.userName?.toLowerCase() || "").includes(searchLower));

    return matchesSearch;
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
      </div>
      <div class="mt-6">
        <OrgUsersTable
          {organization}
          data={filteredUsers}
          usersQuery={$orgMemberUsersInfiniteQuery}
          invitesQuery={$orgInvitesInfiniteQuery}
          currentUserEmail={$currentUser.data?.user.email}
          {currentUserRole}
          billingContact={$billingContactUser?.email}
          {scrollToTopTrigger}
          guestOnly
          onAttemptRemoveBillingContactUser={() => {}}
          onAttemptChangeBillingContactUserRole={() => {}}
          onEditUserGroup={(groupName) => {
            editingUserGroupName = groupName;
            isEditUserGroupDialogOpen = true;
          }}
          onShareProject={(projectName) => {
            sharingProject = projectName;
            isShareProjectDialogOpen = true;
          }}
          onConvertToMember={(user) => {
            convertGuestEmail = user;
            convertGuestDialogOpen = true;
          }}
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

{#if editingUserGroupName}
  <EditUserGroupDialog
    bind:open={isEditUserGroupDialogOpen}
    groupName={editingUserGroupName}
    organizationUsers={allOrgMemberUsersRows}
    currentUserEmail={$currentUser.data?.user.email}
  />
{/if}

{#if sharingProject}
  <ShareProjectDialog
    bind:open={isShareProjectDialogOpen}
    {organization}
    project={sharingProject}
    manageOrgAdmins={organizationPermissions?.manageOrgAdmins}
    manageOrgMembers={organizationPermissions?.manageOrgMembers}
  />
{/if}

<ConvertGuestToMemberDialog
  bind:open={convertGuestDialogOpen}
  email={convertGuestEmail}
  {isSuperUser}
/>
