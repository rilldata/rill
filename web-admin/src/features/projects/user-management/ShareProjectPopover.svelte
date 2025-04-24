<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceListProjectInvites,
    createAdminServiceListProjectMemberUsergroups,
    createAdminServiceListProjectMemberUsers,
    createAdminServiceRemoveProjectMemberUsergroup,
    createAdminServiceAddProjectMemberUsergroup,
    createAdminServiceGetCurrentUser,
  } from "@rilldata/web-admin/client";
  import CopyInviteLinkButton from "@rilldata/web-admin/features/projects/user-management/CopyInviteLinkButton.svelte";
  import UserInviteForm from "@rilldata/web-admin/features/projects/user-management/UserInviteForm.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import AutogroupMembersItem from "./AutogroupMembersItem.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import UserItem from "./UserItem.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { getAdminServiceListProjectMemberUsergroupsQueryKey } from "@rilldata/web-admin/client";

  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Lock from "@rilldata/web-common/components/icons/Lock.svelte";

  export let organization: string;
  export let project: string;
  export let isAdmin: boolean;
  export let isEditor: boolean;

  let open = false;
  let accessDropdownOpen = false;
  let accessType: "everyone" | "invite-only" = "everyone";

  const queryClient = useQueryClient();
  const removeProjectMemberUsergroup =
    createAdminServiceRemoveProjectMemberUsergroup();
  const addProjectMemberUsergroup =
    createAdminServiceAddProjectMemberUsergroup();
  const currentUser = createAdminServiceGetCurrentUser();

  async function setAccessInviteOnly() {
    if (accessType === "invite-only") return;

    // Find the autogroup:members user group
    const autogroup = projectMemberUserGroupsList.find(
      (group) => group.groupName === "autogroup:members",
    );

    if (autogroup) {
      // Remove the autogroup:members user group
      await $removeProjectMemberUsergroup.mutateAsync({
        organization,
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
      organization,
      project,
      usergroup: "autogroup:members",
      data: {
        role: "viewer", // Default role for autogroup:members
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

  $: copyLink = `${$page.url.protocol}//${$page.url.host}/${organization}/${project}`;

  // viewer: "not allowed to list project user groups"
  $: listProjectMemberUsergroups =
    createAdminServiceListProjectMemberUsergroups(
      organization,
      project,
      undefined,
      {
        query: {
          enabled: isAdmin,
          refetchOnMount: true,
          refetchOnWindowFocus: true,
        },
      },
    );

  $: listProjectMemberUsers = createAdminServiceListProjectMemberUsers(
    organization,
    project,
    undefined,
    {
      query: {
        refetchOnMount: true,
        refetchOnWindowFocus: true,
      },
    },
  );

  // viewer: "not authorized to read project members"
  $: listProjectInvites = createAdminServiceListProjectInvites(
    organization,
    project,
    undefined,
    {
      query: {
        enabled: isAdmin,
        refetchOnMount: true,
        refetchOnWindowFocus: true,
      },
    },
  );

  $: projectMemberUserGroupsList =
    $listProjectMemberUsergroups.data?.members ?? [];
  $: projectMemberUsersList = $listProjectMemberUsers.data?.members ?? [];
  $: projectInvitesList = $listProjectInvites.data?.invites ?? [];

  // Sort the list to prioritize the current user
  $: sortedProjectMemberUsersList = projectMemberUsersList.sort((a, b) => {
    if (a.userEmail === $currentUser.data?.user?.email) return -1;
    if (b.userEmail === $currentUser.data?.user?.email) return 1;
    return 0;
  });

  $: hasAutogroupMembers = projectMemberUserGroupsList.some(
    (group) => group.groupName === "autogroup:members",
  );

  // $: hasRegularUserGroups = projectMemberUserGroupsList.some(
  //   (group) => !group.groupManaged,
  // );

  $: accessType = hasAutogroupMembers ? "everyone" : "invite-only";
</script>

<Popover bind:open>
  <PopoverTrigger asChild let:builder>
    <Button builders={[builder]} type="secondary" selected={open}>Share</Button>
  </PopoverTrigger>
  <PopoverContent align="end" class="w-[520px]" padding="0">
    <div class="flex flex-col p-4">
      <div class="flex flex-row items-center mb-4">
        <div class="text-sm font-medium">Share project: {project}</div>
        <div class="grow"></div>
      </div>
      {#if isAdmin || isEditor}
        <UserInviteForm {organization} {project} />
      {/if}
      <!-- 52 * 8 = 416px -->
      <div class="flex flex-col gap-y-1 overflow-y-auto max-h-[416px]">
        <div class={isAdmin ? "mt-4" : ""}>
          <!-- Project Users -->
          {#each sortedProjectMemberUsersList as user}
            <UserItem
              {organization}
              {project}
              {user}
              canChangeRole={isAdmin || isEditor}
            />
          {/each}
          <!-- Pending Invites -->
          {#each projectInvitesList as user}
            <UserItem
              {organization}
              {project}
              {user}
              canChangeRole={isAdmin || isEditor}
            />
          {/each}
          <!-- User Groups -->
          <!-- TODO: revisit when https://www.notion.so/rilldata/User-Management-Role-Based-Access-Control-RBAC-Enhancements-8d331b29d9b64d87bca066e06ef87f54?pvs=4#1acba33c8f5780f38303f01a73e82e60 -->
          <!-- {#if hasRegularUserGroups}
            {#each projectMemberUserGroupsList as group}
              {#if !group.groupManaged}
                <AutogroupMembersItem
                  {organization}
                  {project}
                  {group}
                  avatarName={`Everyone at ${organization}`}
                  {isAdmin}
                />
              {/if}
            {/each}
          {/if} -->
        </div>
        {#if hasAutogroupMembers}
          <div class="mt-2">
            <div class="text-xs text-gray-500 font-semibold uppercase">
              General Access
            </div>
            <!-- NOTE: Only support "autogroup:members" -->
            <!-- https://www.notion.so/rilldata/User-Management-Role-Based-Access-Control-RBAC-Enhancements-8d331b29d9b64d87bca066e06ef87f54?pvs=4#1acba33c8f5780f4903bf16510193dd8 -->
            {#each projectMemberUserGroupsList as group}
              {#if group.groupName === "autogroup:members"}
                <AutogroupMembersItem
                  {organization}
                  {project}
                  {group}
                  avatarName={`Everyone at ${organization}`}
                  {isAdmin}
                />
              {/if}
            {/each}
          </div>
        {/if}
      </div>
    </div>
    <div
      class="flex flex-row items-center px-3.5 py-3 border-t border-gray-200"
    >
      {#if isAdmin}
        <DropdownMenu.Root bind:open={accessDropdownOpen}>
          <DropdownMenu.Trigger>
            <Button
              type="secondary"
              class="flex flex-row items- gap-2"
              forcedStyle="min-height: 28px !important; height: 28px !important; border-color: #D1D5DB !important;"
            >
              {#if accessType === "everyone"}
                <div
                  class="h-4 w-4 flex items-center justify-center bg-primary-600 rounded-sm"
                >
                  <span class="text-[10px] text-white font-semibold"
                    >{organization[0].toUpperCase()}</span
                  >
                </div>
              {:else}
                <Lock size="18px" />
              {/if}
              <span class="text-sm font-medium text-gray-900">
                {#if accessType === "everyone"}
                  Everyone at {organization}
                {:else}
                  Invite only
                {/if}
              </span>
              {#if accessDropdownOpen}
                <CaretUpIcon size="12px" color="text-gray-700" />
              {:else}
                <CaretDownIcon size="12px" color="text-gray-700" />
              {/if}
            </Button>
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="end">
            <DropdownMenu.Item
              on:click={setAccessInviteOnly}
              class="flex flex-col items-start py-2 data-[highlighted]:bg-gray-100 {accessType ===
              'invite-only'
                ? 'bg-gray-50'
                : ''}"
            >
              <div class="flex items-start gap-2">
                <Lock size="20px" color="#374151" />
                <span class="text-xs font-medium text-gray-700"
                  >Invite only</span
                >
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
                  class="h-5 w-5 flex items-center justify-center bg-primary-600"
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
                <span class="text-[11px] text-gray-500"
                  >Org members can access</span
                >
              </div>
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      {:else}
        <a
          href="https://docs.rilldata.com/manage/roles-permissions#project-level-permissions"
          target="_blank"
          class="text-xs text-primary-600">Learn more about sharing</a
        >
      {/if}
      <div class="grow"></div>
      <CopyInviteLinkButton {copyLink} />
    </div>
  </PopoverContent>
</Popover>
