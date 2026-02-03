<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { writable, type Readable, type Writable } from "svelte/store";
  import type {
    TimeSeriesPoint,
    DimensionSeriesData,
    ChartScales,
    PlotBounds,
    ChartSeries,
    ChartMode,
  } from "./types";
  import {
    computeChartConfig,
    computeYExtent,
    computeNiceYExtent,
  } from "./scales";
  import {
    createChartInteractions,
    createVisibilityObserver,
  } from "./interactions";
  import TimeSeriesChart from "@rilldata/web-common/components/time-series-chart/TimeSeriesChart.svelte";
  import BarChart from "@rilldata/web-common/components/time-series-chart/BarChart.svelte";
  import MeasureChartTooltip from "./MeasureChartTooltip.svelte";
  import MeasureChartScrub from "./MeasureChartScrub.svelte";
  import MeasurePan from "./MeasurePan.svelte";
  import ExplainButton from "./ExplainButton.svelte";
  import MeasureChartPointIndicator from "./MeasureChartPointIndicator.svelte";
  import TDDAlternateChart from "@rilldata/web-common/features/dashboards/time-dimension-details/charts/TDDAlternateChart.svelte";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import { adjustTimeInterval } from "../utils";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
  import {
    createQueryServiceMetricsViewTimeSeries,
    type V1Expression,
  } from "@rilldata/web-common/runtime-client";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import { keepPreviousData } from "@tanstack/svelte-query";
  import { scaleLinear } from "d3-scale";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import {
    MainLineColor,
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
    TimeComparisonLineColor,
  } from "../chart-colors";
  import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import type { AvailableTimeGrain } from "@rilldata/web-common/lib/time/types";
  import { DateTime } from "luxon";
  import { transformTimeSeriesData } from "./use-measure-time-series";
  import {
    createDimensionAggregationQuery,
    buildDimensionSeriesData,
  } from "./use-dimension-data";
  import type { Annotation } from "@rilldata/web-common/components/data-graphic/marks/annotations";
  import {
    groupAnnotations,
    findHoveredGroup,
    type AnnotationGroup,
  } from "./annotation-utils";
  import MeasureChartAnnotationMarkers from "./MeasureChartAnnotationMarkers.svelte";
  import MeasureChartAnnotationPopover from "./MeasureChartAnnotationPopover.svelte";
  import { measureSelection } from "../measure-selection/measure-selection";

  const Y_DASH_ARRAY = "1,1";
  const X_PAD = 8;
  const { visible, observe } = createVisibilityObserver("120px");
  const selMeasure = measureSelection.measure;
  const selStart = measureSelection.start;
  const selEnd = measureSelection.end;

  export let measure: MetricsViewSpecMeasure;
  export let instanceId: string;
  export let metricsViewName: string;
  export let where: V1Expression | undefined = undefined;
  export let timeDimension: string | undefined = undefined;
  export let adjustedTimeStart: string | undefined = undefined;
  export let adjustedTimeEnd: string | undefined = undefined;
  export let primaryTimeStart: DateTime | undefined = undefined;
  export let primaryTimeEnd: DateTime | undefined = undefined;
  export let comparisonTimeStart: string | undefined = undefined;
  export let comparisonTimeEnd: string | undefined = undefined;
  export let timeGranularity: V1TimeGrain | undefined = undefined;
  export let timeZone: string = "UTC";
  export let comparisonDimension: string | undefined = undefined;
  export let dimensionValues: (string | null)[] = [];
  export let dimensionWhere: V1Expression | undefined = undefined;
  export let annotations: Readable<Annotation[]> | undefined = undefined;
  export let showComparison: boolean = false;
  export let showTimeDimensionDetail: boolean = false;
  export let ready: boolean = true;
  export let sharedHoverIndex: Writable<number | undefined> =
    writable(undefined);
  export let tableHoverTime: Readable<Date | undefined> = writable(
    undefined,
  ) as Readable<Date | undefined>;
  export let scrubRange: { start: DateTime; end: DateTime } | undefined =
    undefined;
  export let canPanLeft: boolean = false;
  export let canPanRight: boolean = false;
  export let tddChartType: TDDChart = TDDChart.DEFAULT;
  export let onScrub:
    | ((range: {
        start: DateTime;
        end: DateTime;
        isScrubbing: boolean;
      }) => void)
    | undefined = undefined;
  export let onScrubClear: (() => void) | undefined = undefined;
  export let onHover: ((dt: DateTime | undefined) => void) | undefined =
    undefined;
  export let onPanLeft: (() => void) | undefined = undefined;
  export let onPanRight: (() => void) | undefined = undefined;

  let container: HTMLDivElement;
  let clientWidth = 425;
  let unobserve: (() => void) | undefined;
  let tddIsScrubbing = false;
  let prevScrubbing = false;
  let hoveredAnnotationGroup: AnnotationGroup | null = null;
  let annotationPopoverHovered = false;
  let mouseDownX: number | null = null;
  let mouseDownY: number | null = null;
  let suppressScrubClear = false;

  const visibleRangeStore = writable<[number, number]>([0, 0]);
  const scalesStore = writable<ChartScales>({
    x: scaleLinear(),
    y: scaleLinear(),
  });
  const plotBoundsStore = writable<PlotBounds>({
    left: 0,
    top: 0,
    right: 0,
    bottom: 0,
    width: 0,
    height: 0,
  });

  onMount(() => {
    if (container) unobserve = observe(container);
  });
  onDestroy(() => unobserve?.());

  $: measureName = measure.name ?? "";
  $: height = showTimeDimensionDetail ? 245 : 145;
  $: config = computeChartConfig(clientWidth, height, showTimeDimensionDetail);

  // Time series queries
  $: timeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
    instanceId,
    metricsViewName,
    {
      measureNames: [measureName],
      where,
      timeDimension,
      timeStart: adjustedTimeStart,
      timeEnd: adjustedTimeEnd,
      timeGranularity,
      timeZone,
    },
    {
      query: {
        enabled: $visible && ready && !!adjustedTimeStart,
        placeholderData: keepPreviousData,
        refetchOnMount: false,
      },
    },
  );

  $: comparisonTimeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
    instanceId,
    metricsViewName,
    {
      measureNames: [measureName],
      where,
      timeDimension,
      timeStart: comparisonTimeStart,
      timeEnd: comparisonTimeEnd,
      timeGranularity,
      timeZone,
    },
    {
      query: {
        enabled: $visible && ready && showComparison && !!comparisonTimeStart,
        placeholderData: keepPreviousData,
        refetchOnMount: false,
      },
    },
  );

  // Transform query results
  $: comparisonData =
    showComparison && !$comparisonTimeSeriesQuery.isFetching
      ? $comparisonTimeSeriesQuery.data?.data
      : undefined;

  $: data =
    $timeSeriesQuery.isFetching || !$timeSeriesQuery.data?.data
      ? ([] as TimeSeriesPoint[])
      : transformTimeSeriesData(
          $timeSeriesQuery.data.data,
          comparisonData,
          measureName,
          timeZone,
        );

  $: isError = $timeSeriesQuery.isError;
  $: error = ($timeSeriesQuery.error as HTTPError)?.response?.data?.message;

  // Dimension comparison data
  $: hasDimensionComparison =
    !!comparisonDimension && dimensionValues.length > 0 && !!timeDimension;

  $: dimAggQuery = hasDimensionComparison
    ? createDimensionAggregationQuery(
        instanceId,
        metricsViewName,
        measureName,
        comparisonDimension!,
        dimensionValues,
        dimensionWhere,
        timeDimension!,
        adjustedTimeStart,
        adjustedTimeEnd,
        timeGranularity!,
        timeZone,
        $visible && ready && !!adjustedTimeStart,
      )
    : undefined;

  $: dimCompAggQuery =
    hasDimensionComparison && showComparison && !!comparisonTimeStart
      ? createDimensionAggregationQuery(
          instanceId,
          metricsViewName,
          measureName,
          comparisonDimension!,
          dimensionValues,
          dimensionWhere,
          timeDimension!,
          comparisonTimeStart,
          comparisonTimeEnd,
          timeGranularity!,
          timeZone,

          $visible && ready && !!comparisonTimeStart,
        )
      : undefined;

  $: dimIsFetching =
    (dimAggQuery ? $dimAggQuery?.isFetching : false) ||
    (dimCompAggQuery ? $dimCompAggQuery?.isFetching : false);

  $: dimensionData =
    hasDimensionComparison && timeDimension && timeGranularity
      ? buildDimensionSeriesData(
          measureName,
          comparisonDimension!,
          dimensionValues,
          timeDimension,
          timeGranularity,
          timeZone,
          $timeSeriesQuery.data?.data,
          dimAggQuery ? $dimAggQuery?.data?.data : undefined,
          showComparison ? $comparisonTimeSeriesQuery.data?.data : undefined,
          dimCompAggQuery ? $dimCompAggQuery?.data?.data : undefined,
          !!dimIsFetching,
        )
      : [];

  $: isFetching =
    $timeSeriesQuery.isFetching ||
    (showComparison && $comparisonTimeSeriesQuery.isFetching) ||
    !!dimIsFetching;

  // Chart series & mode
  $: chartSeries = buildChartSeries(data, dimensionData, showComparison);

  $: {
    if (mode === "bar" && chartSeries.length === 2) {
      chartSeries[1] = { ...chartSeries[1], opacity: 0.4 };
    }
  }

  $: mode = determineMode(data, dimensionData) as ChartMode;

  // Y extent & scales
  $: yRawExtent = computeYExtent(data, dimensionData, showComparison);
  $: [yMin, yMax] = computeNiceYExtent(yRawExtent[0], yRawExtent[1]);

  // Visible range
  $: visibleStart = computeVisibleStart(data, primaryTimeStart);
  $: visibleEnd = computeVisibleEnd(data, primaryTimeEnd);

  // X/Y scales
  $: barSlotWidth =
    config.plotBounds.width / Math.max(1, visibleEnd - visibleStart + 1);
  $: xRangeStart =
    mode === "line"
      ? config.plotBounds.left + X_PAD
      : config.plotBounds.left + barSlotWidth / 2;
  $: xRangeEnd =
    mode === "line"
      ? config.plotBounds.left + config.plotBounds.width - X_PAD
      : config.plotBounds.left + config.plotBounds.width - barSlotWidth / 2;
  $: xScale = scaleLinear<number>()
    .domain([visibleStart, visibleEnd])
    .range([xRangeStart, xRangeEnd]);
  $: yScale = scaleLinear<number>()
    .domain([yMin, yMax])
    .range([
      config.plotBounds.top + config.plotBounds.height,
      config.plotBounds.top,
    ]);
  $: scales = { x: xScale, y: yScale };
  $: yTicks = yScale.ticks(3);

  // Sync stores for interactions
  $: visibleRangeStore.set([visibleStart, visibleEnd]);
  $: scalesStore.set(scales);
  $: plotBoundsStore.set(config.plotBounds);

  $: ({
    state: interactionState,
    handlers,
    resetScrub,
  } = createChartInteractions(scalesStore, visibleRangeStore, plotBoundsStore));

  // Scrub state
  $: localScrubStartIndex = $interactionState.scrub.startIndex;
  $: localScrubEndIndex = $interactionState.scrub.endIndex;
  $: isScrubbing = $interactionState.scrub.isScrubbing;
  $: externalScrubStartIndex = scrubRange
    ? dateToIndex(data, scrubRange.start)
    : null;
  $: externalScrubEndIndex = scrubRange
    ? dateToIndex(data, scrubRange.end)
    : null;
  $: scrubStartIndex = localScrubStartIndex ?? externalScrubStartIndex;
  $: scrubEndIndex = localScrubEndIndex ?? externalScrubEndIndex;
  $: hasScrubSelection = scrubStartIndex !== null && scrubEndIndex !== null;

  $: if (!scrubRange) {
    resetScrub();
  }

  // Scrub end sync: fire onScrub only when scrubbing ends
  $: {
    const scrub = $interactionState.scrub;
    if (prevScrubbing && !scrub.isScrubbing) {
      const newStart = indexToDateTime(scrub.startIndex);
      const newEnd = indexToDateTime(scrub.endIndex);
      if (newStart && newEnd) {
        onScrub?.({ start: newStart, end: newEnd, isScrubbing: false });
        if (measureName) {
          const s = newStart.toJSDate();
          const e = newEnd.toJSDate();
          const [start, end] = s < e ? [s, e] : [e, s];
          measureSelection.setRange(measureName, start, end);
        }
      } else if (!suppressScrubClear) {
        onScrubClear?.();
      }
      suppressScrubClear = false;
    }
    prevScrubbing = scrub.isScrubbing;
  }

  // Hover state
  $: localHoveredIndex = $interactionState.bisectedPoint.index;
  $: isLocallyHovered =
    $interactionState.hover.isHovered && localHoveredIndex >= 0;
  $: cursorStyle = $interactionState.cursorStyle;
  $: sharedHoverIndex.set(isLocallyHovered ? localHoveredIndex : undefined);
  $: onHover?.(
    isLocallyHovered
      ? (indexToDateTime(localHoveredIndex) ?? undefined)
      : undefined,
  );
  $: tableHoverIndex = $tableHoverTime
    ? dateToIndex(
        data,
        DateTime.fromJSDate($tableHoverTime, { zone: timeZone }),
      )
    : null;
  $: hoveredIndex = isLocallyHovered
    ? localHoveredIndex
    : ($sharedHoverIndex ?? tableHoverIndex ?? -1);
  $: hoveredPoint = data[hoveredIndex] ?? null;

  // Formatters
  $: measureFormatter = createMeasureValueFormatter(measure);
  $: valueFormatter = (value: number | null): string => {
    if (value === null) return "\u2013";
    return measureFormatter(value);
  };
  $: axisFormatter = createMeasureValueFormatter(measure, "axis");

  // Annotations
  $: annotationGroups = annotations
    ? groupAnnotations(
        $annotations ?? [],
        scales,
        data,
        config,
        timeGranularity,
        timeZone,
      )
    : [];

  // Dimension tooltip
  $: isComparingDimension = dimensionData.length > 0;
  $: dimTooltipEntries =
    isComparingDimension && hoveredIndex >= 0
      ? dimensionData
          .map((dim) => ({
            label: dim.dimensionValue ?? "null",
            value: dim.data[hoveredIndex]?.value ?? null,
            color: dim.color,
          }))
          .filter((e) => e.value !== null)
          .sort((a, b) => (b.value ?? 0) - (a.value ?? 0))
      : [];

  // Explain CTA positioning
  $: isThisMeasureSelected = $selMeasure === measureName;
  $: singleSelectIdx =
    isThisMeasureSelected && $selStart && !$selEnd
      ? dateToIndex(data, DateTime.fromJSDate($selStart, { zone: timeZone }))
      : null;
  $: singleSelectX =
    singleSelectIdx !== null ? scales.x(singleSelectIdx) : null;
  $: explainX = (() => {
    if (!isThisMeasureSelected) return null;
    if (
      $selEnd &&
      hasScrubSelection &&
      scrubStartIndex !== null &&
      scrubEndIndex !== null
    ) {
      return Math.round(
        (scales.x(scrubStartIndex) + scales.x(scrubEndIndex)) / 2,
      );
    }
    if (singleSelectX !== null) return singleSelectX;
    return null;
  })();

  function buildChartSeries(
    d: TimeSeriesPoint[],
    dimData: DimensionSeriesData[],
    comparison: boolean,
  ): ChartSeries[] {
    if (dimData.length > 0) {
      return dimData.map((dim, i) => ({
        id: `dim-${dim.dimensionValue ?? i}`,
        values: dim.data.map((pt) => pt.value),
        color: dim.color || COMPARIONS_COLORS[i % COMPARIONS_COLORS.length],
        opacity: dim.isFetching ? 0.5 : 1,
      }));
    }

    const result: ChartSeries[] = [];

    if (d.length > 0) {
      result.push({
        id: "primary",
        values: d.map((pt) => pt.value),
        color: MainLineColor,
        areaGradient: {
          dark: MainAreaColorGradientDark,
          light: MainAreaColorGradientLight,
        },
      });
    }

    if (comparison && d.length > 0) {
      result.push({
        id: "comparison",
        values: d.map((pt) => pt.comparisonValue ?? null),
        color: TimeComparisonLineColor,
      });
    }

    return result;
  }

  function determineMode(
    d: TimeSeriesPoint[],
    dimData: DimensionSeriesData[],
  ): ChartMode {
    if (d.length >= 14) return "line";
    if (dimData.length > 0) return "stacked-bar";
    return "bar";
  }

  function computeVisibleStart(
    d: TimeSeriesPoint[],
    start: DateTime | undefined,
  ): number {
    if (!start || d.length === 0) return 0;
    const ms = start.toMillis();
    for (let i = 0; i < d.length; i++) {
      if (d[i].ts.toMillis() >= ms) return i;
    }
    return 0;
  }

  function computeVisibleEnd(
    d: TimeSeriesPoint[],
    end: DateTime | undefined,
  ): number {
    if (!end || d.length === 0) return Math.max(0, d.length - 1);
    const ms = end.toMillis();
    for (let i = d.length - 1; i >= 0; i--) {
      if (d[i].ts.toMillis() < ms) return i;
    }
    return Math.max(0, d.length - 1);
  }

  function indexToDateTime(idx: number | null): DateTime | null {
    if (idx === null || data.length === 0) return null;
    const snapped = Math.max(0, Math.min(data.length - 1, Math.round(idx)));
    const dt = data[snapped]?.ts;
    return dt?.isValid ? dt : null;
  }

  function dateToIndex(d: TimeSeriesPoint[], dt: DateTime): number | null {
    if (d.length === 0 || !dt.isValid) return null;
    const ms = dt.toMillis();
    let best = 0;
    let bestDist = Infinity;
    for (let i = 0; i < d.length; i++) {
      const dist = Math.abs(d[i].ts.toMillis() - ms);
      if (dist < bestDist) {
        bestDist = dist;
        best = i;
      }
    }
    return best;
  }

  function formatTime(dt: DateTime): string {
    if (!timeGranularity) return dt.toLocaleString(DateTime.DATE_SHORT);
    const grainConfig = TIME_GRAIN[timeGranularity as AvailableTimeGrain];
    if (!grainConfig?.formatDate) return dt.toLocaleString(DateTime.DATE_SHORT);
    return dt.toJSDate().toLocaleDateString(undefined, grainConfig.formatDate);
  }

  function formatScrubLabel(idx: number): string {
    const dt = indexToDateTime(idx);
    if (!dt) return "";
    return formatTime(dt);
  }

  function handleReset() {
    onScrubClear?.();
    resetScrub();
  }

  function handleTddHover(
    _dimension: undefined | string | null,
    ts: Date | undefined,
  ) {
    onHover?.(ts ? DateTime.fromJSDate(ts, { zone: timeZone }) : undefined);
  }

  function handleTddBrush(_interval: { start: Date; end: Date }) {
    tddIsScrubbing = true;
  }

  function handleTddBrushEnd(interval: { start: Date; end: Date }) {
    tddIsScrubbing = false;
    const { start, end } = adjustTimeInterval(interval, timeZone);
    onScrub?.({
      start: DateTime.fromJSDate(start, { zone: timeZone }),
      end: DateTime.fromJSDate(end, { zone: timeZone }),
      isScrubbing: false,
    });
  }

  function handleTddBrushClear() {
    tddIsScrubbing = false;
    onScrubClear?.();
  }

  function checkAnnotationHover(e: MouseEvent) {
    if (isScrubbing || annotationGroups.length === 0) {
      hoveredAnnotationGroup = null;
      return;
    }
    const svg = e.currentTarget as SVGSVGElement;
    const rect = svg.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;
    hoveredAnnotationGroup = findHoveredGroup(annotationGroups, mouseX, mouseY);
  }

  function handleAnnotationPopoverHover(hovered: boolean) {
    annotationPopoverHovered = hovered;
    if (!hovered) {
      setTimeout(() => {
        if (!annotationPopoverHovered) {
          hoveredAnnotationGroup = null;
        }
      }, 150);
    }
  }

  function handleSvgMouseLeave() {
    handlers.onMouseLeave();
    if (!annotationPopoverHovered) {
      hoveredAnnotationGroup = null;
    }
  }

  function handleMouseDown(e: MouseEvent) {
    mouseDownX = e.clientX;
    mouseDownY = e.clientY;
    handlers.onMouseDown(e);
  }

  function handleMouseUp(e: MouseEvent) {
    const wasClick =
      mouseDownX !== null &&
      mouseDownY !== null &&
      Math.abs(e.clientX - mouseDownX) < 4 &&
      Math.abs(e.clientY - mouseDownY) < 4;

    if (wasClick) suppressScrubClear = true;

    handlers.onMouseUp(e);
    mouseDownX = null;
    mouseDownY = null;

    if (!wasClick || !measureName) return;

    const scrub = $interactionState.scrub;
    const hasSubrangeSelected =
      scrub.startIndex !== null && scrub.endIndex !== null;

    if (!hasSubrangeSelected) {
      if (hoveredIndex < 0) return;
      const pt = data[hoveredIndex];
      if (pt?.ts?.isValid) {
        resetScrub();
        onScrubClear?.();
        measureSelection.setStart(measureName, pt.ts.toJSDate());
      }
    }
  }

  function handleChartClick(e: MouseEvent) {
    e.stopPropagation();
    e.preventDefault();

    const scrub = $interactionState.scrub;
    if (
      scrub.startIndex !== null &&
      scrub.endIndex !== null &&
      !scrub.isScrubbing
    ) {
      handlers.onClick(e);
      const updatedScrub = $interactionState.scrub;
      if (updatedScrub.startIndex === null) {
        measureSelection.clear();
      } else if (measureName) {
        const startDt = indexToDateTime(updatedScrub.startIndex);
        const endDt = indexToDateTime(updatedScrub.endIndex);
        if (startDt && endDt) {
          const s = startDt.toJSDate();
          const e2 = endDt.toJSDate();
          const [start, end] = s < e2 ? [s, e2] : [e2, s];
          measureSelection.setRange(measureName, start, end);
        }
      }
    }
  }
