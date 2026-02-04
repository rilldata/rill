<script lang="ts">
  import { extent, max, min } from "d3-array";
  import { scaleLinear } from "d3-scale";
  import { barplotPolyline } from "@rilldata/web-common/components/data-graphic/utils";
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { COLUMN_PROFILE_CONFIG } from "@rilldata/web-common/layout/config";
  import { INTEGERS } from "@rilldata/web-common/lib/duckdb-data-types";
  import type { NumericHistogramBinsBin } from "@rilldata/web-common/runtime-client";

  const gradientId = `spark-gradient-${guidGenerator()}`;
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

  $: xMin = min(data, (d) => d.low);
  $: xMax = max(data, (d) => d.high);
  $: [, yMax] = extent(data, (d) => d.count);

  $: xScale = scaleLinear()
    .domain([xMin ?? 0, xMax ?? 1])
    .range([plotLeft, plotRight]);
  $: yScale = scaleLinear()
    .domain([0, yMax ?? 1])
    .range([plotBottom, plotTop]);

  $: separator = data.length < 20 && INTEGERS.has(type) ? 0.25 : 0;

  $: d = barplotPolyline(data, xScale, yScale, separator, false, 1);
</script>

{#if data}
  <Tooltip location="right" distance={8}>
    <svg class="overflow-visible" {width} {height}>
      <defs>
        <linearGradient id={gradientId} x1="0" x2="0" y1="0" y2="1">
          <stop offset="5%" stop-color="var(--color-primary-600" />
          <stop
            offset="95%"
            stop-color="var(--surface-background)"
            stop-opacity={0.4}
          />
        </linearGradient>
      </defs>
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
          <path {d} fill="url(#{gradientId})" />
          <path {d} stroke="currentColor" fill="none" stroke-width={0.5} />
        {/if}
      </g>
    </svg>
    <TooltipContent slot="tooltip-content">
      the distribution of the values of this column
    </TooltipContent>
  </Tooltip>
{/if}
