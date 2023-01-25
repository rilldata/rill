<script lang="ts">
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithBisector } from "@rilldata/web-common/components/data-graphic/functional-components";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import Spinner from "@rilldata/web-common/features/temp/Spinner.svelte";
  import { removeTimezoneOffset } from "@rilldata/web-common/lib/formatters";
  import {
    useRuntimeServiceMetricsViewTimeSeries,
    useRuntimeServiceMetricsViewTotals,
    V1MetricsViewTimeSeriesResponse,
    V1MetricsViewTotalsResponse,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { useMetaQuery } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import { extent } from "d3-array";
  import { fly } from "svelte/transition";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../../../../application-state-stores/explorer-stores";
  import { convertTimestampPreview } from "../../../../util/convertTimestampPreview";
  import {
    addGrains,
    formatDateByInterval,
  } from "../time-controls/time-range-utils";
  import MeasureBigNumber from "./MeasureBigNumber.svelte";
  import TimeSeriesBody from "./TimeSeriesBody.svelte";
  import TimeSeriesChartContainer from "./TimeSeriesChartContainer.svelte";

  export let metricViewName;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  $: instanceId = $runtimeStore.instanceId;

  // query the `/meta` endpoint to get the measures and the default time grain
  $: metaQuery = useMetaQuery(instanceId, metricViewName);

  $: timeDimension = $metaQuery.data?.timeDimension;

  $: selectedMeasureNames = metricsExplorer?.selectedMeasureNames;

  $: interval = metricsExplorer?.selectedTimeRange?.interval;

  let totalsQuery: UseQueryStoreResult<V1MetricsViewTotalsResponse, Error>;
  $: if (
    metricsExplorer &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching
  ) {
    const totalsQueryParams = {
      measureNames: selectedMeasureNames,
      filter: metricsExplorer?.filters,
      timeStart: metricsExplorer.selectedTimeRange?.start,
      timeEnd: metricsExplorer.selectedTimeRange?.end,
    };

    totalsQuery = useRuntimeServiceMetricsViewTotals(
      instanceId,
      metricViewName,
      totalsQueryParams
    );
  }

  let timeSeriesQuery: UseQueryStoreResult<
    V1MetricsViewTimeSeriesResponse,
    Error
  >;
  $: if (
    metricsExplorer &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching &&
    metricsExplorer.selectedTimeRange
  ) {
    timeSeriesQuery = useRuntimeServiceMetricsViewTimeSeries(
      instanceId,
      metricViewName,
      {
        measureNames: selectedMeasureNames,
        filter: metricsExplorer?.filters,
        timeStart: metricsExplorer.selectedTimeRange?.start,
        timeEnd: metricsExplorer.selectedTimeRange?.end,
        // Quick hack for now, API expects "day" instead of "1 day"
        timeGranularity: metricsExplorer.selectedTimeRange?.interval,
      }
    );
  }

  // When changing the timeseries query and the cache is empty, $timeSeriesQuery.data?.data is
  // temporarily undefined as results are fetched.
  // To avoid unmounting TimeSeriesBody, which would cause us to lose our tween animations,
  // we make a copy of the data that avoids `undefined` transition states.
  // TODO: instead, try using svelte-query's `keepPreviousData = True` option.
  let dataCopy;

  $: if ($timeSeriesQuery?.data?.data) dataCopy = $timeSeriesQuery.data.data;

  // formattedData adjusts the data to account for Javascript's handling of timezones
  let formattedData;
  $: if (dataCopy)
    formattedData = convertTimestampPreview(dataCopy, timeDimension, true)
      // FIXME: we will need to refactor the graph component animations based on the runtime API return
      // signature. Previously, we were returning 0s instead of nulls. This was likely due to re-using
      // the old diagnostic ts code here. Of course, this isn't correct; null is not the same as 0.
      // For now, let's keep the behavior as-is to ship 0.16. Someone will need to go through and
      // update the animations to work with line segments in the future.
      // An ideal way to fix this would be to segmentize the time series per chart and then tween
      // the individual segments. Alternatively, writing a custom array interpolator could help quite
      // a bit; null values within the interpolator could tween from 0 or from a contiguous point.
      .map((di) => {
        // set nulls to 0, as per the FIXME comment above.
        Object.keys(di).forEach((k) => {
          di[k] = di[k] === null ? 0 : di[k];
        });
        return di;
      });

  let mouseoverValue = undefined;

  let startValue: Date;
  let endValue: Date;
  $: if (metricsExplorer?.selectedTimeRange) {
    startValue = removeTimezoneOffset(
      new Date(metricsExplorer?.selectedTimeRange?.start)
    );
    // selectedTimeRange.end is exclusive and rounded to the time grain ("interval").
    // Since values are grouped with DATE_TRUNC, we subtract one grain to get the (inclusive) axis end.
    endValue = new Date(metricsExplorer?.selectedTimeRange?.end);
    endValue = addGrains(
      endValue,
      -1,
      metricsExplorer?.selectedTimeRange?.interval
    );
    endValue = removeTimezoneOffset(endValue);
  }
</script>

<WithBisector
  data={formattedData}
  callback={(datum) => datum.ts}
  value={mouseoverValue?.x}
  let:point
>
  <TimeSeriesChartContainer start={startValue} end={endValue}>
    <!-- mouseover date elements-->
    <div class="bg-white sticky left-0 top-0" />
    <div class="bg-white sticky left-0 top-0">
      <div style:padding-left="24px">
        {#if point?.ts}
          <div
            class="absolute text-gray-500"
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
      {#if metricsExplorer?.selectedTimeRange}
        <SimpleDataGraphic
          height={32}
          top={34}
          bottom={0}
          xMin={startValue}
          xMax={endValue}
        >
          <Axis superlabel side="top" />
        </SimpleDataGraphic>
      {/if}
    </div>
    <!-- bignumbers and line charts -->
    {#if $metaQuery.data?.measures && $totalsQuery?.isSuccess}
      {#each $metaQuery.data?.measures as measure, index (measure.name)}
        <!-- FIXME: I can't select the big number by the measure id. -->
        {@const bigNum = $totalsQuery?.data.data?.[measure.name]}
        {@const yExtents = extent(dataCopy ?? [], (d) => d[`measure_${index}`])}

        <!-- FIXME: I can't select a time series by measure id. -->
        <MeasureBigNumber
          value={bigNum}
          description={measure?.description ||
            measure?.label ||
            measure?.expression}
          formatPreset={measure?.format}
          status={$totalsQuery?.isFetching
            ? EntityStatus.Running
            : EntityStatus.Idle}
        >
          <svelte:fragment slot="name">
            {measure?.label || measure?.expression}
          </svelte:fragment>
        </MeasureBigNumber>
        <div class="time-series-body" style:height="125px">
          {#if $timeSeriesQuery?.isError}
            <div class="p-5"><CrossIcon /></div>
          {:else if formattedData}
            <TimeSeriesBody
              bind:mouseoverValue
              formatPreset={measure?.format}
              data={formattedData}
              accessor={measure.name}
              mouseover={point}
              timeGrain={metricsExplorer.selectedTimeRange?.interval}
              yMin={yExtents[0] < 0 ? yExtents[0] : 0}
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
