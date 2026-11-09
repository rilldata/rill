<script lang="ts">
  import { page } from "$app/stores";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { type V1Bookmark } from "@rilldata/web-admin/client";
  import {
    bookmarkCreateMutationFactory,
    bookmarkRemoveMutationFactory,
    bookmarkUpdateMutationFactory,
  } from "@rilldata/web-admin/features/bookmarks/bookmark-client.ts";
  import BookmarksMenuItem from "@rilldata/web-admin/features/bookmarks/BookmarksMenuItem.svelte";
  import BookmarksFormDialog from "@rilldata/web-admin/features/bookmarks/BookmarksFormDialog.svelte";
  import {
    invalidateBookmarkQueries,
    isHomeBookmark,
  } from "@rilldata/web-admin/features/bookmarks/selectors.ts";
  import {
    type BookmarkEntry,
    type Bookmarks,
    getBookmarkData,
    searchBookmarks,
  } from "@rilldata/web-admin/features/bookmarks/utils.ts";
  import HomeBookmarkButton from "@rilldata/web-admin/features/bookmarks/HomeBookmarkButton.svelte";
  import { RecentlyUsedBookmarks } from "@rilldata/web-admin/features/bookmarks/recently-used-bookmarks.ts";
  import { getProjectPermissions } from "@rilldata/web-admin/features/projects/selectors.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuGroup,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import { Search } from "@rilldata/web-common/components/search";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { createMutationFromFactory } from "@rilldata/web-common/features/scheduled-reports/FormMetadataProvider.svelte.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { BookmarkIcon, BookmarkPlusIcon } from "lucide-svelte";

  export let organization: string;
  export let project: string;
  // Empty in embeds, where the server resolves the project from the embed's admin token.
  export let projectId: string;
  export let resource: { name: string; kind: ResourceKind };
  export let bookmarkData: {
    bookmarks: V1Bookmark[];
    categorizedBookmarks: Bookmarks;
    defaultUrlParams?: URLSearchParams;
    defaultHomeBookmarkUrl?: string;
    showFiltersOnly?: boolean;
  };
  // Whether to show the home bookmark and shared bookmarks, and allow creating them.
  // Embeds only support personal bookmarks.
  export let showSharedBookmarks = true;

  $: ({ name: resourceName, kind: resourceKind } = resource);
  $: ({
    bookmarks,
    categorizedBookmarks,
    defaultUrlParams,
    defaultHomeBookmarkUrl,
    showFiltersOnly,
  } = bookmarkData);

  let showDialog = false;
  let bookmark: BookmarkEntry | null = null;

  $: recentlyUsedBookmarks = new RecentlyUsedBookmarks(organization, project);
  function recordUsage(bookmark: BookmarkEntry) {
    recentlyUsedBookmarks.update(bookmark.resource.id);
  }

  // Only the home bookmark button needs the project permissions, so embeds never request them.
  $: projectPermissions = showSharedBookmarks
    ? getProjectPermissions(organization, project)
    : undefined;
  $: manageProject = !!$projectPermissions?.data?.manageProject;

  $: curUrlParams = $page.url.searchParams;

  const bookmarkCreator = createMutationFromFactory(
    bookmarkCreateMutationFactory,
  );
  const bookmarkUpdater = createMutationFromFactory(
    bookmarkUpdateMutationFactory,
  );
  const bookmarkDeleter = createMutationFromFactory(
    bookmarkRemoveMutationFactory,
  );

  async function createHomeBookmark() {
    const homeBookmarkUrlSearch = getBookmarkData({
      curUrlParams,
      defaultUrlParams,
    });
    const homeBookmark = bookmarks.find(isHomeBookmark);

    if (homeBookmark) {
      await $bookmarkUpdater.mutateAsync({
        data: {
          bookmarkId: homeBookmark.id,
          displayName: "Go to Home",
          description: "",
          shared: true,
          default: true,
          urlSearch: homeBookmarkUrlSearch,
        },
      });
    } else {
      await $bookmarkCreator.mutateAsync({
        data: {
          displayName: "Go to Home",
          description: "",
          projectId,
          resourceKind,
          resourceName,
          shared: true,
          default: true,
          urlSearch: homeBookmarkUrlSearch,
        },
      });
    }

    eventBus.emit("notification", {
      message: m.bookmark_home_created(),
    });
    return invalidateBookmarkQueries();
  }

  function onEdit(editingBookmark: BookmarkEntry) {
    showDialog = true;
    bookmark = editingBookmark;
  }

  async function deleteBookmark(bookmark: BookmarkEntry) {
    // TODO: add confirmation?
    await $bookmarkDeleter.mutateAsync({
      bookmarkId: bookmark.resource.id,
    });
    eventBus.emit("notification", {
      message: m.bookmark_deleted({
        name: bookmark.resource.displayName ?? "",
      }),
    });
    return invalidateBookmarkQueries();
  }

  let open = false;

  let searchText: string;

  $: filteredBookmarks = searchBookmarks(categorizedBookmarks, searchText);
