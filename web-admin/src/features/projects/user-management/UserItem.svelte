<script lang="ts">
  import AvatarListItem from "@rilldata/web-admin/features/organizations/users/AvatarListItem.svelte";
  import UserSetRole from "./UserSetRole.svelte";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import type { V1MemberUser } from "@rilldata/web-admin/client";

  type User =
    | V1MemberUser
    | {
        userName: any;
        userEmail: string;
        userPhotoUrl?: string;
        roleName: string;
        email?: string;
        role?: string;
        invitedBy?: string;
      };

  export let organization: string;
  export let project: string;
  export let user: User;

  $: currentUser = createAdminServiceGetCurrentUser();
</script>

<div class="flex flex-row items-center gap-x-2 justify-between cursor-auto">
  <AvatarListItem
    name={user.userName ?? user.userEmail}
    email={user.userEmail}
    photoUrl={user.userPhotoUrl}
    isCurrentUser={user.userEmail === $currentUser.data?.user.email}
    pendingAcceptance={!user.userName}
  />
  <UserSetRole
    {organization}
    {project}
    {user}
    isCurrentUser={user.userEmail === $currentUser.data?.user.email}
    pendingAcceptance={!user.userName}
  />
</div>
