<script lang="ts">
  import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
  import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
  import {
    resolveSignalField,
    resolveSignalIntervalField,
    resolveSignalTimeField,
  } from "@rilldata/web-common/components/vega/vega-signals";
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
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
  import type { SignalListeners, View } from "svelte-vega";
  import type { Readable } from "svelte/store";
  import { getChroma } from "../../themes/theme-utils";
  import { discoverTemporalBrushSignal } from "./brush-builder";
  import type { ChartDataResult, ChartType } from "./types";
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

  export let temporalField: string | undefined = undefined;
  export let onBrushEnd: ((interval: TimeRange) => void) | undefined =
    undefined;
  export let onBrushClear: (() => void) | undefined = undefined;
  export let onHover:
    | ((dimension: string | null | undefined, time: Date | undefined) => void)
    | undefined = undefined;

  export let view: View;

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

  $: rawSpec = generateSpec(chartType, chartSpec, chartDataWithTheme);

  // Memoize spec with deep equality so VegaLiteRenderer doesn't recreate the
  // view (and kill brush state) on store re-emissions that produce the same spec.
  let spec: ReturnType<typeof generateSpec> = {};
  $: if (JSON.stringify(rawSpec) !== JSON.stringify(spec)) {
    spec = rawSpec;
  }

  $: useBrush = "isInteractive" in chartSpec && !!chartSpec.isInteractive;

  // Read brushTemporalField from the VL spec's usermeta (set by spec generators)
  $: brushTemporalField =
    spec && typeof spec === "object" && "usermeta" in spec
      ? (spec.usermeta as { brushTemporalField?: string })?.brushTemporalField
      : undefined;

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

  // Hover signal listeners (passed declaratively to VegaLiteRenderer)
  $: signalListeners = buildHoverListeners(!!onHover, temporalField);

  function buildHoverListeners(
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
    return listeners;
  }

  // Brush-end and brush-clear detection.
  // The temporal brush signal is discovered from the live view because its name
  // includes a timeUnit prefix that varies (e.g. brush_yearmonthdatehours___time).
  let pointerUpHandler: (() => void) | undefined;
  let clearHandler: ((name: string, value: unknown) => void) | undefined;
  let currentBrushSignal: string | undefined;

  function attachBrushListener(v: View) {
    detachBrushListener();

    const signalName = discoverTemporalBrushSignal(v, brushTemporalField);
    if (!signalName) return;
    currentBrushSignal = signalName;

    // Detect brush-end via DOM pointerup
    pointerUpHandler = () => {
      try {
        const value = v.signal(signalName);
        const interval = resolveSignalIntervalField(value);
        if (interval) {
          onBrushEnd?.(interval);
        }
      } catch {
        // view may have been finalized
      }
    };
    window.addEventListener("pointerup", pointerUpHandler);

    // Detect brush-clear (user clicks outside brush or double-clicks)
    clearHandler = (_name: string, value: unknown) => {
      if (value === null || value === undefined) {
        onBrushClear?.();
      }
    };
    v.addSignalListener(signalName, clearHandler);
  }

  function detachBrushListener() {
    if (pointerUpHandler) {
      window.removeEventListener("pointerup", pointerUpHandler);
      pointerUpHandler = undefined;
    }
    if (view && currentBrushSignal && clearHandler) {
      try {
        view.removeSignalListener(currentBrushSignal, clearHandler);
      } catch {
        // view may have been finalized
      }
    }
    clearHandler = undefined;
    currentBrushSignal = undefined;
  }

  $: if (useBrush && view) {
    attachBrushListener(view);
  }

  onDestroy(() => {
    detachBrushListener();
  });
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
{:else}
  <VegaLiteRenderer
    bind:viewVL={view}
    canvasDashboard={isCanvas}
    data={{ "metrics-view": data }}
    {themeMode}
    {spec}
    {colorMapping}
    {signalListeners}
    renderer={useBrush ? "svg" : "canvas"}
    {expressionFunctions}
    {hasComparison}
    config={getRillTheme(isThemeModeDark, theme)}
  />
{/if}
