<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    useRuntimeServiceGetTimeRangeSummary,
    V1Model,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import Spacer from "../../../components/icons/Spacer.svelte";
  import { SelectMenu } from "../../../components/menu";
  import {
    getSelectableTimeGrains,
    prettyTimeGrain,
    TimeGrainOption,
  } from "../../dashboards/time-controls/time-range-utils";

  export let metricsInternalRep;
  export let selectedModel: V1Model;

  $: selectedTimeRange = $metricsInternalRep.getMetricKey("default_time_range");

  $: defaultTimeGrainValue =
    $metricsInternalRep.getMetricKey("default_time_grain") ||
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

  $: options = [
    {
      key: "__DEFAULT_VALUE__",
      main: "Infer from timerange",
      divider: true,
    },
  ].concat(
    selectableTimeGrains.map((grain) => {
      return {
        divider: false,
        key: grain.timeGrain,
        main: prettyTimeGrain(grain.timeGrain),
        disabled: !grain.enabled,
        description: !grain.enabled
          ? "not valid for this time range"
          : undefined,
      };
    }) as any[]
  );

  function handleDefaultTimeGrainUpdate(event) {
    const selectedTimeGrain = event.detail?.key;

    if (selectedTimeGrain === "") {
      $metricsInternalRep.updateMetricsParams({
        default_time_grain: "",
      });
    } else {
      $metricsInternalRep.updateMetricsParams({
        default_time_grain: selectedTimeGrain,
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
  } else if (!selectedTimeRange) {
    tooltipText = "Default time grain will be inferred from the data";
    dropdownDisabled = true;
  } else {
    tooltipText = undefined;
    dropdownDisabled = false;
  }
</script>

<div class="w-80 flex items-center">
  <Tooltip alignment="middle" distance={8} location="bottom">
    <div class="text-gray-500 font-medium" style="width:10em; font-size:11px;">
      Default Time Grain
    </div>

    <TooltipContent maxWidth="400px" slot="tooltip-content">
      Select a default time grain for the time series charts which will be shown
      on inital load.
    </TooltipContent>
  </Tooltip>
  <div class="grow">
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
        selection={defaultTimeGrainValue}
        tailwindClasses="overflow-hidden px-2 py-2 rounded"
        alignment="start"
        on:select={handleDefaultTimeGrainUpdate}
      >
        {#if dropdownDisabled}
          {#if !selectedTimeRange}
            <span>Infered from data</span>
          {:else}
            <span>Select a timestamp</span>
          {/if}
        {:else}
          <span style:max-width="14em" class="font-bold truncate"
            >{defaultTimeGrainValue === "__DEFAULT_VALUE__"
              ? "Infer from timerange"
              : prettyTimeGrain(defaultTimeGrainValue)}</span
          >
        {/if}
      </SelectMenu>

      <TooltipContent slot="tooltip-content">
        {tooltipText}
      </TooltipContent>
    </Tooltip>
  </div>
  <Spacer size="24px" />
</div>
