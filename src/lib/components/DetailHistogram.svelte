<script lang="ts">
    import { tweened } from 'svelte/motion';
    import { fade } from "svelte/transition";
    import { cubicOut as easing } from 'svelte/easing';
    import { scaleLinear } from 'd3-scale';
    import { format } from "d3-format";

    
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
    export let dataType = 'int';
    export let buffer = 22;

    // what do we have here? min, q25, q50, mean, q75, max
    
    const t1 = tweened(0, { duration: time, easing });
    const t2 = tweened(0, { duration: time * (1 + Math.random() / 5), easing, delay: time / 6 });
    const t3 = tweened(0, { duration: time * (1 + Math.random() / 3), easing, delay: time / 3});
    const t4 = tweened(0, { duration: time * (1 + Math.random() / 1.5), easing });
    
    const lowValue = tweened(0, { duration: time / 2, easing });
    const highValue = tweened(0, { duration: time / 2, easing });

    $: minX = Math.min(...data.map( d => d.low ));
    $: maxX = Math.max(...data.map( d => d.high ));
    $: X = scaleLinear().domain([minX, maxX]).range([0, width]);

    $: yVals = data.map( d => d.count );
    $: minY = Math.min(...yVals);
    $: maxY = Math.max(...yVals);
    $: Y = scaleLinear().domain([minY, maxY]).range([height - 4 - buffer, 4]);
    
    $: t1.set(1);
    $: t2.set(1);
    $: t3.set(1);
    $: t4.set(1);
    
    function s(i, ...ts) {
        return ts[i % ts.length];
    }

    $: tweeningFunction = dataType === 'int' ? (v:number) => ~~v : (v:number) => v;

    let formatter:Function;
    $: formatter = dataType === 'int' ? format('') : format('.2d');
    $: $lowValue = data[0].low;
    $: $highValue = data.slice(-1)[0].high;
    $: formattedLowValue = formatter(tweeningFunction($lowValue));
    $: formattedHighValue = formatter(tweeningFunction($highValue));
    

    // get lanes

    </script>
    <svg {width} height={height} >
        <!-- histogram -->
        <g shape-rendering=crispEdges>
        {#each data as {low, high, count}, i}

            {@const x      = X(low) + .5}
            {@const width  = X(high) - X(low) - 1}
            {@const y      = Y(0) * (1-s(i, $t1, $t2, $t3, $t4)) +  Y(count) * s(i, $t1, $t2, $t3, $t4)}
            {@const height = Y(0) * (s(i, $t1, $t2, $t3, $t4))   -  Y(count) * (s(i, $t1, $t2, $t3, $t4))}

            <rect {x} {width} {y} {height} fill={color} />

        {/each}
        <line x1={0} x2={width * $t1} y1={Y(0) + 4} y2={Y(0) + 4} stroke={color} />
        </g>

        <text in:fade fill={color} font-size=11 x={X(data[0].low)}  y={height - buffer + 12 - 3 + 4}>{formattedLowValue}</text>

        <text in:fade fill={color} text-anchor=end font-size=11 x={X(data.slice(-1)[0].high)}  y={height - buffer + 12 - 3 + 4}>{formattedHighValue}</text>


    </svg>
