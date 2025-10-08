<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { page } from "$app/stores";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { useQueryClient } from "@tanstack/svelte-query";
  import {
    createAdminServiceAddOrganizationMemberUsergroup,
    createAdminServiceRemoveOrganizationMemberUsergroup,
    createAdminServiceSetOrganizationMemberUsergroupRole,
    getAdminServiceListOrganizationMemberUsergroupsQueryKey,
  } from "@rilldata/web-admin/client";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { capitalize } from "@rilldata/web-common/components/table/utils.ts";
  import { ORG_ROLES_DESCRIPTION_MAP } from "@rilldata/web-admin/features/organizations/user-management/constants.ts";

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

  async function handleRoleSelect(selectedRole: string) {
    if (role) {
      return handleSetRole(selectedRole);
    } else {
      return handleAddRole(selectedRole);
    }
  }

  async function handleAddRole(role: string) {
    try {
      await $addUserGroupRole.mutateAsync({
        org: organization,
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
        org: organization,
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
        org: organization,
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
  <DropdownMenu.Content align="start" strategy="fixed" class="w-[200px]">
    {#if manageOrgAdmins}
      <DropdownMenu.Item
        class="font-normal flex flex-col items-start py-2 {role === 'admin'
          ? 'bg-slate-100'
          : ''}"
        on:click={() => handleRoleSelect(OrgUserRoles.Admin)}
      >
        <span class="font-medium">Admin</span>
        <span class="text-xs text-gray-600"
          >{ORG_ROLES_DESCRIPTION_MAP.admin}</span
        >
      </DropdownMenu.Item>
    {/if}
    <DropdownMenu.Item
      class="font-normal flex flex-col items-start py-2 {role === 'editor'
        ? 'bg-slate-100'
        : ''}"
      on:click={() => handleRoleSelect(OrgUserRoles.Editor)}
    >
      <span class="font-medium">Editor</span>
      <span class="text-xs text-gray-600"
        >{ORG_ROLES_DESCRIPTION_MAP.editor}</span
      >
    </DropdownMenu.Item>
    <DropdownMenu.Item
      class="font-normal flex flex-col items-start py-2 {role === 'viewer'
        ? 'bg-slate-100'
        : ''}"
      on:click={() => handleRoleSelect(OrgUserRoles.Viewer)}
    >
      <span class="font-medium">Viewer</span>
      <span class="text-xs text-gray-600"
        >{ORG_ROLES_DESCRIPTION_MAP.viewer}</span
      >
    </DropdownMenu.Item>
    {#if role}
      <DropdownMenu.Separator />
      <DropdownMenu.Item
        class="font-normal flex items-center"
        on:click={handleRevokeRole}
      >
        <span>Remove</span>
      </DropdownMenu.Item>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
