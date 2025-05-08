<script lang="ts">
  import AvatarListItem from "@rilldata/web-admin/features/organizations/users/AvatarListItem.svelte";
  import UserSetRole from "./UserSetRole.svelte";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import type {
    V1ProjectMemberUser,
    V1ProjectInvite,
  } from "@rilldata/web-admin/client";
  import { capitalize } from "@rilldata/web-common/components/table/utils";

  type User = V1ProjectMemberUser | V1ProjectInvite;

  export let organization: string;
  export let project: string;
  export let user: User;
  export let orgRole: string;
  export let canChangeRole: boolean;
  export let manageProjectAdmins: boolean;

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
  $: roleName = user.roleName;
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
  {#if manageProjectAdmins}
    <UserSetRole
      {organization}
      {project}
      {user}
      {isCurrentUser}
      {canChangeRole}
    />
  {:else}
    <div
      class="w-18 flex flex-row gap-1 items-center rounded-sm mr-[10px] cursor-auto"
    >
      {capitalize(roleName)}
    </div>
  {/if}
</div>
