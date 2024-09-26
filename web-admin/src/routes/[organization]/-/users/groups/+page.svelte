<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceAddOrganizationMemberUsergroup,
    createAdminServiceCreateUsergroup,
    createAdminServiceDeleteUsergroup,
    createAdminServiceListOrganizationMemberUsergroups,
    createAdminServiceRemoveOrganizationMemberUsergroup,
    createAdminServiceSetOrganizationMemberUsergroupRole,
    getAdminServiceListOrganizationMemberUsergroupsQueryKey,
  } from "@rilldata/web-admin/client";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import OrgGroupsTable from "@rilldata/web-admin/features/organizations/users/OrgGroupsTable.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Plus } from "lucide-svelte";
  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog-v2";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";

  let userGroupName = "";
  let open = false;

  $: organization = $page.params.organization;
  $: organizationMemberUserGroups =
    createAdminServiceListOrganizationMemberUsergroups(organization);

  const queryClient = useQueryClient();
  const createUserGroup = createAdminServiceCreateUsergroup();
  const deleteUserGroup = createAdminServiceDeleteUsergroup();
  const addUserGroupRole = createAdminServiceAddOrganizationMemberUsergroup();
  const setUserGroupRole =
    createAdminServiceSetOrganizationMemberUsergroupRole();
  const revokeUserGroupRole =
    createAdminServiceRemoveOrganizationMemberUsergroup();

  function onUserGroupNameInput(e: any) {
    userGroupName = e.target.value;
  }

  async function handleCreate() {
    try {
      await $createUserGroup.mutateAsync({
        organization: organization,
        data: {
          name: userGroupName,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      userGroupName = "";
      open = false;

      eventBus.emit("notification", { message: "User group created" });
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error creating user group",
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
</script>

<div class="flex flex-col w-full">
  {#if $organizationMemberUserGroups.isLoading}
    <DelayedSpinner
      isLoading={$organizationMemberUserGroups.isLoading}
      size="1rem"
    />
  {:else if $organizationMemberUserGroups.isError}
    <div class="text-red-500">
      Error loading organization members: {$organizationMemberUserGroups.error}
    </div>
  {:else if $organizationMemberUserGroups.isSuccess}
    <div class="flex flex-col gap-4">
      <OrgGroupsTable
        data={$organizationMemberUserGroups.data.members}
        onDelete={handleDelete}
        onAddRole={handleAddRole}
        onSetRole={handleSetRole}
        onRevokeRole={handleRevokeRole}
      />
      <Button type="primary" large on:click={() => (open = true)}
        ><Plus size="16px" />
        <span>Add user group</span></Button
      >
    </div>
  {/if}
</div>

<Dialog
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
    open = false;
  }}
>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Add user group</DialogTitle>
    </DialogHeader>
    <DialogFooter class="mt-4">
      <div class="flex flex-col gap-2 w-full">
        <Input
          bind:value={userGroupName}
          placeholder="User group name"
          on:input={onUserGroupNameInput}
        />
        <Button type="primary" large on:click={handleCreate}>Create</Button>
      </div>
    </DialogFooter>
  </DialogContent>
</Dialog>
