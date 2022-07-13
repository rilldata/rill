<script lang="ts">
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import { generateMeasuresAndDimensionsApi } from "$lib/redux-store/metrics-definition/metrics-definition-apis";
  import { store } from "$lib/redux-store/store-root";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";

  $: selectedMetricsDef = getMetricsDefReadableById(metricsDefId);

  export let metricsDefId: string;

  function handleGenerateClick() {
    store.dispatch(generateMeasuresAndDimensionsApi(metricsDefId));
  }

  let tooltipText = "";
  let buttonDisabled = true;
  $: if (
    $selectedMetricsDef?.sourceModelId === undefined ||
    $selectedMetricsDef?.timeDimension === undefined
  ) {
    tooltipText = "";
    buttonDisabled = true;
  } else {
    tooltipText = undefined;
    buttonDisabled = false;
  }
</script>

<Tooltip location="right" alignment="middle" distance={5}>
  <button
    disabled={buttonDisabled}
    on:click={handleGenerateClick}
    class={`bg-white
          border-gray-400
          hover:border-gray-900
          transition-tranform
          duration-100
          items-center
          justify-center
          border
          rounded
          flex flex-row gap-x-2
          pl-4 pr-4
          pt-2 pb-2
          ${buttonDisabled ? "cursor-not-allowed" : "cursor-pointer"}
          ${buttonDisabled ? "text-gray-500" : "text-gray-900"}
        `}>quick start</button
  >
  <TooltipContent slot="tooltip-content">
    <div style:width="30em">
      {#if buttonDisabled}
        select a model and a timestamp column before populating this metrics
        cube
      {:else}
        add initial measure <em>events per time period</em>, and add all
        categorical columns as slicing dimensions.
        <br /> <strong>warning:</strong> replaces current measures and dimensions
      {/if}
    </div>
  </TooltipContent>
</Tooltip>
