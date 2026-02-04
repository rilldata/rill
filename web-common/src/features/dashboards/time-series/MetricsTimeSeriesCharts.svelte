<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import DashboardMetricsDraggableList from "@rilldata/web-common/components/menu/DashboardMetricsDraggableList.svelte";
  import ReplacePivotDialog from "@rilldata/web-common/features/dashboards/pivot/ReplacePivotDialog.svelte";
  import { splitPivotChips } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
  import {
    PivotChipType,
    type PivotChipData,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    metricsExplorerStore,
    useExploreState,
  } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import ChartTypeSelector from "@rilldata/web-common/features/dashboards/time-dimension-details/charts/ChartTypeSelector.svelte";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import BackToExplore from "@rilldata/web-common/features/dashboards/time-series/BackToExplore.svelte";
  import { measureSelection } from "@rilldata/web-common/features/dashboards/time-series/measure-selection/measure-selection.ts";
  import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";
  import {
    TimeRangePreset,
    TimeComparisonOption,
    type AvailableTimeGrain,
    type DashboardTimeControls,
  } from "@rilldata/web-common/lib/time/types";
  import { type MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client/gen/index.schemas";
  import { Button } from "../../../components/button";
  import Pivot from "../../../components/icons/Pivot.svelte";
  import { TIME_GRAIN } from "../../../lib/time/config";
  import { DashboardState_ActivePage } from "../../../proto/gen/rill/ui/v1/dashboard_pb";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import Spinner from "../../entity-management/Spinner.svelte";
  import { featureFlags } from "../../feature-flags";
  import MeasureBigNumber from "../big-number/MeasureBigNumber.svelte";
  import MeasureChart from "./measure-chart/MeasureChart.svelte";
  import MeasureChartXAxis from "./measure-chart/MeasureChartXAxis.svelte";
  import { ScrubController } from "./measure-chart/ScrubController";
  import { getAnnotationsForMeasure } from "./annotations-selectors";
  import ChartInteractions from "./ChartInteractions.svelte";
  import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
  import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import { DateTime, Interval } from "luxon";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  const { rillTime } = featureFlags;

  // Singleton scrub controller — shared across all charts
  const scrubController = new ScrubController();

  export let exploreName: string;
  export let hideStartPivotButton = false;

  const StateManagers = getStateManagers();

  const {
    metricsViewName,
    dashboardStore,
    selectors: {
      measures: { allMeasures, visibleMeasures, getMeasureByName },
      dimensionFilters: { includedDimensionValues },
      charts: { canPanLeft, canPanRight, getNewPanRange },
    },
    actions: {
      measures: { setMeasureVisibility },
    },
  } = StateManagers;

  const timeControlsStore = useTimeControlStore(StateManagers);

  let grainDropdownOpen = false;

  $: ({ instanceId } = $runtime);

  $: ({
    selectedTimeRange,
    selectedComparisonTimeRange,
    timeDimension,
    ready,
    showTimeComparison,
    timeEnd,
    timeStart,
    comparisonTimeEnd,
    comparisonTimeStart,
    aggregationOptions,
  } = $timeControlsStore);

  $: ({ whereFilter, dimensionThresholdFilters, selectedTimezone } =
    $dashboardStore);

  // Use the full selected time range for chart data fetching (not modified by scrub)
  $: chartInterval =
    selectedTimeRange?.start && selectedTimeRange?.end
      ? (Interval.fromDateTimes(
          DateTime.fromJSDate(selectedTimeRange.start, {
            zone: selectedTimezone,
          }),
          DateTime.fromJSDate(selectedTimeRange.end, {
            zone: selectedTimezone,
          }),
        ) as Interval<true>)
      : undefined;
  $: chartComparisonInterval =
    selectedComparisonTimeRange?.start && selectedComparisonTimeRange?.end
      ? (Interval.fromDateTimes(
          DateTime.fromJSDate(selectedComparisonTimeRange.start, {
            zone: selectedTimezone,
          }),
          DateTime.fromJSDate(selectedComparisonTimeRange.end, {
            zone: selectedTimezone,
          }),
        ) as Interval<true>)
      : undefined;

  $: exploreState = useExploreState(exploreName);

  $: activePage = $exploreState?.activePage;
  $: showTimeDimensionDetail = Boolean(
    activePage === DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
  );
  $: expandedMeasureName = $exploreState?.tdd?.expandedMeasureName;

  $: comparisonDimension = $exploreState?.selectedComparisonDimension;
  $: showComparison = Boolean(showTimeComparison);
  $: tddChartType = $exploreState?.tdd?.chartType;

  $: activeTimeGrain = selectedTimeRange?.interval;

  $: chartScrubInterval = $exploreState?.lastDefinedScrubRange
    ? (Interval.fromDateTimes(
        DateTime.fromJSDate($exploreState.lastDefinedScrubRange.start, {
          zone: selectedTimezone,
        }),
        DateTime.fromJSDate($exploreState.lastDefinedScrubRange.end, {
          zone: selectedTimezone,
        }),
      ) as Interval<true>)
    : undefined;
  $: includedValuesForDimension = $includedDimensionValues(
    comparisonDimension as string,
  );
  $: chartDimensionValues = includedValuesForDimension.slice(
    0,
    showTimeDimensionDetail ? 11 : 7,
  ) as (string | null)[];

  $: expandedMeasure = $getMeasureByName(expandedMeasureName);
  let renderedMeasures: MetricsViewSpecMeasure[];
  $: {
    renderedMeasures =
      showTimeDimensionDetail && expandedMeasure
        ? [expandedMeasure]
        : $visibleMeasures;
  }

  $: visibleMeasureNames = $visibleMeasures
    .map(({ name }) => name)
    .filter(isDefined);
  $: allMeasureNames = $allMeasures.map(({ name }) => name).filter(isDefined);
  function isDefined(value: string | undefined): value is string {
    return value !== undefined;
  }

  $: chartMetricsViewName = $metricsViewName;
  $: chartWhere = sanitiseExpression(
    mergeDimensionAndMeasureFilters(whereFilter, dimensionThresholdFilters),
    undefined,
  );

  $: chartReady = !!ready;

  // Annotation stores per measure — keyed by measure name
  function getAnnotationsForMeasureStore(measureName: string) {
    return getAnnotationsForMeasure({
      instanceId,
      exploreName,
      measureName,
      selectedTimeRange,
      dashboardTimezone: selectedTimezone,
    });
  }

  // Pan handler
  function handlePan(direction: "left" | "right") {
    const panRange = $getNewPanRange(direction);
    if (!panRange || !activeTimeGrain) return;
    const { start, end } = panRange;
    const comparisonTimeRange = showComparison
      ? ({ name: TimeComparisonOption.CONTIGUOUS } as DashboardTimeControls)
      : undefined;
    metricsExplorerStore.selectTimeRange(
      exploreName,
      { name: TimeRangePreset.CUSTOM, start, end },
      activeTimeGrain,
      comparisonTimeRange,
      {},
    );
  }

  let showReplacePivotModal = false;
  function startPivotForTimeseries() {
    const pivot = $exploreState?.pivot;
    if (!pivot) return;
    const pivotColumns = splitPivotChips(pivot.columns);
    if (
      pivot.rows.length ||
      pivotColumns.measure.length ||
      pivotColumns.dimension.length
    ) {
      showReplacePivotModal = true;
    } else {
      createPivot();
    }
  }

  function getTimeDimension() {
    return {
      id: selectedTimeRange?.interval,
      title: TIME_GRAIN[activeTimeGrain as AvailableTimeGrain]?.label,
      type: PivotChipType.Time,
    } as PivotChipData;
  }

  function createPivot() {
    showReplacePivotModal = false;
    const measures = renderedMeasures
      .filter((m) => m.name !== undefined)
      .map((m) => ({
        id: m.name as string,
        title: m.displayName || (m.name as string),
        type: PivotChipType.Measure,
      }));
    metricsExplorerStore.createPivot(
      exploreName,
      [getTimeDimension()],
      measures,
    );
  }

  function handleScrub(range: {
    start: DateTime;
    end: DateTime;
    isScrubbing: boolean;
  }) {
    metricsExplorerStore.setSelectedScrubRange(exploreName, {
      start: range.start.toJSDate(),
      end: range.end.toJSDate(),
      isScrubbing: range.isScrubbing,
    });
  }

  function maybeClearMeasureSelection() {
    if (!measureSelection.isRangeSelection()) {
      measureSelection.clear();
    }
  }
</script>

<svelte:window on:click={maybeClearMeasureSelection} />

<div class="max-w-full h-fit flex flex-col max-h-full pr-2">
  <div
    class:mb-6={tddChartType !== TDDChart.DEFAULT}
    class="flex items-center gap-x-1 px-2.5"
  >
    {#if showTimeDimensionDetail}
      <BackToExplore />
      <ChartTypeSelector
        hasComparison={Boolean(
          showComparison || includedValuesForDimension.length,
        )}
        {exploreName}
        chartType={tddChartType}
      />
    {:else}
      <DashboardMetricsDraggableList
        type="measure"
        onSelectedChange={(items) =>
          setMeasureVisibility(items, allMeasureNames)}
        allItems={$allMeasures}
        selectedItems={visibleMeasureNames}
      />

      {#if $rillTime && activeTimeGrain}
        <DropdownMenu.Root bind:open={grainDropdownOpen}>
          <DropdownMenu.Trigger asChild let:builder>
            <button
              {...builder}
              use:builder.action
              aria-label="Select aggregation grain"
              class="flex gap-x-1 items-center text-fg-muted hover:text-fg-accent"
            >
              by <b>
                {V1TimeGrainToDateTimeUnit[activeTimeGrain]}
              </b>
              <span
                class:-rotate-90={grainDropdownOpen}
                class="transition-transform"
              >
                <CaretDownIcon />
              </span>
            </button>
          </DropdownMenu.Trigger>

          <DropdownMenu.Content align="start" class="w-48">
            {#each aggregationOptions ?? [] as option (option)}
              <DropdownMenu.CheckboxItem
                checkRight
                role="menuitem"
                checked={option === activeTimeGrain}
                class="text-xs cursor-pointer"
                on:click={() => {
                  metricsExplorerStore.setTimeGrain(exploreName, option);
                }}
              >
                {V1TimeGrainToDateTimeUnit[option]}
              </DropdownMenu.CheckboxItem>
            {/each}
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      {/if}

      {#if !hideStartPivotButton}
        <div class="grow" />
        <Button
          type="toolbar"
          onClick={() => {
            startPivotForTimeseries();
          }}
        >
          <Pivot size="16px" />
          Start Pivot
        </Button>
      {/if}
    {/if}
  </div>

  {#if renderedMeasures}
    <div
      class:pb-4={!showTimeDimensionDetail}
      class="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 overflow-y-scroll h-full max-h-fit"
    >
      {#if activeTimeGrain}
        <div
          class="sticky top-0 z-10 bg-surface-background col-span-2 grid grid-cols-subgrid"
        >
          <div />
          <div class="relative">
            <MeasureChartXAxis
              interval={chartInterval}
              timeGranularity={activeTimeGrain}
            />
            <ChartInteractions
              {exploreName}
              {showComparison}
              timeGrain={activeTimeGrain}
            />
          </div>
        </div>
      {/if}

      {#each renderedMeasures as measure (measure.name)}
        <MeasureBigNumber
          {measure}
          isMeasureExpanded={showTimeDimensionDetail}
          {showComparison}
          {instanceId}
          metricsViewName={chartMetricsViewName}
          where={chartWhere}
          {timeDimension}
          {timeStart}
          {timeEnd}
          {comparisonTimeStart}
          {comparisonTimeEnd}
          ready={chartReady}
        />

        {#if activeTimeGrain}
          <MeasureChart
            {measure}
            {scrubController}
            tddChartType={showTimeDimensionDetail
              ? (tddChartType ?? TDDChart.DEFAULT)
              : TDDChart.DEFAULT}
            {instanceId}
            metricsViewName={chartMetricsViewName}
            where={chartWhere}
            {timeDimension}
            interval={chartInterval}
            comparisonInterval={chartComparisonInterval}
            timeGranularity={activeTimeGrain}
            timeZone={selectedTimezone}
            ready={chartReady}
            {chartScrubInterval}
            {comparisonDimension}
            dimensionValues={chartDimensionValues}
            dimensionWhere={whereFilter}
            annotations={getAnnotationsForMeasureStore(measure.name ?? "")}
            canPanLeft={$canPanLeft}
            canPanRight={$canPanRight}
            onPanLeft={() => handlePan("left")}
            onPanRight={() => handlePan("right")}
            {showComparison}
            {showTimeDimensionDetail}
            onScrub={handleScrub}
            onScrubClear={() => {
              metricsExplorerStore.setSelectedScrubRange(
                exploreName,
                undefined,
              );
            }}
          />
        {:else}
          <div class="flex items-center justify-center w-24">
            <Spinner status={EntityStatus.Running} />
          </div>
        {/if}
      {/each}
    </div>
  {/if}
</div>

<ReplacePivotDialog
  open={showReplacePivotModal}
  onCancel={() => {
    showReplacePivotModal = false;
  }}
  onReplace={createPivot}
/>
