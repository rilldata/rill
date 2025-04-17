<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import {
    createAdminServiceSetOrganizationMemberUserRole,
    getAdminServiceListOrganizationInvitesQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import { page } from "$app/stores";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import OrgUpgradeGuestConfirmDialog from "./OrgUpgradeGuestConfirmDialog.svelte";

  export let email: string;
  export let role: string;
  export let isCurrentUser: boolean;
  export let currentUserRole: string;

  let isDropdownOpen = false;
  let isUpgradeConfirmOpen = false;
  let newRole = "";

  $: organization = $page.params.organization;
  $: isAdmin = currentUserRole === "admin";
  $: isGuest = role === "guest";

  const queryClient = useQueryClient();
  const setOrganizationMemberUserRole =
    createAdminServiceSetOrganizationMemberUserRole();

  async function handleSetRole(role: string) {
    try {
      if (isGuest) {
        newRole = role;
        isUpgradeConfirmOpen = true;
        return;
      }

      await $setOrganizationMemberUserRole.mutateAsync({
        organization: organization,
        email: email,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListOrganizationMemberUsersQueryKey(organization),
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListOrganizationInvitesQueryKey(organization),
      });

      eventBus.emit("notification", {
        message: "User role updated",
      });
    } catch (error) {
      console.error("Error updating user role", error);
      eventBus.emit("notification", {
        message: "Error updating user role",
        type: "error",
      });
    }
  }

  async function handleUpgrade(email: string, role: string) {
    try {
      await $setOrganizationMemberUserRole.mutateAsync({
        organization: organization,
        email: email,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListOrganizationMemberUsersQueryKey(organization),
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListOrganizationInvitesQueryKey(organization),
      });

      eventBus.emit("notification", {
        message: `Guest upgraded to ${role}`,
      });
    } catch (error) {
      console.error("Error upgrading user role", error);
      eventBus.emit("notification", {
        message: "Error upgrading user role",
        type: "error",
      });
    }
  }
</script>

{#if !isCurrentUser && isAdmin}
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger
      class="w-18 flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
        ? 'bg-slate-200'
        : 'hover:bg-slate-100'} px-2 py-1"
    >
      <span class="capitalize">{role ? `${role}` : "-"}</span>
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
          handleSetRole("admin");
        }}
      >
        <span>Admin</span>
      </DropdownMenu.CheckboxItem>
      <DropdownMenu.CheckboxItem
        class="font-normal flex items-center"
        checked={role === "editor"}
        on:click={() => {
          handleSetRole("editor");
        }}
      >
        <span>Editor</span>
      </DropdownMenu.CheckboxItem>
      <DropdownMenu.CheckboxItem
        class="font-normal flex items-center"
        checked={role === "viewer"}
        on:click={() => {
          handleSetRole("viewer");
        }}
      >
        <span>Viewer</span>
      </DropdownMenu.CheckboxItem>
      <DropdownMenu.CheckboxItem
        class="font-normal flex items-center"
        checked={role === "guest"}
        on:click={() => {
          handleSetRole("guest");
        }}
      >
        <span>Guest</span>
      </DropdownMenu.CheckboxItem>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{:else}
  <div class="w-18 flex flex-row gap-1 items-center rounded-sm px-2 py-1">
    <span class="capitalize">{role}</span>
  </div>
{/if}

<OrgUpgradeGuestConfirmDialog
  bind:open={isUpgradeConfirmOpen}
  {email}
  {newRole}
  onUpgrade={handleUpgrade}
/>
