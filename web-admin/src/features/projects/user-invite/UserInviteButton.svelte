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
  import UserInviteGroup from "./UserInviteGroup.svelte";
  import UserInviteOrganization from "./UserInviteOrganization.svelte";

  import UserInviteUserSetRole from "./UserInviteUserSetRole.svelte";
  import UserInviteGroupSetRole from "./UserInviteGroupSetRole.svelte";

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

  $: userGroupsList = $listProjectMemberUsergroups.data?.members ?? [];
  $: usersList = $listProjectMemberUsers.data?.members ?? [];
  $: invitesList = $listProjectInvites.data?.invites ?? [];

  function coerceInvitesToUsers(invites: V1UserInvite[]) {
    return invites.map((invite) => ({
      ...invite,
      userName: null,
      userEmail: invite.email,
      roleName: invite.role,
    }));
  }

  $: usersWithPendingInvites = [
    ...usersList,
    ...coerceInvitesToUsers(invitesList),
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
          {#each userGroupsList as group}
            <div class="flex flex-row items-center gap-x-2 justify-between">
              <UserInviteGroup {organization} usergroup={group.groupName} />
              <UserInviteGroupSetRole {organization} {project} {group} />
            </div>
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
              <UserInviteUserSetRole {organization} {project} {user} />
            </div>
          {/each}
        </div>
      </div>
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>
