<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
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

  export let name: string;
  export let role: string | undefined = undefined;

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

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

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

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      eventBus.emit("notification", { message: "User group role updated" });
    } catch (error) {
      console.error("Error updating user group role", error);
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

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      eventBus.emit("notification", { message: "User group role revoked" });
    } catch (error) {
      console.error("Error revoking user group role", error);
      eventBus.emit("notification", {
        message: "Error revoking user group role",
        type: "error",
      });
    }
  }
</script>

<!-- all-users — "cannot add role for all-users group" -->
{#if name !== "all-users"}
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger
      class="w-18 flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
        ? 'bg-slate-200'
        : 'hover:bg-slate-100'} px-2 py-1"
    >
      {role ? `Org ${role}` : "-"}
      {#if isDropdownOpen}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
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
{:else}
  <Tooltip location="top" alignment="start" distance={8}>
    <div class="w-18 rounded-sm px-2 py-1">
      <span class="cursor-help">-</span>
    </div>
    <TooltipContent maxWidth="400px" slot="tooltip-content">
      Cannot add role for all-users group
    </TooltipContent>
  </Tooltip>
{/if}
