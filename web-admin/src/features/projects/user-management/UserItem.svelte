<script lang="ts">
  import AvatarListItem from "@rilldata/web-common/components/avatar/AvatarListItem.svelte";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import UserSetRole from "./UserSetRole.svelte";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import type {
    V1ProjectMemberUser,
    V1ProjectInvite,
  } from "@rilldata/web-admin/client";

  type User = V1ProjectMemberUser | V1ProjectInvite;

  export let organization: string;
  export let project: string;
  export let user: User;
  export let orgRole: string;
  export let manageProjectMembers: boolean;
  export let manageProjectAdmins: boolean;

  $: currentUser = createAdminServiceGetCurrentUser();

  $: showGuestChip = orgRole === OrgUserRoles.Guest;

  function isProjectMemberUser(user: User): user is V1ProjectMemberUser {
    return "userName" in user;
  }

  function isPendingInvite(user: User): boolean {
    return "invitedBy" in user && user.invitedBy !== undefined;
  }

  $: name = isProjectMemberUser(user) ? user.userName : user.email;
  $: email = isProjectMemberUser(user) ? user.userEmail : user.email;
  $: photoUrl = isProjectMemberUser(user) ? user.userPhotoUrl : null;
  $: isCurrentUser = email === $currentUser.data?.user.email;
</script>

<div class="flex flex-row items-center gap-x-2 justify-between cursor-auto">
  <AvatarListItem
    {name}
    {email}
    {photoUrl}
    {isCurrentUser}
    {showGuestChip}
    pendingAcceptance={isPendingInvite(user)}
  />
  <UserSetRole
    {organization}
    {project}
    {user}
    {isCurrentUser}
    {manageProjectMembers}
    {manageProjectAdmins}
  />
</div>
