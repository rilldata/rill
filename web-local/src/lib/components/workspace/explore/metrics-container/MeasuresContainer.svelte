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
  const MARGIN_TOP = 15;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  $: instanceId = $runtimeStore.instanceId;

  // query the `/meta` endpoint to get the measures and the default time grain
  $: metaQuery = useMetaQuery(instanceId, metricViewName);

  $: selectedMeasureNames = metricsExplorer?.selectedMeasureNames;

  let totalsQuery: UseQueryStoreResult<V1MetricsViewTotalsResponse, Error>;

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

  $: numColumns = 1;

  $: if (metricsContainerHeight) {
    const columns =
      ($metaQuery.data?.measures?.length * (MEASURE_HEIGHT + MARGIN_TOP)) /
      metricsContainerHeight;

    numColumns = columns > 3 ? 3 : Math.ceil(columns);

    // Check if two columns can individually accomodate all measures without scrollbar
    if (numColumns == 2) {
      const maxMeasuresInColumn = Math.ceil(
        $metaQuery.data?.measures?.length / 2
      );

      const extraHeight =
        metricsContainerHeight -
        maxMeasuresInColumn * (MEASURE_HEIGHT + MARGIN_TOP);
      if (extraHeight < 0) numColumns = 3;
    }
  }
</script>

<div class="grid grid-cols-{numColumns} gap-2">
  {#if $metaQuery.data?.measures}
    {#each $metaQuery.data?.measures as measure, index (measure.name)}
      <!-- FIXME: I can't select the big number by the measure id. -->
      {@const bigNum = $totalsQuery?.data.data?.[measure.name]}
      <div
        style:min-width="170px"
        style:max-width="200px"
        style:height="{MEASURE_HEIGHT}px"
        style:margin-top="{MARGIN_TOP}px"
        class="inline-grid"
      >
        <MeasureBigNumber
          value={bigNum}
          description={measure?.description ||
            measure?.label ||
            measure?.expression}
          formatPreset={measure?.format}
          compact={false}
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
