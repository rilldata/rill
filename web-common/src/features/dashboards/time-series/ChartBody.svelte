<script lang="ts">
  import { writable } from "svelte/store";
  import {
    ChunkedLine,
    ClippedChunkedLine,
  } from "@rilldata/web-common/components/data-graphic/marks";
  import { previousValueStore } from "@rilldata/web-common/lib/store-utils";

  export let xMin: Date = undefined;
  export let xMax: Date = undefined;
  export let yExtentMax: number = undefined;
  export let showComparison;
  export let isHovering;
  export let data;
  export let dimensionData;
  export let xAccessor;
  export let yAccessor;
  export let scrubStart;
  export let scrubEnd;

  $: hasSubrangeSelected = Boolean(scrubStart && scrubEnd);

  $: mainLineColor = hasSubrangeSelected
    ? "hsla(217, 10%, 60%, 1)"
    : "hsla(217,60%, 55%, 1)";

  $: areaColor = hasSubrangeSelected
    ? "hsla(225, 20%, 80%, .2)"
    : "hsla(217,70%, 80%, .4)";

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
    {#each dimensionData as dimensionValue}
      <ChunkedLine
        area={false}
        delay={$timeRangeKey !== $previousTimeRangeKey ? 0 : delay}
        duration={hasSubrangeSelected || $timeRangeKey !== $previousTimeRangeKey
          ? 0
          : 200}
        lineClasses={dimensionValue?.strokeClass}
        data={dimensionValue?.data || []}
        {xAccessor}
        {yAccessor}
      />
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
          lineColor={`hsl(217, 10%, 60%)`}
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
        lineColor="hsla(217,60%, 55%, 1)"
        areaColor="hsla(217,70%, 80%, .4)"
        delay={0}
        duration={0}
        {data}
        {xAccessor}
        {yAccessor}
      />
    {/if}
  {/if}
{/key}
