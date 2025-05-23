<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { usePivotForExplore } from "@rilldata/web-common/features/dashboards/pivot/pivot-data-store";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { formatCompactInteger } from "@rilldata/web-common/lib/formatters";
  import { createQueryServiceMetricsViewAggregation } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { get } from "svelte/store";
  import { useExploreState } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import ExportMenu from "../../exports/ExportMenu.svelte";
  import { featureFlags } from "../../feature-flags";
  import { mergeDimensionAndMeasureFilters } from "../filters/measure-filters/measure-filter-utils";
  import type { PivotFilter } from "../pivot/types";
  import RowsViewer from "./RowsViewer.svelte";

  const { exports } = featureFlags;
  const timeControlsStore = useTimeControlStore(getStateManagers());

  export let metricsViewName: string;
  export let exploreName: string;

  const DEFAULT_LABEL = "Model Data";
  const INITIAL_HEIGHT_EXPANDED = 300;
  const MIN_HEIGHT_EXPANDED = 30;
  const MAX_HEIGHT_EXPANDED = 1000;
  const PIVOT_HEIGHT_EXPANDED = 200;

  let isOpen = false;
  let rowCountlabel = "";
  let label = DEFAULT_LABEL;
  let height = INITIAL_HEIGHT_EXPANDED;

  let manualClose = false;

  const stateManagers = getStateManagers();
  const {
    selectors: {
      pivot: { showPivot: showPivotStore },
    },
  } = stateManagers;

  $: ({ instanceId } = $runtime);

  $: exploreState = useExploreState(exploreName);
  $: ({ whereFilter, dimensionThresholdFilters } = $exploreState);
  $: pivotDataStore = usePivotForExplore(stateManagers);
  $: ({ activeCellFilters } = $pivotDataStore);
  $: showPivot = $showPivotStore;
  $: isPivotCellSelected = Boolean(showPivot && activeCellFilters);

  $: label = isPivotCellSelected
    ? "Model data for selected cell"
    : DEFAULT_LABEL;

  $: timeRange = isPivotCellSelected
    ? (activeCellFilters as PivotFilter).timeRange
    : {
        start: $timeControlsStore.timeStart,
        end: $timeControlsStore.timeEnd,
      };

  $: filters = isPivotCellSelected
    ? sanitiseExpression((activeCellFilters as PivotFilter).filters, undefined)
    : sanitiseExpression(
        mergeDimensionAndMeasureFilters(whereFilter, dimensionThresholdFilters),
        undefined,
      );

  $: filteredTotalsQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: [{ name: "count", builtinMeasure: "BUILTIN_MEASURE_COUNT" }],
      timeStart: timeRange.start,
      timeEnd: timeRange.end,
      where: filters,
    },
    {
      query: {
        queryKey: [
          "dashboardFilteredRowsCt",
          metricsViewName,
          {
            timeStart: timeRange.start,
            timeEnd: timeRange.end,
            where: filters,
          },
        ],
        enabled: $timeControlsStore.ready && !!$exploreState?.whereFilter,
      },
    },
  );

  $: totalsQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: [{ name: "count", builtinMeasure: "BUILTIN_MEASURE_COUNT" }],
    },
    {
      query: {
        // This should not be needed, but we are getting occasional query failures where the query gets stuck on status: 'loading' when using the autogenerated key.
        // TODO: investigate this further
        queryKey: ["dashboardAllRowsCt", metricsViewName],
        enabled: true,
      },
    },
  );

  $: {
    if ($filteredTotalsQuery.data && $totalsQuery.data) {
      const numerator = $filteredTotalsQuery.data?.data?.[0]["count"] as number;
      const denominator = $totalsQuery.data.data?.[0]["count"] as number;
      rowCountlabel = `${formatCompactInteger(numerator)} of ${formatCompactInteger(denominator)} rows`;
    }
  }

  // Clicking on a pivot cell should open the rows viewer, if it is not already open and hasn't been manually closed
  $: if (isPivotCellSelected && !isOpen && !manualClose) {
    height = PIVOT_HEIGHT_EXPANDED;
    isOpen = true;
  }

  function toggle() {
    manualClose = true;
    isOpen = !isOpen;
  }

  function getExportQuery() {
    return {
      metricsViewRowsRequest: {
        instanceId: get(runtime).instanceId,
        metricsViewName,
        timeStart: timeRange.start,
        timeEnd: timeRange.end,
        where: sanitiseExpression(
          mergeDimensionAndMeasureFilters(
            $exploreState.whereFilter,
            $exploreState.dimensionThresholdFilters,
          ),
          undefined,
        ),
      },
    };
  }
</script>

<div
  class="relative w-full flex-none overflow-hidden flex flex-col bg-gray-100"
>
  <Resizer
    disabled={!isOpen}
    dimension={height}
    min={MIN_HEIGHT_EXPANDED}
    max={MAX_HEIGHT_EXPANDED}
    basis={INITIAL_HEIGHT_EXPANDED}
    onUpdate={(dimension) => (height = dimension)}
    direction="NS"
  />
  <div class="bar">
    <button
      aria-label="Toggle rows viewer"
      class="text-xs text-gray-800 rounded-sm hover:bg-gray-200 h-6 px-1.5 py-px flex items-center gap-1"
      on:click={toggle}
    >
      <span class:rotate-180={isOpen}>
        <CaretDownIcon size="14px" />
      </span>
      <span class="font-bold">{label}</span>
      {rowCountlabel}
    </button>
    {#if $exports}
      <div class="ml-auto">
        <ExportMenu label="Export model data" getQuery={getExportQuery} />
      </div>
    {/if}
  </div>

  {#if isOpen}
    <RowsViewer
      {filters}
      {timeRange}
      {metricsViewName}
      {exploreName}
      {height}
    />
  {/if}
</div>

<style lang="postcss">
  .bar {
    @apply flex items-center px-2 h-7 w-full bg-gray-100 border-t;
  }
</style>
