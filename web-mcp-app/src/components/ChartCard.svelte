<script lang="ts">
  import type { AdaptedChart } from "../lib/adapt";
  import CardFooter from "./CardFooter.svelte";
  import CardHeader from "./CardHeader.svelte";
  import ChartSurface from "./ChartSurface.svelte";
  import TruncatedBanner from "./TruncatedBanner.svelte";

  let {
    chart,
    isDark,
    fontFamily,
    chartHeight,
    canFullscreen,
    openUrl,
    refreshing = false,
    refreshError,
    onOpenLink,
    onFullscreen,
  }: {
    chart: AdaptedChart;
    isDark: boolean;
    fontFamily: string;
    chartHeight: number;
    canFullscreen: boolean;
    openUrl: string | undefined;
    refreshing?: boolean;
    refreshError?: string;
    onOpenLink: () => void;
    onFullscreen: () => void;
  } = $props();
</script>

<section class="card" aria-busy={refreshing}>
  <CardHeader title={chart.title} caption={chart.caption} />
  <ChartSurface
    {chart}
    {isDark}
    {fontFamily}
    height={chartHeight}
    dimmed={refreshing}
  />
  {#if chart.truncationNote}
    <TruncatedBanner
      note={chart.truncationNote}
      detail={chart.truncationWarning}
    />
  {/if}
  {#if refreshError}
    <p class="notice" role="status">{refreshError}</p>
  {/if}
  <CardFooter
    {openUrl}
    {onOpenLink}
    secondary={canFullscreen ? "fullscreen" : undefined}
    onSecondary={onFullscreen}
  />
</section>

<style lang="postcss">
  .card {
    @apply flex w-full flex-col min-w-0;
  }
  .notice {
    @apply mx-4 my-1 text-xs text-fg-tertiary;
  }
</style>
