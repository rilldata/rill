<script lang="ts">
  import CreateEphemeralMeasureButton from "@rilldata/web-common/features/dashboards/ephemeral-measures/CreateEphemeralMeasureButton.svelte";
  import { ephemeralMeasureDialog } from "@rilldata/web-common/features/dashboards/ephemeral-measures/dialog-store";
  import { ephemeralMeasureNameSet } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
  import {
    precisionStateFor,
    requestExactNames,
  } from "@rilldata/web-common/features/dashboards/exact-measures/exact-measures";
  import { exploreExactSurface } from "@rilldata/web-common/features/dashboards/exact-measures/explore-surfaces";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import MeasureBigNumber from "./MeasureBigNumber.svelte";
  import DashboardMetricsDraggableList from "@rilldata/web-common/components/menu/DashboardMetricsDraggableList.svelte";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";

  export let metricsViewName: string;
  export let whereFilter: V1Expression | undefined;

  const ctx = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      measures: { allMeasures, visibleMeasures },
      tags: { measureTagIndex },
      exactMeasures: { exactMeasureContext },
    },
    actions: {
      measures: { setMeasureVisibility },
      exactMeasures: { toggleExactMeasure },
    },
  } = ctx;

  const timeControlsStore = useTimeControlStore(ctx);

  $: visibleMeasureNames = $visibleMeasures
    .map(({ name }) => name)
    .filter(isDefined);
  // The big numbers are one surface: a click applies to this row only.
  $: exactContext = $exactMeasureContext(exploreExactSurface.measures);
  $: exactRequestNames = requestExactNames(visibleMeasureNames, exactContext);
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
      {@const measureName = measure.name ?? ""}
      <div class="inline-grid">
        <!-- FIXME: I can't select the big number by the measure id. -->
        <MeasureBigNumber
          {measure}
          ephemeralMeasures={$dashboardStore.ephemeralMeasures}
          exact={exactRequestNames.has(measureName)}
          precision={precisionStateFor(measureName, exactContext)}
          onToggleExact={() =>
            toggleExactMeasure(exploreExactSurface.measures, measureName)}
          withTimeseries={false}
          {metricsViewName}
          where={whereFilter}
          ready={chartReady}
        />
      </div>
    {/each}
  </div>
</div>
