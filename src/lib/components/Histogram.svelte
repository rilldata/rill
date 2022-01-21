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

const t1 = tweened(0, { duration: time, easing });
const t2 = tweened(0, { duration: time * (1 + Math.random() / 5), easing, delay: time / 6 });
const t3 = tweened(0, { duration: time * (1 + Math.random() / 3), easing, delay: time / 3});
const t4 = tweened(0, { duration: time * (1 + Math.random() / 1.5), easing });

$: minX = Math.min(...data.map( d => d.low ));
$: maxX = Math.max(...data.map( d => d.high ));
$: X = scaleLinear().domain([minX, maxX]).range([0, width]);

$: yVals = data.map( d => d.count );
$: maxY = Math.max(...yVals);
$: Y = scaleLinear().domain([0, maxY]).range([height - 4, 4]);

$: t1.set(1);
$: t2.set(1);
$: t3.set(1);
$: t4.set(1);

function s(i, ...ts) {
    return ts[i % ts.length];
}

// scales

</script>
<svg {width} {height} shape-rendering=crispEdges>
    {#each data as {low, high, count}, i}
        <rect x={X(low)} width={X(high) - X(low)} 
        y={Y(0) * (1-s(i, $t1, $t2, $t3, $t4)) +  Y(count) * s(i, $t1, $t2, $t3, $t4)} 
        height={Y(0) * (s(i, $t1, $t2, $t3, $t4)) - Y(count) * (s(i, $t1, $t2, $t3, $t4))} fill={color} />
    {/each}
    <line x1={0} x2={width * $t1} y1={Y(0)} y2={Y(0)} stroke={color} />
</svg>
