<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { MEASURE_CONFIG } from "../config";
  import MeasureBigNumber from "./MeasureBigNumber.svelte";
  import DashboardVisibilityDropdown from "@rilldata/web-common/components/menu/DashboardVisibilityDropdown.svelte";
  import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";

  export let metricsViewName: string;
  export let exploreContainerWidth: number;

  const MEASURE_HEIGHT = 60;
  const MEASURE_HEIGHT_MULTILINE = 80;
  const MEASURE_WIDTH = 175;
  const MARGIN_TOP = 36;
  const COLUMN_GAP = 28;
  const GRID_MARGIN_TOP = 8;

  // external sizes
  const MIN_LEADERBOARD_WIDTH = 355;
  const MEASURES_PADDING_LEFT = 44;
  const LEADERBOARD_PADDING_RIGHT = 16;

  $: maxWidthMeasuresContainer =
    exploreContainerWidth -
    MIN_LEADERBOARD_WIDTH -
    MEASURES_PADDING_LEFT -
    LEADERBOARD_PADDING_RIGHT;

  const ctx = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      measures: { allMeasures, visibleMeasures },
    },
    actions: {
      measures: { toggleMeasureVisibility, toggleAllMeasuresVisibility },
    },
  } = ctx;

  $: ({ instanceId } = $runtime);

  const timeControlsStore = useTimeControlStore(ctx);

  let metricsContainerHeight: number;
  let measureNodes: HTMLDivElement[] = [];
  let measuresWrapper;
  let measuresHeight: number[] = [];
  let measureGridHeights: number[] = [];
  let containerWidths = MEASURE_CONFIG.bigNumber.widthWithoutChart;

  $: visibleMeasureNames = $visibleMeasures
    .map(({ name }) => name)
    .filter(isDefined);
  $: allMeasureNames = $allMeasures.map(({ name }) => name).filter(isDefined);
  function isDefined(value: string | undefined): value is string {
    return value !== undefined;
  }

  function getMeasureHeightsForColumn(measuresHeight, numColumns) {
    const recalculatedHeights = [...measuresHeight];
    if (numColumns == 1) return recalculatedHeights;
    for (let i = 0; i < measuresHeight.length; i = i + numColumns) {
      const row = measuresHeight.slice(i, i + numColumns);
      if (row.indexOf(MEASURE_HEIGHT_MULTILINE) != -1) {
        for (let j = i; j < i + numColumns && j < measuresHeight.length; j++) {
          recalculatedHeights[j] = MEASURE_HEIGHT_MULTILINE;
        }
      }
    }
    return recalculatedHeights;
  }

  function calculateGridColumns() {
    measuresHeight = measureNodes.map(
      (measureNode) => measureNode?.offsetHeight,
    );

    const minInMeasures = Math.min(...measuresHeight);
    measuresHeight = measuresHeight.map((height) =>
      height > minInMeasures ? MEASURE_HEIGHT_MULTILINE : MEASURE_HEIGHT,
    );
    const totalMeasuresHeight = measuresHeight.reduce(
      (s, v) => s + v + MARGIN_TOP,
      0,
    );

    if (totalMeasuresHeight && metricsContainerHeight) {
      let columns = totalMeasuresHeight / metricsContainerHeight;
      if (columns <= 1 || columns > 2) {
        numColumns = Math.min(Math.ceil(columns), 3);
        measureGridHeights = getMeasureHeightsForColumn(
          measuresHeight,
          numColumns,
        );
      } else {
        numColumns = 2;
        measureGridHeights = getMeasureHeightsForColumn(measuresHeight, 2);

        // Check if two columns can individually accommodate all measures without scroll
        const measuresHeightInSingleColumn = measureGridHeights
          .filter((_, i) => i % 2 == 0)
          .reduce((s, v) => s + v + MARGIN_TOP, 0);
        const extraHeight =
          metricsContainerHeight - measuresHeightInSingleColumn;
        if (extraHeight < 0) {
          numColumns = 3;
          measureGridHeights = getMeasureHeightsForColumn(measuresHeight, 3);
        }
      }
    }

    // Check if there is any horizontal overlap between measures
    if (maxWidthMeasuresContainer) {
      while (numColumns > 1) {
        const widthPerColumn = maxWidthMeasuresContainer / numColumns;
        if (widthPerColumn < MEASURE_WIDTH + COLUMN_GAP / 2) {
          numColumns = numColumns - 1;
          measureGridHeights = getMeasureHeightsForColumn(
            measuresHeight,
            numColumns,
          );
        } else break;
      }
    }
  }

  $: numColumns = 3;

  // Query-context props for MeasureBigNumber
  $: chartWhere = sanitiseExpression(
    mergeDimensionAndMeasureFilters(
      $dashboardStore?.whereFilter,
      $dashboardStore?.dimensionThresholdFilters,
    ),
    undefined,
  );
  $: chartReady = !!$timeControlsStore.ready;

  $: if (metricsContainerHeight && measureNodes.length) {
    calculateGridColumns();
  }
</script>

<svelte:window on:resize={() => calculateGridColumns()} />
<div
  class="overflow-y-scroll"
  style:height="calc(100% - {GRID_MARGIN_TOP}px)"
  style:width={containerWidths[numColumns]}
  bind:clientHeight={metricsContainerHeight}
>
  <div
    bind:this={measuresWrapper}
    class="grid grid-cols-{numColumns}"
    style:column-gap="{COLUMN_GAP}px"
  >
    <div class="bg-surface-subtle sticky top-0">
      <DashboardVisibilityDropdown
        category="Measures"
        tooltipText="Choose measures to display"
        onSelect={(name) => toggleMeasureVisibility(allMeasureNames, name)}
        selectableItems={$allMeasures.map(({ name, displayName }) => ({
          name: name || "",
          label: displayName || name || "",
        }))}
        selectedItems={visibleMeasureNames}
        onToggleSelectAll={() => {
          toggleAllMeasuresVisibility(allMeasureNames);
        }}
      />
    </div>

    {#each $visibleMeasures as measure, index (measure.name)}
      <div
        bind:this={measureNodes[index]}
        style:width="{MEASURE_WIDTH}px"
        style:height="{measureGridHeights[index]}px"
        style:margin-top="{MARGIN_TOP}px"
        class="inline-grid"
      >
        <!-- FIXME: I can't select the big number by the measure id. -->
        <MeasureBigNumber
          {measure}
          withTimeseries={false}
          {instanceId}
          {metricsViewName}
          where={chartWhere}
          ready={chartReady}
          queryClient={ctx.queryClient}
        />
      </div>
    {/each}
  </div>
</div>
