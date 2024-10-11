<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceListOrganizationMemberUsers,
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizationInvites,
  } from "@rilldata/web-admin/client";
  import type { V1UserInvite } from "@rilldata/web-admin/client";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import OrgUsersTable from "@rilldata/web-admin/features/organizations/users/OrgUsersTable.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Plus } from "lucide-svelte";
  import AddUsersDialog from "@rilldata/web-admin/features/organizations/users/AddUsersDialog.svelte";
  import { Search } from "@rilldata/web-common/components/search";

  let userEmail = "";
  let userRole = "";
  let isSuperUser = false;
  let isAddUserDialogOpen = false;
  let searchText = "";

  $: organization = $page.params.organization;
  $: listOrganizationMemberUsers =
    createAdminServiceListOrganizationMemberUsers(organization);
  $: listOrganizationInvites =
    createAdminServiceListOrganizationInvites(organization);

  function coerceInvitesToUsers(invites: V1UserInvite[]) {
    return invites.map((invite) => ({
      ...invite,
      userEmail: invite.email,
      roleName: invite.role,
    }));
  }

  $: usersWithPendingInvites = [
    ...($listOrganizationMemberUsers.data?.members ?? []),
    ...coerceInvitesToUsers($listOrganizationInvites.data?.invites ?? []),
  ];

  $: filteredUsers = usersWithPendingInvites.filter((user) =>
    user.userEmail.toLowerCase().includes(searchText.toLowerCase()),
  );

  const currentUser = createAdminServiceGetCurrentUser();
</script>

<div class="flex flex-col w-full">
  {#if $listOrganizationMemberUsers.isLoading}
    <DelayedSpinner
      isLoading={$listOrganizationMemberUsers.isLoading}
      size="1rem"
    />
  {:else if $listOrganizationMemberUsers.isError}
    <div class="text-red-500">
      Error loading organization members: {$listOrganizationMemberUsers.error}
    </div>
  {:else if $listOrganizationMemberUsers.isSuccess}
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
