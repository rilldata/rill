<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { get } from "svelte/store";
  import type { TimeSeriesPoint, HoverState } from "./types";
  import {
    computeChartConfig,
    computeYExtent,
    computeNiceYExtent,
  } from "./scales";
  import { createVisibilityObserver, EMPTY_HOVER } from "./interactions";
  import { ScrubController } from "./ScrubController";
  import TimeSeriesChart from "@rilldata/web-common/components/time-series-chart/TimeSeriesChart.svelte";
  import BarChart from "@rilldata/web-common/components/time-series-chart/BarChart.svelte";
  import MeasureChartTooltip from "./MeasureChartTooltip.svelte";
  import MeasureChartHoverTooltip from "./MeasureChartHoverTooltip.svelte";
  import MeasureChartScrub from "./MeasureChartScrub.svelte";
  import MeasurePan from "./MeasurePan.svelte";
  import ExplainButton from "./ExplainButton.svelte";
  import MeasureChartPointIndicator from "./MeasureChartPointIndicator.svelte";
  import TDDAlternateChart from "@rilldata/web-common/features/dashboards/time-dimension-details/charts/TDDAlternateChart.svelte";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import { adjustTimeInterval } from "../utils";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
  import {
    createQueryServiceMetricsViewTimeSeries,
    createQueryServiceMetricsViewAnnotations,
    type V1Expression,
    type V1MetricsViewAnnotationsResponseAnnotation,
  } from "@rilldata/web-common/runtime-client";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import { keepPreviousData } from "@tanstack/svelte-query";
  import { scaleLinear } from "d3-scale";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { DateTime, Interval } from "luxon";
  import { transformTimeSeriesData } from "./use-measure-time-series";
  import {
    createDimensionAggregationQuery,
    buildDimensionSeriesData,
  } from "./use-dimension-data";
  import type { Annotation } from "@rilldata/web-common/components/data-graphic/marks/annotations";
  import { groupAnnotations } from "./annotation-utils";
  import {
    convertV1AnnotationsResponseItemToAnnotation,
    getPeriodFromTimeGrain,
  } from "../annotations-selectors";
  import type { Period } from "@rilldata/web-common/lib/time/types";
  import { AnnotationPopoverController } from "./AnnotationPopoverController";
  import MeasureChartGrid from "./MeasureChartGrid.svelte";
  import MeasureChartAnnotationMarkers from "./MeasureChartAnnotationMarkers.svelte";
  import MeasureChartAnnotationPopover from "./MeasureChartAnnotationPopover.svelte";
  import { measureSelection } from "../measure-selection/measure-selection";
  import { formatGrainBucket } from "@rilldata/web-common/lib/time/ranges/formatter";
  import { snapIndex, dateToIndex } from "./utils";
  import { hoverIndex } from "./hover-index";
  import {
    buildChartSeries,
    determineMode,
    computeTooltipDelta,
  } from "./chart-series";
  import { X_PAD } from "./scales";
  import ComparisonTooltip from "./ComparisonTooltip.svelte";

  const chartId = Math.random().toString(36).slice(2, 11);
  const CLICK_THRESHOLD_PX = 4;
  const VISIBILITY_ROOT_MARGIN = "120px";

  export let measure: MetricsViewSpecMeasure;
  export let instanceId: string;
  export let metricsViewName: string;
  export let where: V1Expression | undefined = undefined;
  export let timeDimension: string | undefined = undefined;
  export let interval: Interval<true> | undefined = undefined;
  export let comparisonInterval: Interval<true> | undefined = undefined;
  export let timeGranularity: V1TimeGrain | undefined = undefined;
  export let timeZone: string = "UTC";
  export let comparisonDimension: string | undefined = undefined;
  export let dimensionValues: (string | null)[] = [];
  export let dimensionWhere: V1Expression | undefined = undefined;
  export let annotationsEnabled: boolean = false;
  export let showComparison: boolean = false;
  export let showTimeDimensionDetail: boolean = false;
  export let ready: boolean = true;
  export let chartScrubInterval: Interval<true> | undefined = undefined;
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
  export let onPanLeft: (() => void) | undefined = undefined;
  export let onPanRight: (() => void) | undefined = undefined;
  export let scrubController: ScrubController;

  const annotationPopover = new AnnotationPopoverController();
  const hoveredAnnotationGroup = annotationPopover.hoveredGroup;
  const { visible, observe } = createVisibilityObserver(VISIBILITY_ROOT_MARGIN);
  const selMeasure = measureSelection.measure;
  const selStart = measureSelection.start;
  const selEnd = measureSelection.end;

  let container: HTMLDivElement;
  let clientWidth = 425;
  let unobserve: (() => void) | undefined;
  let tddIsScrubbing = false;
  let mouseDownX: number | null = null;
  let mouseDownY: number | null = null;
  let mousePageX: number | null = null;
  let mousePageY: number | null = null;
  let wasDragging = false; // Track if we just finished a drag (to skip click handler)
  let hoverState: HoverState = EMPTY_HOVER;

  onMount(() => {
    if (container) unobserve = observe(container);
  });

  onDestroy(() => {
    unobserve?.();
    hoverIndex.clear(chartId);
    annotationPopover.destroy();
  });

  $: measureName = measure.name ?? "";
  $: height = showTimeDimensionDetail ? 245 : 145;
  $: config = computeChartConfig(clientWidth, height, showTimeDimensionDetail);
  $: pb = config.plotBounds;

  // Extract ISO strings for API calls
  $: timeStart = interval?.start?.toISO() ?? undefined;
  $: timeEnd = interval?.end?.toISO() ?? undefined;
  $: comparisonTimeStart = comparisonInterval?.start?.toISO() ?? undefined;
  $: comparisonTimeEnd = comparisonInterval?.end?.toISO() ?? undefined;

  // Time series queries
  $: timeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
    instanceId,
    metricsViewName,
    {
      measureNames: [measureName],
      where,
      timeDimension,
      timeStart,
      timeEnd,
      timeGranularity,
      timeZone,
    },
    {
      query: {
        enabled: $visible && ready && !!timeStart,
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
  $: error = ($timeSeriesQuery.error as HTTPError | undefined)?.response?.data
    ?.message;

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
        timeStart,
        timeEnd,
        timeGranularity!,
        timeZone,
        $visible && ready && !!timeStart,
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
  $: mode = determineMode(data);
  $: chartSeries = buildChartSeries(data, dimensionData, showComparison);

  // For bar mode with time comparison, reverse order so comparison bar is on left
  $: barSeries =
    mode === "bar" && showComparison && chartSeries.length === 2
      ? [chartSeries[1], chartSeries[0]]
      : chartSeries;

  // Y extent & scales
  $: yRawExtent = computeYExtent(data, dimensionData, showComparison);
  $: [yMin, yMax] = computeNiceYExtent(yRawExtent[0], yRawExtent[1]);

  // X/Y scales - domain is always [0, dataLength-1]
  $: dataLastIndex = Math.max(0, data.length - 1);
  $: barSlotWidth = pb.width / Math.max(1, data.length);
  $: xRangeStart =
    mode === "line" ? pb.left + X_PAD : pb.left + barSlotWidth / 2;
  $: xRangeEnd =
    mode === "line"
      ? pb.left + pb.width - X_PAD
      : pb.left + pb.width - barSlotWidth / 2;
  $: xScale = scaleLinear<number>()
    .domain([0, dataLastIndex])
    .range([xRangeStart, xRangeEnd]);
  $: yScale = scaleLinear<number>()
    .domain([yMin, yMax])
    .range([pb.top + pb.height, pb.top]);
  $: scales = { x: xScale, y: yScale };
  $: yTicks = yScale.ticks(3);
  $: xTickIndices =
    dataLastIndex >= 2
      ? [0, Math.floor(dataLastIndex / 2), dataLastIndex]
      : dataLastIndex >= 1
        ? [0, dataLastIndex]
        : [0];
  // Keep scrub controller in sync with data length
  $: scrubController.setDataLength(data.length);

  // Subscribe to scrub state from controller for rendering
  $: scrubStateStore = scrubController.state;
  $: currentScrubState = $scrubStateStore;
  $: isScrubbing = currentScrubState.isScrubbing;

  // Scrub indices: use local (active) state while scrubbing, external (URL) state otherwise
  $: externalScrubStartIndex = chartScrubInterval
    ? dateToIndex(data, chartScrubInterval.start.toMillis())
    : null;
  $: externalScrubEndIndex = chartScrubInterval
    ? dateToIndex(data, chartScrubInterval.end.toMillis())
    : null;
  $: scrubStartIndex = currentScrubState.startIndex ?? externalScrubStartIndex;
  $: scrubEndIndex = currentScrubState.endIndex ?? externalScrubEndIndex;
  $: hasScrubSelection = scrubStartIndex !== null && scrubEndIndex !== null;

  // Hover state
  $: hoverIndex.registerScale(xScale);
  $: isLocallyHovered =
    hoverState.isHovered && hoverState.index !== null && data.length > 0;
  $: if (isLocallyHovered) {
    hoverIndex.set(snapIndex(hoverState.index!, data.length), chartId);
  } else {
    hoverIndex.clear(chartId);
  }
  $: hoveredIndex = $hoverIndex ?? -1;
  $: hoveredPoint = data[hoveredIndex] ?? null;
  $: cursorStyle = scrubController.getCursorStyle(hoverState.screenX, xScale);

  // Formatters
  $: measureFormatter = createMeasureValueFormatter(measure);
  $: valueFormatter = (value: number | null): string => {
    if (value === null) return "\u2013";
    return measureFormatter(value);
  };
  $: axisFormatter = createMeasureValueFormatter(measure, "axis");

  // Annotations
  $: annotationsQuery = createQueryServiceMetricsViewAnnotations(
    instanceId,
    metricsViewName,
    {
      timeRange: {
        start: timeStart,
        end: timeEnd,
        timeDimension,
      },
      timeGrain: timeGranularity,
      measures: [measureName],
    },
    {
      query: {
        enabled:
          annotationsEnabled && !!timeStart && !!timeEnd && !!timeGranularity,
        placeholderData: keepPreviousData,
        refetchOnMount: false,
      },
    },
  );

  $: selectedPeriod = getPeriodFromTimeGrain(timeGranularity);

  $: annotationRows = $annotationsQuery.data?.rows;
  $: annotationsList = buildAnnotationsList(
    annotationRows,
    selectedPeriod,
    timeGranularity,
    timeZone,
  );

  $: annotationGroups = groupAnnotations(
    annotationsList,
    scales,
    data,
    config,
    timeGranularity,
  );

  // Tooltip data
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

  // Time comparison delta
  $: tooltipDelta = computeTooltipDelta(hoveredPoint);
  $: ({
    currentValue: tooltipCurrentValue,
    comparisonValue: tooltipComparisonValue,
    deltaLabel: tooltipDeltaLabel,
    deltaPositive: tooltipDeltaPositive,
  } = tooltipDelta);

  // Explain CTA positioning
  $: isThisMeasureSelected = $selMeasure === measureName;
  $: singleSelectIdx =
    isThisMeasureSelected && $selStart && !$selEnd
      ? dateToIndex(data, $selStart.getTime())
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

  function buildAnnotationsList(
    rows: V1MetricsViewAnnotationsResponseAnnotation[] | undefined,
    period: Period | undefined,
    grain: V1TimeGrain | undefined,
    tz: string,
  ): Annotation[] {
    if (!rows?.length) return [];
    const list = rows.map((a) =>
      convertV1AnnotationsResponseItemToAnnotation(
        a,
        period,
        grain ?? V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
        tz,
      ),
    );
    list.sort((a, b) => a.startTime.toMillis() - b.startTime.toMillis());
    return list;
  }

  function indexToDateTime(idx: number | null): DateTime | null {
    if (idx === null || data.length === 0) return null;
    const dt = data[snapIndex(idx, data.length)]?.ts;
    return dt?.isValid ? dt : null;
  }

  function formatScrubLabel(idx: number): string {
    const dt = indexToDateTime(idx);
    if (!dt) return "";
    return formatGrainBucket(dt, timeGranularity, interval);
  }

  function clampX(offsetX: number) {
    return Math.max(pb.left, Math.min(pb.left + pb.width, offsetX));
  }

  function handleReset() {
    onScrubClear?.();
    scrubController.reset();
  }

  function handleTddHover(
    _dimension: undefined | string | null,
    ts: Date | undefined,
  ) {
    if (ts) {
      const idx = dateToIndex(data, ts.getTime());
      if (idx !== null) hoverIndex.set(idx, "tddChart");
    } else {
      hoverIndex.clear("tddChart");
    }
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

  function handleSvgMouseLeave() {
    hoverState = EMPTY_HOVER;
    mousePageX = null;
    mousePageY = null;
    annotationPopover.scheduleClear();
  }

  function handleMouseDown(e: MouseEvent) {
    if (e.button !== 0) return;
    mouseDownX = e.clientX;
    mouseDownY = e.clientY;
    const x = clampX(e.offsetX);

    // If there's a visual selection from external scrubRange but controller is empty,
    // initialize the controller so edge-resize and move detection work
    const controllerState = get(scrubController.state);
    if (
      controllerState.startIndex === null &&
      externalScrubStartIndex !== null &&
      externalScrubEndIndex !== null
    ) {
      scrubController.initFromExternal(
        externalScrubStartIndex,
        externalScrubEndIndex,
      );
    }

    scrubController.start(x, xScale);
  }

  function finalizeScrubSelection(startIndex: number, endIndex: number) {
    const startDt = indexToDateTime(startIndex);
    const endDt = indexToDateTime(endIndex);
    if (!startDt || !endDt) return;
    onScrub?.({ start: startDt, end: endDt, isScrubbing: false });
    if (measureName) {
      const s = startDt.toJSDate();
      const e = endDt.toJSDate();
      const [start, end] = s < e ? [s, e] : [e, s];
      measureSelection.setRange(measureName, start, end);
    }
    scrubController.reset();
  }

  function handlePointClick(offsetX: number) {
    const clickedIndex = Math.max(
      0,
      Math.min(data.length - 1, Math.round(xScale.invert(clampX(offsetX)))),
    );
    const pt = data[clickedIndex];
    if (pt?.ts?.isValid && measureName) {
      onScrubClear?.();
      measureSelection.setStart(measureName, pt.ts.toJSDate());
    }
  }

  function handleMouseUp(e: MouseEvent) {
    const wasClick =
      mouseDownX !== null &&
      mouseDownY !== null &&
      Math.abs(e.clientX - mouseDownX) < CLICK_THRESHOLD_PX &&
      Math.abs(e.clientY - mouseDownY) < CLICK_THRESHOLD_PX;

    wasDragging = !wasClick;

    // Capture scrub state BEFORE end() which may reset it
    const scrubState = get(scrubController.state);
    const { startIndex, endIndex } = scrubState;
    const selectionKept = scrubController.end();
    mouseDownX = null;
    mouseDownY = null;

    if (!scrubState.isScrubbing) return;

    if (selectionKept && startIndex !== null && endIndex !== null) {
      finalizeScrubSelection(startIndex, endIndex);
    } else if (wasClick) {
      handlePointClick(e.offsetX);
    } else {
      onScrubClear?.();
    }
  }

  function handleChartClick(e: MouseEvent) {
    e.stopPropagation();
    e.preventDefault();

    // Skip if we just finished a drag - only handle actual clicks
    if (wasDragging) {
      wasDragging = false;
      return;
    }

    // Use visual scrub indices (which may come from controller OR external scrubRange)
    if (scrubStartIndex === null || scrubEndIndex === null || isScrubbing) {
      return;
    }

    const x = clampX(e.offsetX);
    const clickIndex = xScale.invert(x);
    const [minIdx, maxIdx] =
      scrubStartIndex < scrubEndIndex
        ? [scrubStartIndex, scrubEndIndex]
        : [scrubEndIndex, scrubStartIndex];
    const isOutside = clickIndex < minIdx || clickIndex > maxIdx;

    if (isOutside) {
      // Click outside selection clears it
      scrubController.reset();
      onScrubClear?.();
      measureSelection.clear();
    } else if (measureName) {
      // Click inside - ensure measureSelection matches the scrub
      const startDt = indexToDateTime(scrubStartIndex);
      const endDt = indexToDateTime(scrubEndIndex);
      if (startDt && endDt) {
        const s = startDt.toJSDate();
        const e2 = endDt.toJSDate();
        const [start, end] = s < e2 ? [s, e2] : [e2, s];
        measureSelection.setRange(measureName, start, end);
      }
    }
  }
</script>

<div
  bind:clientWidth
  bind:this={container}
  class="measure-chart {cursorStyle} select-none size-full relative overflow-visible"
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
      aria-label="Measure Chart for {measureName}"
      class="size w-full overflow-visible"
      height="{height}px"
      on:mousemove={(e) => {
        const x = clampX(e.offsetX);
        const fractionalIndex = xScale.invert(x);

        hoverState = {
          index: fractionalIndex,
          screenX: x,
          screenY: e.offsetY,
          isHovered: true,
        };

        // Update scrub if dragging
        if (get(scrubController.state).isScrubbing) {
          scrubController.update(x, xScale);
        }

        annotationPopover.checkHover(e, annotationGroups, isScrubbing);
        mousePageX = e.pageX;
        mousePageY = e.pageY;
      }}
      on:mouseleave={handleSvgMouseLeave}
      on:mousedown={handleMouseDown}
      on:mouseup={handleMouseUp}
      on:click={handleChartClick}
    >
      <!-- Clip chart body to plot area so lines/bars don't bleed into margins when overplotting -->
      <defs>
        <clipPath id="chart-body-{chartId}">
          <rect x={pb.left} y={pb.top} width={pb.width} height={pb.height} />
        </clipPath>
      </defs>

      <MeasureChartGrid
        {yTicks}
        {xTickIndices}
        {yScale}
        {xScale}
        plotLeft={pb.left}
        plotWidth={pb.width}
        plotTop={pb.top}
        plotHeight={pb.height}
        {axisFormatter}
      />

      <!-- Chart body -->
      <g clip-path="url(#chart-body-{chartId})">
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
            series={barSeries}
            yScale={scales.y}
            stacked={false}
            plotLeft={pb.left}
            plotWidth={pb.width}
            visibleStart={0}
            visibleEnd={dataLastIndex}
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
          {dimensionData}
          {showComparison}
          isComparingDimension={dimensionData.length > 0}
          isBarMode={mode === "bar"}
          visibleStart={0}
          visibleEnd={dataLastIndex}
          formatter={valueFormatter}
        />
      {/if}

      <!-- Data readout in top-left corner (only for non-hovered charts, not in dimension comparison) -->
      {#if !isScrubbing && hoveredPoint && (!showComparison || !isLocallyHovered) && !isComparingDimension}
        {@const showDelta =
          showComparison &&
          tooltipComparisonValue !== null &&
          !!tooltipDeltaLabel}
        <g class="data-readout">
          <!-- Date -->
          <text
            class="fill-fg-muted text-outline text-[10px]"
            aria-label="{measureName} primary time label"
            x={pb.left + 6}
            y={pb.top + 10}
          >
            {formatGrainBucket(hoveredPoint.ts, timeGranularity, interval)}
          </text>

          {#if showComparison}
            <ComparisonTooltip
              x={pb.left + 6}
              y={pb.top + 22}
              {tooltipCurrentValue}
              {tooltipComparisonValue}
              {tooltipDeltaLabel}
              {tooltipDeltaPositive}
              {showDelta}
              {valueFormatter}
            />
          {/if}
        </g>
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
        onReset={handleReset}
      />

      {#if isLocallyHovered}
        <MeasurePan
          plotBounds={pb}
          {canPanLeft}
          {canPanRight}
          {onPanLeft}
          {onPanRight}
        />
      {/if}

      {#if annotationGroups.length > 0}
        <MeasureChartAnnotationMarkers
          groups={annotationGroups}
          hoveredGroup={$hoveredAnnotationGroup}
          plotBounds={pb}
        />
      {/if}
    </svg>

    <!-- Floating tooltip - only shown in comparison modes -->
    {#if !isScrubbing && isLocallyHovered && hoveredPoint && mousePageX !== null && mousePageY !== null && (showComparison || isComparingDimension)}
      <MeasureChartHoverTooltip
        mouseX={mousePageX}
        mouseY={mousePageY}
        currentValue={tooltipCurrentValue}
        comparisonValue={tooltipComparisonValue}
        currentTs={hoveredPoint.ts}
        comparisonTs={hoveredPoint.comparisonTs}
        {timeGranularity}
        {interval}
        {comparisonInterval}
        {isComparingDimension}
        {dimTooltipEntries}
        deltaLabel={tooltipDeltaLabel}
        deltaPositive={tooltipDeltaPositive}
        formatter={valueFormatter}
      />
    {/if}

    {#if annotationGroups.length > 0}
      <MeasureChartAnnotationPopover
        hoveredGroup={$hoveredAnnotationGroup}
        onHover={(h) => annotationPopover.setPopoverHovered(h)}
      />
    {/if}

    <!-- Explain CTA -->
    {#if !isScrubbing && explainX !== null}
      <ExplainButton
        x={explainX}
        plotBounds={pb}
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
