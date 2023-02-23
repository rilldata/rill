<script lang="ts">
  import { SelectMenu } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { useModelNames } from "@rilldata/web-common/features/models/selectors";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type { Readable } from "svelte/store";
  import type { MetricsInternalRepresentation } from "../../metrics-internal-store";
  import {
    CONFIG_SELECTOR,
    CONFIG_TOP_LEVEL_LABEL_CLASSES,
    INPUT_ELEMENT_CONTAINER,
    SELECTOR_CONTAINER,
  } from "../styles";
  import FormattedSelectorText from "./FormattedSelectorText.svelte";

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

<div
  class={INPUT_ELEMENT_CONTAINER.classes}
  style={INPUT_ELEMENT_CONTAINER.style}
>
  <Tooltip alignment="middle" distance={8} location="bottom">
    <div class={CONFIG_TOP_LEVEL_LABEL_CLASSES}>Model</div>

    <TooltipContent slot="tooltip-content">
      Assign a model for the dashboard
    </TooltipContent>
  </Tooltip>

  <div class={SELECTOR_CONTAINER.classes} style={SELECTOR_CONTAINER.style}>
    <SelectMenu
      block
      paddingTop={1}
      paddingBottom={1}
      {options}
      selection={sourceModelDisplayValue}
      tailwindClasses="{CONFIG_SELECTOR.base} {CONFIG_SELECTOR.info}"
      activeTailwindClasses={CONFIG_SELECTOR.active}
      distance={CONFIG_SELECTOR.distance}
      alignment="start"
      on:select={(evt) => {
        updateMetricsDefinitionHandler(evt.detail?.key);
      }}
    >
      <FormattedSelectorText
        value={sourceModelDisplayValue === "__DEFAULT_VALUE__"
          ? "Select a model"
          : sourceModelDisplayValue}
        selected={sourceModelDisplayValue !== "__DEFAULT_VALUE__"}
      />
    </SelectMenu>
  </div>
</div>
