<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListProjectInvites,
    createAdminServiceListProjectMemberUsergroups,
    createAdminServiceListProjectMemberUsers,
  } from "@rilldata/web-admin/client";
  import CopyInviteLinkButton from "@rilldata/web-admin/features/projects/user-invite/CopyInviteLinkButton.svelte";
  import UserInviteAllowlist from "@rilldata/web-admin/features/projects/user-invite/UserInviteAllowlist.svelte";
  import UserInviteForm from "@rilldata/web-admin/features/projects/user-invite/UserInviteForm.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import type { V1UserInvite } from "@rilldata/web-admin/client";
  import AvatarCircleList from "../../organizations/users/AvatarCircleList.svelte";
  import UserInviteOrganization from "./UserInviteOrganization.svelte";
  import UserInviteGroup from "./UserInviteGroup.svelte";
  import UserInviteUserSetRole from "./UserInviteUserSetRole.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let organization: string;
  export let project: string;

  let open = false;

  $: copyLink = `${$page.url.protocol}//${$page.url.host}/${organization}/${project}`;

  $: currentUser = createAdminServiceGetCurrentUser();
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
      roleName: invite.role,
    }));
  }

  $: usersWithPendingInvites = [
    ...projectMemberUsersList,
    ...coerceInvitesToUsers(projectInvitesList),
  ];
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    <Button builders={[builder]} type="secondary">Share</Button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content class="w-[520px] p-4" side="bottom" align="end">
    <div class="flex flex-col">
      <div class="flex flex-row items-center mb-4">
        <div class="text-sm font-medium">{project}</div>
        <div class="grow"></div>
        <CopyInviteLinkButton {copyLink} />
      </div>
      <UserInviteForm
        {organization}
        {project}
        onInvite={() => (open = false)}
      />
      <UserInviteAllowlist {organization} {project} />
      <div class="mt-4">
        <div class="text-xs text-gray-500 font-semibold uppercase">
          Organization
        </div>
        <div class="flex flex-col gap-y-1">
          <UserInviteOrganization {organization} />
        </div>
      </div>
      <div class="mt-2">
        <div class="text-xs text-gray-500 font-semibold uppercase">Groups</div>
        <!-- 52 * 4 = 208px -->
        <div class="flex flex-col gap-y-1 overflow-y-auto max-h-[208px]">
          {#each projectMemberUserGroupsList as group}
            <UserInviteGroup {organization} {project} {group} />
          {/each}
        </div>
      </div>
      <div class="mt-2">
        <div class="text-xs text-gray-500 font-semibold uppercase">Users</div>
        <!-- 52 * 4 = 208px -->
        <div class="flex flex-col gap-y-1 overflow-y-auto max-h-[208px]">
          {#each usersWithPendingInvites as user}
            <div class="flex flex-row items-center gap-x-2 justify-between">
              <AvatarCircleList
                name={user.userName ?? user.userEmail}
                email={user.userEmail}
                isCurrentUser={user.userEmail === $currentUser.data?.user.email}
                pendingAcceptance={!user.userName}
              />
              <!-- If user's role is Viewer and $currentUser.data?.user.email is in a group that has admin role, then hasMultipleAccess is true -->
              {#if user.roleName === "viewer"}
                <div class="flex flex-row items-center gap-x-1">
                  <Tooltip location="bottom" alignment="middle" distance={8}>
                    <div class="text-yellow-600">
                      <InfoCircle size="16px" />
                    </div>
                    <TooltipContent maxWidth="400px" slot="tooltip-content">
                      This person can still edit because they are part of the
                      group Marketing
                    </TooltipContent>
                  </Tooltip>
                  <UserInviteUserSetRole
                    {organization}
                    {project}
                    {user}
                    isCurrentUser={user.userEmail ===
                      $currentUser.data?.user.email}
                  />
                </div>
              {/if}
            </div>
          {/each}
        </div>
      </div>
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>
