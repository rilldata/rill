<script lang="ts">
  import { goto } from "$app/navigation";
  import { IconButton } from "@rilldata/web-common/components/button";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { V1Model } from "@rilldata/web-common/runtime-client";
  import { selectTimestampColumnFromSchema } from "@rilldata/web-local/lib/svelte-query/column-selectors";

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

  function updateMetricsDefinitionHandler(evt: Event) {
    $metricsInternalRep.updateMetricKey(
      "timeseries",
      (<HTMLSelectElement>evt.target).value
    );
  }

  function removeTimeseries() {
    $metricsInternalRep.updateMetricKey("timeseries", "");
  }

  function noTimeseriesCTA() {
    if (timestampColumns?.length) {
      $metricsInternalRep.updateMetricKey("timeseries", timestampColumns[0]);
    } else {
      let sourceModelName = $metricsInternalRep.getMetricKey("model");
      goto(`/model/${sourceModelName}`);
    }
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
  <div class="text-gray-500 font-medium" style="width:10em; font-size:11px;">
    Timestamp
  </div>
  <div>
    <Tooltip
      alignment="middle"
      distance={16}
      location="right"
      suppress={tooltipText === undefined}
    >
      <select
        bind:this={selection}
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
      <IconButton
        compact
        marginClasses="ml-1"
        on:click={() => {
          noTimeseriesCTA();
        }}
      >
        <InfoCircle color="gray" size="16px" />
      </IconButton>
      <TooltipContent slot="tooltip-content" maxWidth="300px">
        Select a column to see the time series charts on the dashboard.
      </TooltipContent>
    </Tooltip>
  {/if}
</div>
