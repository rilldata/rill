<script lang="ts">
    import { tweened } from 'svelte/motion';
    import { fly, fade } from "svelte/transition";
    import { circIn, cubicOut as easing } from 'svelte/easing';
    import { scaleLinear } from 'd3-scale';
    import { format } from "d3-format";

    
    interface HistogramBin {
        bucket:number;
        low:number;
        high:number;
        count:number;
    }
    
    export let min:number;
    export let qlow:number;
    export let median:number;
    export let qhigh:number;
    export let mean:number;
    export let max:number;

    export let data:HistogramBin[];
    
    export let width = 60;
    export let height = 19;
    export let time = 1000;
    export let color = 'hsl(340, 70%, 70%)';
    export let dataType = 'int';
    export let buffer = 22;

    // rowsize for table
    export let left = 36;
    export let right = 16;
    export let fontSize = 20;

    // what do we have here? min, q25, q50, mean, q75, max
    
    const t1 = tweened(0, { duration: time, easing });
    const t2 = tweened(0, { duration: time * (1 + Math.random() / 5), easing, delay: time / 6 });
    const t3 = tweened(0, { duration: time * (1 + Math.random() / 3), easing, delay: time / 3});
    const t4 = tweened(0, { duration: time * (1 + Math.random() / 1.5), easing });
    
    const lowValue = tweened(0, { duration: time / 2, easing });
    const highValue = tweened(0, { duration: time / 2, easing });

    $: minX = Math.min(...data.map( d => d.low ));
    $: maxX = Math.max(...data.map( d => d.high ));
    $: X = scaleLinear().domain([minX, maxX]).range([left, width - right]);

    $: yVals = data.map( d => d.count );
    $: maxY = Math.max(...yVals);
    $: Y = scaleLinear().domain([0, maxY]).range([height - 4 - buffer, 4]);
    
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
    
    </script>
    <svg {width} height={height + 6 * fontSize} >
        <!-- histogram -->
        <g shape-rendering=crispEdges>
        {#each data as {low, high, count}, i}

            {@const x      = X(low) }
            {@const width  = X(high) - X(low)}
            {@const y      = Y(0) * (1-s(i, $t1, $t2, $t3, $t4)) +  Y(count) * s(i, $t1, $t2, $t3, $t4)}
            {@const height = Math.min(Y(0), Y(0) * (s(i, $t1, $t2, $t3, $t4))   -  Y(count) * (s(i, $t1, $t2, $t3, $t4)))}

            <rect x={x} {width} {y} {height} fill={color} />

        {/each}
        <line x1={X(X.domain()[0])} x2={width * $t1} y1={Y(0) + 4} y2={Y(0) + 4} stroke={color} />
        </g>

        <!-- <text in:fade fill={color} font-size=11 x={X(data[0].low)}  y={height - buffer + 12 - 3 + 4}>{formattedLowValue}</text>
        <text in:fade fill={color} text-anchor=end font-size=11 x={X(data.slice(-1)[0].high)}  y={height - buffer + 12 - 3 + 4}>{formattedHighValue}</text> -->
        
        <g style:font-size="12px" class='textElements'>
            {#each [['min', min], ['q25', qlow], ['med', median], ['mean', mean], ['q75', qhigh], ['max', max]] as [label, value], i} 
                {@const y = height + i * fontSize}
                {@const anchor = X(value) < (width / 2) ? 'start' : 'end'}
                {@const anchorBuffer = anchor === 'start' ? 6 : -6}
                <line x1={X($lowValue)}  x2={X($highValue)} y1={y - fontSize / 4 } y2={y - fontSize / 4 } stroke-dasharray=2,1 opacity=.3 stroke={color} />
                <line x1={X(value)}  x2={X(value)} y1={y - fontSize / 4} y2={Y(0) + 4}  opacity=.3 stroke={color} />
                <text text-anchor="end" x={left - 6} y={y}>
                        {label}
                </text>
                <text x={X(value) + anchorBuffer} y={y} text-anchor={anchor}>{value}</text>
                <circle in:fly={{duration: 500, y: -5}} fill={color} cx={X(value)} cy={y - fontSize / 4 } r=3 />
            {/each}
        </g>

    </svg>

    <!-- temp: get json -->

<button on:click={() => {
    navigator.clipboard.writeText(JSON.stringify(data, null, 2));
    console.log('copied to clipboard.')
}}>json</button>