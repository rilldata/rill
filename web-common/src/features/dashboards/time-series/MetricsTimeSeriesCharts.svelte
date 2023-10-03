<script lang="ts">
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import SeachableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SeachableFilterButton.svelte";
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import {
    humanizeDataType,
    FormatPreset,
    nicelyFormattedTypesToNumberKind,
  } from "@rilldata/web-common/features/dashboards/humanize-numbers";
  import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors";
  import { createShowHideMeasuresStore } from "@rilldata/web-common/features/dashboards/show-hide-selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { getAdjustedChartTime } from "@rilldata/web-common/lib/time/ranges";
  import { createQueryServiceMetricsViewTotals } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import Spinner from "../../entity-management/Spinner.svelte";
  import MeasureBigNumber from "../big-number/MeasureBigNumber.svelte";
  import MeasureChart from "./MeasureChart.svelte";
  import MeasureZoom from "./MeasureZoom.svelte";
  import TimeSeriesChartContainer from "./TimeSeriesChartContainer.svelte";
  import BackToOverview from "@rilldata/web-common/features/dashboards/time-series/BackToOverview.svelte";
  import { useTimeSeriesDataStore } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";

  export let metricViewName;
  export let workspaceWidth: number;

  $: dashboardStore = useDashboardStore(metricViewName);

  $: instanceId = $runtime.instanceId;

  // query the `/meta` endpoint to get the measures and the default time grain
  $: metaQuery = useMetaQuery(instanceId, metricViewName);

  $: showHideMeasures = createShowHideMeasuresStore(metricViewName, metaQuery);

  const timeControlsStore = useTimeControlStore(getStateManagers());
  const timeSeriesDataStore = useTimeSeriesDataStore(getStateManagers());

  $: selectedMeasureNames = $dashboardStore?.selectedMeasureNames;
  $: expandedMeasureName = $dashboardStore?.expandedMeasureName;
  $: comparisonDimension = $dashboardStore?.selectedComparisonDimension;
  $: showComparison = !comparisonDimension && $timeControlsStore.showComparison;
  $: interval =
    $timeControlsStore.selectedTimeRange?.interval ??
    $timeControlsStore.minTimeGrain;

  // List of measures which will be shown on the dashboard
  let renderedMeasures = [];
  $: {
    if (expandedMeasureName) {
      renderedMeasures = $metaQuery.data?.measures.filter(
        (measure) => measure.name === expandedMeasureName
      );
    } else {
      renderedMeasures = $metaQuery.data?.measures.filter(
        (_, i) => $showHideMeasures.selectedItems[i]
      );
    }
  }

  $: renderedMeasureNames = renderedMeasures.map((measure) => measure.name);

  // List of measures which will be queried
  // In case we on expanded view, only query the for that measure
  $: queriedMeasureNames = expandedMeasureName
    ? renderedMeasureNames
    : selectedMeasureNames;

  $: totalsQuery = createQueryServiceMetricsViewTotals(
    $runtime.instanceId,
    metricViewName,
    {
      measureNames: queriedMeasureNames,
      timeStart: $timeControlsStore.timeStart,
      timeEnd: $timeControlsStore.timeEnd,
      filter: $dashboardStore?.filters,
    },
    {
      query: {
        enabled:
          queriedMeasureNames?.length > 0 &&
          $timeControlsStore.ready &&
          !!$dashboardStore?.filters,
      },
    }
  );

  $: totalsComparisonQuery = createQueryServiceMetricsViewTotals(
    instanceId,
    metricViewName,
    {
      measureNames: queriedMeasureNames,
      timeStart: $timeControlsStore.comparisonTimeStart,
      timeEnd: $timeControlsStore.comparisonTimeEnd,
      filter: $dashboardStore?.filters,
    },
    {
      query: {
        enabled: Boolean(showComparison && !!$dashboardStore?.filters),
      },
    }
  );

  // get the totalsComparisons.
  $: totalsComparisons = $totalsComparisonQuery?.data?.data;

  let scrubStart;
  let scrubEnd;

  let mouseoverValue = undefined;
  let startValue: Date;
  let endValue: Date;

  // When changing the timeseries query and the cache is empty, $timeSeriesQuery.data?.data is
  // temporarily undefined as results are fetched.
  // To avoid unmounting TimeSeriesBody, which would cause us to lose our tween animations,
  // we make a copy of the data that avoids `undefined` transition states.
  // TODO: instead, try using svelte-query's `keepPreviousData = True` option.

  let dataCopy;
  $: if ($timeSeriesDataStore?.timeSeriesData) {
    dataCopy = $timeSeriesDataStore.timeSeriesData;
  }
  $: formattedData = dataCopy;

  $: dimensionData = $timeSeriesDataStore?.dimensionChartData || [];

  // FIXME: move this logic to a function + write tests.
  $: if ($timeControlsStore.ready) {
    // adjust scrub values for Javascript's timezone changes
    scrubStart = adjustOffsetForZone(
      $dashboardStore?.selectedScrubRange?.start,
      $dashboardStore?.selectedTimezone
    );
    scrubEnd = adjustOffsetForZone(
      $dashboardStore?.selectedScrubRange?.end,
      $dashboardStore?.selectedTimezone
    );

    const adjustedChartValue = getAdjustedChartTime(
      $timeControlsStore.selectedTimeRange?.start,
      $timeControlsStore.selectedTimeRange?.end,
      $dashboardStore?.selectedTimezone,
      interval,
      $timeControlsStore.selectedTimeRange?.name
    );

    startValue = adjustedChartValue?.start;
    endValue = adjustedChartValue?.end;
  }

  const toggleMeasureVisibility = (e) => {
    showHideMeasures.toggleVisibility(e.detail.name);
  };
  const setAllMeasuresNotVisible = () => {
    showHideMeasures.setAllToNotVisible();
  };
  const setAllMeasuresVisible = () => {
    showHideMeasures.setAllToVisible();
  };
