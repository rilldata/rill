<script lang="ts">
  import { format } from "d3-format";
  import { fade } from "svelte/transition";

  import Body from "$lib/components/data-graphic/elements/Body.svelte";
  import GraphicContext from "$lib/components/data-graphic/elements/GraphicContext.svelte";
  import Axis from "$lib/components/data-graphic/guides/Axis.svelte";
  import PointLabel from "$lib/components/data-graphic/guides/PointLabel.svelte";
  import Line from "$lib/components/data-graphic/marks/Line.svelte";
  import SimpleDataGraphic from "$lib/components/data-graphic/SimpleDataGraphic.svelte";
  import WithBisector from "$lib/components/data-graphic/functional-components/WithBisector.svelte";
  import { cubicOut } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import Area from "$lib/components/data-graphic/marks/Area.svelte";

  function smooth(data, accessor, windowSize = 7) {
    return data.map((datum, i) => {
      const window = data.slice(Math.max(0, i - windowSize), i);
      const v = window.reduce((acc, v) => acc + v[accessor], 0);
      return { ...datum, [accessor]: v };
    });
  }

  function makeData(length = 180) {
    let y = 100;
    const data = Array.from({ length }).map((_, i) => {
      y += (Math.random() - 0.5) * 30;
      if (y < 0) y = 1;
      return {
        period: new Date(
          +new Date("2010-01-01 00:01:04") + i * 1000 * 60 * 60 * 24
        ),
        y,
      };
    });
    return smooth(data, 7);
  }
  let mouseoverValues;
  let datasets = tweened(
    Array.from({ length: 36 }).map(() => makeData(60)),
    { duration: 500, easing: cubicOut }
  );
</script>

<section>
  <h1 class="text-xl mb-6">Small Multiples</h1>
  <button
    class="p-3 pt-1 pb-1 bg-gray-100 rounded mb-6 hover:bg-gray-200"
    on:click={() => {
      datasets.set(Array.from({ length: 36 }).map(() => makeData(60)));
    }}>randomize</button
  >

  <GraphicContext xType="date" yType="number">
    <div class="flex flex-row flex-wrap gap-3 w-max-screen">
      {#each $datasets as data, i (i)}
        <div>
          <h2 class="pl-5">Group {i + 1}</h2>
          <SimpleDataGraphic
            width={180}
            height={120}
            left={20}
            right={20}
            top={4}
            bottom={16}
            bind:mouseoverValues
            let:hovered
          >
            <Body bottomBorder>
              {#if hovered}
                <g transition:fade={{ duration: 100 }}>
                  <Area
                    {data}
                    xAccessor="period"
                    yAccessor="y"
                    color="hsla(1, 50%, 90%)"
                  />
                </g>
              {/if}
              <Line
                {data}
                xAccessor="period"
                yAccessor="y"
                color={hovered ? "hsl(1,50%, 50%)" : "hsl(217, 50%, 50%)"}
              />
            </Body>
            {#if i === 0 || (hovered && !(i === 0))}
              <g transition:fade={{ duration: 50 }}>
                <Axis side="bottom" />
              </g>
            {/if}
            {#if mouseoverValues?.x}
              <g transition:fade={{ duration: 50 }}>
                <WithBisector
                  {data}
                  value={mouseoverValues.x}
                  callback={(datum) => datum.period}
                  let:point
                >
                  <PointLabel
                    variant="fixed"
                    lineStart="bodyBottom"
                    lineEnd="point"
                    lineColor="hsla(1,30%, 70%, .3)"
                    lineThickness="scale"
                    format={format(".4r")}
                    x={point.period}
                    y={point.y}
                  />
                </WithBisector>
              </g>
            {/if}
          </SimpleDataGraphic>
        </div>
      {/each}
    </div>
  </GraphicContext>
</section>
