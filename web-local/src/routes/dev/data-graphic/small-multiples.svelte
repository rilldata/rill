<script lang="ts">
  import { format } from "d3-format";
  import { fade } from "svelte/transition";

  import type { DomainCoordinates } from "@rilldata/web-common/components/data-graphic/constants/types";
  import {
    Body,
    GraphicContext,
    SimpleDataGraphic,
  } from "@rilldata/web-common/components/data-graphic/elements";
  import { WithBisector } from "@rilldata/web-common/components/data-graphic/functional-components";
  import {
    Axis,
    PointLabel,
  } from "@rilldata/web-common/components/data-graphic/guides";
  import {
    Area,
    Line,
  } from "@rilldata/web-common/components/data-graphic/marks";
  import { cubicOut } from "svelte/easing";
  import { tweened } from "svelte/motion";

  /** bind the mouseoverValue of a graph to this variable to share
   * with other graphs
   */
  let mouseoverValue: DomainCoordinates;

  function smooth(data, accessor, windowSize = 7) {
    return data.map((datum, i) => {
      const window = data.slice(Math.max(0, i - windowSize), i);
      const v = window.reduce((acc, v) => acc + v[accessor], 0);
      return { ...datum, [accessor]: v };
    });
  }

  function makeData(length = 180) {
    let y = Math.random() * 150;
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

  let datasets = tweened(
    Array.from({ length: 36 }).map(() => makeData(60)),
    { duration: 500, easing: cubicOut }
  );
</script>

<section>
  <h1 class="text-xl">
    Small Multiples <button
      class="text-sm font-normal p-3 pt-1 pb-1 bg-gray-100 rounded mb-6 hover:bg-gray-200"
      on:click={() => {
        datasets.set(Array.from({ length: 36 }).map(() => makeData(60)));
      }}>randomize</button
    >
  </h1>
  <GraphicContext xType="date" yType="number">
    <div class="flex flex-row flex-wrap gap-3 w-max-screen">
      {#each $datasets as data, i (i)}
        <div>
          <h2 class="pl-5">Group {i + 1}</h2>
          <SimpleDataGraphic
            width={180}
            height={116}
            left={32}
            right={20}
            top={4}
            bottom={16}
            fontSize={10}
            bind:mouseoverValue
            let:hovered
          >
            <Body bottomBorder>
              {#if hovered}
                <g transition:fade={{ duration: 100 }}>
                  <Area
                    {data}
                    xAccessor="period"
                    yAccessor="y"
                    color="hsla(1, 50%, 90%, .5)"
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
              <g>
                <Axis side="bottom" />
                <Axis side="left" />
              </g>
            {/if}
            {#if mouseoverValue?.x}
              <g transition:fade={{ duration: 50 }}>
                <WithBisector
                  {data}
                  value={mouseoverValue.x}
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
