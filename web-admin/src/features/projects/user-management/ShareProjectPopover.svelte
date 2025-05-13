<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceListProjectInvites,
    createAdminServiceListProjectMemberUsergroups,
    createAdminServiceListProjectMemberUsers,
    createAdminServiceRemoveProjectMemberUsergroup,
    createAdminServiceAddProjectMemberUsergroup,
    createAdminServiceGetCurrentUser,
    createAdminServiceListUsergroupMemberUsers,
    createAdminServiceListOrganizationMemberUsergroups,
  } from "@rilldata/web-admin/client";
  import CopyInviteLinkButton from "@rilldata/web-admin/features/projects/user-management/CopyInviteLinkButton.svelte";
  import UserInviteForm from "@rilldata/web-admin/features/projects/user-management/UserInviteForm.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import UserItem from "./UserItem.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { getAdminServiceListProjectMemberUsergroupsQueryKey } from "@rilldata/web-admin/client";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import { getRandomBgColor } from "@rilldata/web-common/features/themes/color-config";

  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Lock from "@rilldata/web-common/components/icons/Lock.svelte";
  import UsergroupSetRole from "./UsergroupSetRole.svelte";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import UserGroupItem from "./UserGroupItem.svelte";

  export let organization: string;
  export let project: string;
  export let manageProjectAdmins: boolean;
  export let manageOrgAdmins: boolean;
  export let manageOrgMembers: boolean;

  let open = false;
  let accessDropdownOpen = false;
  let accessType: "everyone" | "invite-only" = "everyone";
  let isHovered = false;

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

  // NOTE: viewer: "not allowed to list project user groups"
  $: listProjectMemberUsergroups =
    createAdminServiceListProjectMemberUsergroups(
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

  // NOTE: viewer: "not authorized to read project members"
  $: listProjectInvites = createAdminServiceListProjectInvites(
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

  // NOTE: editor: "not allowed to list user group members"
  $: listUsergroupMemberUsers = createAdminServiceListUsergroupMemberUsers(
    organization,
    "autogroup:members",
    undefined,
    {
      query: {
        refetchOnMount: true,
        refetchOnWindowFocus: true,
      },
    },
  );

  const PAGE_SIZE = 20;
  $: listOrganizationMemberUsergroups =
    createAdminServiceListOrganizationMemberUsergroups(organization, {
      pageSize: PAGE_SIZE,
      pageToken: "",
      includeCounts: true,
    });
  $: nonManagedGroups =
    $listOrganizationMemberUsergroups.data?.members.filter(
      (group) => !group.groupManaged,
    ) ?? [];

  $: userGroupMemberUsers = $listUsergroupMemberUsers?.data?.members ?? [];
  $: userGroupMemberUsersCount = userGroupMemberUsers?.length ?? 0;

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

  $: accessType = hasAutogroupMembers ? "everyone" : "invite-only";

  function getInitials(name: string) {
    return name.charAt(0).toUpperCase();
  }
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
      <UserInviteForm {organization} {project} />
      <!-- 52 * 8 = 416px -->
      <div class="flex flex-col gap-y-1 overflow-y-auto max-h-[416px]">
        <div class="mt-4">
          {#each sortedProjectMemberUsersList as user}
            <UserItem
              {organization}
              {project}
              {user}
              orgRole={user.orgRoleName}
              {manageProjectAdmins}
              manageProjectMembers={true}
            />
          {/each}
          {#each projectInvitesList as user}
            <UserItem
              {organization}
              {project}
              {user}
              orgRole={user.orgRoleName}
              {manageProjectAdmins}
              manageProjectMembers={true}
            />
          {/each}
          {#each nonManagedGroups as group}
            <UserGroupItem
              {organization}
              {group}
              {manageOrgAdmins}
              {manageOrgMembers}
            />
          {/each}
        </div>
        <div class="mt-2">
          <div class="text-xs text-gray-500 font-semibold uppercase">
            General Access
          </div>
          <Tooltip
            location="right"
            alignment="middle"
            distance={8}
            suppress={accessType !== "everyone"}
          >
            <div
              role="button"
              tabindex="0"
              class="flex flex-row items-center gap-x-2 justify-between rounded-sm cursor-auto"
              data-hovered={isHovered}
              on:mouseover={() => (isHovered = true)}
              on:mouseleave={() => (isHovered = false)}
              on:focus={() => (isHovered = true)}
              on:blur={() => (isHovered = false)}
            >
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
                        <span
                          class="flex flex-row items-center gap-x-1 text-sm font-medium text-gray-900"
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
                        </span>

                        {#if accessType === "everyone"}
                          <div class="flex flex-row items-center gap-x-1">
                            {#if userGroupMemberUsersCount && userGroupMemberUsersCount > 0}
                              <span class="text-xs text-gray-500">
                                {userGroupMemberUsersCount} user{userGroupMemberUsersCount >
                                1
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
                      <span class="text-[11px] text-gray-500"
                        >Org members can access</span
                      >
                    </div>
                  </DropdownMenu.Item>
                </DropdownMenu.Content>
              </DropdownMenu.Root>

              {#if hasAutogroupMembers}
                {#each projectMemberUserGroupsList as group}
                  {#if group.groupName === "autogroup:members"}
                    <UsergroupSetRole {organization} {project} {group} />
                  {/if}
                {/each}
              {/if}
            </div>

            <TooltipContent slot="tooltip-content">
              <ul>
                {#each userGroupMemberUsers.slice(0, 6) as user}
                  <div class="flex items-center gap-1 py-1">
                    <Avatar
                      src={user.userPhotoUrl}
                      avatarSize="h-4 w-4"
                      fontSize="text-[10px]"
                      alt={user.userName}
                      bgColor={getRandomBgColor(user.userEmail)}
                    />
                    <li>{user.userName}</li>
                  </div>
                {/each}
                {#if userGroupMemberUsers.length > 6}
                  <li>and {userGroupMemberUsers.length - 6} more</li>
                {/if}
              </ul>
            </TooltipContent>
          </Tooltip>
        </div>
      </div>
    </div>
    <div
      class="flex flex-row items-center px-3.5 py-3 border-t border-gray-200"
    >
      <a
        href="https://docs.rilldata.com/manage/roles-permissions#project-level-permissions"
        target="_blank"
        class="text-xs text-primary-600">Learn more about sharing</a
      >
      <div class="grow"></div>
      <CopyInviteLinkButton {copyLink} />
    </div>
  </PopoverContent>
</Popover>
