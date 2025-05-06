<script lang="ts">
  import AvatarListItem from "@rilldata/web-admin/features/organizations/users/AvatarListItem.svelte";
  import UserSetRole from "./UserSetRole.svelte";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import type {
    V1ProjectMemberUser,
    V1UserInvite,
  } from "@rilldata/web-admin/client";
  type User = V1ProjectMemberUser | V1UserInvite;

  export let organization: string;
  export let project: string;
  export let user: User;
  export let orgRole: string;
  export let canChangeRole: boolean;

  $: currentUser = createAdminServiceGetCurrentUser();

  $: showGuestChip = orgRole === "guest";

  function isProjectMemberUser(user: User): user is V1ProjectMemberUser {
    return "userName" in user;
  }

  function isPendingInvite(user: User): boolean {
    return "invitedBy" in user && user.invitedBy !== undefined;
  }

  $: name = isProjectMemberUser(user) ? user.userName : user.email;
  $: email = isProjectMemberUser(user) ? user.userEmail : user.email;
  $: photoUrl = isProjectMemberUser(user) ? user.userPhotoUrl : null;
  $: role = isProjectMemberUser(user) ? user.roleName : user.role;
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
    {role}
    {isCurrentUser}
    {canChangeRole}
  />
</div>
