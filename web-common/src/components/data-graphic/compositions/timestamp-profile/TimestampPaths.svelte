<script lang="ts">
  /** the path elements used in the timestamp profiler.
   *
   */
  import type { ScaleLinear } from "d3-scale";

  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";

  import { areaFactory, lineFactory } from "../../utils";
  import type { PlotConfig } from "../../utils";

  export let xAccessor: string;
  export let yAccessor: string;
  export let curve: string;
  export let smooth = false;
  export let data;

  const X: Writable<ScaleLinear<number, number>> = getContext(
    "rill:data-graphic:X"
  );
  const Y: Writable<ScaleLinear<number, number>> = getContext(
    "rill:data-graphic:Y"
  );

  const plotConfig: Writable<PlotConfig> = getContext(
    "rill:data-graphic:plot-config"
  );

  $: lineFcn = lineFactory({
    xScale: $X,
    yScale: $Y,
    curve,
    xAccessor,
  });

  $: areaFcn = areaFactory({
    xScale: $X,
    yScale: $Y,
    curve,
    xAccessor,
  });

  // this adaptive smoothing should be a function.
  $: dataWindow = data.filter(
    (di) => di[xAccessor] >= $X.domain()[0] && di[xAccessor] <= $X.domain()[1]
  );
  $: windowWithoutZeros = dataWindow.filter((di) => {
    return di[yAccessor] !== 0;
  });
  $: windowSize = dataWindow.length < 150 ? 30 : ~~(dataWindow.length / 25);

  $: smoothedData = data.map((di, i, arr) => {
    const dii = { ...di };
    const window = Math.max(3, Math.min(~~windowSize, i));
    const prev = arr.slice(i - ~~(window / 2), i + ~~(window / 2));
    dii._smoothed = prev.reduce((a, b) => a + b.count, 0) / prev.length;
    return dii;
  });

  $: totalTravelDistance = dataWindow
    .map((di, i) => {
      if (i === data.length - 1) {
        return 0;
      }
      const max = Math.max($Y(data[i + 1][yAccessor]), $Y(data[i][yAccessor]));
      const min = Math.min($Y(data[i + 1][yAccessor]), $Y(data[i][yAccessor]));
      return Math.abs(max - min);
    })
    .reduce((acc, v) => acc + v, 0);

  let lineDensity = 0.05;

  $: lineDensity = Math.min(
    1,
    /** to determine the stroke width of the path, let's look at
     * the bigger of two values:
     * 1. the "y-ish" distance travelled
     * the inverse of "total travel distance", which is the Y
     * gap size b/t successive points divided by the zoom window size;
     * 2. time series length / available X pixels
     * the time series divided by the total number of pixels in the existing
     * zoom window.
     *
     * These heuristics could be refined, but this seems to provide a reasonable approximation for
     * the stroke width. (1) excels when lots of successive points are close together in the Y direction,
     * whereas (2) excels when a line is very, very noisy (and thus the X direction is the main constraint).
     */
    Math.max(
      2 /
        (totalTravelDistance /
          (($X.range()[1] - $X.range()[0]) * $plotConfig.devicePixelRatio)),
      (($X.range()[1] - $X.range()[0]) * $plotConfig.devicePixelRatio * 0.7) /
        dataWindow.length /
        1.5
    )
  );
  /** the line opacity calculation is just a function of the available pixels divided
   * by the window length, capped at 1. This seems to work well in practice.
   */
  $: opacity = Math.min(
    1,
    1 +
      (($X.range()[1] - $X.range()[0]) * $plotConfig.devicePixelRatio) /
        dataWindow.length /
        2
  );
  $: smoothedLine = lineFcn("_smoothed")(smoothedData);
</script>

<path d={areaFcn(yAccessor)(data)} fill="rgba(0,0,0,.05)" />
<path
  d={lineFcn(yAccessor)(data)}
  stroke="black"
  stroke-width={lineDensity}
  fill="none"
  style:opacity
  class="transition-opacity"
/>

<!-- smoothed line -->
<g
  style:transition="opacity 300ms"
  style:opacity={smooth &&
  windowWithoutZeros?.length &&
  windowWithoutZeros.length > $plotConfig.width * $plotConfig.devicePixelRatio
    ? 1
    : 0}
>
  <path
    d={smoothedLine}
    stroke="white"
    fill="none"
    stroke-width={3}
    style:opacity={0.5}
  />
  <path
    d={smoothedLine}
    stroke="hsl(217, 80%, 20%)"
    fill="none"
    stroke-width={1.5}
    style:opacity={0.85}
  />
</g>
