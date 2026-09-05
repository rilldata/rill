<script lang="ts">
  import CreateEphemeralMeasureButton from "@rilldata/web-common/features/dashboards/ephemeral-measures/CreateEphemeralMeasureButton.svelte";
  import { ephemeralMeasureDialog } from "@rilldata/web-common/features/dashboards/ephemeral-measures/dialog-store";
  import { ephemeralMeasureNameSet } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
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

  function openEphemeralMeasureEditor(name: string) {
    const def = $dashboardStore?.ephemeralMeasures?.find(
      (d) => d.name === name,
    );
    if (def) ephemeralMeasureDialog.set({ def });
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
    ephemeralNames={ephemeralMeasureNameSet($dashboardStore?.ephemeralMeasures)}
    onEditEphemeral={openEphemeralMeasureEditor}
  >
    <div class="border-t border-border" slot="action" let:close>
      <CreateEphemeralMeasureButton onOpen={close} />
    </div>
  </DashboardMetricsDraggableList>

  <div class="flex flex-row flex-wrap mt-2 gap-x-7 gap-y-9">
    {#each $visibleMeasures as measure (measure.name)}
      <div class="inline-grid">
        <!-- FIXME: I can't select the big number by the measure id. -->
        <MeasureBigNumber
          {measure}
          ephemeralMeasures={$dashboardStore.ephemeralMeasures}
          withTimeseries={false}
          {metricsViewName}
          where={chartWhere}
          ready={chartReady}
        />
      </div>
    {/each}
  </div>
</div>
