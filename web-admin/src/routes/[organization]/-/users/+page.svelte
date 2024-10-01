<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceListOrganizationMemberUsers,
    createAdminServiceAddOrganizationMemberUser,
    getAdminServiceListOrganizationMemberUsersQueryKey,
    createAdminServiceRemoveOrganizationMemberUser,
    createAdminServiceSetOrganizationMemberUserRole,
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizationInvites,
    getAdminServiceListOrganizationInvitesQueryKey,
    createAdminServiceAddUsergroupMemberUser,
    createAdminServiceListOrganizationMemberUsergroups,
    getAdminServiceListUsergroupMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import OrgUsersTable from "@rilldata/web-admin/features/organizations/users/OrgUsersTable.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Plus } from "lucide-svelte";
  import AddUsersDialog from "@rilldata/web-admin/features/organizations/users/AddUsersDialog.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { Search } from "@rilldata/web-common/components/search";

  let userEmail = "";
  let userRole = "";
  let isSuperUser = false;
  let isAddUserDialogOpen = false;
  let searchText = "";

  $: organization = $page.params.organization;
  $: listOrganizationMemberUsers =
    createAdminServiceListOrganizationMemberUsers(organization);
  $: listOrganizationMemberUsergroups =
    createAdminServiceListOrganizationMemberUsergroups(organization);
  $: listOrganizationInvites =
    createAdminServiceListOrganizationInvites(organization);

  $: usersWithPendingInvites = [
    ...($listOrganizationMemberUsers.data?.members ?? []),
    ...($listOrganizationInvites.data?.invites?.map((invite) => ({
      ...invite,
      userEmail: invite.email,
      roleName: invite.role,
    })) ?? []),
  ];

  // TODO: fuzzy search
  $: filteredUsers = usersWithPendingInvites.filter((user) =>
    user.userEmail.toLowerCase().includes(searchText.toLowerCase()),
  );

  const queryClient = useQueryClient();
  const currentUser = createAdminServiceGetCurrentUser();
  const addOrganizationMemberUser =
    createAdminServiceAddOrganizationMemberUser();
  const removeOrganizationMemberUser =
    createAdminServiceRemoveOrganizationMemberUser();
  const setOrganizationMemberUserRole =
    createAdminServiceSetOrganizationMemberUserRole();
  const addUsergroupMemberUser = createAdminServiceAddUsergroupMemberUser();

  async function handleCreate(
    newEmail: string,
    newRole: string,
    isSuperUser: boolean = false,
  ) {
    try {
      await $addOrganizationMemberUser.mutateAsync({
        organization: organization,
        data: {
          email: newEmail,
          role: newRole,
          superuserForceAccess: isSuperUser,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsersQueryKey(organization),
      );

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationInvitesQueryKey(organization),
      );

      userEmail = "";
      userRole = "";
      isSuperUser = false;
      isAddUserDialogOpen = false;

      eventBus.emit("notification", { message: "User added to organization" });
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error adding user to organization",
        type: "error",
      });
    }
  }

  async function handleRemove(email: string) {
    try {
      await $removeOrganizationMemberUser.mutateAsync({
        organization: organization,
        email: email,
        // TODO: what is the default value for keepProjectRoles?
        // params: {
        //   keepProjectRoles: false,
        // },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsersQueryKey(organization),
      );

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationInvitesQueryKey(organization),
      );

      eventBus.emit("notification", {
        message: "User removed from organization",
      });
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error removing user from organization",
        type: "error",
      });
    }
  }

  async function handleSetRole(email: string, role: string) {
    try {
      await $setOrganizationMemberUserRole.mutateAsync({
        organization: organization,
        email: email,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsersQueryKey(organization),
      );

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationInvitesQueryKey(organization),
      );

      eventBus.emit("notification", {
        message: "User role updated",
      });
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error updating user role",
        type: "error",
      });
    }
  }

  async function handleAddUsergroupMemberUser(
    email: string,
    usergroup: string,
  ) {
    try {
      await $addUsergroupMemberUser.mutateAsync({
        organization: organization,
        usergroup: usergroup,
        email: email,
        data: {},
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsersQueryKey(organization),
      );

      await queryClient.invalidateQueries(
        getAdminServiceListUsergroupMemberUsersQueryKey(
          organization,
          usergroup,
        ),
      );

      eventBus.emit("notification", {
        message: "User added to user group",
      });
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error adding user to user group",
        type: "error",
      });
    }
  }
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
        userGroups={$listOrganizationMemberUsergroups.data?.members}
        currentUserEmail={$currentUser.data?.user.email}
        onRemove={handleRemove}
        onSetRole={handleSetRole}
        onAddUsergroupMemberUser={handleAddUsergroupMemberUser}
      />
    </div>
  {/if}
</div>

<AddUsersDialog
  bind:open={isAddUserDialogOpen}
  email={userEmail}
  role={userRole}
  {isSuperUser}
  onCreate={handleCreate}
/>
