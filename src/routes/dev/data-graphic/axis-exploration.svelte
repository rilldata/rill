<script lang="ts">
  import {
    SimpleDataGraphic,
    Body,
  } from "$lib/components/data-graphic/elements";
  import { Line } from "$lib/components/data-graphic/marks";
  import { Axis } from "$lib/components/data-graphic/guides";

  function makeData(intervalSize = 1000) {
    let v = 50;
    const windowSize = 1 + ~~(Math.random() * 150);
    const data = Array.from({ length: 200 }).map((_, i) => {
      v += 100 * (Math.random() - 0.5);
      return {
        period: new Date(+new Date("2010-04-11 08:43:04") + i * intervalSize),
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

  let interval = 1000 * 60 * 60 * 24;
  const intervals = [1000, 1000 * 60, 1000 * 60 * 60, 1000 * 60 * 60 * 24];

  let widthMultiple = 150;
  $: data = makeData(interval);
</script>

Width Multiple:
<input
  autocomplete="off"
  type="range"
  bind:value={widthMultiple}
  min={40}
  max={320}
  class="w-48 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
/>
<span>{widthMultiple}</span>
<div>
  Time Interval:

  <select bind:value={interval}>
    <option value={1000 * 60 * 60 * 24}> day interval </option>
    <option value={1000 * 60 * 60}> hour interval </option>
    <option value={1000 * 60}> min interval </option>
  </select>
</div>

<div style:width="max-content">
  {#each [1, 2, 3, 4] as chart}
    <SimpleDataGraphic
      width={widthMultiple * (chart + 1)}
      height={200}
      xType="date"
      yType="number"
    >
      <Axis side="left" />
      <Axis side="bottom" />
      <Body border>
        <Line
          {data}
          xAccessor="period"
          yAccessor="value"
          color="#3498db"
          lineThickness={1}
        />
      </Body>
    </SimpleDataGraphic>
  {/each}
</div>
