<script lang="ts">
  import { createAdminServiceListOrganizationMemberUsers } from "@rilldata/web-admin/client";
  import AvatarSquareList from "../../organizations/users/AvatarSquareList.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import { getRandomBgColor } from "@rilldata/web-common/features/themes/color-config";

  export let organization: string;

  let isHovered = false;

  $: listOrganizationMemberUsers =
    createAdminServiceListOrganizationMemberUsers(organization);
  $: organizationUsersCount =
    $listOrganizationMemberUsers.data?.members?.length ?? 0;
  $: organizationUsersList = $listOrganizationMemberUsers.data?.members ?? [];
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
    <AvatarSquareList name={organization} count={organizationUsersCount}>
      {@html `Everyone from <span class="font-bold">${organization}</span>`}
    </AvatarSquareList>
  </div>

  <TooltipContent maxWidth="121px" slot="tooltip-content">
    <ul>
      {#each organizationUsersList.slice(0, 6) as user}
        <div class="flex items-center gap-1 py-1">
          <Avatar
            avatarSize="h-4 w-4"
            fontSize="text-[10px]"
            alt={user.userName}
            bgColor={getRandomBgColor()}
          />
          <li>{user.userName}</li>
        </div>
      {/each}
      {#if organizationUsersList.length > 6}
        <li>and {organizationUsersList.length - 6} more</li>
      {/if}
    </ul>
  </TooltipContent>
</Tooltip>
