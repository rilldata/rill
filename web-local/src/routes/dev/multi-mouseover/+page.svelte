<script lang="ts">
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithTween } from "@rilldata/web-common/components/data-graphic/functional-components";
  import Axis from "@rilldata/web-common/components/data-graphic/guides/Axis.svelte";
  import { preventVerticalOverlap } from "@rilldata/web-common/components/data-graphic/marks/prevent-vertical-overlap";
</script>

<SimpleDataGraphic
  xType="number"
  yType="number"
  width={600}
  height={300}
  yMin={0}
  yMax={100}
  xMin={0}
  xMax={100}
  let:xScale
  let:yScale
  let:config
  let:mouseoverValue
>
  <text x={30} y={30}>{mouseoverValue?.x}</text>
  {#if mouseoverValue?.x}
    {@const points = [
      { key: 0, value: yScale(Math.sin(mouseoverValue?.x / 10) * 10 + 50) },
      { key: 1, value: yScale(Math.cos(mouseoverValue?.x / 10) * 10 + 50) },
    ]}
    <!-- create the overlap points-->
    {@const locations = preventVerticalOverlap(
      points,
      config.plotTop,
      config.plotBottom,
      12,
      4
    )}
    <!--  -->
    {#each locations as location, i (location.key)}
      <WithTween
        value={{ ...location, x: xScale(mouseoverValue?.x) }}
        tweenProps={{ duration: 100 }}
        let:output
      >
        <text x={output.x + 6} y={output.value}>label - {output.key}</text>
        <circle cx={output.x} cy={output.value} r={4} />
      </WithTween>
    {/each}
  {/if}
  <Axis side="left" />
  <Axis side="bottom" />
</SimpleDataGraphic>
