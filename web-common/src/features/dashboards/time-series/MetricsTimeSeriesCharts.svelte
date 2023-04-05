<script lang="ts">
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithBisector } from "@rilldata/web-common/components/data-graphic/functional-components";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import SeachableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SeachableFilterButton.svelte";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import {
    humanizeDataType,
    NicelyFormattedTypes,
    nicelyFormattedTypesToNumberKind,
  } from "@rilldata/web-common/features/dashboards/humanize-numbers";
  import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { removeTimezoneOffset } from "@rilldata/web-common/lib/formatters";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { getOffset } from "@rilldata/web-common/lib/time/transforms";
  import { TimeOffsetType } from "@rilldata/web-common/lib/time/types";
  import {
    useQueryServiceMetricsViewTimeSeries,
    useQueryServiceMetricsViewTotals,
    V1MetricsViewTimeSeriesResponse,
    V1MetricsViewTotalsResponse,
  } from "@rilldata/web-common/runtime-client";
  import { convertTimestampPreview } from "@rilldata/web-local/lib/util/convertTimestampPreview";
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";
  import Spinner from "../../entity-management/Spinner.svelte";
  import MeasureBigNumber from "../big-number/MeasureBigNumber.svelte";
  import MeasureChart from "./MeasureChart.svelte";
  import TimeSeriesChartContainer from "./TimeSeriesChartContainer.svelte";
  export let metricViewName;
  export let workspaceWidth: number;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  $: instanceId = $runtime.instanceId;

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

    totalsQuery = useQueryServiceMetricsViewTotals(
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
    timeSeriesQuery = useQueryServiceMetricsViewTimeSeries(
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
  $: if (dataCopy && dataCopy?.length) {
    formattedData = convertTimestampPreview(dataCopy, true).map((di, _i) => {
      di = { ts: di.ts, bin: di.bin, ...di.records };
      return di;
    });
  }

  let mouseoverValue = undefined;

  let startValue: Date;
  let endValue: Date;

  // FIXME: move this logic to a function + write tests.
  $: if (
    metricsExplorer?.selectedTimeRange &&
    metricsExplorer?.selectedTimeRange?.start
  ) {
    startValue = removeTimezoneOffset(
      new Date(metricsExplorer?.selectedTimeRange?.start)
    );

    // selectedTimeRange.end is exclusive and rounded to the time grain ("interval").
    // Since values are grouped with DATE_TRUNC, we subtract one grain to get the (inclusive) axis end.
    endValue = new Date(metricsExplorer?.selectedTimeRange?.end);

    endValue = getOffset(
      new Date(metricsExplorer?.selectedTimeRange?.end),
      TIME_GRAIN[metricsExplorer?.selectedTimeRange?.interval].duration,
      TimeOffsetType.SUBTRACT
    );

    endValue = removeTimezoneOffset(endValue);
  }

  let availableMeasureLabels = [];
  let visibleMeasures = [];

  $: availableMeasureLabels =
    $totalsQuery?.isSuccess && $metaQuery.data?.measures.map((m) => m.label);

  $: visibleMeasures = metricsExplorer.visibleMeasures;
  const toggleMeasureVisibility = (e) =>
    metricsExplorerStore.toggleMeasureVisibility(metricViewName, e.detail);
  const setAllMeasuresNotVisible = () =>
    metricsExplorerStore.setAllMeasuresVisibility(metricViewName, false);
  const setAllMeasuresVisible = () =>
    metricsExplorerStore.setAllMeasuresVisibility(metricViewName, true);
</script>

<WithBisector
  data={formattedData}
  callback={(datum) => datum.ts}
  value={mouseoverValue?.x}
  let:point
>
  <TimeSeriesChartContainer {workspaceWidth} start={startValue} end={endValue}>
    <div class="bg-white sticky left-0 top-0" style="z-index:100">
      <SeachableFilterButton
        selectableItems={availableMeasureLabels}
        selectedItems={visibleMeasures}
        on:itemClicked={toggleMeasureVisibility}
        on:deselectAll={setAllMeasuresNotVisible}
        on:selectAll={setAllMeasuresVisible}
        label="Measures"
        tooltipText="Choose measures to display"
      />
    </div>
    <div class="bg-white sticky left-0 top-0">
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
      {#each $metaQuery.data?.measures.filter((_, i) => visibleMeasures[i]) as measure, index (measure.name)}
        <!-- FIXME: I can't select the big number by the measure id. -->
        {@const bigNum = $totalsQuery?.data.data?.[measure.name]}
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
          {#if $timeSeriesQuery?.isError}
            <div class="p-5"><CrossIcon /></div>
          {:else if formattedData}
            <MeasureChart
              bind:mouseoverValue
              data={formattedData}
              xAccessor="ts"
              yAccessor={measure.name}
              xMin={startValue}
              xMax={endValue}
              mouseoverTimeFormat={(value) => {
                /** format the date according to the time grain */
                return new Date(value).toLocaleDateString(
                  undefined,
                  TIME_GRAIN[metricsExplorer?.selectedTimeRange?.interval]
                    .formatDate
                );
              }}
              numberKind={nicelyFormattedTypesToNumberKind(measure?.format)}
              mouseoverFormat={(value) =>
                formatPreset === NicelyFormattedTypes.NONE
                  ? `${value}`
                  : humanizeDataType(value, measure?.format, {
                      excludeDecimalZeros: true,
                    })}
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
