<script lang="ts">
  import { store } from "$lib/redux-store/store-root";
  import { fetchManyMeasuresApi } from "$lib/redux-store/measure-definition/measure-definition-apis";
  import { updateLeaderboardMeasure } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-apis";
  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";

  export let metricsDefId;

  $: measures = getMeasuresByMetricsId(metricsDefId);
  $: if (metricsDefId) {
    store.dispatch(fetchManyMeasuresApi({ metricsDefId }));
  }

  function handleMeasureUpdate(measureID) {
    updateLeaderboardMeasure(
      store.dispatch,
      metricsDefId,
      measureID,
      $measures.find((measure) => measure.id === measureID)?.expression
    );
    selectedValue = measureID;
  }

  let selectedValue;
  /** select the first measure available if no value has been selected on initialization. */
  $: if (selectedValue === undefined && $measures?.length) {
    handleMeasureUpdate($measures[0].id);
  }
</script>

{#if $measures}
  <select
    class="pl-1 mb-2"
    value={selectedValue}
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
