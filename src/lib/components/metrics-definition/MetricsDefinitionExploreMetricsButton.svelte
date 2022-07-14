<script lang="ts">
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { getDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-readables";
  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import ExploreIcon from "$lib/components/icons/Explore.svelte";

  $: selectedMetricsDef = getMetricsDefReadableById(metricsDefId);
  $: measures = getMeasuresByMetricsId(metricsDefId);
  $: dimensions = getDimensionsByMetricsId(metricsDefId);

  export let metricsDefId: string;

  let tooltipText = "";
  let buttonDisabled = true;
  let buttonStatus = "OK";
  $: if (
    $selectedMetricsDef?.sourceModelId === undefined ||
    $selectedMetricsDef?.timeDimension === undefined
  ) {
    tooltipText = "";
    buttonDisabled = true;
    buttonStatus = "MISSING_MODEL_OR_TIMESTAMP";
  } else if ($measures.length === 0 || $dimensions.length === 0) {
    tooltipText = "";
    buttonDisabled = true;
    buttonStatus = "MISSING_MEASURES_OR_DIMENSIONS";
  } else {
    tooltipText = undefined;
    buttonDisabled = false;
  }
</script>

<Tooltip location="right" alignment="middle" distance={5}>
  <button
    disabled={buttonDisabled}
    on:click={() => {
      dataModelerService.dispatch("setActiveAsset", [
        EntityType.MetricsLeaderboard,
        metricsDefId,
      ]);
    }}
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
        `}>explore metrics <ExploreIcon /></button
  >
  <TooltipContent slot="tooltip-content">
    <div>
      {#if buttonStatus === "MISSING_MODEL_OR_TIMESTAMP"}
        select a model and a timestamp column before exploring metrics
      {:else if buttonStatus === "MISSING_MEASURES_OR_DIMENSIONS"}
        add measures and dimensions before exploring metrics
      {:else}
        explore your metrics dashboard
      {/if}
    </div>
  </TooltipContent>
</Tooltip>
