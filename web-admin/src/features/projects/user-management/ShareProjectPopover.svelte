<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceListProjectInvites,
    createAdminServiceListProjectMemberUsergroups,
    createAdminServiceListProjectMemberUsers,
  } from "@rilldata/web-admin/client";
  import CopyInviteLinkButton from "@rilldata/web-admin/features/projects/user-management/CopyInviteLinkButton.svelte";
  import UserInviteForm from "@rilldata/web-admin/features/projects/user-management/UserInviteForm.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import type { V1UserInvite } from "@rilldata/web-admin/client";
  import OrganizationItem from "./OrganizationItem.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import UsergroupItem from "./UsergroupItem.svelte";
  import UserItem from "./UserItem.svelte";

  export let organization: string;
  export let project: string;

  let open = false;

  $: copyLink = `${$page.url.protocol}//${$page.url.host}/${organization}/${project}`;

  $: listProjectMemberUsergroups =
    createAdminServiceListProjectMemberUsergroups(organization, project);
  $: listProjectMemberUsers = createAdminServiceListProjectMemberUsers(
    organization,
    project,
  );
  $: listProjectInvites = createAdminServiceListProjectInvites(
    organization,
    project,
  );

  $: projectMemberUserGroupsList =
    $listProjectMemberUsergroups.data?.members ?? [];
  $: projectMemberUsersList = $listProjectMemberUsers.data?.members ?? [];
  $: projectInvitesList = $listProjectInvites.data?.invites ?? [];

  function coerceInvitesToUsers(invites: V1UserInvite[]) {
    return invites.map((invite) => ({
      ...invite,
      userName: null,
      userEmail: invite.email,
      userPhotoUrl: null,
      roleName: invite.role,
    }));
  }

  $: usersWithPendingInvites = [
    ...projectMemberUsersList,
    ...coerceInvitesToUsers(projectInvitesList),
  ];

  $: showOrganizationSection = projectMemberUserGroupsList.some(
    (group) => group.groupName === "all-users",
  );

  $: showAllUsersGroup = projectMemberUserGroupsList.find(
    (group) => group.groupName === "all-users",
  );

  $: showGroupsSection =
    projectMemberUserGroupsList.length > 0 &&
    projectMemberUserGroupsList.length === 1 &&
    projectMemberUserGroupsList[0].groupName !== "all-users";
</script>

<Popover bind:open>
  <PopoverTrigger asChild let:builder>
    <Button builders={[builder]} type="secondary" selected={open}>Share</Button>
  </PopoverTrigger>
  <PopoverContent align="end" class="w-[520px] p-4">
    <div class="flex flex-col">
      <div class="flex flex-row items-center mb-4">
        <div class="text-sm font-medium">{project}</div>
        <div class="grow"></div>
        <CopyInviteLinkButton {copyLink} />
      </div>
      <UserInviteForm {organization} {project} />
      {#if showOrganizationSection}
        <div class="mt-4">
          <div class="text-xs text-gray-500 font-semibold uppercase">
            Org Users
          </div>
          <div class="flex flex-col gap-y-1">
            <OrganizationItem
              {organization}
              {project}
              group={showAllUsersGroup ?? null}
            />
          </div>
        </div>
      {/if}
      {#if showGroupsSection}
        <div class="mt-2">
          <div class="text-xs text-gray-500 font-semibold uppercase">
            User Groups
          </div>
          <!-- 52 * 5 = 260px -->
          <div class="flex flex-col gap-y-1 overflow-y-auto max-h-[260px]">
            {#each projectMemberUserGroupsList as group}
              <UsergroupItem {organization} {project} {group} />
            {/each}
          </div>
        </div>
      {/if}
      <div class="mt-2">
        <div class="text-xs text-gray-500 font-semibold uppercase">
          Project Users
        </div>
        <!-- 52 * 5 = 260px -->
        <div class="flex flex-col gap-y-1 overflow-y-auto max-h-[260px]">
          {#each usersWithPendingInvites as user}
            <UserItem {organization} {project} {user} />
          {/each}
        </div>
      </div>
    </div>
  </PopoverContent>
</Popover>
