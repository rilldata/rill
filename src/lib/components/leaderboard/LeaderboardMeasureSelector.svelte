<script lang="ts">
  import { store } from "$lib/redux-store/store-root";
  import { fetchManyMeasuresApi } from "$lib/redux-store/measure-definition/measure-definition-apis";
  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import type { Readable } from "svelte/store";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import { setMeasureIdAndUpdateLeaderboard } from "$lib/redux-store/explore/explore-apis";
  import type { MetricsExploreEntity } from "$lib/redux-store/explore/explore-slice";
  import { getMetricsExploreById } from "$lib/redux-store/explore/explore-readables";

  export let metricsDefId;

  let measures: Readable<Array<MeasureDefinitionEntity>>;
  $: measures = getMeasuresByMetricsId(metricsDefId);
  $: if (metricsDefId) {
    store.dispatch(fetchManyMeasuresApi({ metricsDefId }));
  }

  let metricsExplore: Readable<MetricsExploreEntity>;
  $: metricsExplore = getMetricsExploreById(metricsDefId);

  function handleMeasureUpdate(measureID) {
    setMeasureIdAndUpdateLeaderboard(store.dispatch, metricsDefId, measureID);
  }
</script>

{#if $measures}
  Dimension Leaders by
  <select
    class="pl-1 font-bold"
    value={$metricsExplore?.leaderboardMeasureId}
    on:change={(event) => {
      handleMeasureUpdate(event.target.value);
    }}
  >
    {#each $measures as measure (measure.id)}
      <option value={measure.id}
        >{measure.label.length ? measure.label : measure.expression}</option
      >
    {/each}
  </select>
{/if}
