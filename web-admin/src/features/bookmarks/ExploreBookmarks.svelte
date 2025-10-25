<script lang="ts">
  import { page } from "$app/stores";
  import Bookmarks from "@rilldata/web-admin/features/bookmarks/Bookmarks.svelte";
  import { createExploreBookmarkLegacyDataTransformer } from "@rilldata/web-admin/features/bookmarks/explore-bookmark-legacy-data-transformer.ts";
  import {
    categorizeBookmarks,
    parseBookmarks,
  } from "@rilldata/web-admin/features/bookmarks/utils.ts";
  import { getBookmarksQueryOptions } from "@rilldata/web-admin/features/bookmarks/selectors.ts";
  import { createUrlForExploreYAMLDefaultState } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config.ts";
  import {
    getFilterStateFromNameStore,
    getTimeControlsStateFromNameStore,
  } from "@rilldata/web-common/features/dashboards/stores/utils.ts";
  import { createRillDefaultExploreUrlParamsV2 } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { createQuery } from "@tanstack/svelte-query";
  import { writable } from "svelte/store";

  export let organization: string;
  export let project: string;
  export let metricsViewName: string;
  export let exploreName: string;

  const orgAndProjectNameStore = writable({ organization, project });
  $: orgAndProjectNameStore.set({ organization, project });

  const exploreNameStore = writable(exploreName);
  $: exploreNameStore.set(exploreName);

  const urlForExploreYAMLDefaultState =
    createUrlForExploreYAMLDefaultState(exploreNameStore);

  const rillDefaultExploreURLParams =
    createRillDefaultExploreUrlParamsV2(exploreNameStore);

  const filtersStore = getFilterStateFromNameStore(exploreNameStore);
  const timeControlsStore = getTimeControlsStateFromNameStore(exploreNameStore);

  const exploreBookmarkDataTransformer =
    createExploreBookmarkLegacyDataTransformer(exploreNameStore);

  const bookmarksQuery = createQuery(
    getBookmarksQueryOptions(
      orgAndProjectNameStore,
      ResourceKind.Explore,
      exploreNameStore,
    ),
  );
  $: bookmarks = $bookmarksQuery.data?.bookmarks ?? [];

  $: parsedBookmarks = parseBookmarks(
    bookmarks,
    $page.url.searchParams,
    $rillDefaultExploreURLParams,
    $exploreBookmarkDataTransformer,
  );
  $: categorizedBookmarks = categorizeBookmarks(parsedBookmarks);
</script>

<Bookmarks
  {organization}
  {project}
  metricsViewNames={[metricsViewName]}
  resourceKind={ResourceKind.Explore}
  resourceName={exploreName}
  {bookmarks}
  {categorizedBookmarks}
  defaultUrlParams={$rillDefaultExploreURLParams}
  defaultHomeBookmarkUrl={$urlForExploreYAMLDefaultState}
  filtersState={$filtersStore}
  timeControlState={$timeControlsStore}
/>
