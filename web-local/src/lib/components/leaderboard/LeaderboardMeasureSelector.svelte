<script lang="ts">
  import { SelectMenu } from "@rilldata/web-common/components/menu";
  import type { MetricsViewMeasure } from "@rilldata/web-common/runtime-client";
  import { crossfade, fly } from "svelte/transition";
  import { runtimeStore } from "../../application-state-stores/application-store";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../../application-state-stores/explorer-stores";
  import { useMetaQuery } from "../../svelte-query/dashboards";
  import { EntityStatus } from "../../temp/entity";
  import Spinner from "../Spinner.svelte";

  export let metricViewName;

  $: metaQuery = useMetaQuery($runtimeStore.instanceId, metricViewName);

  $: measures = $metaQuery.data?.measures;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  function handleMeasureUpdate(event: CustomEvent) {
    metricsExplorerStore.setLeaderboardMeasureName(
      metricViewName,
      event.detail.key
    );
  }

  function formatForSelector(measure: MetricsViewMeasure) {
    if (!measure) return undefined;
    return {
      ...measure,
      key: measure.name,
      main: measure.label?.length ? measure.label : measure.expression,
    };
  }

  let [send, receive] = crossfade({ fallback: fly });

  /** this should be a single element */
  // reset selections based on the active leaderboard measure
  let activeLeaderboardMeasure;
  $: activeLeaderboardMeasure =
    measures?.length &&
    metricsExplorer?.leaderboardMeasureName &&
    formatForSelector(
      measures.find(
        (measure) => measure.name === metricsExplorer?.leaderboardMeasureName
      ) ?? undefined
    );

  /** this controls the animation direction */

  $: options =
    measures?.map((measure) => {
      let main = measure.label?.length ? measure.label : measure.expression;
      return {
        ...measure,
        key: measure.name,
        main,
      };
    }) || [];

  /** set the selection only if measures is not undefined */
  $: selection = measures ? activeLeaderboardMeasure : [];
</script>

<div>
  {#if measures && options.length && selection}
    <div
      class="flex flex-row items-center ui-copy-muted"
      style:grid-column-gap=".4rem"
      in:send={{ key: "leaderboard-metric" }}
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
      in:receive={{ key: "loading-leaderboard-metric" }}
    >
      pulling leaderboards <Spinner status={EntityStatus.Running} />
    </div>
  {/if}
</div>
