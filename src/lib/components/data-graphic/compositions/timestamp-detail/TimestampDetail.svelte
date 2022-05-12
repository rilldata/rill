<script lang="ts">
/**
 * TimestampDetail.svelte
 * ----------------------
 * This component is a diagnostic plot of the count(*) over time of a timestamp column.
 * The goal is to enable users to understand abnormalities and trends in the timestamp columns
 * of a dataset. As such, this component can:
 * - zoom into a specified scrub region – if the user alt + clicks + drags, the component
 * will zoom into a specific region, enabling the user to better understand weird data.
 * - panning – after zooming, the user may pan around to better situate the viewport.
 * - shift + clicking – users can copy the timestamp value.
 * 
 * The graph will contain an unsmoothed series (showing noise * abnormalities) by default, and 
 * a smoothed series (showing the trend) if the time series merits it.
*/
import { onMount } from "svelte";
import { guidGenerator } from '$lib/util/guid';
import { spring } from 'svelte/motion';
import { fly, fade } from "svelte/transition";
import { cubicOut as easing } from 'svelte/easing';
import { scaleLinear } from 'd3-scale';
import { lineFactory, areaFactory } from "./utils";
import { DEFAULT_COORDINATES } from '$lib/components/data-graphic/constants';
import { createScrubAction } from '$lib/components/data-graphic/create-scrub-action';
import { extent, bisector, max, min } from "d3-array";
import { outline } from "$lib/components/data-graphic/outline";
import { datePortion, timePortion, formatInteger, intervalToTimestring } from '$lib/util/formatters';
import type { Interval } from "$lib/util/formatters"
import { writable } from 'svelte/store';
import { createExtremumResolutionStore } from '../../extremum-resolution-store';

import TimestampBound from './TimestampBound.svelte';

import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TimestampTooltipContent from "./TimestampTooltipContent.svelte";

import { createShiftClickAction } from "$lib/util/shift-click-action";
import notifications from "$lib/components/notifications";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";

const plotID = guidGenerator();

export let data;
export let spark;

export let width = 360;
export let height = 120;
export let curve = 'curveLinear';
export let mouseover = false;
export let smooth = true;

export let separate = true; 
$: separateQuantity = separate ? .25 : 0;

export let xAccessor:string;
export let yAccessor:string;

// rowsize for table
export let left = 1;
export let right = 1;
export let top = 12;
export let bottom = 4;
export let buffer = 0

/** text elements */
// the gap b/t text nodes
export let fontSize = 12;
export let textGap = 4;

/** zoom elements */
export let zoomWindowColor = "hsla(217, 90%, 60%, .2)";

/** rollup grain, time range, etc. */
export let interval:Interval;
export let rollupGrain:string;
export let estimatedSmallestTimeGrain:string;

let scale:number;
onMount(() => {
    scale = window.devicePixelRatio;
})

const X = writable(undefined);
const Y = writable(undefined);
let coordinates = writable(DEFAULT_COORDINATES);

$: plotTop = top + buffer;
$: plotBottom = height - buffer - bottom;
$: plotLeft = left + buffer;
$: plotRight = width - right - buffer;

/**
 * The scrub action creates a scrubbing event that enables the user to 
 */
const { coordinates: zoomCoords, scrubAction, isScrubbing: isZooming } = createScrubAction({
    plotLeft, plotRight, plotTop, plotBottom,
    startPredicate: (event) => event.altKey,
    movePredicate: (event) => event.altKey,
    completedEventName: 'scrub',
});

/**
 * This scroll action creates a scrolling event that will be used in the svg container.
 * The main requirement is this event does not have the shiftKey in use.
 */
const { scrubAction: scrollAction, isScrubbing: isScrolling } = createScrubAction({
    plotLeft, plotRight, plotTop, plotBottom,
    startPredicate: (event) => !event.altKey && !event.shiftKey,
    movePredicate: (event) => !event.altKey && !event.shiftKey,
    moveEventName: "scrolling"
});

let isZoomed = false;

let zoomedXStart:Date;
let zoomedXEnd:Date;
// establish basis values
let xExtents = extent(data, (d) => d[xAccessor]);

