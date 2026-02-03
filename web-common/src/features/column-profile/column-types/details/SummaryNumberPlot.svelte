<script lang="ts">
  import { INTEGERS } from "@rilldata/web-common/lib/duckdb-data-types";
  import { justEnoughPrecision } from "@rilldata/web-common/lib/formatters";
  import type { V1NumericStatistics } from "@rilldata/web-common/runtime-client";
  import type { ScaleLinear } from "d3-scale";
  import { format } from "d3-format";
  import { tweened } from "svelte/motion";

  export let summary: V1NumericStatistics | undefined;
  export let rowHeight = 24;
  export let type: string;
  export let xScale: ScaleLinear<number, number>;
  export let plotRight: number;

  $: formatter = INTEGERS.has(type) ? format(".0f") : justEnoughPrecision;
  $: values = summary
    ? [
        { label: "min", value: summary.min, format: formatter },
        { label: "q25", value: summary.q25, format: formatter },
        { label: "q50", value: summary.q50, format: formatter },
        { label: "q75", value: summary.q75, format: formatter },
        { label: "max", value: summary.max, format: formatter },
        {
          label: "mean",
          value: summary.mean,
          format: justEnoughPrecision,
        },
      ].reverse()
    : undefined;

  $: totalHeight = (values?.length ?? 0) * rowHeight;

  // Label positioning: flip label to left side if it overflows plotRight
  const LABEL_BUFFER = 6;

  function dynamicLabel(node: SVGTextElement) {
    const offset = tweened(LABEL_BUFFER, { duration: 0 });

    function update() {
      const bb = node.getBBox();
      if (bb.x + bb.width > plotRight) {
        void offset.set(-bb.width - LABEL_BUFFER);
      } else {
        void offset.set(LABEL_BUFFER);
      }
    }

    const resize = new ResizeObserver(() => update());
    const mutation = new MutationObserver(() => update());
    mutation.observe(node, { attributes: true, childList: true });
    resize.observe(node);
    update();

    const unsub = offset.subscribe((v) => {
      const g = node.parentElement;
      if (g) g.setAttribute("transform", `translate(${v} 0)`);
    });

    return {
      destroy() {
        resize.disconnect();
        mutation.disconnect();
        unsub();
      },
    };
  }
</script>

{#if values}
  <svg class="overflow-visible" width="100%" height={totalHeight}>
    {#each values as { label, value, format = undefined }, i (i)}
      {@const px = xScale(value ?? 0)}
      {@const rowY = (values.length - i - 1) * rowHeight}
      <g transform="translate(0 {rowY})">
        <circle
          cx={px}
          cy={rowHeight / 2}
          r="2.5"
          class="fill-destructive"
          opacity={0.8}
        />
        <line
          x1={px}
          x2={px}
          y1={rowHeight / 2 - 1}
          y2={-rowY}
          class="stroke-destructive"
          opacity={0.5}
        />
        <g>
          <text
            use:dynamicLabel
            class="fill-fg-secondary"
            x={px}
            y={rowHeight / 2}
            dy=".35em"
          >
            <tspan class="font-semibold text-fg-muted">{label}</tspan>
            <tspan class="text-fg-disabled"
              >{format && value !== undefined
                ? format(value)
                : (value ?? "—")}</tspan
            >
          </text>
        </g>
      </g>
    {/each}
  </svg>
{/if}
