<script lang="ts">
  import {
    createAdminServiceRemoveProjectMemberUser,
    createAdminServiceSetProjectMemberUserRole,
    getAdminServiceListProjectInvitesQueryKey,
    getAdminServiceListProjectMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import type { V1MemberUser } from "@rilldata/web-admin/client";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { capitalize } from "@rilldata/web-common/components/table/utils";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

  type User =
    | V1MemberUser
    | {
        userName: any;
        userEmail: string;
        userPhotoUrl?: string;
        roleName: string;
        email?: string;
        role?: string;
        invitedBy?: string;
      };

  export let organization: string;
  export let project: string;
  export let user: User;
  export let isCurrentUser = false;
  export let pendingAcceptance: boolean = false;

  let isOpen = false;

  const queryClient = useQueryClient();

  $: setProjectMemberUserRole = createAdminServiceSetProjectMemberUserRole();
  $: removeProjectMemberUser = createAdminServiceRemoveProjectMemberUser();

  async function handleSetRole(email: string, role: string) {
    try {
      await $setProjectMemberUserRole.mutateAsync({
        organization: organization,
        project: project,
        email: email,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListProjectMemberUsersQueryKey(organization, project),
      );

      await queryClient.invalidateQueries(
        getAdminServiceListProjectInvitesQueryKey(organization, project),
      );

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

  async function handleRemove(email: string) {
    try {
      await $removeProjectMemberUser.mutateAsync({
        organization: organization,
        project: project,
        email: email,
      });

      await queryClient.invalidateQueries(
        getAdminServiceListProjectMemberUsersQueryKey(organization, project),
      );

      await queryClient.invalidateQueries(
        getAdminServiceListProjectInvitesQueryKey(organization, project),
      );

      eventBus.emit("notification", {
        message: "User removed",
      });
    } catch (error) {
      console.error("Error removing user", error);
      eventBus.emit("notification", {
        message: "Error removing user",
        type: "error",
      });
    }
  }
</script>

{#if !isCurrentUser && !pendingAcceptance}
  <DropdownMenu.Root bind:open={isOpen}>
    <DropdownMenu.Trigger
      class="w-18 flex flex-row gap-1 items-center rounded-sm mr-[10px] {isOpen
        ? 'bg-slate-200'
        : 'hover:bg-slate-100'} px-2 py-1"
    >
      {capitalize(user.roleName)}
      {#if isOpen}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      <DropdownMenu.CheckboxItem
        class="font-normal flex items-center"
        checked={user.roleName === "admin"}
        on:click={() => {
          handleSetRole(user.userEmail, "admin");
        }}
      >
        <span>Admin</span>
      </DropdownMenu.CheckboxItem>
      <DropdownMenu.CheckboxItem
        class="font-normal flex items-center"
        checked={user.roleName === "editor"}
        on:click={() => {
          handleSetRole(user.userEmail, "editor");
        }}
      >
        <span>Editor</span>
      </DropdownMenu.CheckboxItem>
      <DropdownMenu.CheckboxItem
        class="font-normal flex items-center"
        checked={user.roleName === "viewer"}
        on:click={() => {
          handleSetRole(user.userEmail, "viewer");
        }}
      >
        <span>Viewer</span>
      </DropdownMenu.CheckboxItem>
      {#if !isCurrentUser}
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() => {
            handleRemove(user.userEmail);
          }}
        >
          <span class="ml-6 text-red-600">Remove</span>
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{:else}
  <div
    class="w-18 flex flex-row gap-1 items-center rounded-sm px-2 py-1 mr-[10px]"
  >
    <span>{capitalize(user.roleName)}</span>
  </div>
{/if}