</script>

<div
  bind:clientWidth
  bind:this={container}
  class="measure-chart {cursorStyle} select-none size-full"
>
  {#if !$visible || (isFetching && data.length === 0)}
    <div
      class="flex items-center justify-center h-[145px]"
      class:h-[245px]={showTimeDimensionDetail}
    >
      <Spinner status={EntityStatus.Running} size="24px" />
    </div>
  {:else if isError}
    <div
      class="flex items-center justify-center text-red-500 text-xs h-[145px]"
      class:h-[245px]={showTimeDimensionDetail}
    >
      {error ?? "Error loading data"}
    </div>
  {:else if tddChartType !== TDDChart.DEFAULT && data.length > 0}
    <div class="w-full" style:height="{height}px">
      <TDDAlternateChart
        chartType={tddChartType}
        expandedMeasureName={measureName}
        timeSeriesPoints={data}
        dimensionSeriesData={dimensionData}
        timeGrain={timeGranularity}
        xMin={undefined}
        xMax={undefined}
        isTimeComparison={showComparison}
        isScrubbing={tddIsScrubbing}
        onChartHover={handleTddHover}
        onChartBrush={handleTddBrush}
        onChartBrushEnd={handleTddBrushEnd}
        onChartBrushClear={handleTddBrushClear}
      />
    </div>
  {:else if data.length > 0}
    <svg
      role="presentation"
      class="size w-full overflow-visible"
      height="{height}px"
      on:mousemove={(e) => {
        handlers.onMouseMove(e);
        checkAnnotationHover(e);
      }}
      on:mouseleave={handleSvgMouseLeave}
      on:mousedown={handleMouseDown}
      on:mouseup={handleMouseUp}
      on:click={handleChartClick}
    >
      <!-- Clip chart body to plot area so lines/bars don't bleed into margins when overplotting -->
      <defs>
        <clipPath id="chart-body-{measure.name}">
          <rect
            x={config.plotBounds.left}
            y={config.plotBounds.top}
            width={config.plotBounds.width}
            height={config.plotBounds.height}
          />
        </clipPath>
      </defs>

      <g class="y-axis">
        {#each yTicks as tick (tick)}
          <text
            class="fill-fg-muted text-[11px]"
            text-anchor="start"
            x={config.plotBounds.left + config.plotBounds.width + 4}
            y={yScale(tick) + 4}
          >
            {axisFormatter(tick)}
          </text>
          <line
            class="stroke-gray-300"
            x1={config.plotBounds.left}
            x2={config.plotBounds.left + config.plotBounds.width}
            y1={yScale(tick)}
            y2={yScale(tick)}
            stroke-width="0.5"
            stroke-dasharray={Y_DASH_ARRAY}
          />
        {/each}
      </g>

      <!-- Zero line -->
      <line
        class="stroke-gray-300"
        x1={config.plotBounds.left}
        x2={config.plotBounds.left + config.plotBounds.width}
        y1={yScale(0)}
        y2={yScale(0)}
      />

      <!-- Chart body -->
      <g clip-path="url(#chart-body-{measure.name})">
        {#if mode === "line"}
          <TimeSeriesChart
            series={chartSeries}
            {scales}
            {hasScrubSelection}
            {scrubStartIndex}
            {scrubEndIndex}
          />
        {:else}
          <BarChart
            series={chartSeries}
            yScale={scales.y}
            stacked={mode === "stacked-bar"}
            plotLeft={config.plotBounds.left}
            plotWidth={config.plotBounds.width}
            {visibleStart}
            {visibleEnd}
            {scrubStartIndex}
            {scrubEndIndex}
          />
        {/if}
      </g>

      {#if !isScrubbing && hoveredPoint}
        <MeasureChartTooltip
          {scales}
          {config}
          {hoveredIndex}
          {hoveredPoint}
          timeGrain={timeGranularity}
          {dimensionData}
          {showComparison}
          isComparingDimension={dimensionData.length > 0}
          formatter={valueFormatter}
        />
      {/if}

      <!-- Single-point selection indicator -->
      {#if singleSelectIdx !== null && singleSelectX !== null && isThisMeasureSelected}
        {@const selPt = data[singleSelectIdx]}
        {#if selPt?.value !== null && selPt?.value !== undefined}
          <MeasureChartPointIndicator
            x={singleSelectX}
            y={scales.y(selPt.value)}
            zeroY={scales.y(0)}
            value={valueFormatter(selPt.value)}
            selected
          />
        {/if}
      {/if}

      <MeasureChartScrub
        {scales}
        {config}
        startIndex={scrubStartIndex}
        endIndex={scrubEndIndex}
        {isScrubbing}
        showLabels={showTimeDimensionDetail}
        formatLabel={formatScrubLabel}
        on:reset={handleReset}
      />

      {#if isLocallyHovered}
        <MeasurePan
          plotBounds={config.plotBounds}
          {canPanLeft}
          {canPanRight}
          {onPanLeft}
          {onPanRight}
        />
      {/if}

      {#if annotationGroups.length > 0}
        <MeasureChartAnnotationMarkers
          groups={annotationGroups}
          hoveredGroup={hoveredAnnotationGroup}
          plotBounds={config.plotBounds}
        />
      {/if}
    </svg>

    <!-- Floating dimension comparison tooltip (HTML, positioned inside chart container) -->
    {#if !isScrubbing && isComparingDimension && isLocallyHovered && hoveredPoint && dimTooltipEntries.length > 0}
      {@const tooltipX = scales.x(hoveredIndex) + 10}
      {@const tooltipY = config.plotBounds.top + config.plotBounds.height / 2}
      <div
        class="w-fit text-[10px] font-semibold flex flex-col z-10 shadow-sm bg-surface-subtle text-fg-secondary -translate-y-1/2 py-0.5 border rounded-sm px-1 absolute pointer-events-none"
        style:top="{tooltipY}px"
        style:left="{tooltipX}px"
      >
        {#each dimTooltipEntries as entry (entry.label)}
          <div class="flex gap-x-1 items-center">
            <span
              class="size-[6.5px] rounded-full"
              style:background-color={entry.color}
            />
            <span>{entry.label}:</span>
            <span>{valueFormatter(entry.value)}</span>
          </div>
        {/each}
      </div>
    {/if}

    {#if annotationGroups.length > 0}
      <MeasureChartAnnotationPopover
        hoveredGroup={hoveredAnnotationGroup}
        onHover={handleAnnotationPopoverHover}
      />
    {/if}

    <!-- Explain CTA -->
    {#if !isScrubbing && explainX !== null}
      <ExplainButton
        x={explainX}
        plotBounds={config.plotBounds}
        onClick={() =>
          measureSelection.startAnomalyExplanationChat(metricsViewName)}
      />
    {/if}
  {:else}
    <div class="flex items-center justify-center h-full text-gray-400 text-sm">
      No data available
    </div>
  {/if}
</div>

<style>
  .measure-chart {
    position: relative;
    overflow: visible;
  }

  .cursor-crosshair {
    cursor: crosshair;
  }

  .cursor-ew-resize {
    cursor: ew-resize;
  }

  .cursor-grab {
    cursor: grab;
  }

  .cursor-grabbing {
    cursor: grabbing;
  }

  .cursor-pointer {
    cursor: pointer;
  }
</style>
