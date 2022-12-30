<script lang="ts">
  import {
    Body,
    GraphicContext,
    SimpleDataGraphic,
  } from "@rilldata/web-common/components/data-graphic/elements";
  import {
    WithSimpleLinearScale,
    WithTween,
  } from "@rilldata/web-common/components/data-graphic/functional-components";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import {
    Area,
    Line,
  } from "@rilldata/web-common/components/data-graphic/marks";
  import { cubicOut, elasticOut } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import { makeTimeSeries } from "./_utils";

  const data1 = makeTimeSeries();
  const data2 = makeTimeSeries();

  const width = tweened(300, { duration: 100, easing: cubicOut });
  const height = tweened(300, { duration: 100, easing: cubicOut });
  const margin = tweened(24, { duration: 400, easing: elasticOut });
  const bodyBuffer = tweened(4, { duration: 400, easing: elasticOut });
</script>

<section>
  <h1 class="text-xl pb-8">
    <span style:text-decoration="underline">All</span> graph parameters are reactive
  </h1>

  <div class="grid grid-cols-2 w-max pb-8">
    Width
    <input
      autocomplete="off"
      type="range"
      bind:value={$width}
      min={100}
      max={1500}
      class="w-48 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
    />
    Height
    <input
      autocomplete="off"
      type="range"
      bind:value={$height}
      min={100}
      max={600}
      class="w-48 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
    />
    Margin
    <input
      autocomplete="off"
      type="range"
      bind:value={$margin}
      min={0}
      max={64}
      class="w-48 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
    />
    bodyBuffer
    <input
      autocomplete="off"
      type="range"
      bind:value={$bodyBuffer}
      min={0}
      max={64}
      class="w-48 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
    />
  </div>

  <div class="border border-gray-200 w-max">
    <GraphicContext
      xType="date"
      yType="number"
      height={$height}
      width={$width}
      right={$margin}
      left={$margin}
      top={$margin}
      bottom={$margin}
      bodyBuffer={$bodyBuffer}
      let:config
    >
      <SimpleDataGraphic
        right={$margin}
        let:mouseoverValue
        let:xScale={outerXScale}
        let:hovered
        let:config
      >
        <Axis side="bottom" />
        <Axis side="left" />

        <Body border>
          <Area data={data1} xAccessor="period" yAccessor="value" />
          <Line data={data1} xAccessor="period" yAccessor="value" />
          <GraphicContext
            xType="number"
            yType="number"
            width={Math.max(50, $width / 4)}
            height={Math.max(50, $height / 5)}
            left={2}
            right={2}
            top={2}
            bottom={2}
            bodyBuffer={$bodyBuffer / 2}
          >
            <WithTween
              tweenProps={{ duration: 500, easing: elasticOut }}
              value={hovered
                ? outerXScale(new Date(mouseoverValue.x))
                : $margin}
              let:output
            >
              <g
                transform="translate({Math.min(
                  output,
                  config.width - Math.max(50, $width / 4)
                )} {$margin + $bodyBuffer})"
              >
                <WithSimpleLinearScale
                  domain={[100, 400]}
                  range={[100, 60]}
                  let:scale
                >
                  <Body bg bgColor="hsl(217,5%,{~~scale($height)}%)">
                    <Line
                      data={data2}
                      xAccessor="period"
                      yAccessor="value"
                      color="hsl({~~scale($height * 5)}, 50%, 50%)"
                    />
                  </Body>
                </WithSimpleLinearScale>
              </g>
            </WithTween>
          </GraphicContext>
        </Body>
      </SimpleDataGraphic>
    </GraphicContext>
  </div>
</section>
