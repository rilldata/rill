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
  } from "@rilldata/web-admin/client";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import OrgUsersTable from "@rilldata/web-admin/features/organizations/users/OrgUsersTable.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Plus } from "lucide-svelte";
  import AddUserDialog from "@rilldata/web-admin/features/organizations/users/AddUserDialog.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  let userEmail = "";
  let userRole = "";
  let isSuperUser = false;
  let isAddUserDialogOpen = false;

  $: organization = $page.params.organization;
  $: listOrganizationMemberUsers =
    createAdminServiceListOrganizationMemberUsers(organization);
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

  $: console.log(usersWithPendingInvites);

  const queryClient = useQueryClient();
  const currentUser = createAdminServiceGetCurrentUser();
  const addOrganizationMemberUser =
    createAdminServiceAddOrganizationMemberUser();
  const removeOrganizationMemberUser =
    createAdminServiceRemoveOrganizationMemberUser();
  const setOrganizationMemberUserRole =
    createAdminServiceSetOrganizationMemberUserRole();

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
      <OrgUsersTable
        data={usersWithPendingInvites}
        currentUserEmail={$currentUser.data?.user.email}
        onRemove={handleRemove}
        onSetRole={handleSetRole}
      />
      <Button
        type="primary"
        large
        on:click={() => (isAddUserDialogOpen = true)}
      >
        <Plus size="16px" />
        <span>Add user</span>
      </Button>
    </div>
  {/if}
</div>

<AddUserDialog
  bind:open={isAddUserDialogOpen}
  email={userEmail}
  role={userRole}
  {isSuperUser}
  onCreate={handleCreate}
/>
