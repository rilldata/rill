<script lang="ts">
  import {
    createAdminServiceRemoveProjectMemberUser,
    createAdminServiceSetProjectMemberUserRole,
    getAdminServiceListProjectInvitesQueryKey,
    getAdminServiceListProjectMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import type {
    V1ProjectMemberUser,
    V1ProjectInvite,
  } from "@rilldata/web-admin/client";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { capitalize } from "@rilldata/web-common/components/table/utils";
  import { ProjectUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { PROJECT_ROLES_DESCRIPTION_MAP } from "../constants";

  type User = V1ProjectMemberUser | V1ProjectInvite;

  export let organization: string;
  export let project: string;
  export let user: User;
  export let isCurrentUser = false;
  export let manageProjectMembers: boolean;
  export let manageProjectAdmins: boolean;

  let isOpen = false;

  const queryClient = useQueryClient();
  const setProjectMemberUserRole = createAdminServiceSetProjectMemberUserRole();
  const removeProjectMemberUser = createAdminServiceRemoveProjectMemberUser();

  function getUserEmail(user: User): string {
    if ("userEmail" in user) return user.userEmail;
    if ("email" in user) return user.email;
    return "";
  }

  function getUserRole(user: User): string {
    if ("roleName" in user) return user.roleName;
    return "";
  }

  async function handleSetRole(email: string, role: string) {
    try {
      await $setProjectMemberUserRole.mutateAsync({
        org: organization,
        project: project,
        email: email,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectMemberUsersQueryKey(
          organization,
          project,
        ),
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectInvitesQueryKey(
          organization,
          project,
        ),
      });

      eventBus.emit("notification", {
        message: "User role updated",
      });
    } catch {
      eventBus.emit("notification", {
        message: "Error updating user role",
        type: "error",
      });
    }
  }

  async function handleRemove(email: string) {
    try {
      await $removeProjectMemberUser.mutateAsync({
        org: organization,
        project: project,
        email: email,
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectMemberUsersQueryKey(
          organization,
          project,
        ),
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectInvitesQueryKey(
          organization,
          project,
        ),
      });

      eventBus.emit("notification", {
        message: "User removed",
      });
    } catch {
      eventBus.emit("notification", {
        message: "Error removing user",
        type: "error",
      });
    }
  }
</script>

{#if manageProjectMembers && !isCurrentUser}
  <DropdownMenu.Root bind:open={isOpen}>
    <DropdownMenu.Trigger
      class="flex flex-row gap-1 items-center rounded-sm mr-[10px] w-[72px] text-right {isOpen
        ? 'bg-slate-200'
        : 'hover:bg-slate-100'} px-2 py-1"
    >
      {capitalize(getUserRole(user))}
      {#if !(!manageProjectAdmins && getUserRole(user) === ProjectUserRoles.Admin)}
        {#if isOpen}
          <CaretUpIcon size="12px" />
        {:else}
          <CaretDownIcon size="12px" />
        {/if}
      {/if}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start" strategy="fixed">
      {#if manageProjectAdmins}
        <DropdownMenu.Item
          class="font-normal flex flex-col items-start py-2 {getUserRole(
            user,
          ) === 'admin'
            ? 'bg-slate-100'
            : ''}"
          on:click={() =>
            handleSetRole(getUserEmail(user), ProjectUserRoles.Admin)}
        >
          <span class="font-medium">Admin</span>
          <span class="text-xs text-gray-600"
            >{PROJECT_ROLES_DESCRIPTION_MAP.admin}</span
          >
        </DropdownMenu.Item>
      {/if}

      <DropdownMenu.Item
        class="font-normal flex flex-col items-start py-2 {getUserRole(user) ===
        'editor'
          ? 'bg-slate-100'
          : ''}"
        on:click={() =>
          handleSetRole(getUserEmail(user), ProjectUserRoles.Editor)}
      >
        <span class="font-medium">Editor</span>
        <span class="text-xs text-gray-600"
          >{PROJECT_ROLES_DESCRIPTION_MAP.editor}</span
        >
      </DropdownMenu.Item>

      <DropdownMenu.Item
        class="font-normal flex flex-col items-start py-2 {getUserRole(user) ===
        'viewer'
          ? 'bg-slate-100'
          : ''}"
        on:click={() =>
          handleSetRole(getUserEmail(user), ProjectUserRoles.Viewer)}
      >
        <span class="font-medium">Viewer</span>
        <span class="text-xs text-gray-600"
          >{PROJECT_ROLES_DESCRIPTION_MAP.viewer}</span
        >
      </DropdownMenu.Item>

      {#if !isCurrentUser}
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          class="font-normal flex items-center py-2"
          on:click={() => handleRemove(getUserEmail(user))}
        >
          <span class="text-red-600">Remove</span>
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{:else}
  <div
    class="flex flex-row gap-1 items-center rounded-sm px-2 py-1 mr-[10px] w-[72px] text-right"
  >
    <span>{capitalize(getUserRole(user))}</span>
  </div>
{/if}
