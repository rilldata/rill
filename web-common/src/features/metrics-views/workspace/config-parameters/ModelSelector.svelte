<script lang="ts">
  import { SelectMenu } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { useModelNames } from "@rilldata/web-common/features/models/selectors";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { getContext } from "svelte";
  import type { Readable, Writable } from "svelte/store";
  import type { MetricsInternalRepresentation } from "../../metrics-internal-store";
  import {
    CONFIG_SELECTOR,
    CONFIG_TOP_LEVEL_LABEL_CLASSES,
    INPUT_ELEMENT_CONTAINER,
    SELECTOR_CONTAINER,
  } from "../styles";
  import FormattedSelectorText from "./FormattedSelectorText.svelte";

  export let metricsInternalRep: Readable<MetricsInternalRepresentation>;

  let metricsConfigErrorStore = getContext(
    "rill:metrics-config:errors"
  ) as Writable<any>;

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

  let level: "error" | undefined = undefined;

  const TOOLTIP_TEXT = "Assign a model for the dashboard";
  let tooltipStateText = TOOLTIP_TEXT;
  // handle case where model name is not in the list of models
  $: if (sourceModelDisplayValue !== "__DEFAULT_VALUE__") {
    if (!$allModels?.data?.includes(sourceModelDisplayValue)) {
      level = "error";
      tooltipStateText = "Model not found";
    } else {
      level = undefined;
      tooltipStateText = TOOLTIP_TEXT;
    }
  } else {
    level = undefined;
    tooltipStateText = TOOLTIP_TEXT;
  }

  $: metricsConfigErrorStore.update((errors) => {
    errors.model = level === "error" ? tooltipStateText : null;
    return errors;
  });

  $: options = [
    ...(level === "error"
      ? [
          {
            key: sourceModelDisplayValue,
            main: sourceModelDisplayValue,
            description: tooltipStateText,
            divider: true,
          },
        ]
      : []),
    ...($allModels?.data?.map((modelName) => {
      return {
        key: modelName,
        main: modelName,
      };
    }) || []),
  ];
</script>

<div
  class={INPUT_ELEMENT_CONTAINER.classes}
  style={INPUT_ELEMENT_CONTAINER.style}
>
  <Tooltip alignment="start" distance={16} location="bottom">
    <div class={CONFIG_TOP_LEVEL_LABEL_CLASSES}>Model</div>

    <TooltipContent slot="tooltip-content">
      {TOOLTIP_TEXT}
    </TooltipContent>
  </Tooltip>

  <Tooltip alignment="start" distance={8}>
    <div class={SELECTOR_CONTAINER.classes} style={SELECTOR_CONTAINER.style}>
      <SelectMenu
        block
        paddingTop={1}
        paddingBottom={1}
        {options}
        selection={sourceModelDisplayValue}
        tailwindClasses="{CONFIG_SELECTOR.base} {level === 'error'
          ? CONFIG_SELECTOR.error
          : CONFIG_SELECTOR.info}"
        activeTailwindClasses={level === "error"
          ? CONFIG_SELECTOR.activeError
          : CONFIG_SELECTOR.active}
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
    <TooltipContent slot="tooltip-content">{tooltipStateText}</TooltipContent>
  </Tooltip>
</div>
