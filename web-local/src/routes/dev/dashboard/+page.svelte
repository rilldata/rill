<script lang="ts">
  import MeasureChart from "./MeasureChart.svelte";

  const makeData = (n: number, addBreak = false) => {
    const data = [];
    let y = 1000;
    let breakPoint = addBreak ? ~~(Math.random() * n) : undefined;
    for (let i = 0; i < n; i++) {
      y += Math.random() * 100 - 50;
      data.push({
        ts: new Date(2019, 0, i),
        value: addBreak && i > breakPoint && i < breakPoint + n / 5 ? null : y,
      });
    }
    return data;
  };
  let data = makeData(50, true);
  $: start = data[0].ts;
  $: end = data.at(-1).ts;
</script>

<button on:click={() => (data = makeData(50, true))}>randomize</button>

<MeasureChart {data} xMin={start} xMax={end} xAccessor="ts" yAccessor="value" />
