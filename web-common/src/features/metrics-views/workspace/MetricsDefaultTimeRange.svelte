<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    useRuntimeServiceGetTimeRangeSummary,
    V1Model,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type { TimeSeriesTimeRange } from "../../dashboards/time-controls/time-control-types";
  import {
    getSelectableTimeRangeNames,
    makeTimeRanges,
  } from "../../dashboards/time-controls/time-range-utils";

  export let metricsInternalRep;
  export let selectedModel: V1Model;

  let selection;

  $: timeRangeSelectedValue =
    $metricsInternalRep.getMetricKey("default_timerange") ||
    "__DEFAULT_VALUE__";

  $: timeColumn = $metricsInternalRep.getMetricKey("timeseries");

  let timeRangeQuery;
  $: if (selectedModel.name && timeColumn) {
    timeRangeQuery = useRuntimeServiceGetTimeRangeSummary(
      $runtimeStore.instanceId,
      selectedModel.name,
      { columnName: timeColumn }
    );
  }

  let allTimeRange;
  $: if (
    timeRangeQuery &&
    $timeRangeQuery.isSuccess &&
    !$timeRangeQuery.isRefetching
  ) {
    allTimeRange = {
      start: $timeRangeQuery.data.timeRangeSummary.min,
      end: $timeRangeQuery.data.timeRangeSummary.max,
    };
  }

  const getSelectableTimeRanges = (
    allTimeRangeInDataset: TimeSeriesTimeRange
  ) => {
    const selectableTimeRangeNames = getSelectableTimeRangeNames(
      allTimeRangeInDataset
    );
    const selectableTimeRanges = makeTimeRanges(
      selectableTimeRangeNames,
      allTimeRangeInDataset
    );
    return selectableTimeRanges;
  };

  let selectableTimeRanges = [];
  $: if (allTimeRange) {
    selectableTimeRanges = getSelectableTimeRanges(allTimeRange);
  }

  function updateMetricsDefinitionHandler(evt: Event) {
    $metricsInternalRep.updateMetricKey(
      "default_timerange",
      (<HTMLSelectElement>evt.target).value
    );
  }

  let tooltipText = "";
  let dropdownDisabled = true;
  $: if (selectedModel?.name === undefined) {
    tooltipText = "Select a model before selecting a time range";
    dropdownDisabled = true;
  } else if (!timeColumn) {
    tooltipText = "The selected model has no timestamp columns";
    dropdownDisabled = true;
  } else {
    tooltipText = undefined;
    dropdownDisabled = false;
  }
</script>

<div class="flex items-center mb-3">
  <div class="text-gray-500 font-medium" style="width:10em; font-size:11px;">
    Time Range
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
        value={timeRangeSelectedValue}
      >
        <option disabled hidden selected value="__DEFAULT_VALUE__"
          >Select a default time range</option
        >
        {#each selectableTimeRanges as timeRange}
          <option value={timeRange.name}>{timeRange.name}</option>
        {/each}
      </select>

      <TooltipContent slot="tooltip-content">
        {tooltipText}
      </TooltipContent>
    </Tooltip>
  </div>
</div>
