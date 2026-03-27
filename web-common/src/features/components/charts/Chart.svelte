<script lang="ts">
  import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
  import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
  import {
    resolveSignalField,
    resolveSignalIntervalField,
    resolveSignalTimeField,
  } from "@rilldata/web-common/components/vega/vega-signals";
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
  import VegaRenderer from "@rilldata/web-common/components/vega/VegaRenderer.svelte";
  import type { CanvasChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import {
    createMeasureValueFormatter,
    humanizeDataType,
  } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import type { TimeRange } from "@rilldata/web-common/lib/time/types";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
  import { onDestroy } from "svelte";
  import type { SignalListeners, VegaSpec } from "svelte-vega";
  import type { Readable } from "svelte/store";
  import type { View } from "vega-typings";
  import { getChroma } from "../../themes/theme-utils";
  import {
    compileToBrushedVegaSpec,
    createAdaptiveScrubHandler,
  } from "./brush-builder";
  import {
    computeVisibleSlices,
    OTHER_LABEL,
    type OtherGroupingResult,
  } from "./circular/other-grouping";
  import { createPieTooltipFormatter } from "./circular/pie-tooltip-formatter";
  import type { VLTooltipFormatter } from "@rilldata/web-common/components/vega/types";
  import type { ColorMapping, ChartDataResult, ChartType } from "./types";
  import { generateSpec, getColorMappingForChart } from "./util";

  export let chartType: ChartType;
  export let chartSpec: CanvasChartSpec;
  export let chartData: Readable<ChartDataResult>;
  export let measures: MetricsViewSpecMeasure[];
  export let themeMode: "light" | "dark" = "light";
  /**
   * Full theme object with all CSS variables (primary, secondary, background, etc.)
   * If provided, chart uses these directly. If not, falls back to defaults.
   */
  export let theme: Record<string, string> | undefined = undefined;
  export let isCanvas: boolean;

  export let isScrubbing: boolean = false;
  export let temporalField: string | undefined = undefined;
  export let onBrush: ((interval: TimeRange) => void) | undefined = undefined;
  export let onBrushEnd: ((interval: TimeRange) => void) | undefined =
    undefined;
  export let onBrushClear: (() => void) | undefined = undefined;
  export let onHover:
    | ((dimension: string | null | undefined, time: Date | undefined) => void)
    | undefined = undefined;

  export let view: View;

  let vegaSpec: VegaSpec | undefined = undefined;
  let prevVlSpec: unknown = undefined;
  let compileGeneration = 0;

  $: ({ data, domainValues, hasComparison, isFetching, error } = $chartData);

  $: hasNoData = !isFetching && data.length === 0;

  // Override chartData theme with mode-aware colors if theme prop is provided
  $: chartDataWithTheme = theme
    ? {
        ...$chartData,
        theme: {
          primary: theme.primary
            ? getChroma(theme.primary)
            : $chartData.theme.primary,
          secondary: theme.secondary
            ? getChroma(theme.secondary)
            : $chartData.theme.secondary,
        },
      }
    : $chartData;

  // For pie/donut charts: apply "Other" grouping before spec generation
  $: isPieOrDonut = chartType === "pie_chart" || chartType === "donut_chart";

  $: otherGrouping = isPieOrDonut
    ? applyPieGrouping(chartSpec, chartDataWithTheme.data)
    : null;

  // Inject grouped data + "Other" domain value into chartData for spec generation
  $: chartDataForSpec = otherGrouping
    ? {
        ...chartDataWithTheme,
        data: otherGrouping.visibleData,
        domainValues: injectOtherIntoDomain(
          chartDataWithTheme.domainValues,
          chartSpec,
          otherGrouping,
        ),
      }
    : chartDataWithTheme;

  $: spec = generateSpec(chartType, chartSpec, chartDataForSpec);

  // Compile VL spec to Vega spec when brush is enabled.
  // Memoize with deep equality to avoid recompilation on store re-emissions
  // that produce the same spec, which would reset brush selection state.
  $: useBrush = "isInteractive" in chartSpec && !!chartSpec.isInteractive;
  $: {
    if (
      useBrush &&
      spec &&
      JSON.stringify(spec) !== JSON.stringify(prevVlSpec)
    ) {
      prevVlSpec = spec;
      const gen = ++compileGeneration;
      void compileToBrushedVegaSpec(spec, isThemeModeDark, theme).then(
        (compiled) => {
          if (gen === compileGeneration) vegaSpec = compiled;
        },
      );
    }
  }

  // TODO: Move this to a central cached store
  $: measureFormatters = measures.reduce(
    (acc, measure) => ({
      ...acc,
      [sanitizeFieldName(measure.name || "measure")]:
        createMeasureValueFormatter<null | undefined>(measure),
    }),
    {} as Record<string, (value: number | null | undefined) => string>,
  );

  $: expressionFunctions = {
    humanize: {
      fn: (val: number) =>
        humanizeDataType(val, FormatPreset.HUMANIZE, "table"),
    },
    ...measures.reduce((acc, measure) => {
      const fieldName = sanitizeFieldName(measure.name || "measure");
      const formatter = measureFormatters[fieldName];
      return {
        ...acc,
        [fieldName]: {
          fn: (val: number) => (formatter ? formatter(val) : String(val)),
        },
      };
    }, {}),
  };

  // Color mapping needs to be reactive to theme mode changes (light/dark)
  // because colors are resolved differently for each mode
  $: isThemeModeDark = themeMode === "dark";
  $: colorMapping = getColorMappingForChart(
    chartSpec,
    domainValues,
    isThemeModeDark,
  );

  $: pieTooltipFormatter =
    isPieOrDonut && otherGrouping
      ? buildPieTooltipFormatter(
          chartSpec,
          otherGrouping,
          colorMapping,
          measureFormatters,
          isThemeModeDark,
        )
      : undefined;

  const scrubHandler = createAdaptiveScrubHandler((interval) =>
    onBrush?.(interval),
  );
  onDestroy(() => scrubHandler.destroy());

  // Signal listeners for brush and hover events
  $: signalListeners = buildSignalListeners(
    useBrush && !!vegaSpec,
    !!onHover,
    temporalField,
  );

  function buildSignalListeners(
    brushEnabled: boolean,
    hoverEnabled: boolean,
    timeField?: string,
  ): SignalListeners {
    const listeners: SignalListeners = {};

    if (hoverEnabled) {
      listeners.hover = (_name: string, value: unknown) => {
        const dimension = resolveSignalField(value, "dimension");
        const ts = resolveSignalTimeField(value, timeField);
        onHover?.(dimension, ts);
      };
    }

    if (brushEnabled) {
      listeners.brush = (_name: string, value: unknown) => {
        const interval = resolveSignalIntervalField(value);
        // Trigger async rendering to prevent race condition
        void view?.runAsync();
        if (interval) scrubHandler.update(interval);
      };

      listeners.brush_end = (_name: string, value: unknown) => {
        const interval = resolveSignalIntervalField(value);
        if (interval) {
          onBrushEnd?.(interval);
        } else {
          // Brush was cleared by clicking outside the selection
          onBrushClear?.();
        }
      };

      listeners.brush_clear = (_name: string, value: unknown) => {
        if (value) onBrushClear?.();
      };
    }

    return listeners;
  }

  function applyPieGrouping(
    spec: CanvasChartSpec,
    rawData: Record<string, unknown>[],
  ): OtherGroupingResult | null {
    const colorField =
      "color" in spec ? (spec.color?.field as string | undefined) : undefined;
    const measureField =
      "measure" in spec
        ? (spec.measure?.field as string | undefined)
        : undefined;
    if (!colorField || !measureField) return null;

    const showOther =
      "color" in spec ? spec.color?.showOther !== false : true;
    const explicitLimit =
      "color" in spec ? spec.color?.limit : undefined;

    return computeVisibleSlices(rawData, colorField, measureField, {
      explicitLimit,
      showOther,
    });
  }

  function injectOtherIntoDomain(
    domainValues: Record<string, string[] | number[] | undefined> | undefined,
    spec: CanvasChartSpec,
    grouping: OtherGroupingResult,
  ): Record<string, string[] | number[] | undefined> | undefined {
    if (!domainValues || !grouping.otherItems) return domainValues;
    const colorField =
      "color" in spec ? (spec.color?.field as string | undefined) : undefined;
    if (!colorField || !domainValues[colorField]) return domainValues;

    const existing = domainValues[colorField] as string[];
    const visibleLabels = new Set(
      grouping.visibleData
        .filter((d) => !d.__isOther)
        .map((d) => String(d[colorField])),
    );
    const filtered = existing.filter((v) => visibleLabels.has(v));
    filtered.push(OTHER_LABEL);

    return {
      ...domainValues,
      [colorField]: filtered,
    };
  }

  function buildPieTooltipFormatter(
    spec: CanvasChartSpec,
    grouping: OtherGroupingResult,
    colorMappingArr: ColorMapping,
    formatters: Record<string, (value: number | null | undefined) => string>,
    isDark: boolean,
  ): VLTooltipFormatter | undefined {
    const colorField =
      "color" in spec ? (spec.color?.field as string | undefined) : undefined;
    const measureField =
      "measure" in spec
        ? (spec.measure?.field as string | undefined)
        : undefined;
    if (!colorField || !measureField) return undefined;

    const colorMap = new Map<string, string>(
      (colorMappingArr ?? []).map((m) => [m.value, m.color]),
    );

    const measureFormatter = formatters[sanitizeFieldName(measureField)];
    const formatValue = (v: number) =>
      measureFormatter ? measureFormatter(v) : String(v);

    const mutedColor = isDark ? "#374151" : "#e5e7eb";

    return createPieTooltipFormatter({
      colorField,
      measureField,
      total: grouping.total,
      otherItems: grouping.otherItems,
      colorMap,
      formatValue,
      mutedColor,
    });
  }
</script>

{#if isFetching || measures.length === 0}
  <div class="flex items-center justify-center h-full w-full">
    <Spinner status={EntityStatus.Running} size="20px" />
  </div>
{:else if error}
  <ComponentError error={error.message} />
{:else if hasNoData}
  <div
    class="flex w-full h-full p-2 text-xl text-fg-disabled items-center justify-center"
  >
    No Data to Display
  </div>
{:else if useBrush && vegaSpec}
  <VegaRenderer
    bind:view
    data={{ "metrics-view": chartDataForSpec.data }}
    {isScrubbing}
    spec={vegaSpec}
    {colorMapping}
    theme={themeMode}
    {signalListeners}
    renderer="svg"
    {expressionFunctions}
    {hasComparison}
  />
{:else}
  <VegaLiteRenderer
    bind:viewVL={view}
    canvasDashboard={isCanvas}
    data={{ "metrics-view": chartDataForSpec.data }}
    {themeMode}
    {spec}
    {colorMapping}
    {signalListeners}
    renderer="canvas"
    {expressionFunctions}
    {hasComparison}
    tooltipFormatter={pieTooltipFormatter}
    config={getRillTheme(isThemeModeDark, theme)}
  />
{/if}
