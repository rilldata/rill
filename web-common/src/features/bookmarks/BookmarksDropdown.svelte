<script lang="ts">
  import { page } from "$app/stores";
  import type { V1Bookmark } from "@rilldata/web-admin/client";
  import {
    DropdownMenu,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import BookmarksContent from "@rilldata/web-common/features/bookmarks/BookmarksContent.svelte";
  import CreateBookmarkDialog from "@rilldata/web-common/features/bookmarks/CreateBookmarkDialog.svelte";
  import EditBookmarkDialog from "@rilldata/web-common/features/bookmarks/EditBookmarkDialog.svelte";
  import { BookmarkIcon } from "lucide-svelte";

  let createBookmark = false;
  let editBookmark = false;
  let bookmark: V1Bookmark;
</script>

<DropdownMenu>
  <DropdownMenuTrigger>
    <BookmarkIcon class="inline-flex" />
  </DropdownMenuTrigger>
  <BookmarksContent
    on:create={() => (createBookmark = true)}
    on:edit={({ detail }) => {
      editBookmark = true;
      bookmark = detail;
    }}
  />
</DropdownMenu>

<CreateBookmarkDialog
  bind:open={createBookmark}
  metricsViewName={$page.params.dashboard}
/>

{#if bookmark}
  <EditBookmarkDialog
    {bookmark}
    bind:open={editBookmark}
    metricsViewName={$page.params.dashboard}
  />
{/if}
