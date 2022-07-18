<script lang="ts">
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import ModelIcon from "$lib/components/icons/Model.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import Button from "../Button.svelte";

  $: selectedMetricsDef = getMetricsDefReadableById(metricsDefId);

  export let metricsDefId: string;

  let buttonDisabled = true;
  let buttonStatus = "OK";
  $: if ($selectedMetricsDef?.sourceModelId === undefined) {
    buttonDisabled = true;
    buttonStatus = "MISSING_MODEL";
  } else {
    buttonDisabled = false;
    buttonStatus = undefined;
  }
</script>

<Tooltip location="left" alignment="middle" distance={5}>
  <Button
    type="text"
    disabled={buttonDisabled}
    on:click={() => {
      dataModelerService.dispatch("setActiveAsset", [
        EntityType.Model,
        $selectedMetricsDef?.sourceModelId,
      ]);
    }}>Back to Model <ModelIcon size="16px" /></Button
  >
  <TooltipContent slot="tooltip-content">
    <div>
      {#if buttonStatus === "MISSING_MODEL"}
        set a model
      {:else}
        go to the corresponding model
      {/if}
    </div>
  </TooltipContent>
</Tooltip>
