<script lang="ts">
  import { fly } from "svelte/transition";
  import BarAndLabel from "$lib/components/viz/BarAndLabel.svelte";
  import CheckerFull from "$lib/components/icons/CheckerFull.svelte";
  import CheckerHalf from "$lib/components/icons/CheckerHalf.svelte";
  import Close from "$lib/components/icons/Close.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import {
    formatBigNumberPercentage,
    formatInteger,
  } from "$lib/util/formatters";
  import { cubicIn } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import { store } from "$lib/redux-store/store-root";
  import { MetricsLeaderboardEntity } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-slice";
  import { isAnythingSelected } from "$lib/util/isAnythingSelected";
  import LeaderboardMeasureSelector from "$lib/components/leaderboard/LeaderboardMeasureSelector.svelte";
  import { clearLeaderboardAndUpdate } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-apis";
  import { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import { getMeasureById } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import type { Readable } from "svelte/store";
  import { getMetricsLeaderboardById } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-readables";

  export let metricsDefId: string;
  export let whichReferenceValue = "global";

  let metricsLeaderboard: Readable<MetricsLeaderboardEntity>;
  $: metricsLeaderboard = getMetricsLeaderboardById(metricsDefId);

  let measure: Readable<MeasureDefinitionEntity>;
  $: if ($metricsLeaderboard?.measureId) {
    measure = getMeasureById($metricsLeaderboard?.measureId);
  }

  const metricFormatters = {
    simpleSummable: formatInteger,
  };
  let bigNumber;
  const bigNumberTween = tweened(0, {
    duration: 1000,
    delay: 200,
    easing: cubicIn,
  });
  $: bigNumber = $metricsLeaderboard?.bigNumber || 0;
  $: bigNumberTween.set(bigNumber);
  $: anythingSelected = isAnythingSelected($metricsLeaderboard?.activeValues);
  function clearAllFilters() {
    clearLeaderboardAndUpdate(
      store.dispatch,
      metricsDefId,
      $measure.expression
    );
  }
</script>

<header
  style:grid-template-columns="auto max-content"
  class="pb-6 pt-6 grid w-full bg-white"
>
  <div>
    <LeaderboardMeasureSelector {metricsDefId} />
    <h1 style:line-height="1.1">
      <div class="pl-2 text-gray-600 font-normal" style:font-size="1.5rem">
        Total Records
      </div>
      <div style:font-size="2rem" style:width="400px">
        <div class="w-full rounded">
          <BarAndLabel
            justify="stretch"
            showBackground={anythingSelected}
            color={!anythingSelected ? "bg-transparent" : "bg-blue-200"}
            value={$metricsLeaderboard?.bigNumber /
              $metricsLeaderboard?.referenceValue || 0}
          >
            <div
              style:grid-template-columns="auto auto"
              class="grid items-center gap-x-2 w-full text-left pb-2 pt-2"
            >
              <div>
                {metricFormatters.simpleSummable(~~$bigNumberTween)}
              </div>
              <div class="font-normal text-gray-600 italic text-right">
                {#if $bigNumberTween && $metricsLeaderboard?.referenceValue}
                  {formatBigNumberPercentage(
                    $bigNumberTween / $metricsLeaderboard?.referenceValue
                  )}
                {/if}
              </div>
            </div>
          </BarAndLabel>
        </div>
      </div>
    </h1>
  </div>
  <div class="justify-self-end">
    <div
      style:font-size="24px"
      class="grid justify-items-end justify-end grid-flow-col items-center"
    >
      <Tooltip distance={16}>
        <button
          class="m-0 p-1 transition-color"
          class:bg-transparent={whichReferenceValue !== "filtered"}
          class:bg-gray-200={whichReferenceValue === "filtered"}
          class:font-bold={whichReferenceValue === "filtered"}
          class:text-gray-400={whichReferenceValue !== "filtered"}
          on:click={() => (whichReferenceValue = "filtered")}
          ><CheckerHalf /></button
        >
        <TooltipContent slot="tooltip-content">
          scale leaderboard bars by currently-filtered total
        </TooltipContent>
      </Tooltip>
      <Tooltip distance={16}>
        <button
          class="m-0 p-1 transition-color"
          class:bg-transparent={whichReferenceValue !== "global"}
          class:bg-gray-200={whichReferenceValue === "global"}
          class:font-bold={whichReferenceValue === "global"}
          class:text-gray-400={whichReferenceValue !== "global"}
          on:click={() => (whichReferenceValue = "global")}
          ><CheckerFull /></button
        >
        <TooltipContent slot="tooltip-content">
          scale leaderboard bars by total record count
        </TooltipContent>
      </Tooltip>
    </div>
    <div class="pt-3">
      {#if anythingSelected}
        <button
          transition:fly={{ duration: 200, y: 5 }}
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
  </div>
</header>
