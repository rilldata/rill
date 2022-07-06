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
  <select
    class="pl-1 mb-2"
    value={$metricsExplore?.measureId}
    on:change={(event) => {
      handleMeasureUpdate(event.target.value);
    }}
  >
    <option value="">Select One</option>
    {#each $measures as measure (measure.id)}
      <option value={measure.id}>{measure.expression}</option>
    {/each}
  </select>
{/if}
