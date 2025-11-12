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

  // Get url params for the explore based on yaml defaults.
  // This is used for home bookmarks button when there is no explicit home bookmark created.
  const urlForExploreYAMLDefaultState =
    createUrlForExploreYAMLDefaultState(exploreNameStore);

  // Rill opinionated url params that are removed from url to keep the url short.
  // To keep bookmarks exhaustive, these are added on top of current url params while creating bookmarks.
  const rillDefaultExploreURLParams =
    createRillDefaultExploreUrlParamsV2(exploreNameStore);

  // Transformer for legacy bookmark data that was stored in proto format.
  const exploreBookmarkLegacyDataTransformer =
    createExploreBookmarkLegacyDataTransformer(exploreNameStore);

  // Stable query object for bookmarks for this explore
  const bookmarksQuery = createQuery(
    getBookmarksQueryOptions(
      orgAndProjectNameStore,
      ResourceKind.Explore,
      exploreNameStore,
    ),
  );
  $: bookmarks = $bookmarksQuery.data?.bookmarks ?? [];

  // Parse bookmarks and fill in metadata based on bookmark data.
  $: parsedBookmarks = parseBookmarks(
    bookmarks,
    $page.url.searchParams,
    $rillDefaultExploreURLParams,
    $exploreBookmarkLegacyDataTransformer,
  );
  // Categorize bookmarks into home, shared and personal bookmarks.
  $: categorizedBookmarks = categorizeBookmarks(parsedBookmarks);
</script>

<Bookmarks
  {organization}
  {project}
  resource={{ name: exploreName, kind: ResourceKind.Explore }}
  bookmarkData={{
    bookmarks,
    categorizedBookmarks,
    defaultUrlParams: $rillDefaultExploreURLParams,
    defaultHomeBookmarkUrl: $urlForExploreYAMLDefaultState,
  }}
  metricsViewNames={[metricsViewName]}
/>
