<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import type { V1UserInvite } from "@rilldata/web-admin/client";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import OrgUsersTable from "@rilldata/web-admin/features/organizations/users/OrgUsersTable.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Plus } from "lucide-svelte";
  import AddUsersDialog from "@rilldata/web-admin/features/organizations/users/AddUsersDialog.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import { createAdminServiceListOrganizationMemberUsersInfiniteQuery } from "@rilldata/web-admin/features/organizations/users/create-infinite-query-org-users";
  import { createAdminServiceListOrganizationInvitesInfiniteQuery } from "@rilldata/web-admin/features/organizations/users/create-infinite-query-org-invites";

  const PAGE_SIZE = 12;

  let userEmail = "";
  let userRole = "";
  let isSuperUser = false;
  let isAddUserDialogOpen = false;
  let searchText = "";

  $: organization = $page.params.organization;

  $: orgMemberUsersInfiniteQuery =
    createAdminServiceListOrganizationMemberUsersInfiniteQuery(organization, {
      pageSize: PAGE_SIZE,
    });
  $: orgInvitesInfiniteQuery =
    createAdminServiceListOrganizationInvitesInfiniteQuery(organization, {
      pageSize: PAGE_SIZE,
    });

  $: allOrgMemberUsersRows =
    $orgMemberUsersInfiniteQuery.data?.pages.flatMap(
      (page) => page.members ?? [],
    ) ?? [];
  $: allOrgInvitesRows =
    $orgInvitesInfiniteQuery.data?.pages.flatMap(
      (page) => page.invites ?? [],
    ) ?? [];

  function coerceInvitesToUsers(invites: V1UserInvite[]) {
    return allOrgInvitesRows.map((invite) => ({
      ...invite,
      userEmail: invite.email,
      roleName: invite.role,
    }));
  }

  $: combinedRows = [
    ...allOrgMemberUsersRows,
    ...coerceInvitesToUsers(allOrgInvitesRows),
  ];

  $: filteredUsers = combinedRows.filter((user) =>
    user.userEmail.toLowerCase().includes(searchText.toLowerCase()),
  );

  const currentUser = createAdminServiceGetCurrentUser();
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
    <div class="flex flex-col gap-4">
      <div class="flex flex-row gap-x-4">
        <Search
          placeholder="Search"
          bind:value={searchText}
          large
          autofocus={false}
          showBorderOnFocus={false}
        />
        <Button
          type="primary"
          large
          on:click={() => (isAddUserDialogOpen = true)}
        >
          <Plus size="16px" />
          <span>Add users</span>
        </Button>
      </div>
      <OrgUsersTable
        data={filteredUsers}
        usersQuery={$orgMemberUsersInfiniteQuery}
        invitesQuery={$orgInvitesInfiniteQuery}
        currentUserEmail={$currentUser.data?.user.email}
      />
    </div>
  {/if}
</div>

<AddUsersDialog
  bind:open={isAddUserDialogOpen}
  email={userEmail}
  role={userRole}
  {isSuperUser}
/>
