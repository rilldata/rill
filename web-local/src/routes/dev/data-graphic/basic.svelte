<script lang="ts">
  import {
    Body,
    SimpleDataGraphic,
  } from "@rilldata/web-common/components/data-graphic/elements";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import { Line } from "@rilldata/web-common/components/data-graphic/marks";

  function makeData(intervalSize = 1000) {
    let v = 50;
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
        value: v / (window.length || 1),
      };
    });
  }
</script>

<div style:width="max-content">
  {#each [1000, 1000 * 60, 1000 * 60 * 60, 1000 * 60 * 60 * 24] as intervalSize}
    {@const data = makeData(intervalSize)}
    <SimpleDataGraphic width={800} height={200} xType="date" yType="number">
      <Axis side="left" />
      <Axis side="right" />
      <Axis side="top" />
      <Axis side="bottom" />
      <Body border>
        <Line
          {data}
          xAccessor="period"
          yAccessor="value"
          color="hsl(90, 50%, 70%)"
          lineThickness={2}
        />
      </Body>
    </SimpleDataGraphic>
  {/each}
</div>