</script>

{#if showSharedBookmarks}
  <HomeBookmarkButton
    {organization}
    {project}
    {resource}
    homeBookmark={categorizedBookmarks.home}
    {defaultHomeBookmarkUrl}
    onCreate={createHomeBookmark}
    onDelete={deleteBookmark}
    onOpen={recordUsage}
    {manageProject}
  />
{/if}

<DropdownMenu bind:open>
  <DropdownMenuTrigger>
    {#snippet child({ props })}
      <Button
        {...props}
        compact
        square
        type="secondary"
        label={m.bookmark_other_dropdown_label()}
        active={open}
      >
        <BookmarkIcon class="flex-none" size="16px" />
      </Button>
    {/snippet}
  </DropdownMenuTrigger>
  <!--
    The content is bounded to the space available below the trigger (capped at 600px)
    so long bookmark lists scroll instead of running off the viewport.
    The "bookmark current view" action and search stay pinned; only the lists scroll.
  -->
  <DropdownMenuContent
    class="flex flex-col w-[450px] max-h-[min(600px,var(--bits-floating-available-height))] overflow-hidden p-0"
  >
    <div class="flex-none p-1.5">
      <DropdownMenuItem onclick={() => (showDialog = true)}>
        <div class="flex flex-row gap-x-2 items-center">
          <BookmarkPlusIcon size="16px" strokeWidth={1.5} />
          <div class="text-xs">{m.bookmark_current_view()}</div>
        </div>
      </DropdownMenuItem>
      <DropdownMenuSeparator />
      <div class="p-2">
        <Search
          autofocus={false}
          bind:value={searchText}
          showBorderOnFocus={false}
        />
      </div>
    </div>
    {#if filteredBookmarks}
      <div class="flex-1 min-h-0 overflow-y-auto p-1.5 pt-0">
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuLabel
            class="sticky top-0 z-10 bg-popover text-fg-secondary text-[10px] h-6 uppercase"
          >
            {m.bookmark_your_bookmarks()}
          </DropdownMenuLabel>
          {#if filteredBookmarks.personal?.length}
            {#each filteredBookmarks.personal as bookmark (bookmark.resource.id)}
              {#key bookmark.resource.id}
                <BookmarksMenuItem
                  {bookmark}
                  {onEdit}
                  onDelete={deleteBookmark}
                  onClick={() => recordUsage(bookmark)}
                />
              {/key}
            {/each}
          {:else}
            <div class="my-2 text-fg-muted text-center">
              {m.bookmark_no_bookmarks()}
            </div>
          {/if}
        </DropdownMenuGroup>
        {#if showSharedBookmarks}
          <DropdownMenuSeparator />
          <DropdownMenuGroup>
            <DropdownMenuLabel
              class="sticky top-0 z-10 bg-popover text-fg-secondary"
            >
              <div class="text-[10px] h-4 uppercase">
                {m.bookmark_managed_bookmarks()}
              </div>
              <div class="text-[11px] font-normal">
                {m.bookmark_created_by_admin()}
              </div>
            </DropdownMenuLabel>
            {#if filteredBookmarks.shared?.length}
              {#each filteredBookmarks.shared as bookmark (bookmark.resource.id)}
                {#key bookmark.resource.id}
                  <BookmarksMenuItem
                    {bookmark}
                    {onEdit}
                    onDelete={deleteBookmark}
                    onClick={() => recordUsage(bookmark)}
                    readOnly={!bookmark.resource.canManage}
                  />
                {/key}
              {/each}
            {:else}
              <div class="my-2 text-fg-muted text-center">
                {m.bookmark_no_shared()}
              </div>
            {/if}
          </DropdownMenuGroup>
        {/if}
      </div>
    {/if}
  </DropdownMenuContent>
</DropdownMenu>

{#if showDialog && resource}
  {#key `${resource.kind}:${resource.name}`}
    <BookmarksFormDialog
      {organization}
      {project}
      {projectId}
      {bookmark}
      {resource}
      {defaultUrlParams}
      {showFiltersOnly}
      {showSharedBookmarks}
      onClose={() => {
        showDialog = false;
        bookmark = null;
      }}
    />
  {/key}
{/if}
