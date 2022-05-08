<script>
import { onMount } from 'svelte'
import { tweened } from 'svelte/motion';
import { cubicOut } from 'svelte/easing';
import { randomNormal } from 'd3-random';
import { scaleLinear } from 'd3-scale';
import { line as lineGenerator } from "d3-shape";
import { blur, fade } from 'svelte/transition'

import { createScrubAction } from './create-scrub-action';


let r = randomNormal();

// critically, we transform the data with these static-ish scales.
// Ideally, the domain and the range would shouldn't change unless
// the time series does.
// Scrubbing should operate entirely in the "window" space, not
// in the data space.
const x = scaleLinear().domain([0, 1]).range([0, 100]);
const y = scaleLinear().domain([-4, 4]).range([0, 100]);

const l = ys => lineGenerator().x(d=>x(d[xAccessor])).y(d=>y(d[ys]));

let length = 6000;

let xAccessor = 'x';
let yAccessor = 'y';

let v = 0;
let whichSeries = 'sin';

// generate random data.
$: data = Array.from({length }).map((di,i) => {
    if (whichSeries === 'sin') return {x: i / length, y: Math.sin(i / 400) + + (i > length *.45 && i < length * .55 ? r() / 10 : r())}
    v = v + r() / 10;
    if (v < -4) {
        v = -4
    } else if (v > 4) {
        v = 4
    }
    return {x: i / length, y: v }
})

// not currently using this tween, but if we were to update it, we would expect
// 
let width = tweened(400, { duration: 500, easing: cubicOut });

const { coordinates: scrubCoords, scrubAction, isScrubbing } = createScrubAction({
    plotLeft:0, plotRight: $width, plotTop:0, plotBottom: 200,
    completedEventName: 'scrub',
});

let scrubbedXStart = undefined;
let scrubbedXEnd = undefined;

let viewportWidth = tweened($width, { duration: 300, easing: cubicOut });
let leftOffset = tweened(0, { duration: 300, easing: cubicOut });
let rightOffset = tweened($width, { duration: 300, easing: cubicOut });

// here is where we establish the viewport width, as well as the offsets.
$: if (scrubbedXStart && scrubbedXEnd) {
    $viewportWidth = (rangeToViewbox(scrubbedXEnd) - rangeToViewbox(scrubbedXStart));
    $leftOffset = scrubbedXStart;
    $rightOffset = scrubbedXEnd;
} else {
    $viewportWidth = 100;
    $leftOffset = 0;
    $rightOffset = $width;
}

// I should probably clean this up, but essentially the idea is we have a domain, a range (the full pixels),
// and a window (0, 100).
$: rangeToViewbox = scaleLinear().domain([0, $width]).range([0, 100]);
$: rangeToWindow = scaleLinear().domain([$leftOffset, $rightOffset]).range([0, $width]);
$: domainToRange = scaleLinear().domain([0, 1]).range([0, $width]);
$: filteredData = data.filter(di => {
    return di[xAccessor] >= domainToRange.invert($leftOffset) && di[xAccessor] <= domainToRange.invert($rightOffset)
})

/** 
this is the line thickness heuristic that enables us to achieve something like alpha overplotting.
IT doesn't QUITE work here. For instance, the "random walk" option is a bit too thin.
So it works best when the line is VERY noisy and points are not near each oher, as is the 
case with "sin".
*/
$: heuristic = Math.max(.05, Math.min(1, scale * $width / filteredData.length));

let scale = 1;

onMount(() => {
    scale = window.devicePixelRatio;
});

</script>
<h1>
    click+drag to scrub:
</h1>
<div>
    <button on:click={() => whichSeries="random"}>
        random walk
    </button>
        <button on:click={() => whichSeries="sin"}>
        sin
    </button>
</div>
<div>
    <input 
        type="range"			 
        min="100"
        max="120000"
        bind:value={length}		 
    />
    points
</div>
<svg 
        use:scrubAction
        on:scrub={(event) => {
            scrubbedXStart = rangeToWindow.invert(Math.min(event.detail.start.x, event.detail.stop.x));
            scrubbedXEnd = rangeToWindow.invert(Math.max(event.detail.start.x, event.detail.stop.x));
        }}
        width={$width} 
        height={200} 
>
        
        <g 
                width={$width}
                height={200}
                viewbox="0 0 100 100"
        >
            <!-- here is the critical element. We scale the group by the $viewportWidth, then
                shift it over to get the correct left offset starting point.
            -->
            <g style="transform: scale({$width / $viewportWidth}, 2) translateX({-rangeToViewbox($leftOffset)}px)">
            <!-- the magic comes from vector-effect="non-scaling-stroke". -->
                <path d={l(yAccessor)(data)} opacity=1 
                                            stroke-width={heuristic} 
                                            stroke=black fill=none vector-effect="non-scaling-stroke" />
            </g>
    </g>
    <!-- generate the scrubbing UX. this is just a click-to-drag interaction. -->
    <g>
        {#if $scrubCoords.start.x && $scrubCoords.stop.x}
            <rect
                in:blur={{duration: 150}}
                out:fade={{duration: 50}}
                x={Math.min($scrubCoords.start.x, $scrubCoords.stop.x)}
                y={0}
                width={Math.abs($scrubCoords.start.x - $scrubCoords.stop.x)}
                height={200}
                fill="hsla(1, 90%, 50%, .1)"

                style:mix-blend-mode=screen
            />
            <line 
                in:blur={{duration: 150}}
                out:fade={{duration: 50}}
                x1={$scrubCoords.stop.x} 
                x2={$scrubCoords.stop.x} 
                y1={0}
                y2={200}
                stroke-width={1}
                stroke="hsl(1,50%, 85%)"
            />
            <line 
                in:blur={{duration: 150}}
                out:fade={{duration: 50}}
                x1={$scrubCoords.start.x} 
                x2={$scrubCoords.start.x} 
                y1={0}
                y2={200}
                stroke-width={1}
                stroke="hsl(1,50%, 85%)"
            />
        {/if}
    </g>
</svg>

<button on:click={() => {
    scrubbedXStart = undefined;
    scrubbedXEnd = undefined;
    }}>
    reset
</button>

<dl>
    <dt>
        visible points
    </dt>	
    <dd>
        {filteredData.length}	
    </dd>
    
    <dt>
    line width 
    </dt>
    <dd>
        {heuristic}
    </dd>
    <dt>
        left 
    </dt>
    <dd>
        {$leftOffset} pixels
    </dd>
    <dt>
        right
    </dt>
    <dd>
        {$rightOffset} pixels
    </dd>
    <dd>
        scale factor
    </dd>
    <dt>
        {$width / $viewportWidth}
    </dt>
</dl>

<style>
* {
    user-select:none;
}

dl {
    display:grid;
    grid-template-columns: max-content max-content;
    width: max-content;
    
}

</style>