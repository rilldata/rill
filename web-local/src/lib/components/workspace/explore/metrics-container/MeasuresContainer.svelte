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

  export let metricViewName;
  export let metricsContainerHeight;

  const MEASURE_HEIGHT = 60;
  const MEASURE_HEIGHT_MULTILINE = 80;
  const MEASURE_WIDTH = 175;
  const MARGIN_TOP = 20;
  const CHARACTER_LIMIT_FOR_WRAPPING = 26;
  const GRID_MARGIN_TOP = 20;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  $: instanceId = $runtimeStore.instanceId;

  // query the `/meta` endpoint to get the measures and the default time grain
  $: metaQuery = useMetaQuery(instanceId, metricViewName);

  $: selectedMeasureNames = metricsExplorer?.selectedMeasureNames;

  function getMeasureHeightsForColumn(measuresHeight, numColumns) {
    const recalculatedHeights = [...measuresHeight];
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

  let totalsQuery: UseQueryStoreResult<V1MetricsViewTotalsResponse, Error>;
  $: numColumns = 1;
  let measureGridHeights = [];

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

    const measures = $metaQuery.data?.measures;

    let measuresHeight = measures.map((measure) => {
      const label = measure?.label || measure?.expression;
      if (label.length > CHARACTER_LIMIT_FOR_WRAPPING)
        return MEASURE_HEIGHT_MULTILINE;
      else return MEASURE_HEIGHT;
    });

    const totalMeasuresHeight = measuresHeight.reduce(
      (s, v) => s + v + MARGIN_TOP,
      0
    );

    if (metricsContainerHeight) {
      const measuresContainerHeight = metricsContainerHeight - GRID_MARGIN_TOP;

      const columns = totalMeasuresHeight / measuresContainerHeight;
      if (columns <= 1) {
        numColumns = 1;
        measureGridHeights = [...measuresHeight];
      } else if (columns > 2) {
        numColumns = 3;
        measureGridHeights = getMeasureHeightsForColumn(measuresHeight, 3);
      } else {
        numColumns = 2;
        measureGridHeights = getMeasureHeightsForColumn(measuresHeight, 2);

        // Check if two columns can individually accommodate all measures without scroll
        const measuresHeightInSingleColumn = measureGridHeights
          .filter((_, i) => i % 2 == 0)
          .reduce((s, v) => s + v + MARGIN_TOP, 0);
        const extraHeight =
          measuresContainerHeight - measuresHeightInSingleColumn;
        if (extraHeight < 0) {
          numColumns = 3;
          measureGridHeights = getMeasureHeightsForColumn(measuresHeight, 3);
        }
      }
    }
  }
</script>

<div
  class="grid grid-cols-{numColumns} gap-x-1"
  style:margin-top="{GRID_MARGIN_TOP}px"
>
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
