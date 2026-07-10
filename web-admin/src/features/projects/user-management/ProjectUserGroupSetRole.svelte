<script lang="ts">
  import {
    createAdminServiceSetProjectMemberUsergroupRole,
    createAdminServiceAddProjectMemberUsergroup,
    getAdminServiceListProjectMemberUsergroupsQueryKey,
    createAdminServiceRemoveProjectMemberUsergroup,
  } from "@rilldata/web-admin/client";
  import type { V1MemberUsergroup } from "@rilldata/web-admin/client";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { capitalize } from "@rilldata/web-common/components/table/utils";
  import { ProjectUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

  export let organization: string;
  export let group: V1MemberUsergroup;
  export let project: string;
  export let manageOrgAdmins: boolean;

  let isOpen = false;

  const queryClient = useQueryClient();
  const addProjectMemberUsergroup =
    createAdminServiceAddProjectMemberUsergroup();
  const setProjectMemberUsergroupRole =
    createAdminServiceSetProjectMemberUsergroupRole();
  const removeProjectMemberUsergroup =
    createAdminServiceRemoveProjectMemberUsergroup();

  async function handleRoleSelect(selectedRole: string) {
    if (group.roleName) {
      return handleSetRole(selectedRole);
    } else {
      return handleAddRole(selectedRole);
    }
  }

  async function handleAddRole(role: string) {
    try {
      await $addProjectMemberUsergroup.mutateAsync({
        org: organization,
        project: project,
        usergroup: group.groupName,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectMemberUsergroupsQueryKey(
          organization,
          project,
        ),
      });

      eventBus.emit("notification", {
        message: m.groups_role_added(),
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
      await $setProjectMemberUsergroupRole.mutateAsync({
        org: organization,
        project: project,
        usergroup: group.groupName,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectMemberUsergroupsQueryKey(
          organization,
          project,
        ),
      });

      eventBus.emit("notification", {
        message: m.groups_role_updated(),
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
      await $removeProjectMemberUsergroup.mutateAsync({
        org: organization,
        project: project,
        usergroup: group.groupName,
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectMemberUsergroupsQueryKey(
          organization,
          project,
        ),
      });

      eventBus.emit("notification", {
        message: m.groups_removed(),
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
    class="w-18 flex flex-row gap-1 items-center rounded-sm mr-[10px] {isOpen
      ? 'bg-gray-200'
      : 'hover:bg-surface-hover'} px-2 py-1"
    disabled={!manageOrgAdmins && group.roleName === ProjectUserRoles.Admin}
  >
    {group.roleName ? capitalize(group.roleName) : "-"}
    {#if !(!manageOrgAdmins && group.roleName === ProjectUserRoles.Admin)}
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
          ? 'bg-gray-100'
          : ''}"
        onclick={() => handleRoleSelect(ProjectUserRoles.Admin)}
      >
        <span class="font-medium">{m.project_share_role_admin()}</span>
        <span class="text-xs text-fg-secondary"
          >{m.project_share_role_admin_description()}</span
        >
      </DropdownMenu.Item>
    {/if}

    <DropdownMenu.Item
      class="font-normal flex flex-col items-start py-2 {group.roleName ===
      'editor'
        ? 'bg-gray-100'
        : ''}"
      onclick={() => handleRoleSelect(ProjectUserRoles.Editor)}
    >
      <span class="font-medium">{m.project_share_role_editor()}</span>
      <span class="text-xs text-fg-secondary"
        >{m.project_share_role_editor_description()}</span
      >
    </DropdownMenu.Item>

    <DropdownMenu.Item
      class="font-normal flex flex-col items-start py-2 {group.roleName ===
      'viewer'
        ? 'bg-gray-100'
        : ''}"
      onclick={() => handleRoleSelect(ProjectUserRoles.Viewer)}
    >
      <span class="font-medium">{m.project_share_role_viewer()}</span>
      <span class="text-xs text-fg-secondary"
        >{m.project_share_role_viewer_description()}</span
      >
    </DropdownMenu.Item>

    {#if group.roleName}
      <DropdownMenu.Separator />
      <DropdownMenu.Item
        class="font-normal flex items-center py-2"
        onclick={handleRemove}
      >
        <span class="text-red-600">{m.users_remove()}</span>
      </DropdownMenu.Item>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
