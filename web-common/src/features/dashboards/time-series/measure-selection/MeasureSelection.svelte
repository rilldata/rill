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

  export let data;
  export let measureName: string;
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

  $: ({ top, plotTop, plotBottom } = $plotConfig);
  $: y1 = plotTop + top + 5;
  $: y2 = plotBottom - 5;

  $: ({ measure, start, end } = measureSelection);
  $: hasSelection = $measure && $start;
  $: forThisMeasure = hasSelection && $measure === measureName;
  $: showLine = hasSelection && !$end;
  $: showBox = forThisMeasure && $end;

  $: if ($start)
    measureSelection.calculatePoint($start, $end, $xScale, $plotConfig);
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
