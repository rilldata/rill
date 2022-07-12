<script lang="ts">
  import SimpleDataGraphic from "$lib/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithTween } from "$lib/components/data-graphic/functional-components";
  import { Axis, Grid, PointLabel } from "$lib/components/data-graphic/guides";
  import { Area, Line } from "$lib/components/data-graphic/marks";
  export let data;
  export let accessor: string;
  export let mouseover = undefined;

  // bind and send up to parent to create global mouseover
  export let mouseoverValue = undefined;
</script>

<div>
  <SimpleDataGraphic shareYScale={false} bind:mouseoverValue>
    <!-- <WithTween
    value={data}
    tweenProps={{ duration: 400 }}
    let:output={tweenedDataset}
  > -->
    <Area {data} yAccessor={accessor} xAccessor="ts" />
    <Line {data} yAccessor={accessor} xAccessor="ts" />
    <!-- </WithTween> -->

    <Axis side="right" />
    <Grid />
    {#if mouseover}
      <PointLabel
        tweenProps={{ duration: 50 }}
        x={mouseover.ts}
        y={mouseover[accessor]}
      />
    {/if}
  </SimpleDataGraphic>
</div>
