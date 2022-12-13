<script lang="ts">
  import {
    GraphicContext,
    SimpleDataGraphic,
  } from "$lib/components/data-graphic/elements";
  import { format } from "d3-format";
  import { DynamicallyPlacedLabel } from "../../../data-graphic/guides";
  export let min;
  export let max;
  export let q25;
  export let q50;
  export let q75;
  export let mean;
  export let rowHeight = 24;

  $: values = [
    { label: "min", value: min },
    { label: "max", value: max },
    { label: "q25", value: q25 },
    { label: "q50", value: q50 },
    { label: "q75", value: q75 },
    { label: "mean", value: mean, format: format(".2f") },
  ].reverse();
</script>

{#if values}
  <!-- note: this currently inherits its settings from a parent GraphicContext -->
  <SimpleDataGraphic
    top={0}
    buffer={0}
    height={values?.length * rowHeight}
    let:xScale
  >
    {#each values as { label, value, format = undefined }, i}
      <g transform="translate(0 {(values.length - i - 1) * rowHeight})">
        <GraphicContext height={rowHeight}>
          <circle cx={xScale(value)} cy={rowHeight / 2} r="4" fill="red" />
          <line
            x1={xScale(value)}
            x2={xScale(value)}
            y1={rowHeight / 2 - 8}
            y2={-(rowHeight * (values.length - i - 1))}
            stroke="red"
            opacity={0.5}
          />
          <DynamicallyPlacedLabel
            dy=".35em"
            x={value}
            ry={rowHeight / 2}
            buffer={8}
            colorClass="ui-copy-muted"
          >
            <tspan>
              <tspan class="font-semibold ui-copy-muted">{label}</tspan>
              <tspan class="ui-copy-disabled"
                >{format ? format(value) : value}</tspan
              >
            </tspan>
          </DynamicallyPlacedLabel>
        </GraphicContext>
      </g>
    {/each}
  </SimpleDataGraphic>
{/if}