const xMin = createExtremumResolutionStore(xExtents[0], { duration: 300, easing, direction: 'min' });
const xMax = createExtremumResolutionStore(xExtents[1], { duration: 300, easing });

$: xExtents = extent(data, (d) => d[xAccessor]);

$: xMin.setWithKey('x', zoomedXStart || xExtents[0]);
$: xMax.setWithKey('x', zoomedXEnd || xExtents[1]);
let dataWindow;

// this adaptive smoothing should be a function?
$: dataWindow = data
    .filter(di => di[xAccessor] >= $xMin && di[xAccessor] <= $xMax);
$: windowWithoutZeros = dataWindow.filter(di => {
    return di[yAccessor] !== 0;
})
$: windowSize = dataWindow.length < 150 ? 30 : ~~(dataWindow.length / (25));

$: smoothedData = data.map((di, i, arr) => {
    const dii = {...di}
    const window = Math.max(3, Math.min(~~(windowSize), i));
    const prev = arr.slice(i - ~~(window /2), i + ~~(window / 2));
    dii._smoothed = prev.reduce((a, b) => a + b.count, 0) / prev.length;
    return dii;
})

// Let's set the X Scale based on the $xMin and $xMax.
$: $X = scaleLinear().domain([$xMin, $xMax]).range([left + buffer, width - right - buffer]);

// Generate the line density by dividing the total available pixels by the window length.
// We will scale by window.pixelDensityRatio.

$: totalTravelDistance = dataWindow.map((di,i) => {
     if (i === data.length - 1) { return 0 };
     let max = Math.max($Y(data[i+1][yAccessor]), $Y(data[i][yAccessor]));
     let min = Math.min($Y(data[i+1][yAccessor]), $Y(data[i][yAccessor]));
     return Math.abs(max - min);
 }).reduce((acc,v) => acc+v, 0)

let lineDensity = .05;
$: lineDensity = Math.min(1, 
    /** to determine the stroke width of the path, let's look at 
     * the bigger of two values:
     * 1. the "y-ish" distance travelled
     * the inverse of "total travel distance", which is the Y
     * gap size b/t successive points divided by the zoom window size;
     * 2. time serires length / available X pixels
     * the time series divided by the total number of pixels in the existing
     * zoom window.
     * 
     * These heuristics could be refined, but this seems to provide a reasonable approximation for
     * the stroke width. (1) excels when lots of successive points are close together in the Y direction,
     * whereas (1) excels when a line is very
    */
    Math.max(
        2 / (totalTravelDistance / (($X($xMax) - $X($xMin)) * scale)),
        (($X($xMax)  - $X($xMin)) * scale * .7) / dataWindow.length / 1.5
    )
    
);
/** the line opacity calculation is just a function of the availble pixels divided
 * by the window length, capped at 1. This seems to work well in practice.
*/
$: opacity = Math.min(1, 1 + (($X($xMax)  - $X($xMin)) * scale) / dataWindow.length / 2);
 
// Generate our Y Scale.
let yExtents = extent(data, d => d[yAccessor]);
$: yExtents = extent(data, d => d[yAccessor]);
const yMax = createExtremumResolutionStore(Math.max(5, yExtents[1]))
$: $Y = scaleLinear().domain([0, $yMax]).range([plotBottom, plotTop]);

$: lineFcn = lineFactory({
    xScale: $X,
    yScale: $Y,
    curve,
    xAccessor
});

$: areaFcn = areaFactory({
    xScale: $X,
    yScale: $Y,
    curve,
    xAccessor
});

let nearestPoint = undefined;

// get the nearest point to where the cursor is.
let bisectDate = bisector((d) => d[xAccessor]).center;
$: nearestPoint = data[bisectDate(data, $X.invert($coordinates.x))]

// let's create the final version of the smoothedData d attribute.
$: smoothedLine = lineFcn("_smoothed")(smoothedData);

function clearMouseMove() { coordinates.set(DEFAULT_COORDINATES); }

function handleMouseMove(event) {
    if (event.offsetX > plotLeft && event.offsetX < plotRight) {
        coordinates.set({x: event.offsetX, y: event.offsetY});
    }
}

function setCursor(isZooming, isScrolling) {
    if (isZooming) return 'text';
    if (isScrolling) return 'grab';
    return 'inherit';
}

