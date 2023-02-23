<script lang="ts">
  import { SelectMenu } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    getTimeGrainOptions,
    timeGrainEnumToYamlString,
    TimeGrainOption,
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

  $: defaultTimeGrainValue =
    $metricsInternalRep.getMetricKey("smallest_time_grain") ||
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

  $: options = [
    {
      key: "__DEFAULT_VALUE__",
      main: "Infer from data",
      divider: true,
    },
  ].concat(
    selectableTimeGrains.map((grain, i) => {
      const isGrainPossible = i <= maxTimeGrainPossibleIndex;
      return {
        divider: false,
        key: timeGrainEnumToYamlString(grain.timeGrain),
        main: timeGrainEnumToYamlString(grain.timeGrain),
        disabled: !isGrainPossible,
        description: !isGrainPossible
          ? "not valid for this time range"
          : undefined,
      };
    }) as any[]
  );

  function handleDefaultTimeGrainUpdate(event) {
    const selectedTimeGrain = event.detail?.key;

    if (selectedTimeGrain === "" || selectedTimeGrain === "__DEFAULT_VALUE__") {
      $metricsInternalRep.updateMetricsParams({
        smallest_time_grain: "",
        default_time_range: "",
      });
    } else {
      $metricsInternalRep.updateMetricsParams({
        smallest_time_grain: timeGrainEnumToYamlString(selectedTimeGrain),
        default_time_range: "",
      });
    }
  }

  let tooltipText = "";
  let dropdownDisabled = true;
  // FIXME: we won't show this element if there's no time column
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

  let active = false;
</script>

<div
  class={INPUT_ELEMENT_CONTAINER.classes}
  style={INPUT_ELEMENT_CONTAINER.style}
>
  <Tooltip alignment="start" distance={8} location="bottom">
    <div class={CONFIG_TOP_LEVEL_LABEL_CLASSES}>Smallest Time Grain</div>

    <TooltipContent maxWidth="280px" slot="tooltip-content">
      The smallest allowable time unit that can be displayed on the dashboard
      line charts
    </TooltipContent>
  </Tooltip>
  <div class={SELECTOR_CONTAINER.classes} style={SELECTOR_CONTAINER.style}>
    <Tooltip
      alignment="middle"
      distance={8}
      location="right"
      suppress={tooltipText === undefined}
    >
      <SelectMenu
        paddingTop={1}
        paddingBottom={1}
        bind:active
        block
        {options}
        disabled={dropdownDisabled}
        selection={defaultTimeGrainValue}
        tailwindClasses="{CONFIG_SELECTOR.base} {CONFIG_SELECTOR.info}"
        activeTailwindClasses={CONFIG_SELECTOR.active}
        distance={CONFIG_SELECTOR.distance}
        alignment="start"
        on:select={handleDefaultTimeGrainUpdate}
      >
        <FormattedSelectorText
          value={defaultTimeGrainValue === "__DEFAULT_VALUE__"
            ? "Infer from data"
            : defaultTimeGrainValue}
          selected={defaultTimeGrainValue !== "__DEFAULT_VALUE__"}
        />
      </SelectMenu>

      <TooltipContent slot="tooltip-content">
        {tooltipText}
      </TooltipContent>
    </Tooltip>
  </div>
</div>
