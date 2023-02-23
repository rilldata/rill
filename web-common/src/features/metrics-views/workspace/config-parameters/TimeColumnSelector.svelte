<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import { SelectMenu } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { selectTimestampColumnFromSchema } from "@rilldata/web-common/features/metrics-views/column-selectors";
  import type { V1Model } from "@rilldata/web-common/runtime-client";
  import {
    CONFIG_SELECTOR,
    CONFIG_TOP_LEVEL_LABEL_CLASSES,
    INPUT_ELEMENT_CONTAINER,
    SELECTOR_BUTTON_TEXT_CLASSES,
    SELECTOR_CONTAINER,
  } from "../styles";

  export let metricsInternalRep;
  export let selectedModel: V1Model;

  let selection;

  $: timeColumnSelectedValue =
    $metricsInternalRep.getMetricKey("timeseries") || "__DEFAULT_VALUE__";

  let timestampColumns: Array<string>;
  $: if (selectedModel) {
    timestampColumns = selectTimestampColumnFromSchema(selectedModel?.schema);
  } else {
    timestampColumns = [];
  }

  function removeTimeseries() {
    $metricsInternalRep.updateMetricsParams({
      timeseries: "",
      smallest_time_grain: "",
      default_time_range: "",
    });
  }

  let tooltipText = "";
  let dropdownDisabled = true;
  $: if (selectedModel?.name === undefined) {
    tooltipText = "Select a model before selecting a timestamp column";
    dropdownDisabled = true;
  } else if (timestampColumns.length === 0) {
    tooltipText = "The selected model has no timestamp columns";
    dropdownDisabled = true;
  } else {
    tooltipText = undefined;
    dropdownDisabled = false;
  }

  $: options =
    timestampColumns.map((columnName) => {
      return {
        key: columnName,
        main: columnName,
      };
    }) || [];
</script>

<div
  class={INPUT_ELEMENT_CONTAINER.classes}
  style={INPUT_ELEMENT_CONTAINER.style}
>
  <Tooltip alignment="middle" distance={8} location="bottom">
    <div class={CONFIG_TOP_LEVEL_LABEL_CLASSES}>Timestamp</div>

    <TooltipContent maxWidth="400px" slot="tooltip-content">
      Select a timestamp column to see the time series charts on the dashboard.
    </TooltipContent>
  </Tooltip>
  <div class={SELECTOR_CONTAINER.classes} style={SELECTOR_CONTAINER.style}>
    <Tooltip
      alignment="middle"
      distance={16}
      location="right"
      suppress={tooltipText === undefined}
    >
      <SelectMenu
        block
        {options}
        disabled={dropdownDisabled}
        selection={timeColumnSelectedValue}
        tailwindClasses={CONFIG_SELECTOR.base}
        activeTailwindClasses={CONFIG_SELECTOR.active}
        distance={CONFIG_SELECTOR.distance}
        alignment="start"
        on:select={(evt) => {
          $metricsInternalRep.updateMetricsParams({
            timeseries: evt.detail?.key,
          });
        }}
      >
        {#if timeColumnSelectedValue === "__DEFAULT_VALUE__"}
          <span class="text-gray-500">Select a time column</span>
        {:else}
          <span class={SELECTOR_BUTTON_TEXT_CLASSES}
            >{timeColumnSelectedValue}</span
          >
        {/if}
      </SelectMenu>

      <TooltipContent slot="tooltip-content">
        {tooltipText}
      </TooltipContent>
    </Tooltip>

    {#if timeColumnSelectedValue !== "__DEFAULT_VALUE__"}
      <Tooltip location="bottom" distance={8}>
        <IconButton
          compact
          marginClasses="ml-1"
          on:click={() => {
            removeTimeseries();
          }}
        >
          <CancelCircle color="gray" size="16px" />
        </IconButton>
        <TooltipContent slot="tooltip-content" maxWidth="300px">
          Remove the timestamp column to remove the time series charts on the
          dashboard.
        </TooltipContent>
      </Tooltip>
    {:else}
      <Tooltip location="bottom" distance={8}>
        <IconButton compact marginClasses="ml-1" disabled>
          <InfoCircle color="gray" size="16px" />
        </IconButton>
        <TooltipContent slot="tooltip-content" maxWidth="300px">
          Select a column to see the time series charts on the dashboard.
        </TooltipContent>
      </Tooltip>
    {/if}
  </div>
</div>
