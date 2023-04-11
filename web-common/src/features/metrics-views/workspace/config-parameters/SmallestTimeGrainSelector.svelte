<script lang="ts">
  import { SelectMenu } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { getTimeGrainOptions } from "@rilldata/web-common/lib/time/grains";
  import type { TimeGrainOption } from "@rilldata/web-common/lib/time/types";
  import {
    createQueryServiceColumnTimeRange,
    V1Model,
  } from "@rilldata/web-common/runtime-client";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { runtime } from "../../../../runtime-client/runtime-store";
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

  $: defaultTimeGrainValue =
    $metricsInternalRep.getMetricKey("smallest_time_grain") ||
    "__DEFAULT_VALUE__";

  $: timeColumn = $metricsInternalRep.getMetricKey("timeseries");

  let timeRangeQuery;
  $: if (selectedModel?.name && timeColumn) {
    timeRangeQuery = createQueryServiceColumnTimeRange(
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

  let selectableTimeGrains: TimeGrainOption[] = [];
  let maxTimeGrainPossibleIndex = 0;
  $: if (allTimeRange) {
    selectableTimeGrains = getTimeGrainOptions(
      allTimeRange.start,
      allTimeRange.end
    );

    maxTimeGrainPossibleIndex =
      selectableTimeGrains.length -
      1 -
      selectableTimeGrains
        .slice()
        .reverse()
        .findIndex((grain) => grain.enabled);
  }

  $: isValidTimeGrain =
    defaultTimeGrainValue === "__DEFAULT_VALUE__" ||
    Object.values(TIME_GRAIN).some(
      (timeGrain) => timeGrain.label === defaultTimeGrainValue
    );

  $: level = isValidTimeGrain ? "" : "error";

  $: metricsConfigErrorStore.update((errors) => {
    errors.smallestTimeGrain =
      level === "error" ? "Invalid smallest time grain" : null;
    return errors;
  });

  $: options = [
    {
      key: "__DEFAULT_VALUE__",
      main: "Infer from data",
      divider: true,
    },
    ...(!isValidTimeGrain
      ? [
          {
            key: defaultTimeGrainValue,
            description: "selected time grain is not valid",
            main: defaultTimeGrainValue,
            divider: false,
          },
        ]
      : []),
    ...(selectableTimeGrains.map((grain, i) => {
      const isGrainPossible = i <= maxTimeGrainPossibleIndex;
      return {
        divider: false,
        key: grain.label,
        main: grain.label,
        disabled: !isGrainPossible,
        description: !isGrainPossible
          ? "not valid for this time range"
          : undefined,
      };
    }) as any[]),
  ];

  function handleSelectSmallestTimeGrain(event) {
    const selectedTimeGrain = event.detail?.key;
    if (selectedTimeGrain === "" || selectedTimeGrain === "__DEFAULT_VALUE__") {
      $metricsInternalRep.updateMetricsParams({
        smallest_time_grain: "",
        default_time_range: "",
      });
    } else {
      $metricsInternalRep.updateMetricsParams({
        smallest_time_grain: selectedTimeGrain,
        default_time_range: "",
      });
    }
  }

  const TOOLTIP_WIDTH = "280px";
  const DEFAULT_TOOLTIP_TEXT =
    "The smallest allowable time unit that can be displayed on the dashboard line charts";

  let tooltipText = "";
  let dropdownDisabled = true;
  // FIXME: we won't show this element if there's no time column
  $: if (selectedModel?.name === undefined) {
    tooltipText = "Select a model before selecting a time grain";
    dropdownDisabled = true;
  } else if (!timeColumn) {
    tooltipText = "The selected model has no timestamp columns";
    dropdownDisabled = true;
  } else if (!isValidTimeGrain) {
    tooltipText = "The selected time grain is not valid";
    dropdownDisabled = false;
  } else {
    tooltipText = DEFAULT_TOOLTIP_TEXT;
    dropdownDisabled = false;
  }

  let active = false;
</script>

<div
  class={INPUT_ELEMENT_CONTAINER.classes}
  style={INPUT_ELEMENT_CONTAINER.style}
>
  <Tooltip alignment="start" distance={16} location="bottom">
    <div class={CONFIG_TOP_LEVEL_LABEL_CLASSES}>Smallest Time Grain</div>

    <TooltipContent maxWidth={TOOLTIP_WIDTH} slot="tooltip-content">
      {DEFAULT_TOOLTIP_TEXT}
    </TooltipContent>
  </Tooltip>
  <div class={SELECTOR_CONTAINER.classes} style={SELECTOR_CONTAINER.style}>
    <Tooltip alignment="start" distance={16} location="right" suppress={active}>
      <SelectMenu
        paddingTop={1}
        paddingBottom={1}
        bind:active
        block
        {options}
        disabled={dropdownDisabled}
        selection={defaultTimeGrainValue}
        tailwindClasses="{CONFIG_SELECTOR.base} {level === 'error'
          ? CONFIG_SELECTOR.error
          : CONFIG_SELECTOR.info}"
        activeTailwindClasses={level === "error"
          ? CONFIG_SELECTOR.activeError
          : CONFIG_SELECTOR.active}
        distance={CONFIG_SELECTOR.distance}
        alignment="start"
        on:select={handleSelectSmallestTimeGrain}
      >
        <FormattedSelectorText
          value={defaultTimeGrainValue === "__DEFAULT_VALUE__"
            ? "Infer from data"
            : defaultTimeGrainValue}
          selected={defaultTimeGrainValue !== "__DEFAULT_VALUE__"}
        />
      </SelectMenu>

      <TooltipContent maxWidth={TOOLTIP_WIDTH} slot="tooltip-content">
        {tooltipText}
      </TooltipContent>
    </Tooltip>
  </div>
</div>
