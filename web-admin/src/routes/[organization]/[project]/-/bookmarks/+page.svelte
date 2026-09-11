<script lang="ts">
  import { page } from "$app/state";
  import BookmarksTable from "@rilldata/web-admin/features/bookmarks/listing/BookmarksTable.svelte";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let { project } = $derived(page.params);
</script>

<svelte:head>
  <title>{m.project_bookmarks_page_title({ project })}</title>
</svelte:head>

<ContentContainer maxWidth={800} title={m.project_bookmarks_title()}>
  <div class="flex flex-col items-center gap-y-4">
    <BookmarksTable />
  </div>
</ContentContainer>