// when zooming / panning, get the total number of zoomed rows.
let zoomedRows;

$: if ($zoomCoords.start.x && $zoomCoords.stop.x) {
    let xStart = $X.invert(Math.min($zoomCoords.start.x, $zoomCoords.stop.x));
    let xEnd = $X.invert(Math.max($zoomCoords.start.x, $zoomCoords.stop.x));
    zoomedRows = ~~data.filter(di => {
        return di[xAccessor] >= xStart && di[xAccessor] <= xEnd;
    }).reduce((sum, di) => sum += di[yAccessor], 0);
} else if (zoomedXStart && zoomedXEnd) {
    zoomedRows = ~~data.filter(di => {
        return di[xAccessor] >= zoomedXStart && di[xAccessor] <= zoomedXEnd;
    }).reduce((sum, di) => sum += di[yAccessor], 0);
}

// Tooltip & timestamp range variables.
let tooltipSparkWidth = 84;
let tooltipSparkHeight = 12;
let tooltipPanShakeAmount = spring(0, { stiffness:.1, damping:.9 });
let movementTimeout:ReturnType<typeof setTimeout>;

$: zoomMinBound = ($zoomCoords.start.x ?
                        $X.invert(Math.min($zoomCoords.start.x, $zoomCoords.stop.x)) :
                        min([zoomedXStart, zoomedXEnd])) ||
                        xExtents[0];

$: zoomMaxBound = ($zoomCoords.start.x ?
                            $X.invert(Math.max($zoomCoords.start.x, $zoomCoords.stop.x)) :
                        max([zoomedXStart, zoomedXEnd])) ||
                        xExtents[1];

/**
 * Use this shiftClickAction to copy the timestamp that is currently moused over.
 */
