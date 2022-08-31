<script lang="ts">
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import { Button } from "$lib/components/button";
  import ExploreIcon from "$lib/components/icons/Explore.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import { useMetaQuery } from "$lib/svelte-query/queries/metric-view";

  export let metricsDefId: string;

  // query the `/meta` endpoint to get the valid measures and dimensions
  $: metaQuery = useMetaQuery(metricsDefId);
  $: measures = $metaQuery.data?.measures;
  $: dimensions = $metaQuery.data?.dimensions;

  $: selectedMetricsDef = getMetricsDefReadableById(metricsDefId);

  let buttonDisabled = true;
  let buttonStatus = "OK";
  $: if (
    $selectedMetricsDef?.sourceModelId === undefined ||
    $selectedMetricsDef?.timeDimension === undefined
  ) {
    buttonDisabled = true;
    buttonStatus = "MISSING_MODEL_OR_TIMESTAMP";
  } else if (measures?.length === 0 || dimensions?.length === 0) {
    buttonDisabled = true;
    buttonStatus = "MISSING_MEASURES_OR_DIMENSIONS";
  } else {
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
        EntityType.MetricsExplorer,
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
