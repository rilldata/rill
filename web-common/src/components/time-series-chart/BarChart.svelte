<script lang="ts">
  import type { ChartSeries } from "@rilldata/web-common/features/dashboards/time-series/measure-chart/types";
  import { computeBarSlotGeometry } from "@rilldata/web-common/features/dashboards/time-series/measure-chart/utils";
  import type { ScaleLinear } from "d3-scale";

  export let series: ChartSeries[];
  export let yScale: ScaleLinear<number, number>;
  export let stacked: boolean = false;
  export let plotLeft: number;
  export let plotWidth: number;
  export let visibleStart: number;
  export let visibleEnd: number;
  export let scrubStartIndex: number | null = null;
  export let scrubEndIndex: number | null = null;

  $: hasScrub = scrubStartIndex !== null && scrubEndIndex !== null;
  $: scrubMin = hasScrub ? Math.min(scrubStartIndex!, scrubEndIndex!) : 0;
  $: scrubMax = hasScrub ? Math.max(scrubStartIndex!, scrubEndIndex!) : 0;

  $: visibleCount = Math.max(1, visibleEnd - visibleStart + 1);
  $: geo = computeBarSlotGeometry(plotWidth, visibleCount, series.length);
  $: zeroY = yScale(0);

  function isInScrub(ptIdx: number): boolean {
    if (!hasScrub) return true;
    return ptIdx >= Math.round(scrubMin) && ptIdx <= Math.round(scrubMax);
  }
</script>

{#if stacked}
  {#each { length: visibleCount } as _, slot (slot)}
    {@const ptIdx = visibleStart + slot}
    {@const cx = plotLeft + (slot + 0.5) * geo.slotWidth}
    {@const bx = cx - geo.bandWidth / 2}
    {@const stackValues = series.map((s) => ({
      value: s.values[ptIdx] ?? 0,
      color: s.color,
      id: s.id,
    }))}
    {#each stackValues as seg, segIdx (seg.id)}
      {#if seg.value !== 0}
        {@const yBottom = yScale(
          stackValues.slice(0, segIdx).reduce((sum, sv) => sum + sv.value, 0),
        )}
        {@const yTop = yScale(
          stackValues
            .slice(0, segIdx + 1)
            .reduce((sum, sv) => sum + sv.value, 0),
        )}
        <rect
          x={bx}
          y={Math.min(yBottom, yTop)}
          width={geo.bandWidth}
          height={Math.abs(yBottom - yTop)}
          fill={isInScrub(ptIdx) ? seg.color : "var(--color-gray-400)"}
          opacity={1}
          rx={1}
        />
      {/if}
    {/each}
  {/each}
{:else}
  {@const radius = 4}
  {#each { length: visibleCount } as _, slot (slot)}
    {@const ptIdx = visibleStart + slot}
    {@const cx = plotLeft + (slot + 0.5) * geo.slotWidth}
    {#each series as s, sIdx (s.id)}
      {@const v = s.values[ptIdx] ?? null}
      {#if v !== null}
        {@const bx =
          cx - geo.bandWidth / 2 + sIdx * (geo.singleBarWidth + geo.barGap)}
        {@const by = Math.min(zeroY, yScale(v))}
        {@const bh = Math.abs(zeroY - yScale(v))}
        {@const r = Math.min(radius, geo.singleBarWidth / 2, bh / 2)}
        {@const isPositive = v >= 0}
        <path
          d={isPositive
            ? `M${bx},${by + bh}
               V${by + r}
               Q${bx},${by} ${bx + r},${by}
               H${bx + geo.singleBarWidth - r}
               Q${bx + geo.singleBarWidth},${by} ${bx + geo.singleBarWidth},${by + r}
               V${by + bh}
               Z`
            : `M${bx},${by}
               V${by + bh - r}
               Q${bx},${by + bh} ${bx + r},${by + bh}
               H${bx + geo.singleBarWidth - r}
               Q${bx + geo.singleBarWidth},${by + bh} ${bx + geo.singleBarWidth},${by + bh - r}
               V${by}
               Z`}
          fill={isInScrub(ptIdx) ? s.color : "var(--color-gray-400)"}
          opacity={isInScrub(ptIdx) ? (s.opacity ?? 1) : 0.5}
        />
      {/if}
    {/each}
  {/each}
{/if}
