<script lang="ts">
  import { tweened, spring } from "svelte/motion";
  import { cubicOut as easing, sineInOut } from "svelte/easing";
  import DataGraphic from "$lib/components/data-graphic/DataGraphic.svelte";
  import Body from "$lib/components/data-graphic/elements/Body.svelte";
  import Line from "$lib/components/data-graphic/marks/Line.svelte";
  import Axis from "$lib/components/data-graphic/guides/Axis.svelte";

  function makeData(intervalSize = 1000) {
    let v = 50;
    let offset = ~~(Math.random() * 100);
    const windowSize = 1 + ~~(Math.random() * 150);
    const data = Array.from({ length: 200 }).map((_, i) => {
      v += 100 * (Math.random() - 0.5);
      return {
        period: new Date(+new Date("2010-01-01 00:01:04") + i * intervalSize),
        value: v,
      };
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
</script>

<button
  on:click={() => {
    data1.set(makeData());
  }}>randomize</button
>

<div style:width="max-content">
  {#each [1000, 1000 * 60, 1000 * 60 * 60, 1000 * 60 * 60 * 24] as intervalSize}
    {@const data = makeData(intervalSize)}
    <DataGraphic
      width={800}
      height={200}
      let:plotConfig
      xType="date"
      yType="number"
    >
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
            {data}
            xAccessor="period"
            yAccessor="value"
            color="hsl(90, 50%, 70%)"
            lineThickness={2}
          />
        </Body>
      </g>
    </DataGraphic>
  {/each}
</div>
