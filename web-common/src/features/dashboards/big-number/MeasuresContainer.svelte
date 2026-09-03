<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import MeasureBigNumber from "./MeasureBigNumber.svelte";
  import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
  import DashboardMetricsDraggableList from "@rilldata/web-common/components/menu/DashboardMetricsDraggableList.svelte";

  export let metricsViewName: string;

  const ctx = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      measures: { allMeasures, visibleMeasures },
      tags: { measureTagIndex },
    },
    actions: {
      measures: { setMeasureVisibility },
    },
  } = ctx;

  const timeControlsStore = useTimeControlStore(ctx);

  $: visibleMeasureNames = $visibleMeasures
    .map(({ name }) => name)
    .filter(isDefined);
  $: allMeasureNames = $allMeasures.map(({ name }) => name).filter(isDefined);
  function isDefined(value: string | undefined): value is string {
    return value !== undefined;
  }

  // Query-context props for MeasureBigNumber
  $: chartWhere = sanitiseExpression(
    mergeDimensionAndMeasureFilters(
      $dashboardStore?.whereFilter,
      $dashboardStore?.dimensionThresholdFilters,
    ),
    undefined,
  );
  $: chartReady = !!$timeControlsStore.ready;
</script>

<div class="overflow-y-scroll">
  <DashboardMetricsDraggableList
    type="measure"
    onSelectedChange={(items) => setMeasureVisibility(items, allMeasureNames)}
    allItems={$allMeasures}
    tagIndex={$measureTagIndex}
    selectedItems={visibleMeasureNames}
  />

  <div class="flex flex-row flex-wrap mt-2 gap-x-7 gap-y-9">
    {#each $visibleMeasures as measure (measure.name)}
      <div class="inline-grid">
        <!-- FIXME: I can't select the big number by the measure id. -->
        <MeasureBigNumber
          {measure}
          withTimeseries={false}
          {metricsViewName}
          where={chartWhere}
          ready={chartReady}
        />
      </div>
    {/each}
  </div>
</div>
