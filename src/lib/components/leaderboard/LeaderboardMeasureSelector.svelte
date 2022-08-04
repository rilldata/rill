<script lang="ts">
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import { setMeasureIdAndUpdateLeaderboard } from "$lib/redux-store/explore/explore-apis";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { store } from "$lib/redux-store/store-root";
  import type { Readable } from "svelte/store";

  export let metricsDefId;

  let measures: Readable<Array<MeasureDefinitionEntity>>;
  $: measures = getMeasuresByMetricsId(metricsDefId);

  let metricsExplorer: Readable<MetricsExplorerEntity>;
  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  function handleMeasureUpdate(measureID) {
    setMeasureIdAndUpdateLeaderboard(store.dispatch, metricsDefId, measureID);
  }
</script>

{#if $measures}
  Dimension Leaders by
  <select
    class="pl-1 font-bold"
    value={$metricsExplorer?.leaderboardMeasureId}
    on:change={(event) => {
      handleMeasureUpdate(event.target.value);
    }}
  >
    {#each $measures as measure (measure.id)}
      <option value={measure.id}
        >{measure.label?.length ? measure.label : measure.expression}</option
      >
    {/each}
  </select>
{/if}
