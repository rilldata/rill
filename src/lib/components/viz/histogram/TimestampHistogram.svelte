<script>
import HistogramBase from "./HistogramBase.svelte";
import { datePortion, timePortion, intervalToTimestring, removeTimezoneOffset } from "$lib/util/formatters";
import { TIMESTAMP_TOKENS } from "$lib/duckdb-data-types";
export let data;
export let type;
export let interval;
export let width;
export let height = 100;

$: effectiveWidth = Math.max(width - 8, 120);

let fontSize = 12;
</script>

{#if interval}
<div class="italic pt-1 pb-2">{intervalToTimestring(type === "DATE" ? {days: interval, months: 0, micros: 0 } : interval)}</div>
{/if}
<HistogramBase separate={width > 300} fillColor={TIMESTAMP_TOKENS.vizFillClass} baselineStrokeColor={TIMESTAMP_TOKENS.vizStrokeClass} {data} left={0} right={0} width={effectiveWidth} {height} bottom={40}>
    <svelte:fragment let:x let:y let:buffer>
        {@const yStart = y.range()[0] + fontSize + buffer * 1.5}
        {@const yEnd = y.range()[0] + fontSize *2 + buffer * 1.75}
        {@const xStart = x.range()[0]}
        {@const xEnd = x.range()[1]}
        {@const start = removeTimezoneOffset(new Date(x.domain()[0] * 1000))}
        {@const end = removeTimezoneOffset(new Date(x.domain()[1] * 1000))}
        {@const isSameDay = 
            start.getFullYear() === end.getFullYear() &&
            start.getMonth() === end.getMonth() &&
            start.getDate() === end.getDate()}
        {@const emphasize = 'font-semibold'}
        {@const deEmphasize = 'fill-gray-500 italic'}
        <g>
            <text 
                x={xStart} y={yStart}
                class={isSameDay ?  deEmphasize : emphasize}
            >
                {datePortion(start)}
            </text>
            {#if type !== 'DATE'}
            <text 
                x={xStart} 
                y={yEnd}
                class={isSameDay ?  emphasize : deEmphasize}
            >
                {timePortion(start)}
            </text>
            {/if}
        </g>
        <g>
            <text
                text-anchor=end 
                x={xEnd} 
                y={yStart}
                class={isSameDay ?  deEmphasize : emphasize}

            >
                {datePortion(end)}
            </text>
            {#if type !== 'DATE'}
            <text 
                text-anchor=end 
                x={xEnd} 
                y={yEnd}
                class={isSameDay ?  emphasize : deEmphasize}
            >
                {timePortion(end)}
            </text>
            {/if}

        </g>
    </svelte:fragment>
</HistogramBase>