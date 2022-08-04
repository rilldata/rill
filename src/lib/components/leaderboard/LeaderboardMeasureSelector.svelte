<script lang="ts">
  import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import { SelectMenu } from "$lib/components/menu";
  import { setMeasureIdAndUpdateLeaderboard } from "$lib/redux-store/explore/explore-apis";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { store } from "$lib/redux-store/store-root";
  import type { Readable } from "svelte/store";
  import { crossfade, fly } from "svelte/transition";
  import Spinner from "../Spinner.svelte";

  export let metricsDefId;

  let measures: Readable<Array<MeasureDefinitionEntity>>;
  $: measures = getMeasuresByMetricsId(metricsDefId);

  let metricsExplorer: Readable<MetricsExplorerEntity>;
  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  function handleMeasureUpdate(measureID) {
    setMeasureIdAndUpdateLeaderboard(store.dispatch, metricsDefId, measureID);
  }

  function formatForSelector(measure) {
    return {
      ...measure,
      key: measure.id,
      main: measure.label?.length ? measure.label : measure.expression,
    };
  }

  let [send, receive] = crossfade({ fallback: fly });

  /** this should be a single element */
  let selections = [];
  // reset selections based on the active leaderboard measure
  let activeLeaderboardMeasure;
  $: activeLeaderboardMeasure =
    $measures?.length &&
    $metricsExplorer?.leaderboardMeasureId &&
    formatForSelector(
      $measures.find(
        (measure) => measure.id === $metricsExplorer?.leaderboardMeasureId
      ) ?? undefined
    );

  /** this controls the animation direction */

  $: options =
    $measures?.map((measure) => {
      let main = measure.label?.length ? measure.label : measure.expression;
      return {
        ...measure,
        key: measure.id,
        main,
      };
    }) || [];
  $: selection = $measures ? activeLeaderboardMeasure : [];
</script>

<div>
  {#if $measures && options.length && selection}
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
        on:select={(event) => {
          const key = event.detail.key;
          handleMeasureUpdate(key);
        }}
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
