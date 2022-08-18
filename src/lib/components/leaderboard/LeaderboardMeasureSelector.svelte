<script lang="ts">
  import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import type { MetricViewMetaResponse } from "$common/rill-developer-service/MetricViewActions";
  import { SelectMenu } from "$lib/components/menu";
  import { setMeasureIdAndUpdateLeaderboard } from "$lib/redux-store/explore/explore-apis";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import { store } from "$lib/redux-store/store-root";
  import {
    getMetricViewMetadata,
    getMetricViewMetaQueryKey,
  } from "$lib/svelte-query/queries/metric-view";
  import { useQuery } from "@sveltestack/svelte-query";
  import type { Readable } from "svelte/store";
  import { crossfade, fly } from "svelte/transition";
  import Spinner from "../Spinner.svelte";

  export let metricsDefId;

  // query the `/meta` endpoint to get the valid measures
  let queryKey = getMetricViewMetaQueryKey(metricsDefId);
  const queryResult = useQuery<MetricViewMetaResponse, Error>(queryKey, () =>
    getMetricViewMetadata(metricsDefId)
  );
  $: {
    queryKey = getMetricViewMetaQueryKey(metricsDefId);
    queryResult.setOptions(queryKey, () => getMetricViewMetadata(metricsDefId));
  }
  let measures: MeasureDefinitionEntity[];
  $: measures = $queryResult.data.measures;

  let metricsExplorer: Readable<MetricsExplorerEntity>;
  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  function handleMeasureUpdate(event: CustomEvent) {
    setMeasureIdAndUpdateLeaderboard(
      store.dispatch,
      metricsDefId,
      event.detail.key
    );
  }

  function formatForSelector(measure: MeasureDefinitionEntity) {
    return {
      ...measure,
      key: measure.id,
      main: measure.label?.length ? measure.label : measure.expression,
    };
  }

  let [send, receive] = crossfade({ fallback: fly });

  /** this should be a single element */
  // reset selections based on the active leaderboard measure
  let activeLeaderboardMeasure;
  $: activeLeaderboardMeasure =
    measures?.length &&
    $metricsExplorer?.leaderboardMeasureId &&
    formatForSelector(
      measures.find(
        (measure) => measure.id === $metricsExplorer?.leaderboardMeasureId
      ) ?? undefined
    );

  /** this controls the animation direction */

  $: options =
    measures?.map((measure) => {
      let main = measure.label?.length ? measure.label : measure.expression;
      return {
        ...measure,
        key: measure.id,
        main,
      };
    }) || [];

  /** set the selection only if $measures is not undefined */
  $: selection = measures ? activeLeaderboardMeasure : [];
</script>

<div>
  {#if measures && options.length && selection}
    <div
      class="flex flex-row items-center"
      style:grid-column-gap=".4rem"
      in:send={{ key: "leaderboard-metric", y: 8 }}
    >
      <div>Dimension Leaders by</div>

      <SelectMenu
        {options}
        {selection}
        alignment="end"
        on:select={handleMeasureUpdate}
      >
        <span class="font-bold">{selection?.main}</span>
      </SelectMenu>
    </div>
  {:else}
    <div
      class="flex flex-row items-center"
      style:grid-column-gap=".4rem"
      in:receive={{ key: "loading-leaderboard-metric", y: 8 }}
    >
      pulling leaderboards <Spinner status={EntityStatus.Running} />
    </div>
  {/if}
</div>
