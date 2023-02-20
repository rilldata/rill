<script lang="ts">
  import { useModelNames } from "@rilldata/web-common/features/models/selectors";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type { Readable } from "svelte/store";
  import Spacer from "../../../components/icons/Spacer.svelte";
  import { SelectMenu } from "../../../components/menu";
  import Tooltip from "../../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../../components/tooltip/TooltipContent.svelte";
  import type { MetricsInternalRepresentation } from "../metrics-internal-store";

  export let metricsInternalRep: Readable<MetricsInternalRepresentation>;

  $: sourceModelDisplayValue =
    $metricsInternalRep.getMetricKey("model") || "__DEFAULT_VALUE__";

  $: allModels = useModelNames($runtimeStore.instanceId);

  function updateMetricsDefinitionHandler(modelName: string) {
    // Reset time selectors as some models might not have a timeseries
    $metricsInternalRep.updateMetricsParams({
      model: modelName,
      timeseries: "",
      smallest_time_grain: "",
      default_time_range: "",
    });
  }

  $: options =
    $allModels?.data?.map((modelName) => {
      return {
        key: modelName,
        main: modelName,
      };
    }) || [];
</script>

<div class="w-80 flex items-center">
  <Tooltip alignment="middle" distance={8} location="bottom">
    <div class="text-gray-500 font-medium" style="width:10em; font-size:11px;">
      Model
    </div>

    <TooltipContent slot="tooltip-content">
      Assign a model for the dashboard
    </TooltipContent>
  </Tooltip>

  <div class="grow">
    <SelectMenu
      block
      {options}
      selection={sourceModelDisplayValue}
      tailwindClasses="overflow-hidden px-2 py-2 rounded"
      alignment="start"
      on:select={(evt) => {
        updateMetricsDefinitionHandler(evt.detail?.key);
      }}
    >
      {#if sourceModelDisplayValue === "__DEFAULT_VALUE__"}
        <span class="text-gray-500">Select a model...</span>
      {:else}
        <span style:max-width="14em" class="font-bold truncate"
          >{sourceModelDisplayValue}</span
        >
      {/if}
    </SelectMenu>
  </div>
  <Spacer size="24px" />
</div>
