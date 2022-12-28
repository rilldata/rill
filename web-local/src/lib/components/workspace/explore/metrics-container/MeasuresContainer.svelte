<script lang="ts">
  import {
    useRuntimeServiceMetricsViewTotals,
    V1MetricsViewTotalsResponse,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { useMetaQuery } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import { EntityStatus } from "@rilldata/web-local/lib/temp/entity";
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../../../../application-state-stores/explorer-stores";
  import MeasureBigNumber from "./MeasureBigNumber.svelte";

  export let metricViewName;

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
      timeStart: metricsExplorer.selectedTimeRange?.start,
      timeEnd: metricsExplorer.selectedTimeRange?.end,
    };

    totalsQuery = useRuntimeServiceMetricsViewTotals(
      instanceId,
      metricViewName,
      totalsQueryParams
    );

    console.log("totals called");
  }

  $: console.log($totalsQuery?.data.data);
</script>

<div class="flex flex-col">
  {#if $metaQuery.data?.measures}
    {#each $metaQuery.data?.measures as measure, index (measure.name)}
      <!-- FIXME: I can't select the big number by the measure id. -->
      {@const bigNum = $totalsQuery?.data.data?.[measure.name]}
      <div class="mt-5">
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
