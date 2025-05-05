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
  export let isBillingContact: boolean;
  // Changing billing contact is not an action for this user. So handle it upstream
  // This also avoids rendering the modal per row.
  export let onAttemptChangeBillingContactUserRole: () => void;

  let isDropdownOpen = false;
  let isUpgradeConfirmOpen = false;
  let newRole = "";

  $: organization = $page.params.organization;
  $: isAdmin = currentUserRole === "admin";
  $: isEditor = currentUserRole === "editor";
  $: isGuest = role === "guest";
  $: canManageUser =
    !isCurrentUser &&
    (isAdmin ||
      (isEditor &&
        (role === "editor" || role === "viewer" || role === "guest")));

  const OPTION_DESCRIPTION_MAP = {
    admin: "Full access to org settings, members, and all projects",
    editor: "Can create/manage projects and non-admin members",
    viewer: "Read-only access to all org projects",
    guest: "Access to invited projects only",
  };

  const queryClient = useQueryClient();
  const setOrganizationMemberUserRole =
    createAdminServiceSetOrganizationMemberUserRole();

  async function handleSetRole(role: string) {
    if (role !== "admin" && isBillingContact) {
      // We cannot change a billing contact's role to a non-admin one.
      onAttemptChangeBillingContactUserRole();
      return;
    }

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

{#if canManageUser}
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
    <DropdownMenu.Content align="start" class="w-[200px]">
      {#if isAdmin}
        <DropdownMenu.Item
          class="font-normal flex flex-col items-start hover:bg-slate-50 {role ===
          'admin'
            ? 'bg-slate-100'
            : ''}"
          on:click={() => {
            handleSetRole("admin");
          }}
        >
          <span class="text-xs font-medium text-slate-700">Admin</span>
          <span class="text-[11px] text-slate-500"
            >{OPTION_DESCRIPTION_MAP.admin}</span
          >
        </DropdownMenu.Item>
      {/if}
      <DropdownMenu.Item
        class="font-normal flex flex-col items-start hover:bg-slate-50 {role ===
        'editor'
          ? 'bg-slate-100'
          : ''}"
        on:click={() => {
          handleSetRole("editor");
        }}
      >
        <span class="text-xs font-medium text-slate-700">Editor</span>
        <span class="text-[11px] text-slate-500"
          >{OPTION_DESCRIPTION_MAP.editor}</span
        >
      </DropdownMenu.Item>
      <DropdownMenu.Item
        class="font-normal flex flex-col items-start hover:bg-slate-50 {role ===
        'viewer'
          ? 'bg-slate-100'
          : ''}"
        on:click={() => {
          handleSetRole("viewer");
        }}
      >
        <span class="text-xs font-medium text-slate-700">Viewer</span>
        <span class="text-[11px] text-slate-500"
          >{OPTION_DESCRIPTION_MAP.viewer}</span
        >
      </DropdownMenu.Item>
      {#if isAdmin}
        <DropdownMenu.Item
          class="font-normal flex flex-col items-start hover:bg-slate-50 {role ===
          'guest'
            ? 'bg-slate-100'
            : ''}"
          on:click={() => {
            handleSetRole("guest");
          }}
        >
          <span class="text-xs font-medium text-slate-700">Guest</span>
          <span class="text-[11px] text-slate-500"
            >{OPTION_DESCRIPTION_MAP.guest}</span
          >
        </DropdownMenu.Item>
      {/if}
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
