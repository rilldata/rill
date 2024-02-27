<script lang="ts">
  import { page } from "$app/stores";
  import {
    DropdownMenuContent,
    DropdownMenuGroup,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
  } from "@rilldata/web-common/components/dropdown-menu/index";
  import { getBookmarks } from "@rilldata/web-common/features/bookmarks/selectors";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { BookmarkPlusIcon } from "lucide-svelte";
  import BookmarkItem from "@rilldata/web-common/features/bookmarks/BookmarkItem.svelte";
  import { createEventDispatcher } from "svelte";

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
  <DropdownMenuItem on:click={() => dispatch("new-bookmark")}>
    <div class="flex flex-row items-center">
      <BookmarkPlusIcon size="16px" />
      <div>Bookmark current view</div>
    </div>
  </DropdownMenuItem>
  <DropdownMenuItem>
    <div class="flex flex-row items-center">
      <BookmarkPlusIcon size="16px" />
      <div>Bookmark current view as Home</div>
    </div>
  </DropdownMenuItem>
  <DropdownMenuSeparator />
  {#if $bookmarks.data}
    <DropdownMenuGroup>
      <DropdownMenuLabel class="capitalize text-gray-500 text-sm">
        Your bookmarks
      </DropdownMenuLabel>
      {#each $bookmarks.data.own as bookmark}
        <DropdownMenuItem><BookmarkItem {bookmark} /></DropdownMenuItem>
      {/each}
    </DropdownMenuGroup>
    <DropdownMenuSeparator />
    <DropdownMenuGroup>
      <DropdownMenuLabel class="capitalize text-gray-500 text-sm">
        Default bookmarks
      </DropdownMenuLabel>
      {#if $bookmarks.data.home}
        <DropdownMenuItem>
          <BookmarkItem bookmark={$bookmarks.data.home} />
        </DropdownMenuItem>
      {/if}
      {#each $bookmarks.data.global as bookmark}
        <DropdownMenuItem><BookmarkItem {bookmark} /></DropdownMenuItem>
      {/each}
    </DropdownMenuGroup>
  {/if}
</DropdownMenuContent>
