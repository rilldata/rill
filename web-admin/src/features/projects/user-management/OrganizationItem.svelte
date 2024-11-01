<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import { getRandomBgColor } from "@rilldata/web-common/features/themes/color-config";
  import AvatarListItem from "../../organizations/users/AvatarListItem.svelte";
  import UserManagementOrganizationSetRole from "./OrganizationSetRole.svelte";
  import {
    type V1MemberUsergroup,
    createAdminServiceListUsergroupMemberUsers,
  } from "@rilldata/web-admin/client";
  import { isMetricsExplorerPage } from "../../navigation/nav-utils";
  import { page } from "$app/stores";

  export let group: V1MemberUsergroup;
  export let organization: string;
  export let project: string;
  export let canManage: boolean = false;

  let isHovered = false;

  $: onMetricsExplorerPage = isMetricsExplorerPage($page);

  $: listUsergroupMemberUsers = createAdminServiceListUsergroupMemberUsers(
    organization,
    group.groupName,
  );

  $: userGroupMemberUsers = $listUsergroupMemberUsers.data?.members ?? [];
  $: userGroupMemberUsersCount = userGroupMemberUsers?.length ?? 0;
</script>

{#if group}
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
      <AvatarListItem
        shape="square"
        name={organization}
        count={userGroupMemberUsersCount}
        isEveryoneFromText
        canManage={onMetricsExplorerPage}
      />
      {#if !canManage}
        <UserManagementOrganizationSetRole {organization} {project} {group} />
      {/if}
    </div>

    <TooltipContent maxWidth="121px" slot="tooltip-content">
      <ul>
        {#each userGroupMemberUsers.slice(0, 6) as user}
          <div class="flex items-center gap-1 py-1">
            <Avatar
              src={user.userPhotoUrl}
              avatarSize="h-4 w-4"
              fontSize="text-[10px]"
              alt={user.userName}
              bgColor={getRandomBgColor(user.userEmail)}
            />
            <li>{user.userName}</li>
          </div>
        {/each}
        {#if userGroupMemberUsers.length > 6}
          <li>and {userGroupMemberUsers.length - 6} more</li>
        {/if}
      </ul>
    </TooltipContent>
  </Tooltip>
{/if}
