<script lang="ts">
  import Close from "$lib/components/icons/Close.svelte";
  import { clearSelectedLeaderboardValuesAndUpdate } from "$lib/redux-store/explore/explore-apis";
  import { store } from "$lib/redux-store/store-root";
  import { isAnythingSelected } from "$lib/util/isAnythingSelected";
  import { fly } from "svelte/transition";
  export let metricsDefId;
  export let values;

  function clearAllFilters() {
    clearSelectedLeaderboardValuesAndUpdate(store.dispatch, metricsDefId);
  }

  $: hasFilters = isAnythingSelected(values);
</script>

<div class="pt-3 pb-3" style:min-height="50px">
  {#if hasFilters}
    <button
      transition:fly|local={{ duration: 200, y: 5 }}
      on:click={clearAllFilters}
      class="
            grid gap-x-2 items-center font-bold
            bg-red-100
            text-red-900
            p-1
            pl-2 pr-2
            rounded
        "
      style:grid-template-columns="auto max-content"
    >
      clear all filters <Close />
    </button>
  {/if}
</div>
