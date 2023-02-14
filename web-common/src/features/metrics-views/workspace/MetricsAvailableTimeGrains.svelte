<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    useRuntimeServiceGetTimeRangeSummary,
    V1Model,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { removeIfExists } from "@rilldata/web-local/lib/util/arrayUtils";
  import Spacer from "../../../components/icons/Spacer.svelte";
  import { SelectMenu } from "../../../components/menu";
  import {
    getAvailableTimeGrains,
    timeGrainEnumToYamlString,
  } from "../../dashboards/time-controls/time-range-utils";

  export let metricsInternalRep;
  export let selectedModel: V1Model;

  $: selectedTimeRange = $metricsInternalRep.getMetricKey("default_time_range");

  $: timeGrainsInYaml = $metricsInternalRep.getMetricKey("time_grains") || [];
  $: availableTimeGrains = timeGrainsInYaml.length
    ? timeGrainsInYaml
    : ["__DEFAULT_VALUE__"];

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

  let selectableTimeGrains: V1TimeGrain[] = [];
  $: if (selectedTimeRange) {
    // get all available time grains
    selectableTimeGrains = getAvailableTimeGrains(allTimeRange);
  }

  $: options = [
    { key: "__DEFAULT_VALUE__", main: "Infer from timerange", divider: true },
  ].concat(
    selectableTimeGrains.map((grain) => {
      return {
        key: timeGrainEnumToYamlString(grain),
        main: timeGrainEnumToYamlString(grain),
        divider: false,
      };
    })
  );

  function handleAvailableTimeGrainsUpdate(event) {
    const selectedTimeGrain = event.detail?.key;

    if (selectedTimeGrain === "__DEFAULT_VALUE__") {
      $metricsInternalRep.updateMetricsParams({
        time_grains: [],
      });
      availableTimeGrains = ["__DEFAULT_VALUE__"];
      return;
    } else {
      availableTimeGrains = availableTimeGrains.filter(
        (timeGrain) => timeGrain !== "__DEFAULT_VALUE__"
      );
    }

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

  const prettyTimeGrains = (timeGrains) => {
    if (timeGrains[0] === "__DEFAULT_VALUE__") {
      return "Infer from timerange";
    }
    return timeGrains.join(", ");
  };

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

<div class="w-80 flex items-center">
  <Tooltip alignment="middle" distance={8} location="bottom">
    <div class="text-gray-500 font-medium" style="width:10em; font-size:11px;">
      Available Time Grains
    </div>

    <TooltipContent slot="tooltip-content">
      Select the timegrains that will be available in the dashboard
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
        multiSelect={true}
        disabled={dropdownDisabled}
        selection={availableTimeGrains}
        tailwindClasses="overflow-hidden px-2 py-2 rounded"
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
          <span style:max-width="14em" class="font-bold text-left"
            >{availableTimeGrains.length
              ? prettyTimeGrains(availableTimeGrains)
              : "Infer from timerange"}</span
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
