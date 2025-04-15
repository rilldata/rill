<script lang="ts">
  import BookmarkItem from "@rilldata/web-admin/features/bookmarks/BookmarksDropdownMenuItem.svelte";
  import {
    type BookmarkEntry,
    type Bookmarks,
    searchBookmarks,
  } from "@rilldata/web-admin/features/bookmarks/selectors";
  import {
    DropdownMenuContent,
    DropdownMenuGroup,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
  } from "@rilldata/web-common/components/dropdown-menu";
  import { Search } from "@rilldata/web-common/components/search";
  import { BookmarkPlusIcon } from "lucide-svelte";

  export let categorizedBookmarks: Bookmarks;
  export let manageProject: boolean;

  export let onCreate: () => void;
  export let onEdit: (bookmark: BookmarkEntry) => void;
  export let onDelete: (bookmark: BookmarkEntry) => Promise<void>;

  let searchText: string;
  $: filteredBookmarks = searchBookmarks(categorizedBookmarks, searchText);
</script>

<DropdownMenuContent class="w-[450px]">
  <DropdownMenuItem on:click={onCreate}>
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
        {#each filteredBookmarks.personal as bookmark}
          {#key bookmark.resource.id}
            <BookmarkItem {bookmark} {onEdit} {onDelete} on:select />
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
        {#each filteredBookmarks.shared as bookmark}
          {#key bookmark.resource.id}
            <BookmarkItem
              {bookmark}
              {onEdit}
              {onDelete}
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
