<script lang="ts">
  import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
  import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
  import { compileFlintSpec } from "@rilldata/web-common/features/custom-viz/flint/compile";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    createMeasureValueFormatter,
    humanizeDataType,
  } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { AlertCircle } from "lucide-svelte";
  import { onMount, untrack } from "svelte";
  import type { AdaptedFlintChart } from "../lib/adapt";
  import type { CompileNotices } from "../lib/notices";
  import Skeleton from "./Skeleton.svelte";

  /** Vega draws text on a canvas, so the first paint waits for the host font (briefly). */
  const FONT_WAIT_MS = 300;
  /** Vega resizes the live view at once; Flint only needs the settled size to recompute its layout. */
  const COMPILE_DEBOUNCE_MS = 150;

  let {
    chart,
    isDark,
    fontFamily,
    height,
    fill = false,
    dimmed = false,
    onCompiled,
  }: {
    chart: AdaptedFlintChart;
    isDark: boolean;
    fontFamily: string;
    /** Fixed height in pixels; ignored when `fill` is set. */
    height?: number;
    /** Fill the parent (fullscreen layout) instead of using a fixed height. */
    fill?: boolean;
    /** Fades the current chart under a skeleton while new data loads. */
    dimmed?: boolean;
    /** Reports the compile outcome so the card can show Flint's warnings or the error. */
    onCompiled?: (notices: CompileNotices) => void;
  } = $props();

  let fontsReady = $state(false);
  let width = $state(0);
  let boxHeight = $state(0);
  // The measured box Flint lays out for, updated once a resize settles.
  let compileSize = $state.raw<{ width: number; height: number } | undefined>(
    undefined,
  );

  onMount(() => {
    let cancelled = false;
    const fonts = document.fonts?.ready ?? Promise.resolve();
    const timeout = new Promise<void>((resolve) =>
      setTimeout(resolve, FONT_WAIT_MS),
    );
    void Promise.race([fonts, timeout]).then(() => {
      if (!cancelled) fontsReady = true;
    });
    return () => {
      cancelled = true;
    };
  });

  // The first measurement compiles at once, so the chart embeds a single time; later
  // size changes wait for the resize to settle before the spec is recompiled.
  $effect(() => {
    const next =
      width > 0 && boxHeight > 0 ? { width, height: boxHeight } : undefined;
    if (!next) return;
    const current = untrack(() => compileSize);
    if (!current) {
      compileSize = next;
      return;
    }
    if (next.width === current.width && next.height === current.height) return;
    const timer = setTimeout(() => {
      compileSize = next;
    }, COMPILE_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  });

  // Flint reads the rows and the box to make layout decisions, so the spec is compiled once
  // the box is measured and recompiled per settled size.
  const compiled = $derived(
    compileSize
      ? compileFlintSpec({
          spec: chart.flint.spec,
          rows: chart.rows,
          fields: chart.flint.fields,
          size: compileSize,
          measureFields: chart.flint.measureFields,
        })
      : undefined,
  );

  $effect(() => {
    if (!compiled) return;
    const notices: CompileNotices = {
      error: compiled.error ?? undefined,
      warnings: compiled.warnings
        .filter((warning) => warning.severity !== "info")
        .map((warning) => warning.message),
    };
    untrack(() => onCompiled?.(notices));
  });

  // One Vega expression function per measure (axis and tooltip formatting) plus `humanize`,
  // the same registration `Chart.svelte` performs; `compileFlintSpec` points the measure
  // axes at them (`rill_<measure>` format types).
  const expressionFunctions = $derived.by(() => {
    const functions: Record<string, { fn: (value: number) => string }> = {
      humanize: {
        fn: (value: number) =>
          humanizeDataType(value, FormatPreset.HUMANIZE, "table"),
      },
    };
    for (const measure of chart.measures) {
      const formatter = createMeasureValueFormatter<null | undefined>(measure);
      functions[sanitizeFieldName(measure.name ?? "measure")] = {
        fn: (value: number) => formatter(value) ?? String(value),
      };
    }
    return functions;
  });

  const config = $derived(getRillTheme(isDark, chart.theme, { fontFamily }));
</script>

<div
  class="surface"
  class:fill
  class:dimmed
  style:height={fill ? undefined : `${height ?? 300}px`}
  aria-busy={dimmed}
  bind:clientWidth={width}
  bind:clientHeight={boxHeight}
>
  {#if compiled?.error}
    <div class="compile-error" role="alert">
      <AlertCircle size="18" aria-hidden="true" />
      <p>{m.mcp_chart_error_render({ message: compiled.error })}</p>
    </div>
  {:else if fontsReady && compiled?.spec}
    <div class="chart">
      <VegaLiteRenderer
        spec={compiled.spec}
        renderer="canvas"
        themeMode={isDark ? "dark" : "light"}
        {config}
        {expressionFunctions}
      />
    </div>
  {/if}
  {#if dimmed}
    <div class="overlay">
      <Skeleton compact label={m.mcp_chart_updating()} />
    </div>
  {/if}
</div>

<style lang="postcss">
  .surface {
    @apply relative w-full min-w-0 overflow-hidden px-2;
  }
  .surface.fill {
    @apply flex-1 min-h-0;
  }
  .chart {
    @apply size-full transition-opacity;
  }
  .dimmed .chart {
    opacity: 0.35;
  }
  .overlay {
    @apply absolute inset-0 flex items-center justify-center p-4;
  }
  .compile-error {
    @apply flex size-full flex-col items-center justify-center gap-1.5 px-6 text-center text-fg-tertiary;
  }
  .compile-error p {
    @apply m-0 max-w-sm break-words text-xs leading-4;
  }
</style>
