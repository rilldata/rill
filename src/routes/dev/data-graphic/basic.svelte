<script lang="ts">
  import { tweened, spring } from "svelte/motion";
  import { cubicOut as easing, sineInOut } from "svelte/easing";
  import DataGraphic from "$lib/components/data-graphic/DataGraphic.svelte";
  import Body from "$lib/components/data-graphic/elements/Body.svelte";
  import Line from "$lib/components/data-graphic/marks/Line.svelte";
  import Axis from "$lib/components/data-graphic/guides/Axis.svelte";

  function makeData() {
    let v = 50;
    let offset = ~~(Math.random() * 100);
    const windowSize = 1 + ~~(Math.random() * 150);
    const data = Array.from({ length: 2000 }).map((_, i) => {
      v += 100 * (Math.random() - 0.5);
      return { period: new Date(i * 10000 + offset + 1656020175), value: v };
    });
    return data.map(({ period }, i) => {
      const window = data.slice(Math.max(0, i - windowSize), i);
      const v = window.reduce((acc, v) => acc + v.value, 0);
      return {
        period,
        value: v / window.length,
      };
    });
  }
  let data1 = tweened(makeData(), { easing });
  let data2 = tweened(makeData(), { easing, duration: 400 });
  let data3 = tweened(makeData(), { easing, duration: 600 });
  let data4 = tweened(makeData(), { easing, duration: 400 });
</script>

<button
  on:click={() => {
    data1.set(makeData());
    data2.set(makeData());
    data3.set(makeData());
    data4.set(makeData());
  }}>randomize</button
>

<div style:width="max-content">
  <DataGraphic width={800} height={400} let:plotConfig xType="date">
    <line
      x1={plotConfig.plotLeft}
      x2={plotConfig.plotLeft}
      y1={plotConfig.plotTop}
      y2={plotConfig.plotBottom}
      stroke="hsl(1, 50%, 80%)"
    />
    <line
      x1={plotConfig.plotRight}
      x2={plotConfig.plotRight}
      y1={plotConfig.plotTop}
      y2={plotConfig.plotBottom}
      stroke="hsl(90, 50%, 80%)"
    />
    <line
      x1={plotConfig.plotLeft}
      x2={plotConfig.plotRight}
      y1={plotConfig.plotTop}
      y2={plotConfig.plotTop}
      stroke="hsl(180, 50%, 80%)"
    />
    <line
      x1={plotConfig.plotLeft}
      x2={plotConfig.plotRight}
      y1={plotConfig.plotBottom}
      y2={plotConfig.plotBottom}
      stroke="hsl(270, 50%, 80%)"
    />
    <g>
      <Axis side="left" />
      <Axis side="right" />
      <Axis side="top" />
      <Axis side="bottom" />
      <Body>
        <Line
          data={$data2}
          xAccessor="period"
          yAccessor="value"
          color="hsl(90, 50%, 70%)"
          lineThickness={2}
        />
        <Line
          data={$data1}
          xAccessor="period"
          yAccessor="value"
          color="hsl(135, 50%, 70%)"
        />
        <Line
          data={$data3}
          xAccessor="period"
          yAccessor="value"
          color="hsl(180, 50%, 70%)"
        />
        <Line
          data={$data4}
          xAccessor="period"
          yAccessor="value"
          lineThickness={3}
          color="hsl(215, 50%, 70%)"
        />
      </Body>
    </g>
  </DataGraphic>
</div>
