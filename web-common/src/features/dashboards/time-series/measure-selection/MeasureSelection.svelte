<script lang="ts">
  import { contexts } from "@rilldata/web-common/components/data-graphic/constants";
  import type {
    ScaleStore,
    SimpleDataGraphicConfiguration,
  } from "@rilldata/web-common/components/data-graphic/state/types";
  import { measureSelection } from "@rilldata/web-common/features/dashboards/time-series/measure-selection/measure-selection.ts";
  import MeasureValueMouseover from "@rilldata/web-common/features/dashboards/time-series/MeasureValueMouseover.svelte";
  import { NumberKind } from "@rilldata/web-common/lib/number-formatting/humanizer-types.ts";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import WithBisector from "web-common/src/components/data-graphic/functional-components/WithBisector.svelte";
  import { Bot } from "lucide-svelte";

  export let data;
  export let measureName: string;
  export let metricsViewName: string;
  export let xAccessor: string;
  export let yAccessor: string;
  export let internalXMin;
  export let internalXMax;
  export let mouseoverFormat;
  export let numberKind: NumberKind;
  export let inBounds: (min, max, value) => boolean;

  const plotConfig: Writable<SimpleDataGraphicConfiguration> = getContext(
    contexts.config,
  );
  const xScale = getContext<ScaleStore>(contexts.scale("x"));

  $: ({ top, bottom, plotTop, plotBottom, bodyBuffer } = $plotConfig);
  $: y1 = plotTop + top + 5;
  $: y2 = plotBottom - bottom - 1;

  $: ({ measure, start, end } = measureSelection);
  $: hasSelection = $measure && $start;
  $: forThisMeasure = hasSelection && $measure === measureName;
  $: showLine = hasSelection && !$end;
  $: showBox = forThisMeasure && $end;

  function onExplain(e) {
    e.stopPropagation();
    e.preventDefault();
    measureSelection.startAnomalyExplanationChat(metricsViewName);
  }
</script>

{#if showLine}
  <WithBisector {data} callback={(d) => d[xAccessor]} value={$start} let:point>
    {#if point && inBounds(internalXMin, internalXMax, point[xAccessor])}
      <MeasureValueMouseover
        {point}
        {xAccessor}
        {yAccessor}
        {mouseoverFormat}
        {numberKind}
        colorClass={forThisMeasure
          ? "stroke-primary-500"
          : "stroke-primary-300"}
        strokeWidth={forThisMeasure ? 3 : 2}
      />

      {#if forThisMeasure}
        <Bot
          size={16}
          class="stroke-primary cursor-pointer"
          x={$xScale(point[xAccessor]) - 35}
          y={plotBottom - 3 + bodyBuffer}
          on:click={onExplain}
        />
        <text
          role="presentation"
          class="fill-primary stroke-surface cursor-pointer hover:underline"
          style:paint-order="stroke"
          stroke-width="1px"
          x={$xScale(point[xAccessor]) - 15}
          y={plotBottom + 10 + bodyBuffer}
          on:click={onExplain}
        >
          Explain (E)
        </text>
      {/if}
    {/if}
  </WithBisector>
{:else if showBox && $start && $end}
  {@const xStart = $xScale($start)}
  {@const xEnd = $xScale($end)}
  <g role="presentation" opacity="0.2">
    <rect
      x={xStart}
      y={y1}
      width={xEnd - xStart}
      height={y2 - y1}
      fill="url('#scrub-selection-gradient')"
    />
  </g>

  <Bot
    size={16}
    class="stroke-primary cursor-pointer"
    x={(xStart + xEnd) / 2 - 35}
    y={plotBottom - 3 + bodyBuffer}
    on:click={onExplain}
  />
  <text
    role="presentation"
    class="fill-primary stroke-surface cursor-pointer hover:underline"
    style:paint-order="stroke"
    stroke-width="1px"
    x={(xStart + xEnd) / 2 - 15}
    y={plotBottom + 10 + bodyBuffer}
    on:click={onExplain}
  >
    Explain (E)
  </text>

  <defs>
    <linearGradient
      gradientUnits="userSpaceOnUse"
      id="scrub-selection-gradient"
    >
      <stop stop-color="var(--color-theme-400)" />
      <stop offset="0.36" stop-color="var(--color-theme-300)" />
      <stop offset="1" stop-color="var(--color-theme-200)" />
    </linearGradient>
  </defs>
{/if}
