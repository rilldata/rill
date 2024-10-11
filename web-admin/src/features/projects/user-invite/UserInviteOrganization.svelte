<script lang="ts">
  import AvatarListItem from "@rilldata/web-common/components/avatar/AvatarListItem.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import { getRandomBgColor } from "@rilldata/web-common/features/themes/color-config";
  import type { V1MemberUser } from "@rilldata/web-admin/client";

  export let organization: string;
  export let memberUsers: V1MemberUser[];

  let isHovered = false;

  $: organizationUsersCount = memberUsers?.length ?? 0;
</script>

<Tooltip location="right" alignment="middle" distance={8}>
  <div
    role="button"
    tabindex="0"
    class="flex flex-row items-center gap-x-2 justify-between data-[hovered=true]:bg-slate-50 rounded-sm"
    data-hovered={isHovered}
    on:mouseover={() => (isHovered = true)}
    on:mouseleave={() => (isHovered = false)}
    on:focus={() => (isHovered = true)}
    on:blur={() => (isHovered = false)}
  >
    <AvatarListItem
      shape="square"
      name={organization}
      count={organizationUsersCount}
    >
      {@html `Everyone from <span class="font-bold">${organization}</span>`}
    </AvatarListItem>
  </div>

  <TooltipContent maxWidth="121px" slot="tooltip-content">
    <ul>
      {#each memberUsers.slice(0, 6) as user}
        <div class="flex items-center gap-1 py-1">
          <Avatar
            avatarSize="h-4 w-4"
            fontSize="text-[10px]"
            alt={user.userName}
            bgColor={getRandomBgColor(user.userEmail)}
          />
          <li>{user.userName}</li>
        </div>
      {/each}
      {#if memberUsers.length > 6}
        <li>and {memberUsers.length - 6} more</li>
      {/if}
    </ul>
  </TooltipContent>
</Tooltip>
