<script lang="ts">
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithBisector } from "@rilldata/web-common/components/data-graphic/functional-components";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import { removeTimezoneOffset } from "@rilldata/web-common/lib/formatters";
  import {
    useRuntimeServiceMetricsViewTimeSeries,
    useRuntimeServiceMetricsViewTotals,
    V1MetricsViewTimeSeriesResponse,
    V1MetricsViewTotalsResponse,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { useMetaQuery } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import { EntityStatus } from "@rilldata/web-local/lib/temp/entity";
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import { extent } from "d3-array";
  import { fly } from "svelte/transition";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../../../../application-state-stores/explorer-stores";
  import { convertTimestampPreview } from "../../../../util/convertTimestampPreview";
  import {
    humanizeDataType,
    NicelyFormattedTypes,
  } from "../../../../util/humanize-numbers";
  import Spinner from "../../../Spinner.svelte";
  import { formatDateByInterval } from "../time-controls/time-range-utils";
  import MeasureBigNumber from "./MeasureBigNumber.svelte";
  import MeasureChart from "./MeasureChart.svelte";
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
    totalsQuery = useRuntimeServiceMetricsViewTotals(
      instanceId,
      metricViewName,
      {
        measureNames: selectedMeasureNames,
        filter: metricsExplorer?.filters,
        timeStart: metricsExplorer.selectedTimeRange?.start,
        timeEnd: metricsExplorer.selectedTimeRange?.end,
      }
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
  $: if (dataCopy && dataCopy?.length)
    formattedData = convertTimestampPreview(dataCopy, timeDimension, true);

  let mouseoverValue = undefined;

  $: startValue = removeTimezoneOffset(
    new Date(metricsExplorer?.selectedTimeRange?.start)
  );
  $: endValue = removeTimezoneOffset(
    new Date(metricsExplorer?.selectedTimeRange?.end)
  );
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
      <SimpleDataGraphic
        height={32}
        top={34}
        bottom={0}
        xMin={startValue}
        xMax={endValue}
      >
        <Axis superlabel side="top" />
      </SimpleDataGraphic>
    </div>
    <!-- bignumbers and line charts -->
    {#if $metaQuery.data?.measures && $totalsQuery?.isSuccess}
      {#each $metaQuery.data?.measures as measure, index (measure.name)}
        <!-- FIXME: I can't select the big number by the measure id. -->
        {@const bigNum = $totalsQuery?.data.data?.[measure.name]}
        {@const yExtents = extent(dataCopy ?? [], (d) => d[`measure_${index}`])}
        {@const formatPreset =
          NicelyFormattedTypes[measure?.format] ||
          NicelyFormattedTypes.HUMANIZE}
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
          {#if $timeSeriesQuery.isError}
            <div class="p-5"><CrossIcon /></div>
          {:else if formattedData}
            <MeasureChart
              bind:mouseoverValue
              data={formattedData}
              xAccessor="ts"
              yAccessor={measure.name}
              timeGrain={metricsExplorer.selectedTimeRange?.interval}
              xMin={startValue}
              xMax={endValue}
              yMin={yExtents[0] < 0 ? yExtents[0] : 0}
              start={startValue}
              end={endValue}
              mouseoverFormat={(value) =>
                formatPreset === NicelyFormattedTypes.NONE
                  ? `${value}`
                  : humanizeDataType(value, formatPreset, {
                      excludeDecimalZeros: true,
                    })}
            />
            {#if false}
              <TimeSeriesBody
                bind:mouseoverValue
                formatPreset={NicelyFormattedTypes[measure?.format] ||
                  NicelyFormattedTypes.HUMANIZE}
                data={formattedData}
                accessor={measure.name}
                mouseover={point}
                timeGrain={metricsExplorer.selectedTimeRange?.interval}
                yMin={yExtents[0] < 0 ? yExtents[0] : 0}
                start={startValue}
                end={endValue}
              />
            {/if}
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
