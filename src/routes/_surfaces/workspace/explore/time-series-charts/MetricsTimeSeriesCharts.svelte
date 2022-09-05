<script lang="ts">
  import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type { TimeSeriesValue } from "$common/database-service/DatabaseTimeSeriesActions";
  import type { MetricsViewTimeSeriesResponse } from "$common/rill-developer-service/MetricsViewActions";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "$lib/application-state-stores/explorer-stores";
  import SimpleDataGraphic from "$lib/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithBisector } from "$lib/components/data-graphic/functional-components";
  import { Axis } from "$lib/components/data-graphic/guides";
  import CrossIcon from "$lib/components/icons/CrossIcon.svelte";
  import Spinner from "$lib/components/Spinner.svelte";
  import {
    useMetaQuery,
    useTimeSeriesQuery,
    useTotalsQuery,
  } from "$lib/svelte-query/queries/metrics-view";
  import { convertTimestampPreview } from "$lib/util/convertTimestampPreview";
  import { removeTimezoneOffset } from "$lib/util/formatters";
  import { NicelyFormattedTypes } from "$lib/util/humanize-numbers";
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import { extent } from "d3-array";
  import { fly } from "svelte/transition";
  import { formatDateByInterval } from "../time-controls/time-range-utils";
  import MeasureBigNumber from "./MeasureBigNumber.svelte";
  import TimeSeriesBody from "./TimeSeriesBody.svelte";
  import TimeSeriesChartContainer from "./TimeSeriesChartContainer.svelte";

  export let metricsDefId;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricsDefId];

  // query the `/meta` endpoint to get the measures and the default time grain
  $: metaQuery = useMetaQuery(metricsDefId);

  $: interval =
    metricsExplorer?.selectedTimeRange?.interval ||
    $metaQuery.data?.timeDimension?.timeRange?.interval;

  let totalsQuery;
  $: if (
    metricsExplorer &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching
  ) {
    totalsQuery = useTotalsQuery(metricsDefId, {
      measures: metricsExplorer?.selectedMeasureIds,
      filter: metricsExplorer?.filters,
      time: {
        start: metricsExplorer?.selectedTimeRange?.start,
        end: metricsExplorer?.selectedTimeRange?.end,
      },
    });
  }

  let timeSeriesQuery: UseQueryStoreResult<
    MetricsViewTimeSeriesResponse,
    Error
  >;
  $: if (
    metricsExplorer &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching
  ) {
    timeSeriesQuery = useTimeSeriesQuery(metricsDefId, {
      measures: metricsExplorer?.selectedMeasureIds,
      filter: metricsExplorer?.filters,
      time: {
        start: metricsExplorer?.selectedTimeRange?.start,
        end: metricsExplorer?.selectedTimeRange?.end,
        granularity: metricsExplorer?.selectedTimeRange?.interval,
      },
    });
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
  $: if (dataCopy) formattedData = convertTimestampPreview(dataCopy, true);

  let mouseoverValue = undefined;

  $: key = `${startValue}` + `${endValue}`;

  $: [minVal, maxVal] = extent(dataCopy ?? [], (d: TimeSeriesValue) => d.ts);
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
    {#if $metaQuery.data?.measures && $totalsQuery?.isSuccess}
      {#each $metaQuery.data?.measures as measure, index (measure.id)}
        <!-- FIXME: I can't select the big number by the measure id. -->
        {@const bigNum = $totalsQuery?.data.data?.[measure.sqlName]}
        {@const yExtents = extent(dataCopy ?? [], (d) => d[`measure_${index}`])}

        <!-- FIXME: I can't select a time series by measure id. -->
        <MeasureBigNumber
          value={bigNum}
          description={measure?.description ||
            measure?.label ||
            measure?.expression}
          formatPreset={measure?.formatPreset || NicelyFormattedTypes.HUMANIZE}
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
            <TimeSeriesBody
              bind:mouseoverValue
              formatPreset={measure?.formatPreset ||
                NicelyFormattedTypes.HUMANIZE}
              data={formattedData}
              accessor={measure.sqlName}
              mouseover={point}
              {key}
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
