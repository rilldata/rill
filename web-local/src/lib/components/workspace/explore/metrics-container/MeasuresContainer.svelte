<script lang="ts">
  import {
    useRuntimeServiceMetricsViewTotals,
    V1MetricsViewTotalsResponse,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { useMetaQuery } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import { EntityStatus } from "@rilldata/web-common/lib/entity";
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../../../../application-state-stores/explorer-stores";
  import MeasureBigNumber from "./MeasureBigNumber.svelte";
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";

  export let metricViewName;

  const MEASURE_HEIGHT = 60;
  const MEASURE_HEIGHT_MULTILINE = 80;
  const MEASURE_WIDTH = 175;
  const MARGIN_TOP = 20;
  const GRID_MARGIN_TOP = 20;

  let measuresWrapper;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  $: instanceId = $runtimeStore.instanceId;

  // query the `/meta` endpoint to get the measures and the default time grain
  $: metaQuery = useMetaQuery(instanceId, metricViewName);

  $: selectedMeasureNames = metricsExplorer?.selectedMeasureNames;

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();
  $: metricsContainerHeight = $observedNode?.offsetHeight || 0;
  $: metricsContainerWidth = $observedNode?.offsetWidth || 0;

  let measuresHeight = [];
  let measureGridHeights = [];

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

    if (metricsContainerHeight) {
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
    if (metricsContainerWidth) {
      while (numColumns > 1) {
        const widthPerColumn = metricsContainerWidth / numColumns;
        // gap-x-4 = 16px
        if (widthPerColumn < MEASURE_WIDTH + 16 / 2) {
          numColumns = numColumns - 1;
          measureGridHeights = getMeasureHeightsForColumn(
            measuresHeight,
            numColumns
          );
        } else break;
      }
    }
  }

  let totalsQuery: UseQueryStoreResult<V1MetricsViewTotalsResponse, Error>;
  $: numColumns = 3;

  $: if (
    metricsExplorer &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching
  ) {
    let totalsQueryParams = {
      measureNames: selectedMeasureNames,
      filter: metricsExplorer?.filters,
    };

    totalsQuery = useRuntimeServiceMetricsViewTotals(
      instanceId,
      metricViewName,
      totalsQueryParams
    );
  }

  $: measureNodes = [...(measuresWrapper?.children || [])];

  $: if (metricsContainerWidth && measureNodes && $metaQuery?.data?.measures) {
    calculateGridColumns();
  }
</script>

<svelte:window on:resize={() => calculateGridColumns()} />
<div
  use:listenToNodeResize
  style:height="calc(100% - {GRID_MARGIN_TOP}px)"
  style:margin-top="{GRID_MARGIN_TOP}px"
>
  <div bind:this={measuresWrapper} class="grid grid-cols-{numColumns} gap-x-4">
    {#if $metaQuery.data?.measures}
      {#each $metaQuery.data?.measures as measure, index (measure.name)}
        <!-- FIXME: I can't select the big number by the measure id. -->
        {@const bigNum = $totalsQuery?.data.data?.[measure.name]}
        <div
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
