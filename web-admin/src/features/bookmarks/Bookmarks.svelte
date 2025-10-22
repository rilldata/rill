<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateBookmark,
    createAdminServiceRemoveBookmark,
    createAdminServiceUpdateBookmark,
    getAdminServiceListBookmarksQueryKey,
    type V1Bookmark,
  } from "@rilldata/web-admin/client";
  import BookmarksMenuItem from "@rilldata/web-admin/features/bookmarks/BookmarksMenuItem.svelte";
  import BookmarksFormBookmarksFormDialog from "@rilldata/web-admin/features/bookmarks/BookmarksFormDialog.svelte";
  import { isHomeBookmark } from "@rilldata/web-admin/features/bookmarks/selectors.ts";
  import {
    type BookmarkEntry,
    type Bookmarks,
    getBookmarkData,
    searchBookmarks,
  } from "@rilldata/web-admin/features/bookmarks/utils.ts";
  import HomeBookmarkButton from "@rilldata/web-admin/features/bookmarks/HomeBookmarkButton.svelte";
  import {
    getProjectPermissions,
    useProjectId,
  } from "@rilldata/web-admin/features/projects/selectors.ts";
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
  import type { FiltersState } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
  import type { TimeControlState } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { BookmarkIcon, BookmarkPlusIcon } from "lucide-svelte";

  export let organization: string;
  export let project: string;
  export let metricsViewNames: string[];
  export let resourceKind: ResourceKind;
  export let resourceName: string;
  export let bookmarksResp: V1Bookmark[];
  export let categorizedBookmarks: Bookmarks;
  export let defaultUrlParams: URLSearchParams | undefined = undefined;
  export let defaultHomeBookmarkUrl: string = "";
  export let filtersState: FiltersState;
  export let timeControlState: TimeControlState;
  export let disableFiltersOnly: boolean = false;

  let showDialog = false;
  let bookmark: BookmarkEntry | null = null;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: curUrlParams = $page.url.searchParams;

  $: projectId = useProjectId(organization, project);

  $: projectPermissions = getProjectPermissions(organization, project);
  $: manageProject = $projectPermissions.data?.manageProject;

  const queryClient = useQueryClient();
  const bookmarkCreator = createAdminServiceCreateBookmark();
  const bookmarkUpdater = createAdminServiceUpdateBookmark();
  const bookmarkDeleter = createAdminServiceRemoveBookmark();

  async function createHomeBookmark() {
    const homeBookmarkUrlSearch = getBookmarkData({
      curUrlParams,
      defaultUrlParams,
    });
    const homeBookmark = bookmarksResp.find(isHomeBookmark);

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
          projectId: $projectId.data,
          resourceKind,
          resourceName,
          shared: true,
          default: true,
          urlSearch: homeBookmarkUrlSearch,
        },
      });
    }

    eventBus.emit("notification", {
      message: "Home bookmark created",
    });
    return queryClient.refetchQueries({
      queryKey: getAdminServiceListBookmarksQueryKey({
        projectId: $projectId.data ?? "",
        resourceKind,
        resourceName,
      }),
    });
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
      message: `Bookmark ${bookmark.resource.displayName} deleted`,
    });
    return queryClient.refetchQueries({
      queryKey: getAdminServiceListBookmarksQueryKey({
        projectId: $projectId.data ?? "",
        resourceKind,
        resourceName,
      }),
    });
  }

  let open = false;

  let searchText: string;
  $: filteredBookmarks = searchBookmarks(categorizedBookmarks, searchText);
</script>

<HomeBookmarkButton
  {organization}
  {project}
  {resourceKind}
  {resourceName}
  homeBookmark={categorizedBookmarks.home}
  {defaultHomeBookmarkUrl}
  onCreate={createHomeBookmark}
  onDelete={deleteBookmark}
  {manageProject}
/>

<DropdownMenu bind:open typeahead={false}>
  <DropdownMenuTrigger asChild let:builder>
    <Button
      builders={[builder]}
      compact
      type="secondary"
      label="Other bookmark dropdown"
      active={open}
    >
      <BookmarkIcon class="inline-flex" size="16px" />
    </Button>
  </DropdownMenuTrigger>
  <DropdownMenuContent class="w-[450px]">
    <DropdownMenuItem on:click={() => (showDialog = true)}>
      <div class="flex flex-row gap-x-2 items-center">
        <BookmarkPlusIcon size="16px" strokeWidth={1.5} />
        <div class="text-xs">Bookmark current view</div>
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
    {#if filteredBookmarks}
      <DropdownMenuSeparator />
      <DropdownMenuGroup>
        <DropdownMenuLabel class="text-gray-500 text-[10px] h-6 uppercase">
          Your bookmarks
        </DropdownMenuLabel>
        {#if filteredBookmarks.personal?.length}
          {#each filteredBookmarks.personal as bookmark (bookmark.resource.id)}
            {#key bookmark.resource.id}
              <BookmarksMenuItem
                {bookmark}
                {onEdit}
                onDelete={deleteBookmark}
                on:select
              />
            {/key}
          {/each}
        {:else}
          <div class="my-2 ui-copy-disabled text-center">
            You have no bookmarks for this dashboard.
          </div>
        {/if}
      </DropdownMenuGroup>
      <DropdownMenuSeparator />
      <DropdownMenuGroup>
        <DropdownMenuLabel class="text-gray-500">
          <div class="text-[10px] h-4 uppercase">Managed bookmarks</div>
          <div class="text-[11px] font-normal">Created by project admin</div>
        </DropdownMenuLabel>
        {#if filteredBookmarks.shared?.length}
          {#each filteredBookmarks.shared as bookmark (bookmark.resource.id)}
            {#key bookmark.resource.id}
              <BookmarksMenuItem
                {bookmark}
                {onEdit}
                onDelete={deleteBookmark}
                readOnly={!manageProject}
              />
            {/key}
          {/each}
        {:else}
          <div class="my-2 ui-copy-disabled text-center">
            There are no shared bookmarks for this dashboard.
          </div>
        {/if}
      </DropdownMenuGroup>
    {/if}
  </DropdownMenuContent>
</DropdownMenu>

{#if showDialog}
  <BookmarksFormBookmarksFormDialog
    {bookmark}
    {metricsViewNames}
    {resourceKind}
    {resourceName}
    {defaultUrlParams}
    {filtersState}
    {timeControlState}
    {disableFiltersOnly}
    onClose={() => {
      showDialog = false;
      bookmark = null;
    }}
  />
{/if}
