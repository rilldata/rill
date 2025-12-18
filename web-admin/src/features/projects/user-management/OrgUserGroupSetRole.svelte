<script lang="ts">
  import {
    createAdminServiceRemoveOrganizationMemberUsergroup,
    createAdminServiceSetOrganizationMemberUsergroupRole,
    createAdminServiceAddOrganizationMemberUsergroup,
    getAdminServiceListOrganizationMemberUsergroupsQueryKey,
  } from "@rilldata/web-admin/client";
  import type { V1MemberUsergroup } from "@rilldata/web-admin/client";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { capitalize } from "@rilldata/web-common/components/table/utils";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { PROJECT_ROLES_DESCRIPTION_MAP } from "../constants";

  export let organization: string;
  export let group: V1MemberUsergroup;
  export let manageOrgAdmins: boolean;

  let isOpen = false;

  const queryClient = useQueryClient();
  const addUserGroupRole = createAdminServiceAddOrganizationMemberUsergroup();
  const setUserGroupRole =
    createAdminServiceSetOrganizationMemberUsergroupRole();
  const removeOrganizationMemberUsergroup =
    createAdminServiceRemoveOrganizationMemberUsergroup();

  async function handleRoleSelect(selectedRole: string) {
    if (group.roleName) {
      return handleSetRole(selectedRole);
    } else {
      return handleAddRole(selectedRole);
    }
  }

  async function handleAddRole(role: string) {
    try {
      await $addUserGroupRole.mutateAsync({
        org: organization,
        usergroup: group.groupName,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      });

      eventBus.emit("notification", {
        message: "User group role added",
      });
    } catch (error) {
      eventBus.emit("notification", {
        message: `Error: ${error.response.data.message}`,
        type: "error",
      });
    }
  }

  async function handleSetRole(role: string) {
    try {
      await $setUserGroupRole.mutateAsync({
        org: organization,
        usergroup: group.groupName,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      });

      eventBus.emit("notification", {
        message: "User group role updated",
      });
    } catch (error) {
      eventBus.emit("notification", {
        message: `Error: ${error.response.data.message}`,
        type: "error",
      });
    }
  }

  async function handleRemove() {
    try {
      await $removeOrganizationMemberUsergroup.mutateAsync({
        org: organization,
        usergroup: group.groupName,
      });

      await queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      });

      eventBus.emit("notification", {
        message: "User group removed",
      });
    } catch (error) {
      eventBus.emit("notification", {
        message: `Error: ${error.response.data.message}`,
        type: "error",
      });
    }
  }
</script>

<DropdownMenu.Root bind:open={isOpen}>
  <DropdownMenu.Trigger
    class="w-18 flex flex-row gap-1 items-center rounded-sm mr-[10px] w-[72px] text-right {isOpen
      ? 'bg-slate-200'
      : 'hover:bg-slate-100'} px-2 py-1"
    disabled={!manageOrgAdmins && group.roleName === OrgUserRoles.Admin}
  >
    {group.roleName ? capitalize(group.roleName) : "-"}
    {#if !(!manageOrgAdmins && group.roleName === OrgUserRoles.Admin)}
      {#if isOpen}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    {/if}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" strategy="fixed">
    {#if manageOrgAdmins}
      <DropdownMenu.Item
        class="font-normal flex flex-col items-start py-2 {group.roleName ===
        'admin'
          ? 'bg-slate-100'
          : ''}"
        on:click={() => handleRoleSelect(OrgUserRoles.Admin)}
      >
        <span class="font-medium">Admin</span>
        <span class="text-xs text-gray-600"
          >{PROJECT_ROLES_DESCRIPTION_MAP.admin}</span
        >
      </DropdownMenu.Item>
    {/if}

    <DropdownMenu.Item
      class="font-normal flex flex-col items-start py-2 {group.roleName ===
      'editor'
        ? 'bg-slate-100'
        : ''}"
      on:click={() => handleRoleSelect(OrgUserRoles.Editor)}
    >
      <span class="font-medium">Editor</span>
      <span class="text-xs text-gray-600"
        >{PROJECT_ROLES_DESCRIPTION_MAP.editor}</span
      >
    </DropdownMenu.Item>

    <DropdownMenu.Item
      class="font-normal flex flex-col items-start py-2 {group.roleName ===
      'viewer'
        ? 'bg-slate-100'
        : ''}"
      on:click={() => handleRoleSelect(OrgUserRoles.Viewer)}
    >
      <span class="font-medium">Viewer</span>
      <span class="text-xs text-gray-600"
        >{PROJECT_ROLES_DESCRIPTION_MAP.viewer}</span
      >
    </DropdownMenu.Item>

    {#if group.roleName}
      <DropdownMenu.Separator />
      <DropdownMenu.Item
        class="font-normal flex items-center py-2"
        on:click={handleRemove}
      >
        <span class="text-red-600">Remove</span>
      </DropdownMenu.Item>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
