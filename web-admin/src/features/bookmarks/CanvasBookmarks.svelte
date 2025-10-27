<script lang="ts">
  import Bookmarks from "@rilldata/web-admin/features/bookmarks/Bookmarks.svelte";
  import { getCanvasCategorisedBookmarks } from "@rilldata/web-admin/features/bookmarks/selectors.ts";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers.ts";
  import type { FiltersState } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
  import type { TimeControlState } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import { writable } from "svelte/store";

  export let organization: string;
  export let project: string;
  export let canvasName: string;

  $: ({ instanceId } = $runtime);

  const orgAndProjectNameStore = writable({ organization, project });
  $: orgAndProjectNameStore.set({ organization, project });

  const canvasNameStore = writable(canvasName);
  $: canvasNameStore.set(canvasName);

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

  const categorizedBookmarksStore = getCanvasCategorisedBookmarks(
    orgAndProjectNameStore,
    canvasNameStore,
  );
  $: ({
    data: { bookmarks, categorizedBookmarks },
  } = $categorizedBookmarksStore);
</script>

<Bookmarks
  {organization}
  {project}
  resource={{ name: canvasName, kind: ResourceKind.Canvas }}
  bookmarkData={{
    bookmarks,
    categorizedBookmarks,
    disableFiltersOnly: true,
  }}
  dashboardState={{
    metricsViewNames: $metricViewNames,
    filtersState,
    timeControlState,
  }}
/>
