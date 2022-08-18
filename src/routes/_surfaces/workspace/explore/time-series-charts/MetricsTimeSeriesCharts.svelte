<script lang="ts">
  import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type {
    MetricViewMetaResponse,
    MetricViewTimeSeriesResponse,
  } from "$common/rill-developer-service/MetricViewActions";
  import SimpleDataGraphic from "$lib/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithBisector } from "$lib/components/data-graphic/functional-components";
  import { Axis } from "$lib/components/data-graphic/guides";
  import CrossIcon from "$lib/components/icons/CrossIcon.svelte";
  import Spinner from "$lib/components/Spinner.svelte";
  import { getBigNumberById } from "$lib/redux-store/big-number/big-number-readables";
  import type { BigNumberEntity } from "$lib/redux-store/big-number/big-number-slice";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import type { TimeSeriesValue } from "$lib/redux-store/timeseries/timeseries-slice";
  import {
    getMetricViewMetadata,
    getMetricViewMetaQueryKey,
    getMetricViewTimeSeries,
    getMetricViewTimeSeriesQueryKey,
  } from "$lib/svelte-query/queries/metric-view";
  import { convertTimestampPreview } from "$lib/util/convertTimestampPreview";
  import { removeTimezoneOffset } from "$lib/util/formatters";
  import { NicelyFormattedTypes } from "$lib/util/humanize-numbers";
  import { useQuery } from "@sveltestack/svelte-query";
  import { extent } from "d3-array";
  import type { Readable } from "svelte/store";
  import { fly } from "svelte/transition";
  import { formatDateByInterval } from "../time-controls/time-range-utils";
  import MeasureBigNumber from "./MeasureBigNumber.svelte";
  import TimeSeriesBody from "./TimeSeriesBody.svelte";
  import TimeSeriesChartContainer from "./TimeSeriesChartContainer.svelte";

  export let metricsDefId;

  let metricsExplorer: Readable<MetricsExplorerEntity>;
  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  // query the `/meta` endpoint to get the measures and the default time grain
  let queryKey = getMetricViewMetaQueryKey(metricsDefId);
  const queryResult = useQuery<MetricViewMetaResponse, Error>(queryKey, () =>
    getMetricViewMetadata(metricsDefId)
  );
  $: {
    queryKey = getMetricViewMetaQueryKey(metricsDefId);
    queryResult.setOptions(queryKey, () => getMetricViewMetadata(metricsDefId));
  }

  $: interval =
    $metricsExplorer?.selectedTimeRange?.interval ||
    $queryResult.data?.timeDimension?.timeRange?.interval;

  let bigNumbers: Readable<BigNumberEntity>;
  $: bigNumbers = getBigNumberById(metricsDefId);

  // query the `/timeseries` endpoint
  let timeSeriesQueryKey = getMetricViewTimeSeriesQueryKey(metricsDefId);
  let timeSeriesQueryFn = () =>
    getMetricViewTimeSeries(metricsDefId, {
      measures: $metricsExplorer.measureIds,
      time: {
        start: $metricsExplorer.selectedTimeRange.start, // TODO: make selectedTimeRange a required field (use "default" functions in time-range-utils.ts)
        end: $metricsExplorer.selectedTimeRange.end,
        granularity: $metricsExplorer.selectedTimeRange.interval,
      },
      filter: { include: [], exclude: [] }, // TODO: remove the filter, once we've designated it as optional in the API
    });
  const timeSeriesQueryResult = useQuery<MetricViewTimeSeriesResponse, Error>(
    timeSeriesQueryKey,
    timeSeriesQueryFn
  );
  $: {
    timeSeriesQueryKey = getMetricViewTimeSeriesQueryKey(metricsDefId);
    timeSeriesQueryResult.setOptions(timeSeriesQueryKey, timeSeriesQueryFn);
  }

  $: formattedData = $timeSeriesQueryResult.data.data
    ? convertTimestampPreview($timeSeriesQueryResult.data.data, true)
    : undefined;

  let mouseoverValue = undefined;

  $: key = `${startValue}` + `${endValue}`;

  $: [minVal, maxVal] = extent(
    $timeSeriesQueryResult.data.data ?? [],
    (d: TimeSeriesValue) => d.ts
  );
  $: startValue = removeTimezoneOffset(new Date(minVal));
  $: endValue = removeTimezoneOffset(new Date(maxVal));
</script>

<WithBisector
  data={formattedData}
  callback={(datum) => datum.ts}
  value={mouseoverValue?.x}
  let:point
>
  <TimeSeriesChartContainer start={startValue} end={endValue}>
    <!-- mouseover date elements-->
    <div />
    <div style:padding-left="24px">
      {#if point?.ts}
        <div
          class="absolute italic text-gray-600"
          transition:fly|local={{ duration: 100, y: 4 }}
        >
          {formatDateByInterval(interval, point.ts)}
        </div>
        &nbsp;
      {:else}
        &nbsp;
      {/if}
    </div>
    <!-- top axis element -->
    <div />
    <SimpleDataGraphic
      height={32}
      top={34}
      bottom={0}
      xMin={startValue}
      xMax={endValue}
    >
      <Axis superlabel side="top" />
    </SimpleDataGraphic>
    <!-- bignumbers and line charts -->
    {#if $queryResult.isSuccess}
      {#each $queryResult.data.measures as measure, index (measure.id)}
        <!-- FIXME: I can't select the big number by the measure id. -->
        {@const bigNum = $bigNumbers?.bigNumbers?.[`measure_${index}`]}

        <!-- FIXME: I can't select a time series by measure id. -->
        <MeasureBigNumber
          value={bigNum}
          description={measure?.description ||
            measure?.label ||
            measure?.expression}
          formatPreset={measure?.formatPreset || NicelyFormattedTypes.HUMANIZE}
          status={$bigNumbers?.status}
        >
          <svelte:fragment slot="name">
            {measure?.label || measure?.expression}
          </svelte:fragment>
        </MeasureBigNumber>
        <div class="time-series-body" style:height="125px">
          {#if $timeSeriesQueryResult.isError}
            <div class="p-5"><CrossIcon /></div>
          {:else if formattedData}
            <TimeSeriesBody
              bind:mouseoverValue
              formatPreset={measure?.formatPreset ||
                NicelyFormattedTypes.HUMANIZE}
              data={formattedData}
              accessor={`measure_${index}`}
              mouseover={point}
              {key}
              start={startValue}
              end={endValue}
            />
          {:else}
            <div>
              <Spinner status={EntityStatus.Running} />
            </div>
          {/if}
        </div>
      {/each}
    {/if}
  </TimeSeriesChartContainer>
</WithBisector>
