<script lang="ts">
  import {
    createAdminServiceRemoveProjectMemberUser,
    createAdminServiceSetProjectMemberUserRole,
    getAdminServiceListProjectInvitesQueryKey,
    getAdminServiceListProjectMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import type {
    V1ProjectMemberUser,
    V1UserInvite,
  } from "@rilldata/web-admin/client";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { capitalize } from "@rilldata/web-common/components/table/utils";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

  type User = V1ProjectMemberUser | V1UserInvite;

  export let organization: string;
  export let project: string;
  export let user: User;
  export let role: string;
  export let isCurrentUser = false;
  export let canChangeRole: boolean;

  let isOpen = false;

  $: console.log("role: ", role);

  const queryClient = useQueryClient();
  // const listProjectMemberUsers = createAdminServiceListProjectMemberUsers(
  //   organization,
  //   project,
  //   undefined,
  //   {
  //     query: {
  //       refetchOnMount: true,
  //       refetchOnWindowFocus: true,
  //     },
  //   },
  // );

  $: setProjectMemberUserRole = createAdminServiceSetProjectMemberUserRole();
  $: removeProjectMemberUser = createAdminServiceRemoveProjectMemberUser();
  // $: projectMemberUsersList = $listProjectMemberUsers.data?.members ?? [];

  function getUserEmail(user: User): string {
    if ("userEmail" in user) return user.userEmail;
    if ("email" in user) return user.email;
    return "";
  }

  function getUserRole(user: User): string {
    if ("role" in user && user.role) return user.role;
    if ("roleName" in user) return user.roleName;
    return "";
  }

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
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
    } catch (_) {
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
    } catch (error) {
      console.error("Error removing user", error);
      eventBus.emit("notification", {
        message: "Error removing user",
        type: "error",
      });
    }
  }
</script>

{#if canChangeRole}
  <DropdownMenu.Root bind:open={isOpen}>
    <DropdownMenu.Trigger
      class="w-18 flex flex-row gap-1 items-center rounded-sm mr-[10px] {isOpen
        ? 'bg-slate-200'
        : 'hover:bg-slate-100'} px-2 py-1"
    >
      {capitalize(getUserRole(user))}
      {#if isOpen}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start" strategy="fixed">
      <!-- TODO: what happens when admin removes themselves as admin? -->
      {#if role === "admin"}
        <DropdownMenu.Item
          class="font-normal flex flex-col items-start py-2 {getUserRole(
            user,
          ) === 'admin'
            ? 'bg-slate-100'
            : ''}"
          on:click={() => handleSetRole(getUserEmail(user), "admin")}
        >
          <span class="font-medium">Admin</span>
          <span class="text-xs text-gray-600"
            >Full control of project settings and members</span
          >
        </DropdownMenu.Item>
      {/if}

      <DropdownMenu.Item
        class="font-normal flex flex-col items-start py-2 {getUserRole(user) ===
        'editor'
          ? 'bg-slate-100'
          : ''}"
        on:click={() => handleSetRole(getUserEmail(user), "editor")}
      >
        <span class="font-medium">Editor</span>
        <span class="text-xs text-gray-600"
          >Can create and edit dashboards; manage non-admin access</span
        >
      </DropdownMenu.Item>

      <DropdownMenu.Item
        class="font-normal flex flex-col items-start py-2 {getUserRole(user) ===
        'viewer'
          ? 'bg-slate-100'
          : ''}"
        on:click={() => handleSetRole(getUserEmail(user), "viewer")}
      >
        <span class="font-medium">Viewer</span>
        <span class="text-xs text-gray-600"
          >Read-only access to project dashboards</span
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
    class="w-18 flex flex-row gap-1 items-center rounded-sm px-2 py-1 mr-[10px]"
  >
    <span>{capitalize(getUserRole(user))}</span>
  </div>
{/if}
