<script lang="ts">
  import { extent } from "d3-array";
  import { scaleLinear, scaleTime } from "d3-scale";
  import { area as areaGen, line as lineGen } from "d3-shape";
  import { timeFormat } from "d3-time-format";
  import { cubicInOut as easing } from "svelte/easing";
  import { tweened } from "svelte/motion";

  export let width = 400;
  export let height = 120;
  export let data: any;
  export let xAccessor = "_ts";
  export let xMin = undefined;
  export let xMax = undefined;
  export let yAccessor: string;
  export let left = 20;
  export let right = 40;
  export let buffer = 6;
  export let tickLength = 6;
  export let tickBuffer = 3;
  export let top = 0;
  export let bottom = 0;
  export let zeroBound = true;
  export let color = "black";
  export let xAxis = false;

  export let hoveredDate;

  const fmt = timeFormat("%Y-%m-%d");
  const axisFmt = timeFormat("%b %d");
  const secondaryFmt = timeFormat("%Y");

  function splitOn(data, splitCriterion) {
    const output = [];
    let current = [];
    data.forEach((d) => {
      const dn = Object.assign({}, d);
      if (splitCriterion(dn)) {
        if (current.length) {
          output.push(current.slice(0));
          current = [];
        }
      } else {
        current.push(dn);
      }
    });
    if (current.length) {
      output.push(current.slice(0));
    }
    return output;
  }

  function splitOnNull(p) {
    return (
      p[yAccessor] === undefined ||
      p[yAccessor] === null ||
      Number.isNaN(p[yAccessor])
    );
  }

  function getPoint(x) {
    //const xi = x?.toISOString();
    return data?.find((di) => {
      return di[xAccessor].getTime() === x?.getTime();
    });
  }

  const cheapID = ~~(Math.random() * 10000000);

  $: plotLeft = left + buffer;
  $: plotRight = width - right - buffer;
  $: plotTop = top + buffer;
  $: plotBottom = height - bottom - buffer;

  $: xDomain = extent(data, (point) => point[xAccessor]);
  $: yDomain = extent(data, (point) => point[yAccessor]);

  const innerXMin = tweened(data[0]._ts, { duration: 500, easing });
  const innerXMax = tweened(data.slice(-1)[0]._ts, { duration: 500, easing });

  $: innerXMin.set(xMin || xDomain[0]);
  $: innerXMax.set(xMax || xDomain[1]);

  $: X = scaleTime()
    .domain([$innerXMin, $innerXMax])
    .range([plotLeft, plotRight]);
  $: Y = scaleLinear()
    .domain([zeroBound ? 0 : yDomain[0], yDomain[1]])
    .range([plotBottom, plotTop]);
  $: lineGenerator = lineGen()
    .x((d) => X(d[xAccessor]) || 0)
    .y((d) => Y(d[yAccessor]) || 0);
  $: areaGenerator = areaGen()
    .x((d) => X(d[xAccessor]) || 0)
    .y0(Y(0))
    .y1((d) => Y(d[yAccessor]) || 0);

  $: plotData = splitOn(data, splitOnNull);

  $: hoveredPoint = getPoint(hoveredDate);

  //$: whichTimeRange =
</script>

{#if xAxis}
  <svg {width} height={24}>
    {#each X.ticks(5) as xTick}
      {@const formattedTick = axisFmt(xTick)}
      <text x={X(xTick)} y={22}
        >{formattedTick === "Jan 01"
          ? secondaryFmt(xTick)
          : formattedTick}</text
      >
      <!-- <text x={X(xTick)} y={18}>{xTick.getMonth() + ' ' + xTick.getDay()}</text> -->
    {/each}
  </svg>
{/if}

<svg
  {width}
  {height}
  on:mousemove={(event) => {
    const offsetX = event.offsetX;
    const dt = X.invert(offsetX);
    if (offsetX >= plotLeft && offsetX <= plotRight) {
      const anotherDay = dt.getHours() >= 12 ? 1 : 0;
      dt.setHours(0, 0, 0, 0);
      dt.setDate(dt.getDate() + anotherDay);
      hoveredDate = dt;
    } else {
      hoveredDate = undefined;
    }
  }}
  on:blur
  on:mouseleave={() => {
    hoveredDate = undefined;
  }}
>
  <clipPath id="explore-{cheapID}">
    <rect
      x={plotLeft}
      y={plotTop}
      width={plotRight - plotLeft}
      height={plotBottom - plotTop}
    />
  </clipPath>
  {#each X.ticks(5) as xTick}
    <line
      x1={X(xTick)}
      x2={X(xTick)}
      y1={0}
      y2={height}
      stroke="hsl(1,0%,80%)"
    />
  {/each}
  {#each plotData as series}
    {@const datum = series[0]}
    {#if series.length > 5}
      <path
        clip-path="url(#explore-{cheapID})"
        d={lineGenerator(series)}
        fill="none"
        stroke={color}
      />
      {#if zeroBound}
        <path
          clip-path="url(#explore-{cheapID})"
          d={areaGenerator(series)}
          fill="black"
          opacity=".2"
        />
      {/if}
    {:else}
      <circle
        cx={X(datum[xAccessor])}
        cy={Y(datum[yAccessor])}
        r="1"
        fill={color}
      />
    {/if}
  {/each}

  {#if hoveredDate}
    {#if hoveredPoint}
      <text x={left} y={20}>{fmt(hoveredPoint[xAccessor])}</text>
      <text x={left} y={10}>{hoveredPoint[yAccessor]}</text>
    {/if}
    <line
      x1={X(hoveredDate)}
      x2={X(hoveredDate)}
      y1={0}
      y2={height}
      stroke="black"
    />
    {#if hoveredPoint}
      <circle
        cx={X(hoveredPoint[xAccessor])}
        cy={Y(hoveredPoint[yAccessor])}
        r="3"
        fill="black"
      />
    {/if}
  {/if}

  <!-- right axis -->
  <g transform="translate({width - right} 0)">
    {#each Y.ticks(3) as tick}
      <text dy=".35em" x={tickLength + tickBuffer} y={Y(tick)}>{tick}</text>
      <line x1={0} x2={tickLength} y1={Y(tick)} y2={Y(tick)} stroke="black" />
    {/each}
  </g>
</svg>

<style>
  text {
    font-size: 10px;
    user-select: none;
  }
</style>
