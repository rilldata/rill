<script lang="ts">
  import { page } from "$app/stores";
  import {
    DropdownMenu,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import BookmarksContent from "@rilldata/web-common/features/bookmarks/BookmarksContent.svelte";
  import CreateBookmarkDialog from "@rilldata/web-common/features/bookmarks/CreateBookmarkDialog.svelte";
  import { BookmarkIcon } from "lucide-svelte";

  let createBookmark = false;
</script>

<DropdownMenu>
  <DropdownMenuTrigger>
    <BookmarkIcon class="inline-flex" />
  </DropdownMenuTrigger>
  <BookmarksContent on:new-bookmark={() => (createBookmark = true)} />
</DropdownMenu>

<CreateBookmarkDialog
  bind:open={createBookmark}
  metricsViewName={$page.params.dashboard}
/>
