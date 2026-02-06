<script lang="ts">
  import { onDestroy } from "svelte";
  import { get } from "svelte/store";
  import type { TimeSeriesPoint, HoverState } from "./types";
  import {
    computeChartConfig,
    computeYExtent,
    computeNiceYExtent,
    computeXTickIndices,
  } from "./scales";
  import { EMPTY_HOVER } from "./interactions";
  import { ScrubController } from "./ScrubController";
  import TimeSeriesChart from "@rilldata/web-common/components/time-series-chart/TimeSeriesChart.svelte";
  import BarChart from "@rilldata/web-common/components/time-series-chart/BarChart.svelte";
  import MeasureChartTooltip from "./MeasureChartTooltip.svelte";
  import MeasureChartHoverTooltip from "./MeasureChartHoverTooltip.svelte";
  import MeasureChartScrub from "./MeasureChartScrub.svelte";
  import MeasurePan from "./MeasurePan.svelte";
  import ExplainButton from "./ExplainButton.svelte";
  import MeasureChartPointIndicator from "./MeasureChartPointIndicator.svelte";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
  import { scaleLinear } from "d3-scale";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import type { Interval } from "luxon";
  import { DateTime } from "luxon";
  import { groupAnnotations } from "./annotation-utils";
  import type { Annotation } from "@rilldata/web-common/components/data-graphic/marks/annotations";
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
  import type { DimensionSeriesData } from "./types";

  const chartId = Math.random().toString(36).slice(2, 11);
  const CLICK_THRESHOLD_PX = 4;

  export let measure: MetricsViewSpecMeasure;
  export let measureName: string;
  export let data: TimeSeriesPoint[];
  export let dimensionData: DimensionSeriesData[];
  export let annotations: Annotation[];
  export let showComparison: boolean;
  export let showTimeDimensionDetail: boolean;
  export let timeGranularity: V1TimeGrain | undefined;
  export let interval: Interval<true> | undefined;
  export let comparisonInterval: Interval<true> | undefined;
  export let chartScrubInterval: Interval<true> | undefined;
  export let canPanLeft: boolean;
  export let canPanRight: boolean;
  export let onPanLeft: (() => void) | undefined;
  export let onPanRight: (() => void) | undefined;
  export let onScrub:
    | ((range: {
        start: DateTime;
        end: DateTime;
        isScrubbing: boolean;
      }) => void)
    | undefined;
  export let onScrubClear: (() => void) | undefined;
  export let scrubController: ScrubController;
  export let metricsViewName: string;
  export let connectNulls: boolean = true;

  const annotationPopover = new AnnotationPopoverController();
  const hoveredAnnotationGroup = annotationPopover.hoveredGroup;
  const selMeasure = measureSelection.measure;
  const selStart = measureSelection.start;
  const selEnd = measureSelection.end;

  let clientWidth = 425;
  let mouseDownX: number | null = null;
  let mouseDownY: number | null = null;
  let mousePageX: number | null = null;
  let mousePageY: number | null = null;
  let wasDragging = false;
  let hoverState: HoverState = EMPTY_HOVER;

  onDestroy(() => {
    hoverIndex.clear(chartId);
    annotationPopover.destroy();
  });

  $: height = showTimeDimensionDetail ? 245 : 145;
  $: config = computeChartConfig(clientWidth, height, showTimeDimensionDetail);
  $: pb = config.plotBounds;

  // Chart series & mode
  $: mode = determineMode(data);
  $: chartSeries = buildChartSeries(data, dimensionData, showComparison);
  $: barSeries =
    mode === "bar" && showComparison && chartSeries.length === 2
      ? [chartSeries[1], chartSeries[0]]
      : chartSeries;

  // Y extent & scales
  $: yRawExtent = computeYExtent(data, dimensionData, showComparison);
  $: [yMin, yMax] = computeNiceYExtent(yRawExtent[0], yRawExtent[1]);

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
  $: xTickIndices = computeXTickIndices(mode, data.length);

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
  } else if (
    hasScrubSelection &&
    scrubStartIndex !== null &&
    scrubEndIndex !== null
  ) {
    // Scrub active: highlight the full range in TDD table
    hoverIndex.setRange(scrubStartIndex, scrubEndIndex, chartId);
  } else {
    hoverIndex.clear(chartId);
  }

  $: hoveredIndex = $hoverIndex?.start ?? -1;
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
  $: annotationGroups = groupAnnotations(
    annotations,
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

  function indexToDateTime(idx: number | null): DateTime | null {
    if (idx === null || data.length === 0) return null;
    const dt = data[snapIndex(idx, data.length)]?.ts;
    return dt?.isValid ? dt : null;
  }

  function clampX(offsetX: number) {
    return Math.max(pb.left, Math.min(pb.left + pb.width, offsetX));
  }

  function handleReset() {
    onScrubClear?.();
    scrubController.reset();
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
      // Check external scrub (from store/URL), not the reactive hasScrubSelection
      // which includes transient controller state from mousedown
      const hasExternalScrub =
        externalScrubStartIndex !== null && externalScrubEndIndex !== null;
      if (!hasExternalScrub) {
        handlePointClick(e.offsetX);
      } else {
        onScrubClear?.();
      }
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
  class="measure-chart {cursorStyle} select-none size-full relative overflow-visible"
>
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
          {connectNulls}
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
      />
    {/if}

    <!-- Date label in top-left corner -->
    {#if !isScrubbing && hoveredPoint && !isComparingDimension}
      <g class="data-readout">
        <text
          class="fill-fg-muted text-outline text-[11px]"
          aria-label="{measureName} primary time label"
          x={pb.left + 6}
          y={pb.top + 10}
        >
          {formatGrainBucket(hoveredPoint.ts, timeGranularity, interval)}
        </text>

        <!-- Comparison values only on non-hovered charts (hovered chart shows floating tooltip) -->
        {#if showComparison && !isLocallyHovered}
          {@const showDelta =
            tooltipComparisonValue !== null && !!tooltipDeltaLabel}
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

  <!-- Floating tooltip -->
  {#if !isScrubbing && isLocallyHovered && hoveredPoint && mousePageX !== null && mousePageY !== null}
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
      {showComparison}
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
</div>
