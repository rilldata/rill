<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceListProjectInvites,
    createAdminServiceListProjectMemberUsergroups,
    createAdminServiceListProjectMemberUsers,
    createAdminServiceGetCurrentUser,
    createAdminServiceListUsergroupMemberUsers,
    createAdminServiceListOrganizationInvitesInfinite,
    createAdminServiceListOrganizationMemberUsersInfinite,
    createAdminServiceListOrganizationMemberUsergroups,
  } from "@rilldata/web-admin/client";
  import CopyInviteLinkButton from "@rilldata/web-admin/features/projects/user-management/CopyInviteLinkButton.svelte";
  import UserAndGroupInviteForm from "@rilldata/web-admin/features/projects/user-management/UserAndGroupInviteForm.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import UserItem from "./UserItem.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import { getRandomBgColor } from "@rilldata/web-common/features/themes/color-config";
  import ProjectUserGroupItem from "./ProjectUserGroupItem.svelte";
  import UsergroupSetRole from "./UsergroupSetRole.svelte";
  import GeneralAccessSelectorDropdown from "./GeneralAccessSelectorDropdown.svelte";
  import { buildSearchList, buildCopyLink } from "./helpers";

  export let organization: string;
  export let project: string;
  export let manageProjectAdmins: boolean;
  export let manageOrgAdmins: boolean;
  export let manageOrgMembers: boolean;

  let open = false;
  let accessType: "everyone" | "invite-only" = "everyone";
  let isHovered = false;

  const currentUser = createAdminServiceGetCurrentUser();

  const PAGE_SIZE = 50;

  $: orgMemberUsersInfiniteQuery =
    createAdminServiceListOrganizationMemberUsersInfinite(
      organization,
      {
        pageSize: PAGE_SIZE,
      },
      {
        query: {
          enabled: open,
          getNextPageParam: (lastPage) => {
            if (lastPage.nextPageToken !== "") {
              return lastPage.nextPageToken;
            }
            return undefined;
          },
        },
      },
    );
  $: orgInvitesInfiniteQuery =
    createAdminServiceListOrganizationInvitesInfinite(
      organization,
      {
        pageSize: PAGE_SIZE,
      },
      {
        query: {
          enabled: open,
          getNextPageParam: (lastPage) => {
            if (lastPage.nextPageToken !== "") {
              return lastPage.nextPageToken;
            }
            return undefined;
          },
        },
      },
    );
  $: listOrganizationMemberUsergroups =
    createAdminServiceListOrganizationMemberUsergroups(
      organization,
      {
        pageSize: PAGE_SIZE,
        includeCounts: true,
      },
      {
        query: {
          enabled: open,
        },
      },
    );
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
  $: listProjectMemberUsers = createAdminServiceListProjectMemberUsers(
    organization,
    project,
    undefined,
    {
      query: {
        enabled: open,
        refetchOnMount: true,
        refetchOnWindowFocus: true,
        select: (data) => {
          if (!data?.members) return data;
          const currentUserEmail = $currentUser.data?.user?.email;
          if (!currentUserEmail) return data;

          return {
            ...data,
            members: [...data.members].sort((a, b) => {
              if (a.userEmail === currentUserEmail) return -1;
              if (b.userEmail === currentUserEmail) return 1;
              return 0;
            }),
          };
        },
      },
    },
  );
  $: listProjectInvites = createAdminServiceListProjectInvites(
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

  $: allOrgMemberUsersRows =
    $orgMemberUsersInfiniteQuery?.data?.pages?.flatMap(
      (page) => page?.members ?? [],
    ) ?? [];
  $: allOrgInvitesRows =
    $orgInvitesInfiniteQuery?.data?.pages?.flatMap(
      (page) => page?.invites ?? [],
    ) ?? [];

  async function ensureAllUsersLoaded() {
    if (
      $orgMemberUsersInfiniteQuery?.hasNextPage &&
      !$orgMemberUsersInfiniteQuery?.isFetchingNextPage
    ) {
      await $orgMemberUsersInfiniteQuery.fetchNextPage();
      // Recursively call until all pages are loaded
      if ($orgMemberUsersInfiniteQuery?.hasNextPage) {
        await ensureAllUsersLoaded();
      }
    }
  }

  // Load all users when popover opens
  $: if (open) {
    ensureAllUsersLoaded();
  }

  $: orgMemberUsergroups =
    $listOrganizationMemberUsergroups?.data?.members ?? [];
  $: userGroupMemberUsers = $listUsergroupMemberUsers?.data?.members ?? [];
  $: projectMemberUserGroupsList =
    $listProjectMemberUsergroups.data?.members ?? [];
  $: projectMemberUsersList = $listProjectMemberUsers?.data?.members ?? [];
  $: projectInvitesList = $listProjectInvites?.data?.invites ?? [];

  // Memoized Sets for efficient O(1) lookups instead of expensive O(n) .some() operations
  $: projectMemberEmailSet = new Set(
    projectMemberUsersList.map((pm) => pm.userEmail),
  );
  $: projectInviteEmailSet = new Set(
    projectInvitesList.map((invite) => invite.email),
  );
  $: projectUserGroups = projectMemberUserGroupsList.filter(
    (group) => !group.groupManaged,
  );
  $: projectUserGroupNameSet = new Set(
    projectUserGroups.map((pg) => pg.groupName),
  );

  $: searchList = buildSearchList(
    allOrgMemberUsersRows,
    allOrgInvitesRows,
    orgMemberUsergroups,
    projectMemberEmailSet,
    projectInviteEmailSet,
    projectUserGroupNameSet,
  );

  $: copyLink = buildCopyLink($page.url, organization, project);

  $: hasAutogroupMembers = projectMemberUserGroupsList.some(
    (group) => group.groupName === "autogroup:members",
  );

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
      <UserAndGroupInviteForm {organization} {project} {searchList} />
      <div class="flex flex-col gap-y-1 overflow-y-auto max-h-[350px] mt-2">
        <div class="mt-2">
          {#each projectMemberUsersList as user}
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
          {#each projectUserGroups as group}
            <ProjectUserGroupItem
              {organization}
              {group}
              {project}
              {manageOrgAdmins}
              {manageOrgMembers}
            />
          {/each}
        </div>
      </div>
      <div class="mt-2 general-access-container bg-white pt-2">
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
            <GeneralAccessSelectorDropdown {organization} {project} />

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
    <div
      class="flex flex-row items-center px-3.5 py-3 border-t border-gray-200"
    >
      <a
        href="https://docs.rilldata.com/manage/roles-permissions#project-level-permissions"
        target="_blank"
        class="text-xs text-primary-600 hover::text-primary-700"
        >Learn more about sharing</a
      >
      <div class="grow"></div>
      <CopyInviteLinkButton {copyLink} />
    </div>
  </PopoverContent>
</Popover>
