<script lang="ts">
  import { SelectMenu } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    getRelativeTimeRangeOptions,
    ISODurationToTimeRange,
    isTimeRangeValidForTimeGrain,
    timeRangeToISODuration,
  } from "@rilldata/web-common/features/dashboards/time-controls/time-range-utils";
  import { unitToTimeGrain } from "@rilldata/web-common/lib/time/grains";
  import {
    useQueryServiceColumnTimeRange,
    V1Model,
  } from "@rilldata/web-common/runtime-client";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import { supportedTimeRangeEnums } from "../../../dashboards/time-controls/time-control-types";
  import {
    CONFIG_SELECTOR,
    CONFIG_TOP_LEVEL_LABEL_CLASSES,
    INPUT_ELEMENT_CONTAINER,
    SELECTOR_CONTAINER,
  } from "../styles";
  import FormattedSelectorText from "./FormattedSelectorText.svelte";

  export let metricsInternalRep;
  export let selectedModel: V1Model;

  let metricsConfigErrorStore = getContext(
    "rill:metrics-config:errors"
  ) as Writable<any>;

  // this is the value that is selected in the dropdown.
  $: timeRangeSelectedValue =
    $metricsInternalRep.getMetricKey("default_time_range") ||
    "__DEFAULT_VALUE__";

  $: timeColumn = $metricsInternalRep.getMetricKey("timeseries");

  $: smallestTimeGrain = unitToTimeGrain(
    $metricsInternalRep.getMetricKey("smallest_time_grain")
  );

  $: selectedTimeRangeName =
    ISODurationToTimeRange(timeRangeSelectedValue, false) ||
    timeRangeSelectedValue;

  let timeRangeQuery;
  $: if (selectedModel?.name && timeColumn) {
    timeRangeQuery = useQueryServiceColumnTimeRange(
      $runtime.instanceId,
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
      smallestTimeGrain
    );
  }

  $: isValidTimeRange =
    timeRangeSelectedValue === "__DEFAULT_VALUE__" ||
    supportedTimeRangeEnums.includes(selectedTimeRangeName);

  $: hasValidTimeRangeForGrain =
    isValidTimeRange &&
    (timeRangeSelectedValue === "__DEFAULT_VALUE__" ||
      isTimeRangeValidForTimeGrain(smallestTimeGrain, selectedTimeRangeName));

  $: level = isValidTimeRange && hasValidTimeRangeForGrain ? "" : "error";

  $: errorMenuDescription = !isValidTimeRange
    ? "default time range is not valid"
    : !hasValidTimeRangeForGrain
    ? "default time range not valid for the selected smallest time grain"
    : undefined;

  $: metricsConfigErrorStore.update((errors) => {
    errors.defaultTimeRange = level === "error" ? errorMenuDescription : null;
    return errors;
  });

  // currently selected display label.
  $: timeRangeName = !isValidTimeRange
    ? timeRangeSelectedValue
    : selectedTimeRangeName;

  $: options = [
    { key: "__DEFAULT_VALUE__", main: "Infer from data", divider: true },
    ...(level === "error"
      ? [
          {
            key: timeRangeSelectedValue,
            description: errorMenuDescription,
            main: timeRangeName,
            divider: false,
          },
        ]
      : []),

    ...selectableTimeRanges.map((range) => {
      return {
        divider: false,
        key: timeRangeToISODuration(range.name),
        main: range.name,
      };
    }),
  ];

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
  const MAIN_TOOLTIP = "default time range for the time series charts";
  const TOOLTIP_WIDTH = "300px";
  $: if (selectedModel?.name === undefined) {
    tooltipText = "Select a model before selecting a time range";
    dropdownDisabled = true;
  } else if (!timeColumn) {
    tooltipText = "The selected model has no timestamp columns";
    dropdownDisabled = true;
  } else if (!isValidTimeRange) {
    tooltipText = "The selected time range is not valid";
    dropdownDisabled = false;
  } else if (!hasValidTimeRangeForGrain) {
    tooltipText =
      "The selected time range is not valid for the selected time grain";
    dropdownDisabled = false;
  } else {
    tooltipText = MAIN_TOOLTIP;
    dropdownDisabled = false;
  }

  let active = false;
</script>

<div
  class={INPUT_ELEMENT_CONTAINER.classes}
  style={INPUT_ELEMENT_CONTAINER.style}
>
  <Tooltip alignment="start" distance={16} location="bottom">
    <div class={CONFIG_TOP_LEVEL_LABEL_CLASSES}>Default Time Range</div>

    <TooltipContent maxWidth={TOOLTIP_WIDTH} slot="tooltip-content">
      {MAIN_TOOLTIP}
    </TooltipContent>
  </Tooltip>
  <div class={SELECTOR_CONTAINER.classes} style={SELECTOR_CONTAINER.style}>
    <Tooltip
      alignment="middle"
      distance={16}
      location="right"
      suppress={active}
    >
      <SelectMenu
        bind:active
        block
        paddingTop={1}
        paddingBottom={1}
        {options}
        disabled={dropdownDisabled}
        selection={timeRangeSelectedValue}
        tailwindClasses="{CONFIG_SELECTOR.base} {level === 'error'
          ? CONFIG_SELECTOR.error
          : CONFIG_SELECTOR.info}"
        activeTailwindClasses={level === "error"
          ? CONFIG_SELECTOR.activeError
          : CONFIG_SELECTOR.active}
        distance={CONFIG_SELECTOR.distance}
        alignment="start"
        on:select={handleDefaultTimeRangeUpdate}
      >
        <FormattedSelectorText
          value={timeRangeSelectedValue === "__DEFAULT_VALUE__"
            ? "Infer from data"
            : timeRangeName}
          selected={timeRangeSelectedValue !== "__DEFAULT_VALUE__"}
        />
      </SelectMenu>

      <TooltipContent maxWidth={TOOLTIP_WIDTH} slot="tooltip-content">
        {tooltipText}
      </TooltipContent>
    </Tooltip>
  </div>
</div>
