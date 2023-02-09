<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    useRuntimeServiceGetTimeRangeSummary,
    V1Model,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { SelectMenu } from "../../../components/menu";
  import {
    getSelectableTimeGrains,
    TimeGrainOption,
  } from "../../dashboards/time-controls/time-range-utils";

  export let metricsInternalRep;
  export let selectedModel: V1Model;

  $: selectedTimeRange = $metricsInternalRep.getMetricKey("default_timerange");

  $: defaultTimeGrainValue =
    $metricsInternalRep.getMetricKey("default_timegrain") ||
    "__DEFAULT_VALUE__";

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
      .map((grain) => {
        return {
          key: grain.timeGrain,
          main: grain.timeGrain,
          disabled: !grain.enabled,
          description: !grain.enabled
            ? "not valid for this time range"
            : undefined,
        };
      })
      .concat({
        key: "__DEFAULT_VALUE__",
        main: "Infer from timerange",
      }) || [];

  function handleDefaultTimeGrainUpdate(event) {
    const selectedTimeGrain = event.detail?.key;

    if (selectedTimeGrain === "") {
      $metricsInternalRep.updateMetricsParams({
        default_timegrain: "",
      });
    } else {
      $metricsInternalRep.updateMetricsParams({
        default_timegrain: selectedTimeGrain,
      });
    }
  }

  let tooltipText = "";
  let dropdownDisabled = true;
  $: if (selectedModel?.name === undefined) {
    tooltipText = "Select a model before selecting a time grain";
    dropdownDisabled = true;
  } else if (!timeColumn) {
    tooltipText = "The selected model has no timestamp columns";
    dropdownDisabled = true;
  } else {
    tooltipText = undefined;
    dropdownDisabled = false;
  }
</script>

<div class="w-80 flex items-center">
  <Tooltip alignment="middle" distance={8} location="bottom">
    <div class="text-gray-500 font-medium" style="width:10em; font-size:11px;">
      Time Grain
    </div>

    <TooltipContent slot="tooltip-content">
      Select a default time grain for the time series charts
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
        selection={defaultTimeGrainValue}
        tailwindClasses="overflow-hidden"
        alignment="start"
        on:select={handleDefaultTimeGrainUpdate}
      >
        {#if dropdownDisabled}
          <span>Select a timestamp</span>
        {:else}
          <span style:max-width="16em" class="font-bold truncate"
            >{defaultTimeGrainValue === "__DEFAULT_VALUE__"
              ? "Infer from timerange"
              : defaultTimeGrainValue}</span
          >
        {/if}
      </SelectMenu>

      <TooltipContent slot="tooltip-content">
        {tooltipText}
      </TooltipContent>
    </Tooltip>
  </div>
</div>
