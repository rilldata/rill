<script lang="ts">
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import ExploreIcon from "$lib/components/icons/Explore.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import { getDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-readables";
  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import Button from "../Button.svelte";

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
  <!-- TODO: we need to standardize these buttons. -->
  <Button
    type="primary"
    disabled={buttonDisabled}
    on:click={() => {
      dataModelerService.dispatch("setActiveAsset", [
        EntityType.MetricsLeaderboard,
        metricsDefId,
      ]);
    }}>Go to Dashboard <ExploreIcon size="16px" /></Button
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
