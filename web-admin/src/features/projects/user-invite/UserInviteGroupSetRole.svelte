<script lang="ts">
  import {
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
</script>

<!-- `all-users` is a special group that cannot be deleted or edited -->
{#if group.groupName !== "all-users"}
  <DropdownMenu.Root bind:open={isOpen}>
    <DropdownMenu.Trigger
      class="w-18 flex flex-row gap-1 items-center rounded-sm {isOpen
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
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
