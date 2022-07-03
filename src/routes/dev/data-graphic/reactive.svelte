<script lang="ts">
  import { tweened } from "svelte/motion";
  import GraphicContext from "$lib/components/data-graphic/elements/GraphicContext.svelte";
  import SimpleDataGraphic from "$lib/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import Axis from "$lib/components/data-graphic/guides/Axis.svelte";
  import Line from "$lib/components/data-graphic/marks/Line.svelte";
  import { makeTimeSeries } from "./_utils";
  import { cubicOut, elasticOut } from "svelte/easing";
  import Body from "$lib/components/data-graphic/elements/Body.svelte";
  import Area from "$lib/components/data-graphic/marks/Area.svelte";
  import WithTween from "$lib/components/data-graphic/functional-components/WithTween.svelte";
  import WithSimpleLinearScale from "$lib/components/data-graphic/functional-components/WithSimpleLinearScale.svelte";

  const data1 = makeTimeSeries();
  const data2 = makeTimeSeries();

  const width = tweened(300, { duration: 100, easing: cubicOut });
  const height = tweened(300, { duration: 100, easing: cubicOut });
  const margin = tweened(24, { duration: 400, easing: elasticOut });
</script>

<section>
  <h1 class="text-xl">All graph parameters are reactive</h1>

  <div>
    Width
    <input
      type="range"
      bind:value={$width}
      min={100}
      max={400}
      class="w-48 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
    />
  </div>
  <div>
    Height
    <input
      type="range"
      bind:value={$height}
      min={130}
      max={400}
      class="w-48 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
    />
  </div>
  <div>
    Margin
    <input
      type="range"
      bind:value={$margin}
      min={0}
      max={60}
      class="w-48 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
    />
  </div>

  <GraphicContext
    xType="date"
    yType="number"
    height={$height}
    width={$width}
    right={$margin}
    left={$margin}
    top={$margin}
    bottom={$margin}
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
          height={50}
          left={2}
          right={2}
          top={2}
          bottom={2}
          bodyBuffer={0}
        >
          <WithTween
            tweenProps={{ duration: 1000, easing: elasticOut }}
            value={hovered ? outerXScale(new Date(mouseoverValue.x)) : $margin}
            let:output
          >
            <g
              transform="translate({Math.min(
                output,
                config.width - Math.max(50, $width / 4)
              )} {$margin})"
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
</section>
