<script lang="ts">
  import type { V1Model } from "@rilldata/web-common/runtime-client";
  import { selectTimestampColumnFromSchema } from "@rilldata/web-local/lib/svelte-query/column-selectors";
  import TimestampIcon from "../../icons/TimestampType.svelte";
  import Tooltip from "../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../tooltip/TooltipContent.svelte";

  export let metricsInternalRep;
  export let selectedModel: V1Model;

  $: timeColumnSelectedValue =
    $metricsInternalRep.getMetricKey("timeseries") || "__DEFAULT_VALUE__";

  let timestampColumns: Array<string>;
  $: if (selectedModel) {
    timestampColumns = selectTimestampColumnFromSchema(selectedModel?.schema);
  } else {
    timestampColumns = [];
  }

  function updateMetricsDefinitionHandler(evt: Event) {
    $metricsInternalRep.updateMetricKey(
      "timeseries",
      (<HTMLSelectElement>evt.target).value
    );
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
</script>

<div class="flex items-center">
  <div class="flex items-center gap-x-2" style="width:9em">
    <TimestampIcon size="16px" />Timestamp
  </div>
  <div>
    <Tooltip
      alignment="middle"
      distance={16}
      location="right"
      suppress={tooltipText === undefined}
    >
      <select
        class="hover:bg-gray-100 rounded border border-6 border-transparent hover:border-gray-300"
        disabled={dropdownDisabled}
        on:change={updateMetricsDefinitionHandler}
        style="background-color: #FFF; width:18em;"
        value={timeColumnSelectedValue}
      >
        <option disabled hidden selected value="__DEFAULT_VALUE__"
          >Select a timestamp...</option
        >
        {#each timestampColumns as column}
          <option value={column}>{column}</option>
        {/each}
      </select>

      <TooltipContent slot="tooltip-content">
        {tooltipText}
      </TooltipContent>
    </Tooltip>
  </div>
</div>
