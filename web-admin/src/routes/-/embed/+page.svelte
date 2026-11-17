<script lang="ts">
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import TableToolbarSort from "@rilldata/web-common/components/table-toolbar/TableToolbarSort.svelte";
  import BookmarksTable from "@rilldata/web-admin/features/bookmarks/listing/BookmarksTable.svelte";
  import { BookmarkTableSortOptions } from "@rilldata/web-admin/features/bookmarks/listing/bookmark-listing-utils.ts";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
  import PersonalCanvasesList from "@rilldata/web-admin/features/personal-files/canvas/PersonalCanvasesList.svelte";
  import { embedBookmarksEnabled } from "@rilldata/web-admin/features/embeds/embed-bookmarks.ts";
  import { embedPersonalCanvasesEnabled } from "@rilldata/web-admin/features/embeds/embed-personal-canvases.ts";
  import { getOrgAndProjectForEmbed } from "@rilldata/web-common/features/embeds/selectors.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { UrlParamsState } from "@rilldata/web-common/lib/store-utils/url-params-state.svelte.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const runtimeClient = useRuntimeClient();
  const orgAndProjectQuery = getOrgAndProjectForEmbed(runtimeClient);
  $: ({ organization, project } = $orgAndProjectQuery);

  // Separate param from the dashboards table's sort so the two sections sort independently.
  const bookmarksSortStore = UrlParamsState.createStringParam(
    "bookmarks_sort",
    BookmarkTableSortOptions[0].value,
  );
</script>

<ContentContainer>
  <div class="flex flex-col gap-y-4">
    {#if $embedPersonalCanvasesEnabled && organization && project}
      <PersonalCanvasesList org={organization} {project} />
    {/if}
    <div class="flex flex-col items-center gap-y-4">
      <DashboardsTable isEmbedded />
    </div>
    {#if $embedBookmarksEnabled}
      <div class="flex flex-col gap-y-4">
        <h2 class="flex flex-row gap-x-2 items-center text-lg font-medium">
          <span>{m.home_bookmarks_heading()}</span>
          <TableToolbarSort
            sortStore={bookmarksSortStore}
            sortOptions={BookmarkTableSortOptions}
            size="sm"
            noOutline
          />
        </h2>
        <BookmarksTable isEmbedded isPreview sortStore={bookmarksSortStore} />
      </div>
    {/if}
  </div>
</ContentContainer>
