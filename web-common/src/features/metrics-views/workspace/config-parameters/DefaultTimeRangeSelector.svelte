<script lang="ts">
  import { SelectMenu } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    getRelativeTimeRangeOptions,
    ISODurationToTimeRange,
    timeGrainStringToEnum,
    timeRangeToISODuration,
  } from "@rilldata/web-common/features/dashboards/time-controls/time-range-utils";
  import {
    useRuntimeServiceGetTimeRangeSummary,
    V1Model,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import {
    CONFIG_SELECTOR,
    CONFIG_TOP_LEVEL_LABEL_CLASSES,
    INPUT_ELEMENT_CONTAINER,
    SELECTOR_CONTAINER,
  } from "../styles";
  import FormattedSelectorText from "./FormattedSelectorText.svelte";

  export let metricsInternalRep;
  export let selectedModel: V1Model;

  $: timeRangeSelectedValue =
    $metricsInternalRep.getMetricKey("default_time_range") ||
    "__DEFAULT_VALUE__";

  $: timeColumn = $metricsInternalRep.getMetricKey("timeseries");
  $: smallestTimeGrain = $metricsInternalRep.getMetricKey(
    "smallest_time_grain"
  );

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
      start: new Date($timeRangeQuery.data.timeRangeSummary.min),
      end: new Date($timeRangeQuery.data.timeRangeSummary.max),
    };
  }

  let selectableTimeRanges = [];
  $: if (allTimeRange) {
    selectableTimeRanges = getRelativeTimeRangeOptions(
      allTimeRange,
      timeGrainStringToEnum(smallestTimeGrain)
    );
  }

  $: options = [
    { key: "__DEFAULT_VALUE__", main: "Infer from data", divider: true },
  ].concat(
    selectableTimeRanges.map((range) => {
      return {
        divider: false,
        key: timeRangeToISODuration(range.name),
        main: range.name,
      };
    })
  );

  function handleDefaultTimeRangeUpdate(event) {
    const timeRangeSelectedValue = event.detail?.key;

    if (timeRangeSelectedValue === "__DEFAULT_VALUE__") {
      $metricsInternalRep.updateMetricsParams({
        default_time_range: "",
      });
    } else {
      $metricsInternalRep.updateMetricsParams({
        default_time_range: timeRangeSelectedValue,
      });
    }
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

<div
  class={INPUT_ELEMENT_CONTAINER.classes}
  style={INPUT_ELEMENT_CONTAINER.style}
>
  <Tooltip alignment="middle" distance={8} location="bottom">
    <div class={CONFIG_TOP_LEVEL_LABEL_CLASSES}>Default Time Range</div>

    <TooltipContent slot="tooltip-content">
      Select a default time range for the time series charts
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
        paddingTop={1}
        paddingBottom={1}
        {options}
        disabled={dropdownDisabled}
        selection={timeRangeSelectedValue}
        tailwindClasses="{CONFIG_SELECTOR.base} {CONFIG_SELECTOR.info}"
        activeTailwindClasses={CONFIG_SELECTOR.active}
        distance={CONFIG_SELECTOR.distance}
        alignment="start"
        on:select={handleDefaultTimeRangeUpdate}
      >
        <FormattedSelectorText
          value={timeRangeSelectedValue === "__DEFAULT_VALUE__"
            ? "Infer from data"
            : ISODurationToTimeRange(timeRangeSelectedValue)}
          selected={timeRangeSelectedValue !== "__DEFAULT_VALUE__"}
        />
      </SelectMenu>

      <TooltipContent slot="tooltip-content">
        {tooltipText}
      </TooltipContent>
    </Tooltip>
  </div>
</div>
