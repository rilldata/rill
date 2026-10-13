<script lang="ts">
  import Chart from "@rilldata/web-common/features/components/charts/Chart.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { onMount, type ComponentProps } from "svelte";
  import type { AdaptedChart } from "../lib/adapt";
  import type { CompileNotices } from "../lib/notices";
  import FlintSurface from "./FlintSurface.svelte";
  import Skeleton from "./Skeleton.svelte";

  /** Vega draws text on a canvas, so the first paint waits for the host font (briefly). */
  const FONT_WAIT_MS = 300;

  let {
    chart,
    isDark,
    fontFamily,
    height,
    fill = false,
    dimmed = false,
    onCompiled,
  }: {
    chart: AdaptedChart;
    isDark: boolean;
    fontFamily: string;
    /** Fixed height in pixels; ignored when `fill` is set. */
    height?: number;
    /** Fill the parent (fullscreen layout) instead of using a fixed height. */
    fill?: boolean;
    /** Fades the current chart under a skeleton while new data loads. */
    dimmed?: boolean;
    /** Flint charts report their compile outcome (warnings, error) here. */
    onCompiled?: (notices: CompileNotices) => void;
  } = $props();

  let fontsReady = $state(false);

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

  type ChartSpecProp = ComponentProps<typeof Chart>["chartSpec"];
</script>

{#if chart.engine === "flint"}
  <FlintSurface
    {chart}
    {isDark}
    {fontFamily}
    {height}
    {fill}
    {dimmed}
    {onCompiled}
  />
{:else}
  <div
    class="surface"
    class:fill
    class:dimmed
    style:height={fill ? undefined : `${height ?? 300}px`}
    aria-busy={dimmed}
  >
    {#if fontsReady}
      <div class="chart">
        <Chart
          chartType={chart.chartType}
          chartSpec={chart.spec as ChartSpecProp}
          chartData={chart.chartData}
          measures={chart.measures}
          themeMode={isDark ? "dark" : "light"}
          theme={chart.theme}
          isCanvas={false}
          {fontFamily}
        />
      </div>
    {/if}
    {#if dimmed}
      <div class="overlay">
        <Skeleton compact label={m.mcp_chart_updating()} />
      </div>
    {/if}
  </div>
{/if}

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
</style>
