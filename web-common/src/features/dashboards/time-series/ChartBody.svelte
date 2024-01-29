<script lang="ts">
  import {
    AreaMutedColor,
    MainAreaColor,
    LineMutedColor,
    MainLineColor,
    TimeComparisonLineColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import { writable } from "svelte/store";
  import {
    ChunkedLine,
    ClippedChunkedLine,
  } from "@rilldata/web-common/components/data-graphic/marks";
  import { previousValueStore } from "@rilldata/web-common/lib/store-utils";
  import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";

  export let xMin: Date | undefined = undefined;
  export let xMax: Date | undefined = undefined;
  export let yExtentMax: number | undefined = undefined;
  export let showComparison: boolean;
  export let dimensionValue: string;
  export let isHovering: boolean;
  export let data;
  export let dimensionData: DimensionDataItem[] = [];
  export let xAccessor: string;
  export let yAccessor: string;
  export let scrubStart;
  export let scrubEnd;

  $: hasSubrangeSelected = Boolean(scrubStart && scrubEnd);

  $: mainLineColor = hasSubrangeSelected ? LineMutedColor : MainLineColor;

  $: areaColor = hasSubrangeSelected ? AreaMutedColor : MainAreaColor;

  $: isDimValueHiglighted =
    dimensionValue !== undefined &&
    dimensionData.map((d) => d.value).includes(dimensionValue);

  // we delay the tween if previousYMax < yMax
  let yMaxStore = writable(yExtentMax);
  let previousYMax = previousValueStore(yMaxStore);

  $: yMaxStore.set(yExtentMax);
  const timeRangeKey = writable(`${xMin}-${xMax}`);

  const previousTimeRangeKey = previousValueStore(timeRangeKey);

  // FIXME: move this function to utils.ts
  /** reset the keys to trigger animations on time range changes */
  let syncTimeRangeKey;
  $: {
    timeRangeKey.set(`${xMin}-${xMax}`);
    if ($previousTimeRangeKey !== $timeRangeKey) {
      if (syncTimeRangeKey) clearTimeout(syncTimeRangeKey);
      syncTimeRangeKey = setTimeout(() => {
        previousTimeRangeKey.set($timeRangeKey);
      }, 400);
    }
  }

  $: delay =
    $previousTimeRangeKey === $timeRangeKey && $previousYMax < yExtentMax
      ? 100
      : 0;
</script>

<!-- key on the time range itself to prevent weird tweening animations.
    We'll need to migrate this to a more robust solution once we've figured out
    the right way to "tile" together a time series with multiple pages of data.
    -->
{#key $timeRangeKey}
  {#if dimensionData?.length}
    {#each dimensionData as d}
      {@const isHighlighted = d?.value === dimensionValue}
      <g
        class="transition-opacity"
        class:opacity-20={isDimValueHiglighted && !isHighlighted}
      >
        <ChunkedLine
          area={false}
          isComparingDimension
          delay={$timeRangeKey !== $previousTimeRangeKey ? 0 : delay}
          duration={hasSubrangeSelected ||
          $timeRangeKey !== $previousTimeRangeKey
            ? 0
            : 200}
          lineClasses={d?.strokeClass}
          data={d?.data || []}
          {xAccessor}
          {yAccessor}
        />
      </g>
    {/each}
  {:else}
    {#if showComparison}
      <g
        class="transition-opacity"
        class:opacity-80={isHovering}
        class:opacity-40={!isHovering}
      >
        <ChunkedLine
          area={false}
          lineColor={TimeComparisonLineColor}
          delay={$timeRangeKey !== $previousTimeRangeKey ? 0 : delay}
          duration={hasSubrangeSelected ||
          $timeRangeKey !== $previousTimeRangeKey
            ? 0
            : 200}
          {data}
          {xAccessor}
          yAccessor="comparison.{yAccessor}"
        />
      </g>
    {/if}
    <ChunkedLine
      lineColor={mainLineColor}
      {areaColor}
      delay={$timeRangeKey !== $previousTimeRangeKey ? 0 : delay}
      duration={hasSubrangeSelected || $timeRangeKey !== $previousTimeRangeKey
        ? 0
        : 200}
      {data}
      {xAccessor}
      {yAccessor}
    />
    {#if hasSubrangeSelected}
      <ClippedChunkedLine
        start={Math.min(scrubStart, scrubEnd)}
        end={Math.max(scrubStart, scrubEnd)}
        lineColor={MainLineColor}
        areaColor={MainAreaColor}
        delay={0}
        duration={0}
        {data}
        {xAccessor}
        {yAccessor}
      />
    {/if}
  {/if}
{/key}
