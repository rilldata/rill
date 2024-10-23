<script lang="ts">
  import UserManagementGroupAvatar from "./UserManagementGroupAvatar.svelte";
  import UserManagementGroupSetRole from "./UserManagementGroupSetRole.svelte";
  import {
    createAdminServiceListUsergroupMemberUsers,
    type V1MemberUsergroup,
  } from "@rilldata/web-admin/client";
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getRandomBgColor } from "@rilldata/web-common/features/themes/color-config";

  export let organization: string;
  export let project: string;
  export let group: V1MemberUsergroup;

  let isHovered = false;

  $: listUsergroupMemberUsers = createAdminServiceListUsergroupMemberUsers(
    organization,
    group.groupName,
  );
  $: userGroupMemberUsersList = $listUsergroupMemberUsers.data?.members ?? [];
</script>

<Tooltip location="right" alignment="middle" distance={8}>
  <div
    role="button"
    tabindex="0"
    class="flex flex-row items-center gap-x-2 justify-between data-[hovered=true]:bg-slate-50 rounded-sm cursor-auto"
    data-hovered={isHovered}
    on:mouseover={() => (isHovered = true)}
    on:mouseleave={() => (isHovered = false)}
    on:focus={() => (isHovered = true)}
    on:blur={() => (isHovered = false)}
  >
    <UserManagementGroupAvatar {organization} usergroup={group.groupName} />
    <UserManagementGroupSetRole {organization} {project} {group} />
  </div>

  <TooltipContent maxWidth="121px" slot="tooltip-content">
    <ul>
      {#each userGroupMemberUsersList.slice(0, 6) as user}
        <div class="flex items-center gap-1 py-1">
          <Avatar
            avatarSize="h-4 w-4"
            fontSize="text-[10px]"
            alt={user.userName}
            src={user.userPhotoUrl}
            bgColor={getRandomBgColor(user.userEmail)}
          />
          <li>{user.userName}</li>
        </div>
      {/each}
      {#if userGroupMemberUsersList.length > 6}
        <li>and {userGroupMemberUsersList.length - 6} more</li>
      {/if}
    </ul>
  </TooltipContent>
</Tooltip>
