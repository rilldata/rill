<script lang="ts">
  import { getContext } from "svelte";
  import { outline } from "../actions/outline";
  import WithTween from "../functional-components/WithTween.svelte";
  import type { PointLabelVariant } from "./types";
  import WithSimpleLinearScale from "../functional-components/WithSimpleLinearScale.svelte";
  import { contexts } from "../constants";

  import type { ScaleStore, SimpleConfigurationStore } from "../state/types";

  export let variant: PointLabelVariant = "fixed";
  export let x;
  export let y;
  export let format = (v: any) => v;

  export let line = true;
  export let lineColor = "rgba(0,0,0,.3)";
  export let lineDasharray: string = undefined;
  export let lineThickness: number | "scale" = 1;

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
</script>

<WithSimpleLinearScale
  domain={$yScale.domain()}
  range={[0, 12]}
  let:scale={diameterScale}
>
  <WithTween {tweenProps} value={{ x: $xScale(x), y: $yScale(y) }} let:output>
    {#if line}
      <line
        x1={output.x}
        x2={output.x}
        y1={lineStart === "point" ? output.y : $config[lineStart]}
        y2={lineEnd === "point" ? output.y : $config[lineEnd]}
        stroke={lineColor}
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
    <circle
      fill="hsl(217, 50%, 50%)"
      cx={output.x}
      cy={output.y}
      stroke="hsla(1,90%, 70%, .3)"
      stroke-width={diameterScale(y)}
      r={2}
    />

    {#if variant === "moving"}
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
      stroke="hsla(1,90%, 70%, .3)"
      stroke-width={diameterScale(y)}
      r={2}
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
