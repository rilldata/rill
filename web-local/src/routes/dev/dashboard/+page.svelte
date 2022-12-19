<script lang="ts">
  import { MeasureChart } from "./measure-chart";

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
</script>

<button
  on:click={() => {
    dataSet = Array.from({ length: 10 }).map(() => makeData(SIZE, true));
  }}>randomize</button
>
{#each dataSet as data}
  <MeasureChart
    bind:mouseoverValue
    {data}
    xMin={data[0].ts}
    xMax={data.at(-1).ts}
    yMin={0}
    xAccessor="ts"
    yAccessor="value"
    height={140}
  />
{/each}
