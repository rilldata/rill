<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    useRuntimeServiceGetTimeRangeSummary,
    V1Model,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { removeIfExists } from "@rilldata/web-local/lib/util/arrayUtils";
  import { SelectMenu } from "../../../components/menu";
  import {
    getSelectableTimeGrains,
    TimeGrainOption,
  } from "../../dashboards/time-controls/time-range-utils";

  export let metricsInternalRep;
  export let selectedModel: V1Model;

  $: selectedTimeRange = $metricsInternalRep.getMetricKey("default_time_range");

  $: availableTimeGrains =
    $metricsInternalRep.getMetricKey("time_grains") || [];

  $: timeColumn = $metricsInternalRep.getMetricKey("timeseries");

  let timeRangeQuery;
  $: if (selectedModel?.name && timeColumn) {
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

  let selectableTimeGrains: TimeGrainOption[] = [];
  $: if (selectedTimeRange) {
    selectableTimeGrains = getSelectableTimeGrains(
      selectedTimeRange,
      allTimeRange
    );
  }

  $: options =
    selectableTimeGrains
      .filter((timeGrain) => timeGrain.enabled)
      .map((grain) => {
        return {
          key: grain.timeGrain,
          main: grain.timeGrain,
        };
      }) || [];

  function handleAvailableTimeGrainsUpdate(event) {
    const selectedTimeGrain = event.detail?.key;

    const isPresent = removeIfExists(
      availableTimeGrains,
      (timeGrain) => timeGrain === selectedTimeGrain
    );

    if (!isPresent) {
      availableTimeGrains.push(selectedTimeGrain);
    }

    $metricsInternalRep.updateMetricsParams({
      time_grains: availableTimeGrains,
    });
  }

  let tooltipText = "";
  let dropdownDisabled = true;
  $: if (selectedModel?.name === undefined) {
    tooltipText = "Select a model before selecting a timestamp column";
    dropdownDisabled = true;
  } else if (!timeColumn) {
    tooltipText = "The selected model has no timestamp columns";
    dropdownDisabled = true;
  } else if (!selectedTimeRange) {
    tooltipText = "Time grains will be inferred from the data";
    dropdownDisabled = true;
  } else {
    tooltipText = undefined;
    dropdownDisabled = false;
  }
</script>

<div class="flex items-center">
  <Tooltip alignment="middle" distance={16} location="bottom">
    <div class="text-gray-500 font-medium" style="width:10em; font-size:11px;">
      Available Time Grains
    </div>

    <TooltipContent slot="tooltip-content">
      Select the timegrains that will be available in the dashboard
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
        multiSelect={true}
        disabled={dropdownDisabled}
        selection={availableTimeGrains}
        tailwindClasses="overflow-hidden"
        alignment="start"
        on:select={handleAvailableTimeGrainsUpdate}
      >
        {#if dropdownDisabled}
          {#if !selectedTimeRange}
            <span>Infered from data</span>
          {:else}
            <span>Select a timestamp</span>
          {/if}
        {:else}
          <span style:max-width="16em" class="font-bold truncate"
            >{availableTimeGrains.length
              ? availableTimeGrains.join(",")
              : "Infer from timerange"}</span
          >
        {/if}
      </SelectMenu>

      <TooltipContent slot="tooltip-content">
        {tooltipText}
      </TooltipContent>
    </Tooltip>
  </div>
</div>
