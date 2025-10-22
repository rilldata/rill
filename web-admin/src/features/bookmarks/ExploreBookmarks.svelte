<script lang="ts">
  import { page } from "$app/stores";
  import Bookmarks from "@rilldata/web-admin/features/bookmarks/Bookmarks.svelte";
  import { createExploreBookmarkLegacyDataTransformer } from "@rilldata/web-admin/features/bookmarks/explore-bookmark-legacy-data-transformer.ts";
  import {
    categorizeBookmarks,
    parseBookmarks,
  } from "@rilldata/web-admin/features/bookmarks/utils.ts";
  import { getBookmarks } from "@rilldata/web-admin/features/bookmarks/selectors.ts";
  import { useProjectId } from "@rilldata/web-admin/features/projects/selectors.ts";
  import { getMetricsViewTimeRangeOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
  import type { FiltersState } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
  import { createUrlForExploreYAMLDefaultState } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config.ts";
  import type { TimeControlState } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
  import { createRillDefaultExploreUrlParamsV2 } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { getExploreValidSpecOptions } from "@rilldata/web-common/features/explores/selectors.ts";
  import { createQuery } from "@tanstack/svelte-query";
  import { writable } from "svelte/store";

  export let organization: string;
  export let project: string;
  export let metricsViewName: string;
  export let exploreName: string;

  const exploreNameStore = writable(exploreName);
  $: exploreNameStore.set(exploreName);

  $: projectId = useProjectId(organization, project);
  const validSpecQuery = createQuery(
    getExploreValidSpecOptions(exploreNameStore),
  );

  $: metricsViewSpec = $validSpecQuery.data?.metricsView ?? {};
  $: exploreSpec = $validSpecQuery.data?.explore ?? {};

  const metricsViewTimeRangeQuery = createQuery(
    getMetricsViewTimeRangeOptions(exploreNameStore),
  );
  $: timeRangeSummary = $metricsViewTimeRangeQuery.data?.timeRangeSummary;

  const urlForExploreYAMLDefaultState =
    createUrlForExploreYAMLDefaultState(exploreNameStore);

  const rillDefaultExploreURLParams =
    createRillDefaultExploreUrlParamsV2(exploreNameStore);

  $: exploreState = useExploreState(exploreName);
  $: filtersState = <FiltersState>{
    whereFilter: $exploreState?.whereFilter,
    dimensionsWithInlistFilter: $exploreState?.dimensionsWithInlistFilter,
    dimensionThresholdFilters: $exploreState?.dimensionThresholdFilters,
    dimensionFilterExcludeMode: $exploreState?.dimensionFilterExcludeMode,
  };
  $: exploreTimeControlState = $exploreState
    ? getTimeControlState(
        metricsViewSpec,
        exploreSpec,
        timeRangeSummary,
        $exploreState,
      )
    : undefined;
  $: timeControlState = <TimeControlState>{
    selectedTimeRange: exploreTimeControlState?.selectedTimeRange,
    selectedComparisonTimeRange:
      exploreTimeControlState?.selectedComparisonTimeRange,
    showTimeComparison: exploreTimeControlState?.showTimeComparison,
    selectedTimezone: $exploreState?.selectedTimezone,
  };

  const exploreBookmarkDataTransformer =
    createExploreBookmarkLegacyDataTransformer(exploreNameStore);

  $: bookmarks = getBookmarks(
    $projectId.data,
    ResourceKind.Explore,
    exploreName,
  );
  $: parsedBookmarks = parseBookmarks(
    $bookmarks.data?.bookmarks ?? [],
    $page.url.searchParams,
    $rillDefaultExploreURLParams,
    $exploreBookmarkDataTransformer,
  );
  $: categorizedBookmarks = categorizeBookmarks(parsedBookmarks);
</script>

<Bookmarks
  {organization}
  {project}
  {metricsViewName}
  resourceKind={ResourceKind.Explore}
  resourceName={exploreName}
  bookmarksResp={$bookmarks.data?.bookmarks ?? []}
  {categorizedBookmarks}
  defaultUrlParams={$rillDefaultExploreURLParams}
  defaultHomeBookmarkUrl={$urlForExploreYAMLDefaultState}
  {filtersState}
  {timeControlState}
/>
