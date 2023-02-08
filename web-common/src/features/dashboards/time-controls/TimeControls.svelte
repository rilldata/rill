<!--
@component
Constructs a TimeSeriesTimeRange object – to be used as the filter in MetricsExplorer – by taking as input:
- the time range name (a semantic understanding of the time range, like "Last 6 Hours" or "Last 30 days")
- the time grain (e.g., "hour" or "day")
- the dataset's full time range (so its end time can be used in relative time ranges)

We should rename TimeSeriesTimeRange to a better name.
-->
<script lang="ts">
  import { goto } from "$app/navigation";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import { useModelHasTimeSeries } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    TimeGrain,
    TimeRange,
    TimeRangeName,
    TimeSeriesTimeRange,
  } from "@rilldata/web-common/features/dashboards/time-controls/time-control-types";
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceGetTimeRangeSummary,
    V1GetTimeRangeSummaryResponse,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import { selectTimestampColumnFromSchema } from "../../metrics-views/column-selectors";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../dashboard-stores";
  import {
    addGrains,
    checkValidTimeGrain,
    floorDate,
    getDefaultTimeGrain,
    getDefaultTimeRange,
    getSelectableTimeGrains,
    TimeGrainOption,
  } from "./time-range-utils";
  import TimeGrainSelector from "./TimeGrainSelector.svelte";
  import TimeRangeSelector from "./TimeRangeSelector.svelte";

  export let metricViewName: string;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  // user selections, used to construct TimeSeriesTimeRange
  let selectedTimeRange: TimeRange;
  let selectedTimeGrain: TimeGrain;

  $: selectedTimeRange = {
    name: metricsExplorer?.selectedTimeRange?.name,
    start: new Date(metricsExplorer?.selectedTimeRange?.start),
    end: new Date(metricsExplorer?.selectedTimeRange?.end),
  };
  $: selectedTimeGrain = metricsExplorer?.selectedTimeRange?.interval;

  $: metricsViewQuery = useRuntimeServiceGetCatalogEntry(
    $runtimeStore.instanceId,
    metricViewName,
    {
      query: {
        enabled: !!$runtimeStore.instanceId,
      },
    }
  );

  // once we have the allTimeRange, set the default time range and time grain
  $: if (allTimeRange) {
    const timeRange = getDefaultTimeRange(allTimeRange);
    selectedTimeGrain = getDefaultTimeGrain(timeRange.start, timeRange.end);
    setSelectedTimeRange(
      timeRange.name,
      timeRange.start.toISOString(),
      timeRange.end.toISOString()
    );
  }

  $: metricTimeSeries = useModelHasTimeSeries(
    $runtimeStore.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $metricTimeSeries?.data;

  let timestampColumns: Array<string>;

  let modelQuery;
  $: if ($metricsViewQuery?.data?.entry?.metricsView?.model) {
    modelQuery = useRuntimeServiceGetCatalogEntry(
      $runtimeStore.instanceId,
      $metricsViewQuery?.data?.entry?.metricsView?.model
    );
  }

  $: if ($modelQuery && $modelQuery.isSuccess && !$modelQuery.isRefetching) {
    const model = $modelQuery.data?.entry?.model;
    timestampColumns = selectTimestampColumnFromSchema(model?.schema);
  } else {
    timestampColumns = [];
  }

  $: redirectToScreen = timestampColumns?.length > 0 ? "metrics" : "model";

  let timeRangeQuery: UseQueryStoreResult<V1GetTimeRangeSummaryResponse, Error>;
  $: if (
    hasTimeSeries &&
    !!$runtimeStore?.instanceId &&
    !!$metricsViewQuery?.data?.entry?.metricsView?.model &&
    !!$metricsViewQuery?.data?.entry?.metricsView?.timeDimension
  ) {
    timeRangeQuery = useRuntimeServiceGetTimeRangeSummary(
      $runtimeStore.instanceId,
      $metricsViewQuery.data.entry.metricsView.model,
      {
        columnName: $metricsViewQuery.data.entry.metricsView.timeDimension,
      }
    );
  }

  let allTimeRange: TimeRange;
  $: if (hasTimeSeries && $timeRangeQuery?.data?.timeRangeSummary) {
    allTimeRange = {
      name: TimeRangeName.AllTime,
      start: new Date($timeRangeQuery.data.timeRangeSummary.min),
      end: new Date($timeRangeQuery.data.timeRangeSummary.max),
    };
  }

  // we get the selectableTimeGrains so that we can assess whether or not the
  // existing selectedTimeGrain is valid whenever the selectedTimeRangeName changes
  let selectableTimeGrains: TimeGrainOption[];
  $: selectableTimeGrains = getSelectableTimeGrains(
    selectedTimeRange?.start,
    selectedTimeRange?.end
  );

  function setSelectedTimeRange(
    name: TimeRangeName,
    start: string,
    end: string
  ) {
    selectedTimeRange = {
      name: name,
      start: new Date(start),
      end: new Date(end),
    };
    makeTimeSeriesTimeRangeAndUpdateAppState(
      name,
      selectedTimeRange.start,
      selectedTimeRange.end
    );
  }

  function setSelectedTimeGrain(timeGrain: TimeGrain) {
    selectedTimeGrain = timeGrain;
    makeTimeSeriesTimeRangeAndUpdateAppState(
      selectedTimeRange.name,
      selectedTimeRange.start,
      selectedTimeRange.end
    );
  }

  function makeTimeSeriesTimeRangeAndUpdateAppState(
    name: TimeRangeName,
    start: Date,
    end: Date
  ) {
    // validate time range name + time grain combination
    // (necessary because when the time range name is changed, the current time grain may not be valid for the new time range name)
    selectableTimeGrains = getSelectableTimeGrains(start, end);
    const isValidTimeGrain = checkValidTimeGrain(
      selectedTimeGrain,
      selectableTimeGrains
    );
    if (!isValidTimeGrain) {
      selectedTimeGrain = getDefaultTimeGrain(start, end);
    }

    // Round start time to nearest lower time grain
    const adjustedStart = floorDate(start, selectedTimeGrain);

    // Round end time to start of next grain, since end times are exclusive
    let adjustedEnd = addGrains(
      new Date(allTimeRange?.end),
      1,
      selectedTimeGrain
    );
    adjustedEnd = floorDate(adjustedEnd, selectedTimeGrain);

    // the adjusted time range
    const newTimeRange: TimeSeriesTimeRange = {
      name: name,
      start: adjustedStart.toISOString(),
      end: adjustedEnd.toISOString(),
      interval: selectedTimeGrain,
    };

    metricsExplorerStore.setSelectedTimeRange(metricViewName, newTimeRange);
  }

  function noTimeseriesCTA() {
    if (timestampColumns?.length) {
      goto(`/dashboard/${metricViewName}/edit`);
    } else {
      const modelName = $metricsViewQuery?.data?.entry?.metricsView?.model;
      goto(`/model/${modelName}`);
    }
  }
</script>

<div class="flex flex-row">
  {#if !hasTimeSeries}
    <Tooltip location="bottom" distance={8}>
      <div
        on:click={() => noTimeseriesCTA()}
        class="px-3 py-2 flex flex-row items-center gap-x-3 cursor-pointer"
      >
        <span class="ui-copy-icon"><Calendar size="16px" /></span>
        <span class="ui-copy-disabled">No time dimension specified</span>
      </div>
      <TooltipContent slot="tooltip-content" maxWidth="250px">
        Add a time dimension to your {redirectToScreen} to enable time series plots.
        <TooltipShortcutContainer>
          <div class="capitalize">Edit {redirectToScreen}</div>
          <Shortcut>Click</Shortcut>
        </TooltipShortcutContainer>
      </TooltipContent>
    </Tooltip>
  {:else}
    <TimeRangeSelector
      {metricViewName}
      {allTimeRange}
      {selectedTimeRange}
      on:select-time-range={(e) =>
        setSelectedTimeRange(e.detail.name, e.detail.start, e.detail.end)}
    />
    <TimeGrainSelector
      on:select-time-grain={(e) => setSelectedTimeGrain(e.detail.timeGrain)}
      {selectableTimeGrains}
      {selectedTimeGrain}
    />
  {/if}
</div>
