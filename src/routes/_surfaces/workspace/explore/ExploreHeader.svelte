<script lang="ts">
  import { fly } from "svelte/transition";
  import CheckerFull from "$lib/components/icons/CheckerFull.svelte";
  import CheckerHalf from "$lib/components/icons/CheckerHalf.svelte";
  import Close from "$lib/components/icons/Close.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import { store } from "$lib/redux-store/store-root";
  import type { MetricsExploreEntity } from "$lib/redux-store/explore/explore-slice";
  import { isAnythingSelected } from "$lib/util/isAnythingSelected";
  import { clearSelectedLeaderboardValuesAndUpdate } from "$lib/redux-store/explore/explore-apis";
  import type { Readable } from "svelte/store";
  import { getMetricsExploreById } from "$lib/redux-store/explore/explore-readables";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";

  export let metricsDefId: string;
  export let whichReferenceValue = "global";

  let metricsLeaderboard: Readable<MetricsExploreEntity>;
  $: metricsLeaderboard = getMetricsExploreById(metricsDefId);

  $: anythingSelected = isAnythingSelected($metricsLeaderboard?.activeValues);
  function clearAllFilters() {
    clearSelectedLeaderboardValuesAndUpdate(store.dispatch, metricsDefId);
  }
  $: metricsDefinition = getMetricsDefReadableById(metricsDefId);
</script>

<header
  style:grid-template-columns="auto max-content"
  class="grid w-full bg-white self-stretch"
>
  <div>
    <h1 style:line-height="1.1">
      <div class="pl-2 text-gray-600" style:font-size="24px">
        {#if $metricsDefinition}
          {$metricsDefinition?.metricDefLabel}
        {/if}
      </div>
    </h1>
  </div>
  <div class="justify-self-end">
    <!-- NOTE: place share buttons here -->
  </div>
</header>
