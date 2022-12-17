<script lang="ts">
  import MeasureChart from "./MeasureChart.svelte";

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
        howMany = 50 + ~~((Math.random() * n) / 25);
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
  let data = makeData(1000, true);
  $: start = data[0].ts;
  $: end = data.at(-1).ts;
</script>

<button on:click={() => (data = makeData(800 + ~~(Math.random() * 200), true))}
  >randomize</button
>

<MeasureChart
  groundOnZero={false}
  {data}
  xMin={start}
  xMax={end}
  yMin={0}
  xAccessor="ts"
  yAccessor="value"
/>
