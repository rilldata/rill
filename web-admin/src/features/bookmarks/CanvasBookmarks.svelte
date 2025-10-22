<script lang="ts">
  import { page } from "$app/stores";
  import Bookmarks from "@rilldata/web-admin/features/bookmarks/Bookmarks.svelte";
  import { getBookmarks } from "@rilldata/web-admin/features/bookmarks/selectors.ts";
  import {
    categorizeBookmarks,
    parseBookmarks,
  } from "@rilldata/web-admin/features/bookmarks/utils.ts";
  import { useProjectId } from "@rilldata/web-admin/features/projects/selectors.ts";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers.ts";
  import type { FiltersState } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
  import type { TimeControlState } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";

  export let organization: string;
  export let project: string;
  export let canvasName: string;

  $: ({ instanceId } = $runtime);
  $: projectId = useProjectId(organization, project);

  $: ({
    canvasEntity: {
      metricsView: { metricViewNames },
      filters: {
        whereFilter,
        dimensionsWithInlistFilter,
        dimensionThresholdFilters,
        dimensionFilterExcludeMode,
      },
      timeControls: {
        timeRangeStateStore,
        comparisonRangeStateStore,
        selectedTimezone,
      },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: filtersState = <FiltersState>{
    whereFilter: $whereFilter,
    dimensionsWithInlistFilter: $dimensionsWithInlistFilter,
    dimensionThresholdFilters: $dimensionThresholdFilters,
    dimensionFilterExcludeMode: $dimensionFilterExcludeMode,
  };
  $: timeControlState = <TimeControlState>{
    selectedTimeRange: $timeRangeStateStore?.selectedTimeRange,
    selectedComparisonTimeRange:
      $comparisonRangeStateStore?.selectedComparisonTimeRange,
    showTimeComparison: $comparisonRangeStateStore?.showTimeComparison,
    selectedTimezone: $selectedTimezone,
  };

  $: bookmarks = getBookmarks($projectId.data, ResourceKind.Canvas, canvasName);
  $: parsedBookmarks = parseBookmarks(
    $bookmarks.data?.bookmarks ?? [],
    $page.url.searchParams,
  );
  $: categorizedBookmarks = categorizeBookmarks(parsedBookmarks);
</script>

<Bookmarks
  {organization}
  {project}
  metricsViewNames={$metricViewNames}
  resourceKind={ResourceKind.Canvas}
  resourceName={canvasName}
  bookmarksResp={$bookmarks.data?.bookmarks ?? []}
  {categorizedBookmarks}
  {filtersState}
  {timeControlState}
  disableFiltersOnly
/>
