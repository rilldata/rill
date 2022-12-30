<script lang="ts">
  import { GraphicContext } from "@rilldata/web-common/components/data-graphic/elements";
  import { justEnoughPrecision } from "@rilldata/web-common/lib/formatters";
  import MeasureChart from "@rilldata/web-local/lib/components/workspace/explore/time-series-charts/MeasureChart.svelte";

  const makeData = (n: number, addBreak = false) => {
    let numBreaks = ~~(Math.random() * 5);
    const data = [];
    let y = 1000;
    let fiveRandomPoints = Array.from({ length: numBreaks }).map(
      (_i) => ~~(Math.random() * n)
    );
    let breakPoint = undefined;
    let howMany = 0;
    for (let i = 0; i < n; i++) {
      if (fiveRandomPoints.includes(i)) {
        breakPoint = i;
        howMany = ~~(n / 20) + ~~((Math.random() * n) / 45);
      } else if (i > breakPoint + howMany) {
        breakPoint = undefined;
        howMany = undefined;
      }
      y += Math.random() * 100 - 50;
      if (y < 0) y = -y;
      data.push({
        ts: new Date(2019, 0, i),
        value: breakPoint ? null : y,
      });
    }
    return data;
  };

  let SIZE = 500;
  let GRAPH_COUNT = 10;
  let dataSet = Array.from({ length: 10 }).map(() => makeData(SIZE, true));
  let mouseoverValue;
  let hovered = false;
  let scrubbing;
  let scrubStart;
  let scrubEnd;
</script>

<button
  on:click={() => {
    dataSet = Array.from({ length: 10 }).map(() => makeData(SIZE, true));
  }}>randomize</button
>

<GraphicContext
  xType="number"
  yType="date"
  xMin={dataSet[0][0].ts}
  xMax={dataSet[0].at(-1).ts}
>
  {#each dataSet as data}
    <MeasureChart
      bind:mouseoverValue
      {data}
      xMin={data[0].ts}
      xMax={data.at(-1).ts}
      yMin={0}
      mouseoverFormat={justEnoughPrecision}
      xAccessor="ts"
      yAccessor="value"
      height={140}
      bind:hovered
      bind:scrubbing
      bind:scrubStart
      bind:scrubEnd
    />
  {/each}
</GraphicContext>
