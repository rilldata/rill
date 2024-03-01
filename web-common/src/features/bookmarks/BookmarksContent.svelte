<script lang="ts">
  import { page } from "$app/stores";
  import {
    DropdownMenuContent,
    DropdownMenuGroup,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
  } from "@rilldata/web-common/components/dropdown-menu/index";
  import BookmarkItem from "@rilldata/web-common/features/bookmarks/BookmarkItem.svelte";
  import { getBookmarks } from "@rilldata/web-common/features/bookmarks/selectors";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { BookmarkPlusIcon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import HomeBookmarkPlus from "@rilldata/web-common/components/icons/HomeBookmarkPlus.svelte";

  const dispatch = createEventDispatcher();
  const queryClient = useQueryClient();

  let bookmarks: ReturnType<typeof getBookmarks>;
  $: bookmarks = getBookmarks(
    queryClient,
    $page.params.organization,
    $page.params.project,
    $page.params.dashboard,
  );
</script>

<DropdownMenuContent class="w-[450px]">
  <DropdownMenuItem on:click={() => dispatch("create")}>
    <div class="flex flex-row gap-x-2 items-center">
      <BookmarkPlusIcon size="16px" strokeWidth={1.5} />
      <div>Bookmark current view</div>
    </div>
  </DropdownMenuItem>
  <DropdownMenuItem on:click={() => dispatch("create-home")}>
    <div class="flex flex-row gap-x-2 items-center">
      <HomeBookmarkPlus size="16px" />
      <div>Bookmark current view as Home</div>
    </div>
  </DropdownMenuItem>
  {#if $bookmarks.data}
    {#if $bookmarks.data.personal?.length}
      <DropdownMenuSeparator />
      <DropdownMenuGroup>
        <DropdownMenuLabel class="capitalize text-gray-500 text-sm">
          Your bookmarks
        </DropdownMenuLabel>
        {#each $bookmarks.data.personal as bookmark}
          <BookmarkItem {bookmark} on:edit on:select />
        {/each}
      </DropdownMenuGroup>
    {/if}
    {#if $bookmarks.data.shared?.length || $bookmarks.data.home}
      <DropdownMenuSeparator />
      <DropdownMenuGroup>
        <DropdownMenuLabel class="capitalize text-gray-500 text-sm">
          Default bookmarks
        </DropdownMenuLabel>
        {#if $bookmarks.data.home}
          <BookmarkItem bookmark={$bookmarks.data.home} on:edit on:select />
        {/if}
        {#each $bookmarks.data.shared as bookmark}
          <BookmarkItem {bookmark} on:edit on:select />
        {/each}
      </DropdownMenuGroup>
    {/if}
  {/if}
</DropdownMenuContent>
