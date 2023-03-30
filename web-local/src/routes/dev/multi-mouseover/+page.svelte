<script lang="ts">
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import Axis from "@rilldata/web-common/components/data-graphic/guides/Axis.svelte";
  import MultiPoint from "@rilldata/web-common/components/data-graphic/marks/MultiPoint.svelte";
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
  <rect x={25} y={25} width={100} height={100} fill="gray" />

  <text x={30} y={30} paint-order="stroke" stroke="white" stroke-width="3"
    >{mouseoverValue?.x}</text
  >

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
    <!-- {#each locations as location, i (location.key)}
      <WithTween
        value={{ ...location, x: xScale(mouseoverValue?.x) }}
        tweenProps={{ duration: 100 }}
        let:output
      >
        <text x={output.x + 6} y={output.value}>label - {output.key}</text>
        <circle cx={output.x} cy={output.value} r={4} />
      </WithTween>
    {/each} -->
    <!-- <MultiMetricMouseoverLabel
      point={points.map((point, i) => ({
        key: point.key,
        y: yScale.invert(i === 0 ? point.value : 10),
        x: mouseoverValue?.x,
        label: point.key,
      }))}
    /> -->
    <MultiPoint
      x={mouseoverValue?.x}
      points={points.map((point, i) => ({
        key: point.key,
        y: yScale.invert(point.value),
        x: mouseoverValue?.x,
        label: point.key,
      }))}
    />
  {/if}

  <Axis side="left" />
  <Axis side="bottom" />
</SimpleDataGraphic>
