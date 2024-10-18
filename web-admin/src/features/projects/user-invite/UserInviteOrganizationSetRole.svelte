<script lang="ts">
  import {
    createAdminServiceRemoveProjectMemberUsergroup,
    createAdminServiceSetProjectMemberUsergroupRole,
    getAdminServiceListProjectMemberUsergroupsQueryKey,
  } from "@rilldata/web-admin/client";
  import type { V1MemberUsergroup } from "@rilldata/web-admin/client";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { capitalize } from "@rilldata/web-common/components/table/utils";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

  export let organization: string;
  export let project: string;
  export let group: V1MemberUsergroup;

  let isOpen = false;

  const queryClient = useQueryClient();

  $: setProjectMemberUserGroupRole =
    createAdminServiceSetProjectMemberUsergroupRole();
  $: removeProjectMemberUsergroup =
    createAdminServiceRemoveProjectMemberUsergroup();

  async function handleSetRole(groupName: string, role: string) {
    try {
      await $setProjectMemberUserGroupRole.mutateAsync({
        organization: organization,
        project: project,
        usergroup: groupName,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListProjectMemberUsergroupsQueryKey(
          organization,
          project,
        ),
      );

      eventBus.emit("notification", {
        message: "User group role updated",
      });
    } catch (error) {
      console.error("Error updating user group role", error);
      eventBus.emit("notification", {
        message: "Error updating user group role",
        type: "error",
      });
    }
  }

  async function handleRemove(groupName: string) {
    try {
      await $removeProjectMemberUsergroup.mutateAsync({
        organization: organization,
        project: project,
        usergroup: groupName,
      });

      await queryClient.invalidateQueries(
        getAdminServiceListProjectMemberUsergroupsQueryKey(
          organization,
          project,
        ),
      );

      eventBus.emit("notification", {
        message: "User group removed",
      });
    } catch (error) {
      console.error("Error removing user group", error);
      eventBus.emit("notification", {
        message: "Error removing user group",
        type: "error",
      });
    }
  }
</script>

<DropdownMenu.Root bind:open={isOpen}>
  <DropdownMenu.Trigger
    class="w-18 flex flex-row gap-1 items-center rounded-sm mr-[10px] {isOpen
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
  <DropdownMenu.Content align="start">
    <DropdownMenu.CheckboxItem
      class="font-normal flex items-center"
      checked={group.roleName === "admin"}
      on:click={() => {
        handleSetRole(group.groupName, "admin");
      }}
    >
      <span>Admin</span>
    </DropdownMenu.CheckboxItem>
    <DropdownMenu.CheckboxItem
      class="font-normal flex items-center"
      checked={group.roleName === "viewer"}
      on:click={() => {
        handleSetRole(group.groupName, "viewer");
      }}
    >
      <span>Viewer</span>
    </DropdownMenu.CheckboxItem>
    <DropdownMenu.Separator />
    <DropdownMenu.Item
      class="font-normal flex items-center"
      on:click={() => {
        handleRemove(group.groupName);
      }}
    >
      <span class="ml-6">Remove</span>
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
