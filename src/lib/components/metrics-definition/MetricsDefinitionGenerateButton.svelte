<script lang="ts">
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import { generateMeasuresAndDimensionsApi } from "$lib/redux-store/metrics-definition/metrics-definition-apis";
  import { store } from "$lib/redux-store/store-root";

  export let metricsDefId: string;

  function handleGenerateClick() {
    store.dispatch(generateMeasuresAndDimensionsApi(metricsDefId));
  }
</script>

<Tooltip location="right" alignment="middle" distance={5}>
  <button
    on:click={handleGenerateClick}
    class="
          hover:bg-gray-300
          hover:border-gray-300
          border-black
          transition-tranform
          duration-100
          items-center
          justify-center
          border
          border-transparent
          rounded
          flex flex-row gap-x-2
          pl-4 pr-4
          pt-2 pb-2
        ">Generate</button
  >
  <TooltipContent slot="tooltip-content">
    Infer dimensions and generate count(*) measure
    <br />Warning: It clears all current dimensions and measures
  </TooltipContent>
</Tooltip>
