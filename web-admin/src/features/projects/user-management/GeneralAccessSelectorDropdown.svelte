<script lang="ts">
  import {
    createAdminServiceListProjectMemberUsergroups,
    createAdminServiceRemoveProjectMemberUsergroup,
    createAdminServiceAddProjectMemberUsergroup,
    createAdminServiceListUsergroupMemberUsers,
  } from "@rilldata/web-admin/client";
  import { ProjectUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { getAdminServiceListProjectMemberUsergroupsQueryKey } from "@rilldata/web-admin/client";
  import { getRandomBgColor } from "@rilldata/web-common/features/themes/color-config";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Lock from "@rilldata/web-common/components/icons/Lock.svelte";

  export let organization: string;
  export let project: string;

  let open = false;
  let accessDropdownOpen = false;
  let accessType: "everyone" | "invite-only" = "everyone";

  const queryClient = useQueryClient();
  const removeProjectMemberUsergroup =
    createAdminServiceRemoveProjectMemberUsergroup();
  const addProjectMemberUsergroup =
    createAdminServiceAddProjectMemberUsergroup();

  $: listProjectMemberUsergroups =
    createAdminServiceListProjectMemberUsergroups(
      organization,
      project,
      undefined,
      {
        query: {
          enabled: open,
          refetchOnMount: true,
          refetchOnWindowFocus: true,
        },
      },
    );

  $: listUsergroupMemberUsers = createAdminServiceListUsergroupMemberUsers(
    organization,
    "autogroup:members",
    undefined,
    {
      query: {
        enabled: open,
        refetchOnMount: true,
        refetchOnWindowFocus: true,
      },
    },
  );

  $: userGroupMemberUsers = $listUsergroupMemberUsers?.data?.members ?? [];
  $: userGroupMemberUsersCount = userGroupMemberUsers?.length ?? 0;
  $: projectMemberUserGroupsList =
    $listProjectMemberUsergroups.data?.members ?? [];

  async function setAccessInviteOnly() {
    if (accessType === "invite-only") return;

    // Find the autogroup:members user group
    const autogroup = projectMemberUserGroupsList.find(
      (group) => group.groupName === "autogroup:members",
    );

    if (autogroup) {
      // Remove the autogroup:members user group
      await $removeProjectMemberUsergroup.mutateAsync({
        org: organization,
        project,
        usergroup: autogroup.groupName,
      });

      // Invalidate the query to refresh the list
      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectMemberUsergroupsQueryKey(
          organization,
          project,
        ),
      });

      eventBus.emit("notification", {
        message: "Project access changed to invite-only",
      });
    }

    accessType = "invite-only";
    accessDropdownOpen = false;
  }

  async function setAccessEveryone() {
    if (accessType === "everyone") return;

    // Add the autogroup:members user group back with the viewer role
    // This is the default role for autogroup:members as seen in the tests
    await $addProjectMemberUsergroup.mutateAsync({
      org: organization,
      project,
      usergroup: "autogroup:members",
      data: {
        role: ProjectUserRoles.Viewer, // Default role for autogroup:members
      },
    });

    // Invalidate the query to refresh the list
    await queryClient.invalidateQueries({
      queryKey: getAdminServiceListProjectMemberUsergroupsQueryKey(
        organization,
        project,
      ),
    });

    eventBus.emit("notification", {
      message: "Project access changed to everyone",
    });

    accessType = "everyone";
    accessDropdownOpen = false;
  }

  $: hasAutogroupMembers = projectMemberUserGroupsList.some(
    (group) => group.groupName === "autogroup:members",
  );

  $: accessType = hasAutogroupMembers ? "everyone" : "invite-only";

  function getInitials(name: string) {
    return name.charAt(0).toUpperCase();
  }
</script>

<!-- Only users with admin rights can see and use the dropdown selector -->
<DropdownMenu.Root bind:open={accessDropdownOpen}>
  <DropdownMenu.Trigger>
    <div class="flex flex-row items-center gap-x-2">
      <div class="flex items-center gap-2 py-2 pl-2">
        {#if hasAutogroupMembers}
          <div
            class={cn(
              "h-7 w-7 rounded-sm flex items-center justify-center",
              getRandomBgColor(`Everyone at ${organization}`),
            )}
          >
            <span class="text-sm text-white font-semibold"
              >{getInitials(`Everyone at ${organization}`)}</span
            >
          </div>
        {:else}
          <Lock size="28px" color="#374151" />
        {/if}
        <div class="flex flex-col text-left">
          <div class="flex">
            <div
              class="inline-flex flex-row items-center gap-x-1 text-sm font-medium text-gray-900 hover:bg-gray-100 rounded-sm px-1 py-0.5 -mx-1 -my-0.5"
            >
              {#if accessType === "everyone"}
                Everyone at {organization}
              {:else}
                Invite only
              {/if}
              {#if accessDropdownOpen}
                <CaretUpIcon size="12px" color="text-gray-700" />
              {:else}
                <CaretDownIcon size="12px" color="text-gray-700" />
              {/if}
            </div>
          </div>

          {#if accessType === "everyone"}
            <div class="flex flex-row items-center gap-x-1">
              {#if userGroupMemberUsersCount && userGroupMemberUsersCount > 0}
                <span class="text-xs text-gray-500">
                  {userGroupMemberUsersCount} user{userGroupMemberUsersCount > 1
                    ? "s"
                    : ""}
                </span>
              {/if}
            </div>
          {:else}
            <div class="flex flex-row items-center gap-x-1">
              <span class="text-xs text-gray-500">
                Only admins and invited users can access
              </span>
            </div>
          {/if}
        </div>
      </div>
    </div>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" strategy="fixed">
    <DropdownMenu.Item
      on:click={setAccessInviteOnly}
      class="flex flex-col items-start py-2 data-[highlighted]:bg-gray-100 {accessType ===
      'invite-only'
        ? 'bg-gray-50'
        : ''}"
    >
      <div class="flex items-start gap-2">
        <Lock size="20px" color="#374151" />
        <span class="text-xs font-medium text-gray-700">Invite only</span>
      </div>
      <div class="flex flex-row items-center gap-2">
        <div class="w-[20px]" />
        <span class="text-[11px] text-gray-500"
          >Only admins and invited users can access</span
        >
      </div>
    </DropdownMenu.Item>
    <DropdownMenu.Item
      on:click={setAccessEveryone}
      class="flex flex-col items-start py-2 data-[highlighted]:bg-gray-100 {accessType ===
      'everyone'
        ? 'bg-gray-50'
        : ''}"
    >
      <div class="flex items-start gap-2">
        <div
          class="h-5 w-5 flex items-center justify-center bg-primary-600 rounded-sm"
        >
          <span class="text-xs text-white font-semibold"
            >{organization[0].toUpperCase()}</span
          >
        </div>
        <span class="text-xs font-medium text-gray-700"
          >Everyone at {organization}</span
        >
      </div>
      <div class="flex flex-row items-center gap-2">
        <div class="w-[20px]" />
        <span class="text-[11px] text-gray-500">Org members can access</span>
      </div>
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
