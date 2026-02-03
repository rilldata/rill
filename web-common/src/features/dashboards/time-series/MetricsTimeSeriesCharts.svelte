<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import DashboardMetricsDraggableList from "@rilldata/web-common/components/menu/DashboardMetricsDraggableList.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
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
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import {
    getAllowedGrains,
    isGrainAllowed,
    V1TimeGrainToDateTimeUnit,
  } from "@rilldata/web-common/lib/time/new-grains";
  import {
    TimeRangePreset,
    TimeComparisonOption,
    type AvailableTimeGrain,
    type DashboardTimeControls,
  } from "@rilldata/web-common/lib/time/types";
  import { type MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client/gen/index.schemas";
  import { Tooltip } from "bits-ui";
  import { Button } from "../../../components/button";
  import Pivot from "../../../components/icons/Pivot.svelte";
  import { TIME_GRAIN } from "../../../lib/time/config";
  import { DashboardState_ActivePage } from "../../../proto/gen/rill/ui/v1/dashboard_pb";
  import Spinner from "../../entity-management/Spinner.svelte";
  import { featureFlags } from "../../feature-flags";
  import MeasureBigNumber from "../big-number/MeasureBigNumber.svelte";
  import { MeasureChart } from "./measure-chart";
  import { ScrubController } from "./measure-chart/ScrubController";
  import { getAnnotationsForMeasure } from "./annotations-selectors";
  import ChartInteractions from "./ChartInteractions.svelte";
  import { chartHoveredTime } from "../time-dimension-details/time-dimension-data-store";
  import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
  import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import { derived, writable } from "svelte/store";
  import { DateTime } from "luxon";
  import { tableInteractionStore } from "@rilldata/web-common/features/dashboards/time-dimension-details/time-dimension-data-store";

  // Derive a readable of the table-hovered time (Date | undefined)
  const tableHoverTime = derived(tableInteractionStore, ($s) => $s.time);

  const { rillTime } = featureFlags;

  // Shared hover index store — all MeasureChart instances read/write this
  const sharedHoverIndex = writable<number | undefined>(undefined);

  // Singleton scrub controller — shared across all charts
  const scrubController = new ScrubController();

  export let exploreName: string;
  export let hideStartPivotButton = false;

  const ctx = getStateManagers();
  const {
    selectors: {
      measures: { allMeasures, visibleMeasures, getMeasureByName },
      dimensionFilters: { includedDimensionValues },
      charts: { canPanLeft, canPanRight, getNewPanRange },
    },
    actions: {
      measures: { setMeasureVisibility },
    },
  } = ctx;

  const timeControlsStore = useTimeControlStore(ctx);

  $: ({
    selectedTimeRange,
    selectedComparisonTimeRange,
    timeDimension,
    ready,
    minTimeGrain,
    showTimeComparison,
    timeEnd,
    timeStart,
    comparisonTimeEnd,
    comparisonTimeStart,
  } = $timeControlsStore);

  // Use the full selected time range for chart data fetching (not modified by scrub)
  $: chartTimeStart = selectedTimeRange?.start?.toISOString();
  $: chartTimeEnd = selectedTimeRange?.end?.toISOString();
  $: chartComparisonTimeStart = selectedComparisonTimeRange?.start?.toISOString();
  $: chartComparisonTimeEnd = selectedComparisonTimeRange?.end?.toISOString();

  $: exploreState = useExploreState(exploreName);

  $: activePage = $exploreState?.activePage;
  $: showTimeDimensionDetail = Boolean(
    activePage === DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
  );
  $: expandedMeasureName = $exploreState?.tdd?.expandedMeasureName;

  $: comparisonDimension = $exploreState?.selectedComparisonDimension;
  $: showComparison = Boolean(showTimeComparison);
  $: tddChartType = $exploreState?.tdd?.chartType;

  $: activeTimeGrain = selectedTimeRange?.interval ?? minTimeGrain;

  $: chartScrubRange = $exploreState?.lastDefinedScrubRange
    ? {
        start: DateTime.fromJSDate($exploreState.lastDefinedScrubRange.start, {
          zone: chartTimeZone,
        }),
        end: DateTime.fromJSDate($exploreState.lastDefinedScrubRange.end, {
          zone: chartTimeZone,
        }),
      }
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

  $: timeGrainOptions = getAllowedGrains(minTimeGrain);
  $: grainAllowed = isGrainAllowed(activeTimeGrain, minTimeGrain);

  let grainDropdownOpen = false;
  $: effectiveGrain = grainAllowed ? activeTimeGrain : minTimeGrain;

  // Props for MeasureChart (context-independent)
  const runtimeStore = ctx.runtime;
  const metricsViewNameStore = ctx.metricsViewName;
  const dashboardStore = ctx.dashboardStore;
  $: instanceId = $runtimeStore.instanceId;
  $: chartMetricsViewName = $metricsViewNameStore;
  $: chartWhere = sanitiseExpression(
    mergeDimensionAndMeasureFilters(
      $dashboardStore.whereFilter,
      $dashboardStore.dimensionThresholdFilters,
    ),
    undefined,
  );


  $: chartTimeZone = $dashboardStore.selectedTimezone;
  $: chartReady = !!ready;

  // Annotation stores per measure — keyed by measure name
  function getAnnotationsForMeasureStore(measureName: string) {
    return getAnnotationsForMeasure({
      instanceId,
      exploreName,
      measureName,
      selectedTimeRange,
      dashboardTimezone: chartTimeZone,
    });
  }

  // Pan handler
  function handlePan(direction: "left" | "right") {
    const panRange = $getNewPanRange(direction);
    if (!panRange) return;
    const { start, end } = panRange;
    const comparisonTimeRange = showComparison
      ? ({ name: TimeComparisonOption.CONTIGUOUS } as DashboardTimeControls)
      : undefined;
    metricsExplorerStore.selectTimeRange(
      exploreName,
      { name: TimeRangePreset.CUSTOM, start, end },
      effectiveGrain!,
      comparisonTimeRange,
      {},
    );
  }

  let showReplacePivotModal = false;
  function startPivotForTimeseries() {
    const pivot = $exploreState?.pivot;
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

      {#if $rillTime && effectiveGrain}
        <DropdownMenu.Root bind:open={grainDropdownOpen}>
          <DropdownMenu.Trigger asChild let:builder>
            <button
              {...builder}
              use:builder.action
              class="flex gap-x-1 items-center text-fg-muted hover:text-fg-accent"
            >
              by <b>
                {V1TimeGrainToDateTimeUnit[effectiveGrain]}
              </b>
              <span
                class:-rotate-90={grainDropdownOpen}
                class="transition-transform"
              >
                <CaretDownIcon />
              </span>
              {#if !grainAllowed && minTimeGrain && activeTimeGrain}
                <Tooltip.Root portal="body">
                  <Tooltip.Trigger>
                    <AlertCircleOutline className="size-3.5 " />
                  </Tooltip.Trigger>
                  <Tooltip.Content side="top" class="z-50 w-64" sideOffset={8}>
                    <TooltipContent>
                      <i>{V1TimeGrainToDateTimeUnit[activeTimeGrain]}</i>
                      aggregation not supported on this dashboard. Displaying by
                      <i>{V1TimeGrainToDateTimeUnit[minTimeGrain]}</i> instead.
                    </TooltipContent>
                  </Tooltip.Content>
                </Tooltip.Root>
              {/if}
            </button>
          </DropdownMenu.Trigger>

          <DropdownMenu.Content align="start" class="w-48">
            {#each timeGrainOptions as option (option)}
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
    <div class="relative">
      <ChartInteractions
        {exploreName}
        {showComparison}
        timeGrain={effectiveGrain}
      />
    </div>
    <div
      class:pb-4={!showTimeDimensionDetail}
      class="flex flex-col gap-y-2 overflow-y-scroll h-full max-h-fit"
    >
      {#each renderedMeasures as measure (measure.name)}
        <div class="flex flex-row gap-x-4">
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

          {#if effectiveGrain}
            <MeasureChart
              {measure}
              {scrubController}
              {sharedHoverIndex}
              tddChartType={showTimeDimensionDetail
                ? (tddChartType ?? TDDChart.DEFAULT)
                : TDDChart.DEFAULT}
              {instanceId}
              metricsViewName={chartMetricsViewName}
              where={chartWhere}
              {timeDimension}
              timeStart={chartTimeStart}
              timeEnd={chartTimeEnd}
              comparisonTimeStart={chartComparisonTimeStart}
              comparisonTimeEnd={chartComparisonTimeEnd}
              timeGranularity={effectiveGrain}
              timeZone={chartTimeZone}
              ready={chartReady}
              scrubRange={chartScrubRange}
              {comparisonDimension}
              dimensionValues={chartDimensionValues}
              dimensionWhere={$dashboardStore.whereFilter}
              annotations={getAnnotationsForMeasureStore(measure.name ?? "")}
              canPanLeft={$canPanLeft}
              canPanRight={$canPanRight}
              onPanLeft={() => handlePan("left")}
              onPanRight={() => handlePan("right")}
              {showComparison}
              {showTimeDimensionDetail}
              {tableHoverTime}
              onHover={(dt) => {
                if (dt) {
                  // Convert to JS Date matching table's timezone handling:
                  // keepLocalTime: true preserves wall clock time when shifting to system zone
                  const systemTimeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;
                  chartHoveredTime.set(dt.setZone(systemTimeZone, { keepLocalTime: true }).toJSDate());
                } else {
                  chartHoveredTime.set(undefined);
                }
              }}
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
        </div>
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
