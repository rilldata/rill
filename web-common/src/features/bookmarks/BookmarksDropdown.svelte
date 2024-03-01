<script lang="ts">
  import { page } from "$app/stores";
  import {
    DropdownMenu,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import { createBookmarkApplier } from "@rilldata/web-common/features/bookmarks/applyBookmark";
  import BookmarksContent from "@rilldata/web-common/features/bookmarks/BookmarksContent.svelte";
  import CreateBookmarkDialog from "@rilldata/web-common/features/bookmarks/CreateBookmarkDialog.svelte";
  import { createHomeBookmarkModifier } from "@rilldata/web-common/features/bookmarks/createOrUpdateHomeBookmark";
  import EditBookmarkDialog from "@rilldata/web-common/features/bookmarks/EditBookmarkDialog.svelte";
  import { getBookmarkDataForDashboard } from "@rilldata/web-common/features/bookmarks/getBookmarkDataForDashboard";
  import type { BookmarkEntry } from "@rilldata/web-common/features/bookmarks/selectors";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { BookmarkIcon } from "lucide-svelte";

  $: metricsViewName = $page.params.dashboard;

  let createBookmark = false;
  let editBookmark = false;
  let bookmark: BookmarkEntry;

  $: bookmarkApplier = createBookmarkApplier(
    $runtime?.instanceId,
    metricsViewName,
  );

  $: dashboardStore = useDashboardStore(metricsViewName);

  const homeBookmarkModifier = createHomeBookmarkModifier();

  function onSelect(bookmark: BookmarkEntry) {
    bookmarkApplier(bookmark.resource);
  }

  async function createHomeBookmark() {
    await homeBookmarkModifier(
      getBookmarkDataForDashboard($dashboardStore, false, false),
    );
  }

  let open = false;
</script>

<DropdownMenu bind:open>
  <DropdownMenuTrigger>
    <BookmarkIcon class="inline-flex" fill={open ? "black" : "none"} />
  </DropdownMenuTrigger>
  <BookmarksContent
    on:create={() => (createBookmark = true)}
    on:create-home={() => createHomeBookmark()}
    on:edit={({ detail }) => {
      editBookmark = true;
      bookmark = detail;
    }}
    on:select={({ detail }) => onSelect(detail)}
  />
</DropdownMenu>

<CreateBookmarkDialog bind:open={createBookmark} {metricsViewName} />

{#if bookmark}
  <EditBookmarkDialog {bookmark} bind:open={editBookmark} {metricsViewName} />
{/if}
