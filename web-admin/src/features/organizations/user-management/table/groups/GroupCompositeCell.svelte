<script lang="ts">
  import { page } from "$app/stores";
  import { getRandomBgColor } from "@rilldata/web-common/features/themes/color-config.ts";
  import { cn } from "@rilldata/web-common/lib/shadcn.ts";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import {
    createAdminServiceListUsergroupMemberUsers,
    adminServiceListUsergroupMemberUsers,
  } from "@rilldata/web-admin/client";
  import type { V1UsergroupMemberUser } from "@rilldata/web-admin/client";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  export let name: string;
  export let usersCount: number;
  export let groupName: string;

  let hovered = false;

  const PREVIEW_COUNT = 6;

  $: organization = $page.params.organization;
  $: listUsergroupMemberUsers = createAdminServiceListUsergroupMemberUsers(
    organization,
    groupName,
    // Fetch server default (20) like the share modal, then slice client-side
    undefined,
    {
      query: {
        enabled: hovered && (usersCount ?? 0) > 0,
      },
    },
  );
  let loadedUsers: V1UsergroupMemberUser[] = [];
  let nextPageToken: string = "";
  let isFetchingMore = false;

  $: {
    const data = $listUsergroupMemberUsers.data;
    if (data) {
      // Reset on fresh fetch
      loadedUsers = data.members ?? [];
      nextPageToken = data.nextPageToken ?? "";
    }
  }

  async function ensureMinPreview() {
    if (isFetchingMore) return;
    if (!hovered) return;
    if ((usersCount ?? 0) <= (loadedUsers?.length ?? 0)) return;
    if ((loadedUsers?.length ?? 0) >= PREVIEW_COUNT) return;

    try {
      isFetchingMore = true;
      while (
        nextPageToken &&
        loadedUsers.length < PREVIEW_COUNT &&
        (usersCount ?? 0) > loadedUsers.length
      ) {
        const resp = await adminServiceListUsergroupMemberUsers(
          organization,
          groupName,
          { pageSize: PREVIEW_COUNT, pageToken: nextPageToken },
        );
        loadedUsers = [...loadedUsers, ...(resp.members ?? [])];
        nextPageToken = resp.nextPageToken ?? "";
      }
    } finally {
      isFetchingMore = false;
    }
  }

  $: if (hovered && (usersCount ?? 0) > 0) {
    // Best-effort: ensure we show up to 6 even if first page is short
    // (e.g., due to pagination window)
    void ensureMinPreview();
  }

  $: previewUsers = loadedUsers;
  $: visiblePreviewCount = Math.min(PREVIEW_COUNT, previewUsers.length);

  function getInitials(name: string) {
    return name.charAt(0).toUpperCase();
  }
</script>

<div class="flex items-center gap-2 py-2 pl-2">
  <div
    class={cn(
      "h-7 w-7 rounded-sm flex items-center justify-center",
      getRandomBgColor(name),
    )}
  >
    <span class="text-sm text-white font-semibold">{getInitials(name)}</span>
  </div>
  <div class="flex flex-col text-left">
    <span class="text-sm font-medium text-gray-900 flex flex-row gap-x-1">
      {name}
    </span>
    <Tooltip location="right" alignment="start" distance={8}>
      <div
        class="flex flex-row items-center gap-x-1"
        role="button"
        tabindex="0"
        on:mouseenter={() => (hovered = true)}
        on:focus={() => (hovered = true)}
        on:blur={() => (hovered = false)}
      >
        <span class="text-xs text-gray-500">
          {usersCount} user{usersCount === 1 ? "" : "s"}
        </span>
      </div>
      <TooltipContent slot="tooltip-content">
        {#if (usersCount ?? 0) === 0}
          <div class="text-xs text-gray-300 px-1 py-0.5">No users</div>
        {:else if $listUsergroupMemberUsers.isLoading}
          <div class="px-1 py-0.5">
            <Spinner
              status={EntityStatus.Running}
              size="0.9rem"
              duration={600}
            />
          </div>
        {:else if $listUsergroupMemberUsers.isError}
          <div class="text-xs text-red-300 px-1 py-0.5">
            Failed to load users
          </div>
        {:else}
          <ul>
            {#each previewUsers.slice(0, PREVIEW_COUNT) as u}
              {@const displayName = u.userName}
              {@const colorSeed = u.userEmail}
              <div class="flex items-center gap-1 py-1">
                <!-- Use email first to seed color for stable, unique avatar colors;
                     visible text and alt prefer name for readability. -->
                <Avatar
                  src={u.userPhotoUrl}
                  avatarSize="h-4 w-4"
                  fontSize="text-[10px]"
                  alt={displayName}
                  bgColor={getRandomBgColor(colorSeed)}
                />
                <li>{displayName}</li>
              </div>
            {/each}
            {#if (usersCount ?? 0) > visiblePreviewCount}
              <li>and {(usersCount ?? 0) - visiblePreviewCount} more</li>
            {/if}
          </ul>
        {/if}
      </TooltipContent>
    </Tooltip>
  </div>
</div>
