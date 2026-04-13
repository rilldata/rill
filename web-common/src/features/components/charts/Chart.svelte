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
  import type { SignalListeners, VegaSpec, View } from "svelte-vega";
  import type { Readable } from "svelte/store";
  import { getChroma } from "../../themes/theme-utils";
  import {
    compileToBrushedVegaSpec,
    createAdaptiveScrubHandler,
  } from "./brush-builder";
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
  export let onBrush: ((interval: TimeRange) => void) | undefined = undefined;
  export let onBrushEnd: ((interval: TimeRange) => void) | undefined =
    undefined;
  export let onBrushClear: (() => void) | undefined = undefined;
  export let onHover:
    | ((dimension: string | null | undefined, time: Date | undefined) => void)
    | undefined = undefined;

  export let view: View;

  // Bundled to avoid a race: if vegaSpec and temporalBrushSignal were separate
  // reactive vars, Svelte could flush signalListeners between the two assignments,
  // registering listeners with the wrong (stale) signal name.
  let brushState: { spec: VegaSpec; temporalBrushSignal: string } | undefined =
    undefined;
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

  $: spec = generateSpec(chartType, chartSpec, chartDataWithTheme);

  // Compile VL spec to Vega spec when brush is enabled.
  // Required because the expression interpreter (CSP compliance) does not support
  // Vega-Lite's selection functions (vlSelectionResolve, vlSelectionTest, etc.).
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
          if (gen === compileGeneration) {
            brushState = compiled;
          }
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

  const scrubHandler = createAdaptiveScrubHandler((interval) =>
    onBrush?.(interval),
  );
  onDestroy(() => scrubHandler.destroy());

  // Signal listeners: hover for both renderers, temporal brush signal for Vega renderer.
  // brush_end/brush_clear are handled via DOM events (no injected signals).
  $: signalListeners = buildSignalListeners(
    useBrush && !!brushState,
    !!onHover,
    temporalField,
    brushState?.temporalBrushSignal,
  );

  function buildSignalListeners(
    brushEnabled: boolean,
    hoverEnabled: boolean,
    timeField?: string,
    brushSignal?: string,
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
      listeners[brushSignal ?? "brush_ts"] = (
        _name: string,
        value: unknown,
      ) => {
        const interval = resolveSignalIntervalField(value);
        if (interval) {
          scrubHandler.update(interval);
        } else {
          onBrushClear?.();
        }
      };
    }

    return listeners;
  }

  // DOM event listeners for brush-end (pointerup).
  // These replace the injected brush_end/brush_clear signals, avoiding
  // Vega signal parse-order issues with dynamically named temporal signals.
  let pointerUpHandler: (() => void) | undefined;

  function readBrushInterval(): { start: Date; end: Date } | undefined {
    if (!view || !brushState) return undefined;
    try {
      const value = view.signal(brushState.temporalBrushSignal);
      return resolveSignalIntervalField(value);
    } catch {
      return undefined;
    }
  }

  function attachDomListeners() {
    detachDomListeners();
    pointerUpHandler = () => {
      const interval = readBrushInterval();
      if (interval) {
        onBrushEnd?.(interval);
      }
    };
    window.addEventListener("pointerup", pointerUpHandler);
  }

  function detachDomListeners() {
    if (pointerUpHandler) {
      window.removeEventListener("pointerup", pointerUpHandler);
      pointerUpHandler = undefined;
    }
  }

  // Attach DOM listeners when brush view is ready
  $: if (useBrush && brushState && view) {
    attachDomListeners();
  }

  onDestroy(() => {
    detachDomListeners();
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
{:else if useBrush && brushState}
  <VegaRenderer
    bind:view
    data={{ "metrics-view": data }}
    isScrubbing={false}
    spec={brushState.spec}
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
    data={{ "metrics-view": data }}
    {themeMode}
    {spec}
    {colorMapping}
    {signalListeners}
    renderer="canvas"
    {expressionFunctions}
    {hasComparison}
    config={getRillTheme(isThemeModeDark, theme)}
  />
{/if}
