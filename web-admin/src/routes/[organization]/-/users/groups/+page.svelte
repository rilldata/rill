<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceAddOrganizationMemberUsergroup,
    createAdminServiceCreateUsergroup,
    createAdminServiceDeleteUsergroup,
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizationMemberUsergroups,
    createAdminServiceRemoveOrganizationMemberUsergroup,
    createAdminServiceRemoveUsergroupMemberUser,
    createAdminServiceRenameUsergroup,
    createAdminServiceSetOrganizationMemberUsergroupRole,
    getAdminServiceListOrganizationMemberUsergroupsQueryKey,
    getAdminServiceListUsergroupMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import OrgGroupsTable from "@rilldata/web-admin/features/organizations/users/OrgGroupsTable.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Plus } from "lucide-svelte";
  import CreateUserGroupDialog from "@rilldata/web-admin/features/organizations/users/CreateUserGroupDialog.svelte";

  let userGroupName = "";
  let isCreateUserGroupDialogOpen = false;

  $: organization = $page.params.organization;
  $: listOrganizationMemberUsergroups =
    createAdminServiceListOrganizationMemberUsergroups(organization);
  // $: listOrganizationMemberUsers =
  //   createAdminServiceListOrganizationMemberUsers(organization);

  const queryClient = useQueryClient();
  const currentUser = createAdminServiceGetCurrentUser();
  const createUserGroup = createAdminServiceCreateUsergroup();
  const renameUserGroup = createAdminServiceRenameUsergroup();
  const deleteUserGroup = createAdminServiceDeleteUsergroup();
  const addUserGroupRole = createAdminServiceAddOrganizationMemberUsergroup();
  const setUserGroupRole =
    createAdminServiceSetOrganizationMemberUsergroupRole();
  const revokeUserGroupRole =
    createAdminServiceRemoveOrganizationMemberUsergroup();
  const removeUserGroupMember = createAdminServiceRemoveUsergroupMemberUser();

  async function handleCreate(newName: string) {
    try {
      await $createUserGroup.mutateAsync({
        organization: organization,
        data: {
          name: newName,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      userGroupName = "";
      isCreateUserGroupDialogOpen = false;

      eventBus.emit("notification", { message: "User group created" });
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error creating user group",
        type: "error",
      });
    }
  }

  async function handleRename(groupName: string, newName: string) {
    try {
      await $renameUserGroup.mutateAsync({
        organization: organization,
        usergroup: groupName,
        data: {
          name: newName,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      eventBus.emit("notification", { message: "User group renamed" });
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error renaming user group",
        type: "error",
      });
    }
  }

  async function handleDelete(deletedUserGroupName: string) {
    try {
      await $deleteUserGroup.mutateAsync({
        organization: organization,
        usergroup: deletedUserGroupName,
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      eventBus.emit("notification", { message: "User group deleted" });
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error deleting user group",
        type: "error",
      });
    }
  }

  async function handleAddRole(groupName: string, role: string) {
    try {
      await $addUserGroupRole.mutateAsync({
        organization: organization,
        usergroup: groupName,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      eventBus.emit("notification", { message: "User group role added" });
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error adding role to user group",
        type: "error",
      });
    }
  }

  async function handleSetRole(groupName: string, role: string) {
    try {
      await $setUserGroupRole.mutateAsync({
        organization: organization,
        usergroup: groupName,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      eventBus.emit("notification", { message: "User group role updated" });
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error updating user group role",
        type: "error",
      });
    }
  }

  async function handleRevokeRole(groupName: string) {
    try {
      await $revokeUserGroupRole.mutateAsync({
        organization: organization,
        usergroup: groupName,
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      eventBus.emit("notification", { message: "User group role revoked" });
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error revoking user group role",
        type: "error",
      });
    }
  }

  async function handleRemoveUser(groupName: string, email: string) {
    try {
      await $removeUserGroupMember.mutateAsync({
        organization: organization,
        usergroup: groupName,
        email: email,
      });

      await queryClient.invalidateQueries(
        getAdminServiceListUsergroupMemberUsersQueryKey(
          organization,
          groupName,
        ),
      );

      eventBus.emit("notification", {
        message: "User removed from user group",
      });
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error removing user from user group",
        type: "error",
      });
    }
  }
</script>

<div class="flex flex-col w-full">
  {#if $listOrganizationMemberUsergroups.isLoading}
    <DelayedSpinner
      isLoading={$listOrganizationMemberUsergroups.isLoading}
      size="1rem"
    />
  {:else if $listOrganizationMemberUsergroups.isError}
    <div class="text-red-500">
      Error loading organization user groups: {$listOrganizationMemberUsergroups.error}
    </div>
  {:else if $listOrganizationMemberUsergroups.isSuccess}
    <div class="flex flex-col gap-4">
      <OrgGroupsTable
        data={$listOrganizationMemberUsergroups.data.members}
        currentUserEmail={$currentUser.data?.user.email}
        onRename={handleRename}
        onDelete={handleDelete}
        onAddRole={handleAddRole}
        onSetRole={handleSetRole}
        onRevokeRole={handleRevokeRole}
        onRemoveUser={handleRemoveUser}
      />
      <Button
        type="primary"
        large
        on:click={() => (isCreateUserGroupDialogOpen = true)}
      >
        <Plus size="16px" />
        <span>Create group</span>
      </Button>
    </div>
  {/if}
</div>

<CreateUserGroupDialog
  bind:open={isCreateUserGroupDialogOpen}
  groupName={userGroupName}
  onCreate={handleCreate}
/>
