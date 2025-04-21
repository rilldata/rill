<script lang="ts">
  import AvatarListItem from "@rilldata/web-admin/features/organizations/users/AvatarListItem.svelte";
  import UserSetRole from "./UserSetRole.svelte";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import type {
    V1ProjectMemberUser,
    V1UserInvite,
  } from "@rilldata/web-admin/client";
  import { capitalize } from "@rilldata/web-common/components/table/utils";

  type User = V1ProjectMemberUser | V1UserInvite;

  export let organization: string;
  export let project: string;
  export let user: User;

  $: currentUser = createAdminServiceGetCurrentUser();

  function isProjectMemberUser(user: User): user is V1ProjectMemberUser {
    return "userName" in user;
  }

  function isPendingInvite(user: User): boolean {
    return "invitedBy" in user && user.invitedBy !== undefined;
  }

  $: name = isProjectMemberUser(user) ? user.userName : user.email;
  $: email = isProjectMemberUser(user) ? user.userEmail : user.email;
  $: photoUrl = isProjectMemberUser(user) ? user.userPhotoUrl : null;
  $: roleName = isProjectMemberUser(user) ? user.roleName : user.role;
  $: isCurrentUser = email === $currentUser.data?.user.email;
</script>

<div class="flex flex-row items-center gap-x-2 justify-between cursor-auto">
  <AvatarListItem
    {name}
    {email}
    {photoUrl}
    {isCurrentUser}
    pendingAcceptance={isPendingInvite(user)}
  />
  {#if isProjectMemberUser(user)}
    <UserSetRole
      {organization}
      {project}
      {user}
      {isCurrentUser}
      pendingAcceptance={isPendingInvite(user)}
    />
  {:else}
    <div
      class="w-18 flex flex-row gap-1 items-center rounded-sm px-2 py-1 mr-[10px]"
    >
      <span>{capitalize(roleName)}</span>
    </div>
  {/if}
</div>
