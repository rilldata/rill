<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { COLUMN_PROFILE_CONFIG } from "@rilldata/web-common/layout/config";
  import type { NumericHistogramBinsBin } from "@rilldata/web-common/runtime-client";
  import { createHistogramScales } from "../histogram-utils";

  const height = 18;
  const margins = { top: 4, right: 1, bottom: 1, left: 1 };

  export let compact = false;
  export let data: NumericHistogramBinsBin[];
  export let type: string;

  $: width =
    COLUMN_PROFILE_CONFIG.summaryVizWidth[compact ? "small" : "medium"];
  $: plotLeft = margins.left;
  $: plotRight = width - margins.right;
  $: plotBottom = height - margins.bottom;
  $: plotTop = margins.top;

  $: ({ path: d } = createHistogramScales(data, type, {
    left: plotLeft,
    right: plotRight,
    top: plotTop,
    bottom: plotBottom,
  }));
</script>

{#if data}
  <Tooltip location="right" distance={8}>
    <svg class="overflow-visible" {width} {height}>
      <g class="text-primary-300">
        <line
          x1={plotLeft}
          x2={plotRight}
          y1={plotBottom}
          y2={plotBottom}
          class="text-primary-200"
          stroke="currentColor"
          stroke-width={0.5}
        />
        {#if d?.length}
          <path {d} class="fill-primary-400/40" />
          <path {d} stroke="currentColor" fill="none" stroke-width={0.5} />
        {/if}
      </g>
    </svg>
    <TooltipContent slot="tooltip-content">
      the distribution of the values of this column
    </TooltipContent>
  </Tooltip>
{/if}
