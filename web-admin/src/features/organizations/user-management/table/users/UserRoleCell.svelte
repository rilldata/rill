<script lang="ts">
  import RemoveUserFromOrgConfirmDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/RemoveUserFromOrgConfirmDialog.svelte";
  import {
    canManageOrgUser,
    invalidateAfterUserDelete,
  } from "@rilldata/web-admin/features/organizations/user-management/utils.ts";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import {
    createAdminServiceRemoveOrganizationMemberUser,
    createAdminServiceSetOrganizationMemberUserRole,
    getAdminServiceListOrganizationInvitesQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
    type V1OrganizationPermissions,
  } from "@rilldata/web-admin/client";
  import { page } from "$app/stores";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import OrgUpgradeGuestConfirmDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/OrgUpgradeGuestConfirmDialog.svelte";
  import { ORG_ROLES_DESCRIPTION_MAP } from "@rilldata/web-admin/features/organizations/user-management/constants.ts";

  export let email: string;
  export let role: string;
  export let isCurrentUser: boolean;
  export let organizationPermissions: V1OrganizationPermissions;
  export let isBillingContact: boolean;
  // Changing billing contact is not an action for this user. So handle it upstream
  // This also avoids rendering the modal per row.
  export let onAttemptChangeBillingContactUserRole: () => void;

  let isDropdownOpen = false;
  let isUpgradeConfirmOpen = false;
  let isRemoveConfirmOpen = false;
  let newRole = "";

  $: organization = $page.params.organization;
  $: isGuest = role === OrgUserRoles.Guest;
  $: canManageUser =
    !isCurrentUser && canManageOrgUser(organizationPermissions, role);

  const queryClient = useQueryClient();
  const setOrganizationMemberUserRole =
    createAdminServiceSetOrganizationMemberUserRole();
  const removeOrganizationMemberUser =
    createAdminServiceRemoveOrganizationMemberUser();

  async function handleSetRole(role: string) {
    if (role !== OrgUserRoles.Admin && isBillingContact) {
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
        org: organization,
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
        org: organization,
        email,
        data: {
          role,
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

  async function handleRemove() {
    try {
      await $removeOrganizationMemberUser.mutateAsync({
        org: organization,
        email,
      });

      await invalidateAfterUserDelete(queryClient, organization);

      eventBus.emit("notification", {
        message: "User removed from organization",
      });
    } catch (error) {
      console.error("Error removing user from organization", error);
      eventBus.emit("notification", {
        message: "Error removing user from organization",
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
      {#if organizationPermissions.manageOrgAdmins}
        <DropdownMenu.Item
          class="font-normal flex flex-col items-start hover:bg-slate-50 {role ===
          'admin'
            ? 'bg-slate-100'
            : ''}"
          on:click={() => handleSetRole(OrgUserRoles.Admin)}
        >
          <span class="text-xs font-medium text-slate-700">Admin</span>
          <span class="text-[11px] text-slate-500"
            >{ORG_ROLES_DESCRIPTION_MAP.admin}</span
          >
        </DropdownMenu.Item>
      {/if}
      <DropdownMenu.Item
        class="font-normal flex flex-col items-start hover:bg-slate-50 {role ===
        'editor'
          ? 'bg-slate-100'
          : ''}"
        on:click={() => handleSetRole(OrgUserRoles.Editor)}
      >
        <span class="text-xs font-medium text-slate-700">Editor</span>
        <span class="text-[11px] text-slate-500"
          >{ORG_ROLES_DESCRIPTION_MAP.editor}</span
        >
      </DropdownMenu.Item>
      <DropdownMenu.Item
        class="font-normal flex flex-col items-start hover:bg-slate-50 {role ===
        'viewer'
          ? 'bg-slate-100'
          : ''}"
        on:click={() => handleSetRole(OrgUserRoles.Viewer)}
      >
        <span class="text-xs font-medium text-slate-700">Viewer</span>
        <span class="text-[11px] text-slate-500"
          >{ORG_ROLES_DESCRIPTION_MAP.viewer}</span
        >
      </DropdownMenu.Item>
      <DropdownMenu.Separator />
      <DropdownMenu.Item
        class="font-normal flex items-center py-2"
        on:click={() => {
          if (isGuest) void handleRemove();
          else isRemoveConfirmOpen = true;
        }}
      >
        <span class="text-red-600">Remove</span>
      </DropdownMenu.Item>
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

<RemoveUserFromOrgConfirmDialog
  bind:open={isRemoveConfirmOpen}
  {email}
  onRemove={handleRemove}
/>
