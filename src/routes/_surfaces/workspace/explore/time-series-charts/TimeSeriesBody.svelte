<script lang="ts">
  import { cubicOut } from "svelte/easing";
  import SimpleDataGraphic from "$lib/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithTween } from "$lib/components/data-graphic/functional-components";
  import { Axis, Grid, PointLabel } from "$lib/components/data-graphic/guides";
  import { Area, Line } from "$lib/components/data-graphic/marks";
  import { interpolateArray } from "d3-interpolate";
  export let start;
  export let end;
  export let data;
  export let accessor: string;
  export let mouseover = undefined;
  export let key: string;

  // bind and send up to parent to create global mouseover
  export let mouseoverValue = undefined;
</script>

{#if key && data?.length}
  <div>
    <SimpleDataGraphic
      shareYScale={false}
      bind:mouseoverValue
      xMin={start}
      xMax={end}
    >
      {#key key}
        <WithTween
          value={data}
          let:output={tweenedFormattedData}
          tweenProps={{
            duration: 0,
            easing: cubicOut,
            interpolate: interpolateArray,
          }}
        >
          <Area
            data={tweenedFormattedData}
            yAccessor={accessor}
            xAccessor="ts"
          />
          <Line
            data={tweenedFormattedData}
            yAccessor={accessor}
            xAccessor="ts"
          />
        </WithTween>
      {/key}

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
{/if}
