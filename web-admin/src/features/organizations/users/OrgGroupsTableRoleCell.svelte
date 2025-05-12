<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { page } from "$app/stores";
  import { useQueryClient } from "@tanstack/svelte-query";
  import {
    createAdminServiceAddOrganizationMemberUsergroup,
    createAdminServiceRemoveOrganizationMemberUsergroup,
    createAdminServiceSetOrganizationMemberUsergroupRole,
    getAdminServiceListOrganizationMemberUsergroupsQueryKey,
  } from "@rilldata/web-admin/client";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { capitalize } from "@rilldata/web-common/components/table/utils";

  export let name: string;
  export let role: string | undefined = undefined;
  export let manageOrgAdmins: boolean;
  let isDropdownOpen = false;

  $: organization = $page.params.organization;

  const queryClient = useQueryClient();
  const addUserGroupRole = createAdminServiceAddOrganizationMemberUsergroup();
  const setUserGroupRole =
    createAdminServiceSetOrganizationMemberUsergroupRole();
  const revokeUserGroupRole =
    createAdminServiceRemoveOrganizationMemberUsergroup();

  async function handleAddRole(role: string) {
    try {
      await $addUserGroupRole.mutateAsync({
        organization: organization,
        usergroup: name,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      });

      eventBus.emit("notification", { message: "User group role added" });
    } catch (error) {
      console.error("Error adding role to user group", error);
      eventBus.emit("notification", {
        message: "Error adding role to user group",
        type: "error",
      });
    }
  }

  async function handleSetRole(role: string) {
    try {
      await $setUserGroupRole.mutateAsync({
        organization: organization,
        usergroup: name,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      });

      eventBus.emit("notification", { message: "User group role updated" });
    } catch {
      eventBus.emit("notification", {
        message: "Error updating user group role",
        type: "error",
      });
    }
  }

  async function handleRevokeRole() {
    try {
      await $revokeUserGroupRole.mutateAsync({
        organization: organization,
        usergroup: name,
      });

      await queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      });

      eventBus.emit("notification", { message: "User group role revoked" });
    } catch {
      eventBus.emit("notification", {
        message: "Error revoking user group role",
        type: "error",
      });
    }
  }
</script>

<!-- https://docs.rilldata.com/reference/cli/usergroup/set-role -->
<DropdownMenu.Root bind:open={isDropdownOpen}>
  <DropdownMenu.Trigger
    class="w-18 flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
      ? 'bg-slate-200'
      : 'hover:bg-slate-100'} px-2 py-1"
  >
    {role ? `${capitalize(role)}` : "-"}
    {#if isDropdownOpen}
      <CaretUpIcon size="12px" />
    {:else}
      <CaretDownIcon size="12px" />
    {/if}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    {#if manageOrgAdmins}
      <DropdownMenu.CheckboxItem
        class="font-normal flex items-center"
        checked={role === "admin"}
        on:click={() => {
          if (role) {
            handleSetRole("admin");
          } else {
            handleAddRole("admin");
          }
        }}
      >
        <span>Admin</span>
      </DropdownMenu.CheckboxItem>
    {/if}
    <DropdownMenu.CheckboxItem
      class="font-normal flex items-center"
      checked={role === "editor"}
      on:click={() => {
        if (role) {
          handleSetRole("editor");
        } else {
          handleAddRole("editor");
        }
      }}
    >
      <span>Editor</span>
    </DropdownMenu.CheckboxItem>
    <DropdownMenu.CheckboxItem
      class="font-normal flex items-center"
      checked={role === "viewer"}
      on:click={() => {
        if (role) {
          handleSetRole("viewer");
        } else {
          handleAddRole("viewer");
        }
      }}
    >
      <span>Viewer</span>
    </DropdownMenu.CheckboxItem>
    {#if role}
      <DropdownMenu.Separator />
      <DropdownMenu.Item
        class="font-normal flex items-center"
        on:click={handleRevokeRole}
      >
        <span class="ml-6">Remove</span>
      </DropdownMenu.Item>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
