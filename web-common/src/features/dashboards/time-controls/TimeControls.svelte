<!--
@component
Constructs a TimeSeriesTimeRange object – to be used as the filter in MetricsExplorer – by taking as input:
- a base time range
- a time grain (e.g., "hour" or "day")
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
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
  import {
    addGrains,
    checkValidTimeGrain,
    floorDate,
    getDefaultTimeGrain,
    getDefaultTimeRange,
    getTimeGrainOptions,
    TimeGrainOption,
  } from "./time-range-utils";
  import TimeGrainSelector from "./TimeGrainSelector.svelte";
  import TimeRangeSelector from "./TimeRangeSelector.svelte";

  export let metricViewName: string;

  const dashboardStore = useDashboardStore(metricViewName);

  let baseTimeRange: TimeRange;

  let metricsViewQuery;
  $: if ($runtimeStore.instanceId) {
    metricsViewQuery = useRuntimeServiceGetCatalogEntry(
      $runtimeStore.instanceId,
      metricViewName
    );
  }

  // once we have the allTimeRange, set the default time range and time grain
  $: if (allTimeRange) {
    const timeRange = getDefaultTimeRange(allTimeRange);
    const timeGrain = getDefaultTimeGrain(timeRange.start, timeRange.end);
    makeTimeSeriesTimeRangeAndUpdateAppState(timeRange, timeGrain);
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

  // we get the timeGrainOptions so that we can assess whether or not the
  // activeTimeGrain is valid whenever the baseTimeRange changes
  let timeGrainOptions: TimeGrainOption[];
  $: timeGrainOptions = getTimeGrainOptions(
    new Date($dashboardStore?.selectedTimeRange?.start),
    new Date($dashboardStore?.selectedTimeRange?.end)
  );

  function onSelectTimeRange(name: TimeRangeName, start: string, end: string) {
    baseTimeRange = {
      name,
      start: new Date(start),
      end: new Date(end),
    };
    makeTimeSeriesTimeRangeAndUpdateAppState(
      baseTimeRange,
      $dashboardStore.selectedTimeRange.interval
    );
  }

  function onSelectTimeGrain(timeGrain: TimeGrain) {
    makeTimeSeriesTimeRangeAndUpdateAppState(baseTimeRange, timeGrain);
  }

  function makeTimeSeriesTimeRangeAndUpdateAppState(
    timeRange: TimeRange,
    timeGrain: TimeGrain
  ) {
    const { name, start, end } = timeRange;

    // validate time range name + time grain combination
    // (necessary because when the time range name is changed, the current time grain may not be valid for the new time range name)
    timeGrainOptions = getTimeGrainOptions(start, end);
    const isValidTimeGrain = checkValidTimeGrain(timeGrain, timeGrainOptions);
    if (!isValidTimeGrain) {
      timeGrain = getDefaultTimeGrain(start, end);
    }

    // Round start time to nearest lower time grain
    const adjustedStart = floorDate(start, timeGrain);

    // Round end time to start of next grain, since end times are exclusive
    let adjustedEnd: Date;
    adjustedEnd = addGrains(new Date(end), 1, timeGrain);
    adjustedEnd = floorDate(adjustedEnd, timeGrain);

    // the adjusted time range
    const newTimeRange: TimeSeriesTimeRange = {
      name: name,
      start: adjustedStart.toISOString(),
      end: adjustedEnd.toISOString(),
      interval: timeGrain,
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
      on:select-time-range={(e) =>
        onSelectTimeRange(e.detail.name, e.detail.start, e.detail.end)}
    />
    <TimeGrainSelector
      on:select-time-grain={(e) => onSelectTimeGrain(e.detail.timeGrain)}
      {metricViewName}
      {timeGrainOptions}
    />
  {/if}
</div>
