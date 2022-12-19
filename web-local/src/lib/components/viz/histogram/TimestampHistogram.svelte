<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { TIMESTAMP_TOKENS } from "@rilldata/web-common/lib/duckdb-data-types";
  import {
    datePortion,
    intervalToTimestring,
    removeTimezoneOffset,
    timePortion,
  } from "@rilldata/web-common/lib/formatters";
  import HistogramBase from "./HistogramBase.svelte";

  export let data;
  export let type;
  export let interval;
  export let width;
  export let height = 100;
  export let estimatedSmallestTimeGrain: string;

  $: effectiveWidth = Math.max(width - 8, 120);

  let fontSize = 12;

  $: timeLength = interval
    ? intervalToTimestring(
        type === "DATE" ? { days: interval, months: 0, micros: 0 } : interval
      )
    : undefined;
</script>

{#if interval}
  <div class="grid space-between grid-cols-2 items-baseline pr-6">
    <Tooltip location="top" distance={16}>
      <div class="pt-1 pb-2">
        {timeLength}
      </div>
      <TooltipContent slot="tooltip-content">
        <div style:width="240px">
          This column represents {timeLength} of data.
        </div>
      </TooltipContent>
    </Tooltip>
    {#if estimatedSmallestTimeGrain}
      <Tooltip location="top" distance={16}>
        <div class="justify-self-end text-gray-500 text-right leading-4">
          time grain in {estimatedSmallestTimeGrain}
        </div>
        <TooltipContent slot="tooltip-content">
          <div style:width="240px">
            The smallest estimated time grain of this column is at the {estimatedSmallestTimeGrain}
            level.
          </div>
        </TooltipContent>
      </Tooltip>
    {/if}
  </div>
{/if}
<HistogramBase
  separate={width > 300}
  fillColor={TIMESTAMP_TOKENS.vizFillClass}
  baselineStrokeColor={TIMESTAMP_TOKENS.vizStrokeClass}
  {data}
  left={0}
  right={0}
  width={effectiveWidth}
  {height}
  bottom={40}
>
  <svelte:fragment let:x let:y let:buffer>
    {@const yStart = y.range()[0] + fontSize + buffer * 1.5}
    {@const yEnd = y.range()[0] + fontSize * 2 + buffer * 1.75}
    {@const xStart = x.range()[0]}
    {@const xEnd = x.range()[1]}
    {@const start = removeTimezoneOffset(new Date(x.domain()[0] * 1000))}
    {@const end = removeTimezoneOffset(new Date(x.domain()[1] * 1000))}
    {@const isSameDay =
      start.getFullYear() === end.getFullYear() &&
      start.getMonth() === end.getMonth() &&
      start.getDate() === end.getDate()}
    {@const emphasize = "font-semibold"}
    {@const deEmphasize = "fill-gray-500"}
    <g>
      <text x={xStart} y={yStart} class={isSameDay ? deEmphasize : emphasize}>
        {datePortion(start)}
      </text>
      {#if type !== "DATE"}
        <text x={xStart} y={yEnd} class={isSameDay ? emphasize : deEmphasize}>
          {timePortion(start)}
        </text>
      {/if}
    </g>
    <g>
      <text
        text-anchor="end"
        x={xEnd}
        y={yStart}
        class={isSameDay ? deEmphasize : emphasize}
      >
        {datePortion(end)}
      </text>
      {#if type !== "DATE"}
        <text
          text-anchor="end"
          x={xEnd}
          y={yEnd}
          class={isSameDay ? emphasize : deEmphasize}
        >
          {timePortion(end)}
        </text>
      {/if}
    </g>
  </svelte:fragment>
</HistogramBase>
