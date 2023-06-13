<script lang="ts">
  import {
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { createShowHideMeasuresStore } from "@rilldata/web-common/features/dashboards/show-hide-selectors";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import { createQueryServiceMetricsViewTotals } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { MEASURE_CONFIG } from "../config";
  import { useDashboardStore } from "../dashboard-stores";
  import MeasureBigNumber from "./MeasureBigNumber.svelte";

  import SeachableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SeachableFilterButton.svelte";

  export let metricViewName;
  export let exploreContainerWidth;

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

  $: dashboardStore = useDashboardStore(metricViewName);

  $: instanceId = $runtime.instanceId;

  // query the `/meta` endpoint to get the measures and the default time grain
  $: metaQuery = useMetaQuery(instanceId, metricViewName);

  $: selectedMeasureNames = $dashboardStore?.selectedMeasureNames;

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();
  $: metricsContainerHeight = $observedNode?.offsetHeight || 0;

  let measuresWrapper;
  let measuresHeight = [];
  let measureGridHeights = [];

  let containerWidths = MEASURE_CONFIG.bigNumber.widthWithoutChart;

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
      (measureNode) => measureNode?.offsetHeight
    );

    const minInMeasures = Math.min(...measuresHeight);
    measuresHeight = measuresHeight.map((height) =>
      height > minInMeasures ? MEASURE_HEIGHT_MULTILINE : MEASURE_HEIGHT
    );
    const totalMeasuresHeight = measuresHeight.reduce(
      (s, v) => s + v + MARGIN_TOP,
      0
    );

    if (totalMeasuresHeight && metricsContainerHeight) {
      let columns = totalMeasuresHeight / metricsContainerHeight;
      if (columns <= 1 || columns > 2) {
        numColumns = Math.min(Math.ceil(columns), 3);
        measureGridHeights = getMeasureHeightsForColumn(
          measuresHeight,
          numColumns
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
            numColumns
          );
        } else break;
      }
    }
  }

  $: numColumns = 3;

  $: metricTimeSeries = useModelHasTimeSeries(instanceId, metricViewName);
  $: hasTimeSeries = $metricTimeSeries.data;
  $: timeStart = $dashboardStore?.selectedTimeRange?.start?.toISOString();
  $: timeEnd = $dashboardStore?.selectedTimeRange?.end?.toISOString();
  $: totalsQuery = createQueryServiceMetricsViewTotals(
    instanceId,
    metricViewName,
    {
      measureNames: selectedMeasureNames,
      timeStart: timeStart,
      timeEnd: timeEnd,
      filter: $dashboardStore?.filters,
    },
    {
      query: {
        enabled:
          (hasTimeSeries ? !!timeStart && !!timeEnd : true) &&
          !!$dashboardStore?.filters,
      },
    }
  );

  let measureNodes = [];

  $: if (metricsContainerHeight && measureNodes.length) {
    calculateGridColumns();
  }

  $: showHideMeasures = createShowHideMeasuresStore(metricViewName, metaQuery);

  const toggleMeasureVisibility = (e) => {
    showHideMeasures.toggleVisibility(e.detail.name);
  };
  const setAllMeasuresNotVisible = () => {
    showHideMeasures.setAllToNotVisible();
  };
  const setAllMeasuresVisible = () => {
    showHideMeasures.setAllToVisible();
  };
</script>

<svelte:window on:resize={() => calculateGridColumns()} />
<div
  style:height="calc(100% - {GRID_MARGIN_TOP}px)"
  style:width={containerWidths[numColumns]}
  use:listenToNodeResize
>
  <div
    bind:this={measuresWrapper}
    class="grid grid-cols-{numColumns}"
    style:column-gap="{COLUMN_GAP}px"
  >
    <div class="bg-white sticky top-0" style="z-index:100">
      <SeachableFilterButton
        label="Measures"
        on:deselect-all={setAllMeasuresNotVisible}
        on:item-clicked={toggleMeasureVisibility}
        on:select-all={setAllMeasuresVisible}
        selectableItems={$showHideMeasures.selectableItems}
        selectedItems={$showHideMeasures.selectedItems}
        tooltipText="Choose measures to display"
      />
    </div>
    {#if $metaQuery.data?.measures}
      {#each $metaQuery.data?.measures.filter((_, i) => $showHideMeasures.selectedItems[i]) as measure, index (measure.name)}
        <!-- FIXME: I can't select the big number by the measure id. -->
        {@const bigNum = $totalsQuery?.data?.data?.[measure.name]}
        <div
          bind:this={measureNodes[index]}
          style:width="{MEASURE_WIDTH}px"
          style:height="{measureGridHeights[index]}px"
          style:margin-top="{MARGIN_TOP}px"
          class="inline-grid"
        >
          <MeasureBigNumber
            value={bigNum}
            description={measure?.description ||
              measure?.label ||
              measure?.expression}
            formatPreset={measure?.format}
            withTimeseries={false}
            status={$totalsQuery?.isFetching
              ? EntityStatus.Running
              : EntityStatus.Idle}
          >
            <svelte:fragment slot="name">
              {measure?.label || measure?.expression}
            </svelte:fragment>
          </MeasureBigNumber>
        </div>
      {/each}
    {/if}
  </div>
</div>
