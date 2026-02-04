<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { writable, get, type Readable, type Writable } from "svelte/store";
  import type {
    TimeSeriesPoint,
    DimensionSeriesData,
    ChartSeries,
    ChartMode,
    HoverState,
  } from "./types";
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
  import { COMPARISON_COLORS } from "@rilldata/web-common/features/dashboards/config";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
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
  import { formatDateTimeByGrain } from "@rilldata/web-common/lib/time/ranges/formatter";
  import { V1TimeGrainToOrder } from "@rilldata/web-common/lib/time/new-grains";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { numberPartsToString } from "@rilldata/web-common/lib/number-formatting/utils/number-parts-utils";
  import { snapIndex, dateToIndex } from "./utils";

  const chartId = Math.random().toString(36).slice(2, 11);
  const Y_DASH_ARRAY = "1,1";
  const DAY_GRAIN_ORDER = V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_DAY];
  const X_PAD = 8;
  const CLICK_THRESHOLD_PX = 4; // Max mouse movement to still count as a click
  const LINE_MODE_MIN_POINTS = 6; // Minimum data points to show line instead of bar
  const ANNOTATION_POPOVER_DELAY_MS = 150;
  const VISIBILITY_ROOT_MARGIN = "120px";
  const { visible, observe } = createVisibilityObserver(VISIBILITY_ROOT_MARGIN);
  const selMeasure = measureSelection.measure;
  const selStart = measureSelection.start;
  const selEnd = measureSelection.end;

  export let measure: MetricsViewSpecMeasure;
  export let instanceId: string;
  export let metricsViewName: string;
  export let where: V1Expression | undefined = undefined;
  export let timeDimension: string | undefined = undefined;
  export let timeStart: string | undefined = undefined;
  export let timeEnd: string | undefined = undefined;
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
  export let scrubController: ScrubController;
  export let showAxis: boolean = false;

  let container: HTMLDivElement;
  let clientWidth = 425;
  let unobserve: (() => void) | undefined;
  let tddIsScrubbing = false;
  let hoveredAnnotationGroup: AnnotationGroup | null = null;
  let annotationPopoverHovered = false;
  let annotationPopoverTimeout: ReturnType<typeof setTimeout> | null = null;
  let mouseDownX: number | null = null;
  let mouseDownY: number | null = null;
  let mousePageX: number | null = null;
  let mousePageY: number | null = null;
  let wasDragging = false; // Track if we just finished a drag (to skip click handler)

  const hoverState = writable<HoverState>(EMPTY_HOVER);

  onMount(() => {
    if (container) unobserve = observe(container);
  });
  onDestroy(() => {
    unobserve?.();
    if (annotationPopoverTimeout) clearTimeout(annotationPopoverTimeout);
  });

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
  $: barSlotWidth = config.plotBounds.width / Math.max(1, data.length);
  $: xRangeStart =
    mode === "line"
      ? config.plotBounds.left + X_PAD
      : config.plotBounds.left + barSlotWidth / 2;
  $: xRangeEnd =
    mode === "line"
      ? config.plotBounds.left + config.plotBounds.width - X_PAD
      : config.plotBounds.left + config.plotBounds.width - barSlotWidth / 2;
  $: xScale = scaleLinear<number>()
    .domain([0, dataLastIndex])
    .range([xRangeStart, xRangeEnd]);
  $: yScale = scaleLinear<number>()
    .domain([yMin, yMax])
    .range([
      config.plotBounds.top + config.plotBounds.height,
      config.plotBounds.top,
    ]);
  $: scales = { x: xScale, y: yScale };
  $: yTicks = yScale.ticks(3);
  $: xTickIndices =
    dataLastIndex >= 2
      ? [0, Math.floor(dataLastIndex / 2), dataLastIndex]
      : dataLastIndex >= 1
        ? [0, dataLastIndex]
        : [0];
  // Sub-day grains (hour, minute, etc.) show time + date on separate lines
  $: isSubDayGrain = timeGranularity
    ? V1TimeGrainToOrder[timeGranularity] < DAY_GRAIN_ORDER
    : false;
  $: axisHeight = isSubDayGrain ? 26 : 16;
  $: axisTicks = buildAxisTicks(xTickIndices, data, mode, isSubDayGrain);

  function buildAxisTicks(
    indices: number[],
    d: TimeSeriesPoint[],
    m: ChartMode,
    subDay: boolean,
  ) {
    return indices.map((idx, i) => {
      const dt = d[idx]?.ts;
      const anchor =
        m === "bar"
          ? "middle"
          : i === 0
            ? "start"
            : i === indices.length - 1
              ? "end"
              : "middle";

      if (!dt) return { x: xScale(idx), anchor, timeLine: "", dateLine: "" };

      if (subDay) {
        const grainOrder = V1TimeGrainToOrder[timeGranularity!];
        const fmt: Intl.DateTimeFormatOptions = {
          hour: "numeric",
          hour12: true,
        };
        if (grainOrder < 1) fmt.minute = "2-digit";
        const timeLine = dt.toLocaleString(fmt);

        const prevDt = i > 0 ? d[indices[i - 1]]?.ts : undefined;
        const showDate =
          !prevDt ||
          dt.day !== prevDt.day ||
          dt.month !== prevDt.month ||
          dt.year !== prevDt.year;
        const dateFmt: Intl.DateTimeFormatOptions = {
          month: "short",
          day: "numeric",
        };
        if (d[0]?.ts.year !== d[d.length - 1]?.ts.year)
          dateFmt.year = "numeric";
        const dateLine = showDate ? dt.toLocaleString(dateFmt) : "";

        return { x: xScale(idx), anchor, timeLine, dateLine };
      }

      return {
        x: xScale(idx),
        anchor,
        timeLine: formatDateTimeByGrain(dt, timeGranularity),
        dateLine: "",
      };
    });
  }

  // Keep scrub controller in sync with data length
  $: scrubController.setDataLength(data.length);

  // Subscribe to scrub state from controller for rendering
  $: scrubStateStore = scrubController.state;
  $: currentScrubState = $scrubStateStore;
  $: isScrubbing = currentScrubState.isScrubbing;

  // Scrub indices: use local (active) state while scrubbing, external (URL) state otherwise
  $: externalScrubStartIndex = scrubRange
    ? dateToIndex(data, scrubRange.start.toMillis())
    : null;
  $: externalScrubEndIndex = scrubRange
    ? dateToIndex(data, scrubRange.end.toMillis())
    : null;
  $: scrubStartIndex = currentScrubState.startIndex ?? externalScrubStartIndex;
  $: scrubEndIndex = currentScrubState.endIndex ?? externalScrubEndIndex;
  $: hasScrubSelection = scrubStartIndex !== null && scrubEndIndex !== null;

  // Hover state - snap to nearest data index
  $: localHoveredIndex =
    $hoverState.index !== null ? snapIndex($hoverState.index, data.length) : -1;
  $: isLocallyHovered = $hoverState.isHovered && localHoveredIndex >= 0;
  $: cursorStyle = scrubController.getCursorStyle($hoverState.screenX, xScale);
  $: sharedHoverIndex.set(isLocallyHovered ? localHoveredIndex : undefined);
  $: onHover?.(
    isLocallyHovered
      ? (indexToDateTime(localHoveredIndex) ?? undefined)
      : undefined,
  );
  $: tableHoverIndex = $tableHoverTime
    ? dateToIndex(data, $tableHoverTime.getTime())
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
  $: tooltipCurrentValue = hoveredPoint?.value ?? null;
  $: tooltipComparisonValue = hoveredPoint?.comparisonValue ?? null;
  $: tooltipDelta =
    tooltipCurrentValue !== null &&
    tooltipComparisonValue !== null &&
    tooltipComparisonValue !== 0
      ? (tooltipCurrentValue - tooltipComparisonValue) / tooltipComparisonValue
      : null;
  $: tooltipDeltaLabel =
    tooltipDelta !== null
      ? numberPartsToString(formatMeasurePercentageDifference(tooltipDelta))
      : null;
  $: tooltipDeltaPositive = tooltipDelta !== null && tooltipDelta >= 0;

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

  function buildChartSeries(
    d: TimeSeriesPoint[],
    dimData: DimensionSeriesData[],
    comparison: boolean,
  ): ChartSeries[] {
    if (dimData.length > 0) {
      return dimData.map((dim, i) => ({
        id: `dim-${dim.dimensionValue ?? i}`,
        values: dim.data.map((pt) => pt.value),
        color: dim.color || COMPARISON_COLORS[i % COMPARISON_COLORS.length],
        opacity: dim.isFetching ? 0.5 : 1,
        strokeWidth: 1.5,
      }));
    }

    const result: ChartSeries[] = [];

    // Primary series first (gets area fill in line mode, right bar in grouped bar mode)
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

    // Comparison series second (lighter line, left bar in grouped bar mode)
    if (comparison && d.length > 0) {
      result.push({
        id: "comparison",
        values: d.map((pt) => pt.comparisonValue ?? null),
        color: TimeComparisonLineColor,
        opacity: 0.5,
      });
    }

    return result;
  }

  function determineMode(d: TimeSeriesPoint[]): ChartMode {
    if (d.length >= LINE_MODE_MIN_POINTS) return "line";
    return "bar";
  }

  function indexToDateTime(idx: number | null): DateTime | null {
    if (idx === null || data.length === 0) return null;
    const dt = data[snapIndex(idx, data.length)]?.ts;
    return dt?.isValid ? dt : null;
  }

  function formatScrubLabel(idx: number): string {
    const dt = indexToDateTime(idx);
    if (!dt) return "";
    return formatDateTimeByGrain(dt, timeGranularity);
  }

  function handleReset() {
    onScrubClear?.();
    scrubController.reset();
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
    if (annotationPopoverTimeout) {
      clearTimeout(annotationPopoverTimeout);
      annotationPopoverTimeout = null;
    }
    if (!hovered) {
      annotationPopoverTimeout = setTimeout(() => {
        if (!annotationPopoverHovered) {
          hoveredAnnotationGroup = null;
        }
        annotationPopoverTimeout = null;
      }, ANNOTATION_POPOVER_DELAY_MS);
    }
  }

  function handleSvgMouseLeave() {
    hoverState.set(EMPTY_HOVER);
    mousePageX = null;
    mousePageY = null;
    if (!annotationPopoverHovered) {
      hoveredAnnotationGroup = null;
    }
  }

  function handleMouseDown(e: MouseEvent) {
    if (e.button !== 0) return;
    mouseDownX = e.clientX;
    mouseDownY = e.clientY;
    const x = Math.max(
      config.plotBounds.left,
      Math.min(config.plotBounds.left + config.plotBounds.width, e.offsetX),
    );

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

  function handleMouseUp(e: MouseEvent) {
    // Check if this was a click (minimal mouse movement) vs a drag
    const wasClick =
      mouseDownX !== null &&
      mouseDownY !== null &&
      Math.abs(e.clientX - mouseDownX) < CLICK_THRESHOLD_PX &&
      Math.abs(e.clientY - mouseDownY) < CLICK_THRESHOLD_PX;

    // Track if this was a drag to prevent click handler from also firing
    wasDragging = !wasClick;

    // Get current scrub state BEFORE calling end() which might reset it
    const scrubState = get(scrubController.state);
    const startIndex = scrubState.startIndex;
    const endIndex = scrubState.endIndex;

    // Calculate clicked index for single-point selection
    const clickedX = Math.max(
      config.plotBounds.left,
      Math.min(config.plotBounds.left + config.plotBounds.width, e.offsetX),
    );
    const clickedIndex = Math.max(
      0,
      Math.min(data.length - 1, Math.round(xScale.invert(clickedX))),
    );

    // Finalize the scrub interaction
    const selectionKept = scrubController.end();
    mouseDownX = null;
    mouseDownY = null;

    // If we weren't scrubbing, nothing to sync
    if (!scrubState.isScrubbing) {
      return;
    }

    if (selectionKept && startIndex !== null && endIndex !== null) {
      // Valid range selection - sync to parent
      const startDt = indexToDateTime(startIndex);
      const endDt = indexToDateTime(endIndex);
      if (startDt && endDt) {
        onScrub?.({ start: startDt, end: endDt, isScrubbing: false });
        if (measureName) {
          const s = startDt.toJSDate();
          const e2 = endDt.toJSDate();
          const [start, end] = s < e2 ? [s, e2] : [e2, s];
          measureSelection.setRange(measureName, start, end);
        }
        // Clear controller's local state - the selection will now come from scrubRange prop
        // This ensures external clearing (e.g., "Clear Filters") works properly
        scrubController.reset();
      }
    } else if (wasClick && measureName) {
      // Single point click - set point selection
      if (clickedIndex >= 0 && clickedIndex < data.length) {
        const pt = data[clickedIndex];
        if (pt?.ts?.isValid) {
          onScrubClear?.();
          measureSelection.setStart(measureName, pt.ts.toJSDate());
        }
      }
    } else {
      // Drag that ended at same point - just clear
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

    const x = Math.max(
      config.plotBounds.left,
      Math.min(config.plotBounds.left + config.plotBounds.width, e.offsetX),
    );
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
    {#if showAxis}
      <svg class="w-full overflow-visible" height={axisHeight}>
        {#each axisTicks as tick, tickIdx (tickIdx)}
          <text
            class="fill-fg-secondary text-[11px]"
            text-anchor={tick.anchor}
            x={tick.x}
            y={isSubDayGrain ? 11 : axisHeight - 3}
          >
            {tick.timeLine}
          </text>
          {#if tick.dateLine}
            <text
              class="fill-fg-muted text-[11px]"
              text-anchor={tick.anchor}
              x={tick.x}
              y={23}
            >
              {tick.dateLine}
            </text>
          {/if}
        {/each}
      </svg>
    {/if}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-noninteractive-element-interactions -->
    <svg
      role="application"
      aria-label="Measure Chart for {measureName}"
      class="size w-full overflow-visible"
      height="{height}px"
      on:mousemove={(e) => {
        // Clamp x to plot bounds
        const x = Math.max(
          config.plotBounds.left,
          Math.min(config.plotBounds.left + config.plotBounds.width, e.offsetX),
        );
        const fractionalIndex = xScale.invert(x);

        hoverState.set({
          index: fractionalIndex,
          screenX: x,
          screenY: e.offsetY,
          isHovered: true,
        });

        // Update scrub if dragging
        if (get(scrubController.state).isScrubbing) {
          scrubController.update(x, xScale);
        }

        checkAnnotationHover(e);
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

      <!-- X-axis vertical tick lines -->
      <g class="x-axis">
        {#each xTickIndices as idx (idx)}
          <line
            class="stroke-border"
            x1={xScale(idx)}
            x2={xScale(idx)}
            y1={config.plotBounds.top}
            y2={config.plotBounds.top + config.plotBounds.height}
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
            plotLeft={config.plotBounds.left}
            plotWidth={config.plotBounds.width}
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
          tooltipDeltaLabel}
        <g class="data-readout">
          <!-- Date -->
          <text
            class="fill-fg-muted text-outline text-[10px]"
            x={config.plotBounds.left + 6}
            y={config.plotBounds.top + 10}
          >
            {formatDateTimeByGrain(hoveredPoint.ts, timeGranularity)}
          </text>

          <!-- Value (with comparison on same line if applicable) -->
          {#if showComparison}
            <text
              class="text-outline text-[12px]"
              x={config.plotBounds.left + 6}
              y={config.plotBounds.top + 24}
            >
              <tspan class="fill-theme-700 font-semibold">
                {valueFormatter(tooltipCurrentValue)}
              </tspan>
              <tspan class="fill-fg-muted">
                vs {valueFormatter(tooltipComparisonValue)}
              </tspan>
              {#if showDelta}
                <tspan
                  class={tooltipDeltaPositive
                    ? "fill-green-600"
                    : "fill-red-600"}
                >
                  ({tooltipDeltaPositive ? "+" : ""}{tooltipDeltaLabel})
                </tspan>
              {/if}
            </text>
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
        {isComparingDimension}
        {dimTooltipEntries}
        deltaLabel={tooltipDeltaLabel}
        deltaPositive={tooltipDeltaPositive}
        formatter={valueFormatter}
      />
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
