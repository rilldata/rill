<script lang="ts">
  import {
    createAdminServiceRemoveProjectMemberUsergroup,
    createAdminServiceSetProjectMemberUsergroupRole,
    getAdminServiceListProjectMemberUsergroupsQueryKey,
  } from "@rilldata/web-admin/client";
  import type { V1MemberUsergroup } from "@rilldata/web-admin/client";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { capitalize } from "@rilldata/web-common/components/table/utils";
  import { ProjectUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { PROJECT_ROLES_DESCRIPTION_MAP } from "../constants";

  export let organization: string;
  export let project: string;
  export let group: V1MemberUsergroup;

  let isOpen = false;

  const queryClient = useQueryClient();
  const setProjectMemberUserGroupRole =
    createAdminServiceSetProjectMemberUsergroupRole();
  const removeProjectMemberUsergroup =
    createAdminServiceRemoveProjectMemberUsergroup();

  async function handleSetRole(groupName: string, role: string) {
    try {
      await $setProjectMemberUserGroupRole.mutateAsync({
        org: organization,
        project: project,
        usergroup: groupName,
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
        message: "User group role updated",
      });
    } catch (error) {
      eventBus.emit("notification", {
        message: `Error: ${error.response.data.message}`,
        type: "error",
      });
    }
  }

  async function handleRemove(groupName: string) {
    try {
      await $removeProjectMemberUsergroup.mutateAsync({
        org: organization,
        project: project,
        usergroup: groupName,
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectMemberUsergroupsQueryKey(
          organization,
          project,
        ),
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
    class="flex flex-row gap-1 items-center rounded-sm mr-[10px] w-[72px] text-right {isOpen
      ? 'bg-slate-200'
      : 'hover:bg-slate-100'} px-2 py-1"
  >
    {capitalize(group.roleName)}
    {#if isOpen}
      <CaretUpIcon size="12px" />
    {:else}
      <CaretDownIcon size="12px" />
    {/if}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" strategy="fixed">
    <DropdownMenu.Item
      class="font-normal flex flex-col items-start py-2 {group.roleName ===
      'admin'
        ? 'bg-slate-100'
        : ''}"
      on:click={() => handleSetRole(group.groupName, ProjectUserRoles.Admin)}
    >
      <span class="font-medium">Admin</span>
      <span class="text-xs text-gray-600"
        >{PROJECT_ROLES_DESCRIPTION_MAP.admin}</span
      >
    </DropdownMenu.Item>
    <DropdownMenu.Item
      class="font-normal flex flex-col items-start py-2 {group.roleName ===
      'editor'
        ? 'bg-slate-100'
        : ''}"
      on:click={() => handleSetRole(group.groupName, ProjectUserRoles.Editor)}
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
      on:click={() => handleSetRole(group.groupName, ProjectUserRoles.Viewer)}
    >
      <span class="font-medium">Viewer</span>
      <span class="text-xs text-gray-600"
        >{PROJECT_ROLES_DESCRIPTION_MAP.viewer}</span
      >
    </DropdownMenu.Item>
    <DropdownMenu.Separator />
    <DropdownMenu.Item
      class="font-normal flex items-center"
      on:click={() => handleRemove(group.groupName)}
    >
      <span class="text-red-600">Remove</span>
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
