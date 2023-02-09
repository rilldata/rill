<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { V1Model } from "@rilldata/web-common/runtime-client";
  import { SelectMenu } from "../../../components/menu";
  import { selectTimestampColumnFromSchema } from "../column-selectors";

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
      default_time_grain: "",
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

<div class="w-80 flex items-center">
  <Tooltip alignment="middle" distance={8} location="bottom">
    <div class="text-gray-500 font-medium" style="width:10em; font-size:11px;">
      Timestamp
    </div>

    <TooltipContent slot="tooltip-content">
      Select a timestamp column to see the time series charts on the dashboard.
    </TooltipContent>
  </Tooltip>
  <div>
    <Tooltip
      alignment="middle"
      distance={16}
      location="right"
      suppress={tooltipText === undefined}
    >
      <SelectMenu
        {options}
        disabled={dropdownDisabled}
        selection={timeColumnSelectedValue}
        tailwindClasses="overflow-hidden"
        alignment="start"
        on:select={(evt) => {
          $metricsInternalRep.updateMetricsParams({
            timeseries: evt.detail?.key,
          });
        }}
      >
        {#if timeColumnSelectedValue === "__DEFAULT_VALUE__"}
          <span class="text-gray-500">Select a timestamp column...</span>
        {:else}
          <span style:max-width="14em" class="font-bold truncate"
            >{timeColumnSelectedValue}</span
          >
        {/if}
      </SelectMenu>

      <TooltipContent slot="tooltip-content">
        {tooltipText}
      </TooltipContent>
    </Tooltip>
  </div>
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
