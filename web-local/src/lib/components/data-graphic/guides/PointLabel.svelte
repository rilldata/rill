<script lang="ts">
  import { getContext } from "svelte";
  import { outline } from "../actions/outline";
  import { contexts } from "../constants";
  import { WithSimpleLinearScale, WithTween } from "../functional-components";
  import type { PointLabelVariant } from "./types";

  import type { ScaleStore, SimpleConfigurationStore } from "../state/types";

  export let variant: PointLabelVariant = "fixed";
  export let x;
  export let y;
  export let format = (v: any) => v;

  export let line = true;
  export let lineColor = "stroke-gray-400 dark:stroke-gray-500";
  export let lineDasharray: string = undefined;
  export let lineThickness: number | "scale" = 1;

  export let showMovingPoint = true;

  export let tweenProps = { duration: 0 };

  type LinePositionChoices =
    | "graphicBottom"
    | "bodyBottom"
    | "graphicTop"
    | "bodyTop"
    | "bodyBottom"
    | "plotTop"
    | "plotBottom"
    | "point";
  export let lineStart: LinePositionChoices = "plotBottom";
  export let lineEnd: LinePositionChoices = "plotTop";

  const xScale = getContext(contexts.scale("x")) as ScaleStore;
  const yScale = getContext(contexts.scale("y")) as ScaleStore;
  const config = getContext(contexts.config) as SimpleConfigurationStore;

  $: input = { x: $xScale(x), y: $yScale(y) };
</script>

{#if x !== undefined && y !== undefined}
  <WithSimpleLinearScale
    domain={$yScale.domain()}
    range={[0, 12]}
    let:scale={diameterScale}
  >
    <WithTween {tweenProps} value={input} let:output>
      {#if line && output?.x}
        <line
          x1={output.x}
          x2={output.x}
          y1={lineStart === "point" ? output.y : $config[lineStart]}
          y2={lineEnd === "point" ? output.y : $config[lineEnd]}
          class={lineColor}
          stroke-dasharray={lineDasharray}
          stroke-width={lineThickness === "scale"
            ? diameterScale(y)
            : lineThickness}
        />
      {/if}
      <text
        x1={output.x}
        x2={output.x}
        y1={0}
        y2={$config.height}
        stroke="gray"
      />
      {#if showMovingPoint}
        <circle fill="hsl(217, 50%, 50%)" cx={output.x} cy={output.y} r={3} />
      {/if}
      {#if variant === "moving" && showMovingPoint}
        <text
          use:outline={{ color: "rgba(255,255,255,.7)" }}
          dy=".35em"
          x={output.x + 16}
          y={output.y}>{format(y)}</text
        >
      {/if}
    </WithTween>
    {#if variant === "fixed"}
      <circle
        fill="hsl(217, 50%, 50%)"
        r={4}
        cx={$config.bodyLeft + 2}
        cy={$config.bodyTop + 2}
      />
      <text
        use:outline={{ color: "rgba(255,255,255,.7)" }}
        x={$config.bodyLeft + 16}
        y={$config.bodyTop + 6}>{format(y)}</text
      >
    {/if}
  </WithSimpleLinearScale>
{/if}
