<script lang="ts">
  import {
    fetchManyMeasuresApi,
    manyMeasuresSelector,
  } from "$lib/redux-store/measure-definition-slice";
  import { reduxReadable, store } from "$lib/redux-store/store-root";
  import { updateLeaderboardMeasure } from "$lib/redux-store/metrics-leaderboard-slice.js";

  export let metricsDefId;

  $: measures = manyMeasuresSelector(metricsDefId)($reduxReadable);
  $: if (metricsDefId) {
    store.dispatch(fetchManyMeasuresApi({ metricsDefId }));
  }
</script>

{#if measures}
  <select
    class="pl-1 mb-2"
    on:change={(event) => {
      updateLeaderboardMeasure(
        store.dispatch,
        metricsDefId,
        event.target.value
      );
    }}
  >
    <option value="">Select One</option>
    {#each measures as measure}
      <option value={measure.id}>{measure.expression}</option>
    {/each}
  </select>
{/if}
