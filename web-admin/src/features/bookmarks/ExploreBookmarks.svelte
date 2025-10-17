<script lang="ts">
  import Bookmarks from "@rilldata/web-admin/features/bookmarks/Bookmarks.svelte";
  import {
    categorizeBookmarks,
    parseBookmarks,
  } from "@rilldata/web-admin/features/bookmarks/form-utils.ts";
  import {
    exploreBookmarkDataTransformer,
    getBookmarks,
  } from "@rilldata/web-admin/features/bookmarks/selectors.ts";
  import { useProjectId } from "@rilldata/web-admin/features/projects/selectors.ts";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors.ts";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
  import type { FiltersState } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
  import { createUrlForExploreYAMLDefaultState } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config.ts";
  import type { TimeControlState } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
  import { getRillDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { createQueryServiceMetricsViewSchema } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";

  export let organization: string;
  export let project: string;
  export let metricsViewName: string;
  export let exploreName: string;

  $: ({ instanceId } = $runtime);

  $: projectId = useProjectId(organization, project);
  $: validSpecStore = useExploreValidSpec(instanceId, exploreName);

  $: metricsViewSpec = $validSpecStore.data?.metricsView ?? {};
  $: exploreSpec = $validSpecStore.data?.explore ?? {};

  $: metricsViewTimeRange = useMetricsViewTimeRange(
    instanceId,
    metricsViewName,
    {},
    queryClient,
  );
  $: timeRangeSummary = $metricsViewTimeRange.data?.timeRangeSummary;

  $: schemaResp = createQueryServiceMetricsViewSchema(
    instanceId,
    metricsViewName,
  );
  $: schema = $schemaResp.data?.schema ?? {};

  $: urlForExploreYAMLDefaultState = createUrlForExploreYAMLDefaultState(
    validSpecStore,
    metricsViewTimeRange,
  );

  $: rillDefaultExploreURLParams = getRillDefaultExploreUrlParams(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
  );

  $: exploreState = useExploreState(exploreName);
  $: filtersState = <FiltersState>{
    whereFilter: $exploreState.whereFilter,
    dimensionsWithInlistFilter: $exploreState.dimensionsWithInlistFilter,
    dimensionThresholdFilters: $exploreState.dimensionThresholdFilters,
    dimensionFilterExcludeMode: $exploreState.dimensionFilterExcludeMode,
  };
  $: exploreTimeControlState = getTimeControlState(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
    $exploreState,
  );
  $: timeControlState = <TimeControlState>{
    selectedTimeRange: exploreTimeControlState.selectedTimeRange,
    selectedComparisonTimeRange:
      exploreTimeControlState.selectedComparisonTimeRange,
    showTimeComparison: exploreTimeControlState.showTimeComparison,
    selectedTimezone: $exploreState.selectedTimezone,
  };

  $: bookmarks = getBookmarks(
    $projectId.data,
    ResourceKind.Explore,
    exploreName,
  );
  $: parsedBookmarks = parseBookmarks(
    $bookmarks.data?.bookmarks ?? [],
    (data) => {
      if (data.startsWith("?")) return data; // New format that has the params directly, start with '?'.
      return exploreBookmarkDataTransformer(
        data,
        metricsViewSpec,
        exploreSpec,
        schema,
        $exploreState,
        timeRangeSummary,
        rillDefaultExploreURLParams,
      );
    },
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
  defaultUrlParams={rillDefaultExploreURLParams}
  defaultHomeBookmarkUrl={$urlForExploreYAMLDefaultState}
  {filtersState}
  {timeControlState}
/>
