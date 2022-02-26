<script lang="ts">
import { tweened } from 'svelte/motion';
import { cubicOut as easing } from 'svelte/easing';
import { scaleLinear } from 'd3-scale';

interface HistogramBin {
    bucket:number;
    low:number;
    high:number;
    count:number;
}

export let data:HistogramBin[];
export let width = 60;
export let height = 19;
export let time = 1000;
export let color = 'hsl(340, 70%, 70%)';

const tw = tweened(0, { duration: time, easing });
$: minX = Math.min(...data.map( d => d.low ));
$: maxX = Math.max(...data.map( d => d.high ));
$: X = scaleLinear().domain([minX, maxX]).range([0, width]);

$: yVals = data.map( d => d.count );
$: maxY = Math.max(...yVals);
$: Y = scaleLinear().domain([0, maxY]).range([height - 4, 4]);

$: tw.set(1);

function s(i, ...ts) {
    return ts[i % ts.length];
}

// scales

</script>
<svg {width} {height} shape-rendering=crispEdges>
    {#each data as {low, high, count}, i}
        <rect x={X(low)} width={X(high) - X(low)} 
        y={Y(0) * (1- $tw) +  Y(count) * $tw} 
        height={Y(0) * $tw - Y(count) * $tw} fill={color} />
    {/each}
    <line x1={0} x2={width * $tw} y1={Y(0)} y2={Y(0)} stroke={color} />
</svg>
