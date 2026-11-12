<script lang="ts">
  import EmbedBookmarks from "@rilldata/web-admin/features/bookmarks/EmbedBookmarks.svelte";
  import {
    embedBookmarksEnabled,
    getEmbedBookmarkResource,
  } from "./embed-bookmarks";
  import { page } from "$app/state";

  // Without the full top bar, dashboards get a lightweight header with just the bookmarks.
  let bookmarkResource = $derived(
    $embedBookmarksEnabled
      ? getEmbedBookmarkResource(page.route.id, page.params)
      : null,
  );
</script>

{#if bookmarkResource}
  <div class="fixed top-1 right-1 z-50">
    {#key `${bookmarkResource.kind}:${bookmarkResource.name}`}
      <EmbedBookmarks resource={bookmarkResource} />
    {/key}
  </div>
{/if}