const { shiftClickAction } = createShiftClickAction();
</script>
<div style:max-width="{width}px">
    <div class="text-gray-600" style="
        display: grid;
        grid-template-columns: auto auto;
    ">
        <Tooltip distance={16} location="top">
        <div style="grid-row: 1; grid-column: 1;">
            {#if interval}
            RANGE {intervalToTimestring(interval)}
            {/if}
        </div>
        <TooltipContent slot="tooltip-content">
            <div style:max-width="315px">
                The range of this timestamp is {intervalToTimestring(interval)}.
            </div>
        </TooltipContent>
        </Tooltip>
        <Tooltip distance={16} location="top">
            <div class="text-right" style="grid-row: 1; column: 2;">
                {#if rollupGrain}
                    ROLLUP {rollupGrain}
                {/if}
            </div>
            <TooltipContent slot="tooltip-content">
                <div style:max-width="315px">
                    This timestamp column is aggregated so each point on the time series is <i>{rollupGrain}</i>.
                </div>
            </TooltipContent>
        </Tooltip>
        <Tooltip distance={16} location="top">
        <div class="text-right" style="grid-row: 2; grid-column: 2;">
            {#if estimatedSmallestTimeGrain}
                GRAIN {estimatedSmallestTimeGrain}
            {/if}
        </div>
        <TooltipContent slot="tooltip-content">
            <div style:width="315px">
                The smallest available grain in this column is at the <i>{estimatedSmallestTimeGrain}</i> level.
            </div>
        </TooltipContent>
        </Tooltip>
    </div>
    <Tooltip location="right" alignment="center" distance={32}>
    <svg width={width} height={height}
        style:cursor={setCursor($isZooming, $isScrolling)}
        use:scrubAction
        use:scrollAction
        use:shiftClickAction
        on:shift-click={async () => {
            let exportedValue = `TIMESTAMP '${nearestPoint[xAccessor].toISOString()}'`
            await navigator.clipboard.writeText(exportedValue);
            setTimeout(() => {
                notifications.send({ message: `copied ${exportedValue} to clipboard`});
            }, 200)
            
        }}

        on:scrolling={(event) => {
            if (isZoomed) {

                // clear the tooltip shake effect zeroing timeout.
                clearTimeout(movementTimeout);
                // shake the word "pan" in the tooltip here.
                tooltipPanShakeAmount.set(event.detail.movementX / 8);
                // set this timeout to resolve back to 0 if the user stops dragging.
                movementTimeout = setTimeout(() => {
                    tooltipPanShakeAmount.set(0);
                }, 150)

                let timeDistance = $X.invert(event.detail.clientX + (event.detail.movementX )) - $X.invert(event.detail.clientX);
                let oldXStart = new Date(+zoomedXStart);
                let oldXEnd = new Date(+zoomedXEnd);
                zoomedXStart = new Date(+zoomedXStart - +timeDistance);
                zoomedXEnd = new Date(+zoomedXEnd - +timeDistance);
                
                if (zoomedXStart < xExtents[0] || zoomedXEnd >= xExtents[1]) {
                    zoomedXStart = oldXStart;
                    zoomedXEnd = oldXEnd;
                }
            }
        }}

        on:scrub={(event) => { 
            // set max and min here.
            zoomedXStart = new Date($X.invert(Math.min(event.detail.start.x, event.detail.stop.x)));
            zoomedXEnd = new Date($X.invert(Math.max(event.detail.start.x, event.detail.stop.x)));
            // mark that this graphic has been scrubbed.
            setTimeout(() => {
                isZoomed = true;
            }, 100)
            
        }}
        on:mousemove={mouseover ? handleMouseMove : () => {}}
        on:mouseleave={mouseover ? clearMouseMove : () => {}}
    >
    <defs>
        <linearGradient id="left-side">
            <stop offset="0%" stop-color="white"/>
            <stop offset="100%" stop-color="rgba(255,255,255,0)"/>
        </linearGradient>
        <linearGradient id="right-side">
            <stop offset="0%" stop-color="rgba(255,255,255,0)"/>
            <stop offset="100%" stop-color="white"/>
        </linearGradient>
    </defs>
        <clipPath id="data-graphic-{plotID}" >
            <rect 
                x={plotLeft}
                y={plotTop}
                width={plotRight - plotLeft}
                height={plotBottom - plotTop}
            />
        </clipPath>
        <g clip-path="url(#data-graphic-{plotID})">
            <!-- core geoms -->
            <path
                d={areaFcn(yAccessor)(data)}
                fill=rgba(0,0,0,.05)
            />
            <path
                d={lineFcn(yAccessor)(data)}
                stroke="black" 
                stroke-width={lineDensity}
                fill=none
                style:opacity={opacity}
                class="transition-opacity"
            />

        <!-- smoothed line -->
        <g style:transition="opacity 300ms" style:opacity={smooth && windowWithoutZeros?.length && windowWithoutZeros.length > (width * scale) ? 1 : 0}>
            <path
                d={smoothedLine}
                stroke=white fill=none
                stroke-width={3}
                style:opacity={.5}
            />
            <path
                d={smoothedLine}
                stroke="hsl(217, 80%, 20%)" fill=none
                stroke-width={1.5}
                style:opacity={.85}
            />
        </g>
        {#if isZoomed}
            <!-- fadeout gradients on each side? -->
            <rect 
                transition:fade
                x={plotLeft}
                y={plotTop}
                width={20}
                height={plotBottom - plotTop}
                fill="url(#left-side)"
            />
            <rect 
                transition:fade
                x={plotRight - 20}
                y={plotTop}
                width={20}
                height={plotBottom - plotTop}
                fill="url(#right-side)"
            />
        {/if}
            <line x1={$X?.range()[0]} x2={$X?.range()[1]} y1={$Y && $Y(0)} y2={$Y && $Y(0)} stroke=rgb(100,100,100) />
        </g>
        <g>
            {#if $zoomCoords.start.x && $zoomCoords.stop.x}
                <rect 
                    x={Math.min($zoomCoords.start.x, $zoomCoords.stop.x)}
                    y={plotTop + buffer}
                    width={Math.abs($zoomCoords.start.x - $zoomCoords.stop.x)}
                    height={plotBottom - plotTop}
                    fill={zoomWindowColor}

                    style:mix-blend-mode=darken
                />
                <line 
                    x1={$zoomCoords.start.x} 
                    x2={$zoomCoords.start.x} 
                    y1={plotTop + buffer}
                    y2={plotBottom}
                    stroke="rgb(100,100,100)"
                />
            {/if}
        </g>
        <!-- mouseover information -->
        {#if $coordinates.x}
        <g>
                <line 
                    x1={$X(nearestPoint[xAccessor])} 
                    x2={$X(nearestPoint[xAccessor])} 
                    y1={plotTop + buffer}
                    y2={plotBottom}
                    stroke="rgb(100,100,100)"
                />
                {#each [[yAccessor, 'rgb(100,100,100)']] as [accessor, color]}
                    {@const cx = $X(nearestPoint[xAccessor])}
                    {@const cy = $Y(nearestPoint[accessor])}
                    {#if cx && cy}
                        <circle {cx} {cy} r={3} fill={color} />
                    {/if}
                {/each}
                <g
                    in:fly={{duration: 200, x: -16 }} 
                    out:fly={{duration: 200, x: -16 }} 
                    font-size={fontSize}
                    style:user-select={"none"}
                >
                    <text
                        x={plotLeft} 
                        y={fontSize} 
                        use:outline
                    >
                        {datePortion(nearestPoint[xAccessor])}
                    </text>
                    <text
                        x={plotLeft} 
                        y={(fontSize) * 2 + textGap}
                        use:outline
                >
                    {timePortion(nearestPoint[xAccessor])}
                </text>
                <text
                    x={plotLeft}
                    y={(fontSize) * 3 + textGap * 2} 
                    use:outline

                >
                    {formatInteger(~~nearestPoint[yAccessor])} row{#if nearestPoint[yAccessor] !== 1}s{/if}
                </text>
                </g>
        </g>
        {/if}
        <!-- scrub-clearing click region -->
        {#if zoomedXStart && zoomedXEnd}
            <text
                font-size={fontSize}
                x={plotRight} 
                y={fontSize} 
                text-anchor="end"

                style:font-style="italic"
                style:user-select="none"
                style:cursor="pointer"

                class="transition-color fill-gray-500 hover:fill-black"
                
                in:fly={{duration: 200, x: 16, delay: 200}} 
                out:fly={{duration: 200, x: 16 }} 

                use:outline

                on:click={()=> {
                    zoomedXStart = undefined;
                    zoomedXEnd = undefined;
                    isZoomed = false;
                }}
            >
                clear zoom ✖
            </text>
        {/if}
    </svg>
    <!--
    Graph Tooltip Content
    ---------------------
    We slot in the tooltip content into an encompassing div.
    Ideally, this tooltip would perfectly center in all cases, but we should use a MutationObserver within FloatingElement.svelte
    to additionally listen to the child element mutations before placement.
    This is a workaround, and given that the content does not really redraw the bounds,
    it should work fine in practice.
    -->
    <div slot="tooltip-content"
        in:fly={{duration: 100, y: 4 }}
        out:fly={{duration: 100, y: 4, }}
        style="
            display: grid; 
            justify-content: center; 
            grid-template-columns: max-content;"
    >
        <TimestampTooltipContent 
            data={spark}
            {xAccessor}
            {yAccessor}
            width={tooltipSparkWidth}
            height={tooltipSparkHeight}

            tooltipPanShakeAmount={
                // we will shake the tooltip pan word
                $tooltipPanShakeAmount
            }

            {zoomedRows}
            totalRows={~~data.reduce((a,b) => a+b[yAccessor], 0)}
            zoomed={($zoomCoords.start.x) || (zoomedXStart)}
            zooming={zoomedXStart && !$zoomCoords.start.x}
            zoomWindowXMin={
                $zoomCoords.start.x ?
                $X.invert(Math.min($zoomCoords.start.x, $zoomCoords.stop.x)) :
                min([zoomedXStart, zoomedXEnd])
            }
            zoomWindowXMax={
                $zoomCoords.stop.x ?
                $X.invert(Math.max($zoomCoords.start.x, $zoomCoords.stop.x)) :
                max([zoomedXStart, zoomedXEnd ])
            }
        />
    </div>
    </Tooltip>

    <!-- Bottom time horizon labels -->
    <div class="select-none grid grid-cols-2 space-between">
        <TimestampBound 
            align="left"
            value={zoomMinBound}
            label="Min"
        />
        <TimestampBound 
            align="right"
            value={zoomMaxBound}
            label="Max"
        />
    </div>
</div>


<style>
text {
    user-select: none;
}
</style>