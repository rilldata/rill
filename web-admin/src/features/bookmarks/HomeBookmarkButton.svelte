<script lang="ts">
  import { page } from "$app/stores";
  import BookmarksMenuItem from "@rilldata/web-admin/features/bookmarks/BookmarksMenuItem.svelte";
  import type { BookmarkEntry } from "@rilldata/web-admin/features/bookmarks/utils.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import HomeBookmark from "@rilldata/web-common/components/icons/HomeBookmark.svelte";
  import HomeBookmarkPlus from "@rilldata/web-common/components/icons/HomeBookmarkPlus.svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { clearExploreSessionStore } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

  export let organization: string;
  export let project: string;
  export let resource: { name: string; kind: ResourceKind };
  export let homeBookmark: BookmarkEntry | undefined;
  export let defaultHomeBookmarkUrl: string | undefined;
  export let manageProject: boolean;
  export let onCreate: () => void;
  export let onDelete: (bookmark: BookmarkEntry) => Promise<void>;

  $: ({ name: resourceName, kind: resourceKind } = resource);

  // fullHomeBookmarkUrl contains every param where as homeBookmarkUrl will have params that are equal to rill defaults removed.
  // EG: in explore if all dimensions are shown, fullHomeBookmarkUrl will have `dims=*` where as this will be skipped from homeBookmarkUrl
  $: fullHomeBookmarkUrl =
    homeBookmark?.fullUrl ?? defaultHomeBookmarkUrl ?? "";
  $: homeBookmarkUrl = homeBookmark?.url ?? defaultHomeBookmarkUrl ?? "";
  // Since we remove default params we need to check the current url with homeBookmarkUrl instead of fullHomeBookmarkUrl
  $: isHomeBookmarkActive =
    homeBookmarkUrl === $page.url.searchParams.toString();

  function goToDashboardHome() {
    if (resourceKind === ResourceKind.Explore) {
      // Without clearing sessions empty, DashboardStateDataLoader will load from session for explore view
      clearExploreSessionStore(resourceName, `${organization}__${project}__`);
    }
  }

  let open = false;
</script>

{#if manageProject}
  <DropdownMenu bind:open typeahead={false}>
    <DropdownMenuTrigger asChild let:builder>
      <Button
        builders={[builder]}
        compact
        type="secondary"
        label="Home bookmark dropdown"
        active={open || isHomeBookmarkActive}
      >
        <HomeBookmark
          size="16px"
          className={isHomeBookmarkActive
            ? "text-primary-600"
            : "text-primary-800"}
        />
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent class="w-[330px]">
      {#if homeBookmark}
        <BookmarksMenuItem
          bookmark={homeBookmark}
          {onDelete}
          readOnly={!manageProject}
          showDeleteTooltip
        />
      {:else}
        <DropdownMenuItem class="py-2">
          <a
            href={fullHomeBookmarkUrl}
            on:click={goToDashboardHome}
            class="flex flex-row gap-x-2 w-full min-h-7"
            aria-label="Home Bookmark Entry"
          >
            <HomeBookmark size="16px" className="text-gray-700" />
            <div class="flex flex-col gap-y-0.5">
              <div
                class="text-xs font-medium text-gray-700 h-4 text-ellipsis overflow-hidden"
              >
                Go to Home
              </div>
            </div>
          </a>
        </DropdownMenuItem>
      {/if}
      <DropdownMenuSeparator />
      <DropdownMenuItem on:click={onCreate}>
        <div class="flex flex-row gap-x-2">
          <HomeBookmarkPlus size="16px" />
          <div>
            <div class="text-xs font-medium text-gray-700 h-4">
              Bookmark current view as Home.
            </div>
            <div class="text-[11px] font-normal text-gray-500 h-4">
              This will be everyone’s main view for this dashboard.
            </div>
          </div>
        </div>
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
{:else}
  <Tooltip.Root portal="body">
    <Tooltip.Trigger asChild let:builder>
      <Button
        type="secondary"
        compact
        preload={false}
        href={fullHomeBookmarkUrl}
        onClick={goToDashboardHome}
        class="border border-primary-300"
        builders={[builder]}
        label="Go to home bookmark"
        active={isHomeBookmarkActive}
      >
        <HomeBookmark
          size="16px"
          className={isHomeBookmarkActive
            ? "text-primary-600"
            : "text-primary-800"}
        />
      </Button>
    </Tooltip.Trigger>
    <Tooltip.Content side="bottom">Return to dashboard home</Tooltip.Content>
  </Tooltip.Root>
{/if}
