<script lang="ts">
  import { reduxReadable, store } from "$lib/redux-store/store-root";
  import { fetchManyMeasuresApi } from "$lib/redux-store/measure-definition/measure-definition-apis";
  import { selectMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-selectors";
  import { updateLeaderboardMeasure } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-apis";

  export let metricsDefId;

  $: measures = selectMeasuresByMetricsId(metricsDefId)($reduxReadable);
  $: if (metricsDefId) {
    store.dispatch(fetchManyMeasuresApi({ metricsDefId }));
  }

  function handleMeasureUpdate(event) {
    updateLeaderboardMeasure(
      store.dispatch,
      metricsDefId,
      event.target.value,
      measures.find((measure) => measure.id === event.target.value)?.expression
    );
  }
</script>

{#if measures}
  <select class="pl-1 mb-2" on:change={handleMeasureUpdate}>
    <option value="">Select One</option>
    {#each measures as measure}
      <option value={measure.id}>{measure.expression}</option>
    {/each}
  </select>
{/if}