</script>

<TimeSeriesChartContainer
  enableFullWidth={Boolean(expandedMeasureName)}
  end={endValue}
  start={startValue}
  {workspaceWidth}
>
  <div class="bg-white sticky top-0 flex pl-1" style="z-index:100">
    {#if expandedMeasureName}
      <BackToOverview {metricViewName} />
    {:else}
      <SeachableFilterButton
        label="Measures"
        on:deselect-all={setAllMeasuresNotVisible}
        on:item-clicked={toggleMeasureVisibility}
        on:select-all={setAllMeasuresVisible}
        selectableItems={$showHideMeasures.selectableItems}
        selectedItems={$showHideMeasures.selectedItems}
        tooltipText="Choose measures to display"
      />
    {/if}
  </div>
  <div
    class="bg-white sticky left-0 top-0 overflow-visible"
    style="z-index:101"
  >
    <!-- top axis element -->
    <div />
    <MeasureZoom {metricViewName} />
    {#if $dashboardStore?.selectedTimeRange}
      <SimpleDataGraphic
        height={26}
        overflowHidden={false}
        top={29}
        bottom={0}
        xMin={startValue}
        xMax={endValue}
      >
        <Axis superlabel side="top" placement="start" />
      </SimpleDataGraphic>
    {/if}
  </div>
  <!-- bignumbers and line charts -->
  {#if renderedMeasures.length}
    <!-- FIXME: this is pending the remaining state work for show/hide measures and dimensions -->
    {#each renderedMeasures as measure (measure.name)}
      <!-- FIXME: I can't select the big number by the measure id. -->
      {@const bigNum = $totalsQuery?.data?.data?.[measure.name]}
      {@const comparisonValue = totalsComparisons?.[measure.name]}
      {@const comparisonPercChange =
        comparisonValue && bigNum !== undefined && bigNum !== null
          ? (bigNum - comparisonValue) / comparisonValue
          : undefined}
      {@const formatPreset =
        FormatPreset[measure?.format] || FormatPreset.HUMANIZE}
      <MeasureBigNumber
        on:expand-measure={() => {
          metricsExplorerStore.setExpandedMeasureName(
            metricViewName,
            measure.name
          );
        }}
        value={bigNum}
        {showComparison}
        comparisonOption={$timeControlsStore?.selectedComparisonTimeRange?.name}
        {comparisonValue}
        {comparisonPercChange}
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
        {#if $timeSeriesDataStore?.hasError}
          <div class="p-5"><CrossIcon /></div>
        {:else if formattedData}
          <MeasureChart
            bind:mouseoverValue
            isScrubbing={$dashboardStore?.selectedScrubRange?.isScrubbing}
            {scrubStart}
            {scrubEnd}
            {metricViewName}
            data={formattedData}
            {dimensionData}
            zone={$dashboardStore?.selectedTimezone}
            xAccessor="ts_position"
            labelAccessor="ts"
            timeGrain={interval}
            yAccessor={measure.name}
            xMin={startValue}
            xMax={endValue}
            {showComparison}
            mouseoverTimeFormat={(value) => {
              /** format the date according to the time grain */
              return new Date(value).toLocaleDateString(
                undefined,
                TIME_GRAIN[interval].formatDate
              );
            }}
            numberKind={nicelyFormattedTypesToNumberKind(measure?.format)}
            mouseoverFormat={(value) =>
              formatPreset === FormatPreset.NONE
                ? `${value}`
                : humanizeDataType(value, measure?.format)}
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
