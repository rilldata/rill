<script>
import HistogramBase from "./HistogramBase.svelte";
import { datePortion, timePortion, intervalToTimestring } from "$lib/util/formatters";
export let data;
export let interval;
export let width;
export let height = 100;

$: effectiveWidth = Math.max(width - 8, 120);

let fontSize = 12;
</script>

<div class="italic pt-1 pb-2">{intervalToTimestring(interval)}</div>
<HistogramBase separate={width > 300} color={"#14b8a6"} {data} left={0} right={0} width={effectiveWidth} {height} bottom={40}>
    <svelte:fragment let:x let:y let:buffer>
        {@const yStart = y.range()[0] + fontSize + buffer * 1.5}
        {@const yEnd = y.range()[0] + fontSize *2 + buffer * 1.75}
        {@const xStart = x.range()[0]}
        {@const xEnd = x.range()[1]}
        {@const start = new Date(x.domain()[0] * 1000)}
        {@const end = new Date(x.domain()[1] * 1000)}
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
            <text 
                x={xStart} 
                y={yEnd}
                class={isSameDay ?  emphasize : deEmphasize}
            >
                {timePortion(start)}
            </text>
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
            <text 
                text-anchor=end 
                x={xEnd} 
                y={yEnd}
                class={isSameDay ?  emphasize : deEmphasize}
            >
                {timePortion(end)}
            </text>
        </g>
    </svelte:fragment>
</HistogramBase>