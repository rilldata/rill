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
  import {
    DropdownMenu,
    DropdownMenuTrigger,
    DropdownMenuContent,
  } from "@rilldata/web-common/components/dropdown-menu";
  import type { V1UserInvite } from "@rilldata/web-admin/client";
  import AvatarCircleList from "../../organizations/users/AvatarCircleList.svelte";

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

<DropdownMenu bind:open>
  <DropdownMenuTrigger asChild let:builder>
    <Button builders={[builder]} type="secondary">Share</Button>
  </DropdownMenuTrigger>
  <DropdownMenuContent class="w-[520px] p-4" side="bottom" align="end">
    <div class="flex flex-col gap-y-3">
      <div class="flex flex-row items-center">
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
      <div>
        <div class="text-xs text-gray-500 font-semibold uppercase">
          Organization
        </div>
        <div class="text-xs text-gray-800">
          Everyone from <span class="font-bold">{organization}</span>
        </div>
      </div>
      <div>
        <div class="text-xs text-gray-500 font-semibold uppercase">Groups</div>
        <!-- TODO: create AvatarSquareList -->
        {#each userGroupsList as group}
          <div class="text-xs text-gray-800">
            Everyone from <span class="font-bold">{group.groupName}</span>
          </div>
        {/each}
      </div>
      <div>
        <div class="text-xs text-gray-500 font-semibold uppercase">Users</div>
        <div class="flex flex-col gap-y-1">
          {#each usersWithPendingInvites as user}
            <AvatarCircleList
              name={user.userName ?? user.userEmail}
              email={user.userEmail}
              isCurrentUser={user.userEmail === $currentUser.data?.user.email}
              pendingAcceptance={!user.userName}
            />
          {/each}
        </div>
      </div>
    </div>
  </DropdownMenuContent>
</DropdownMenu>
