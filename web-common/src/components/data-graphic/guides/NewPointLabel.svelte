<script lang="ts">
  import { outline } from "@rilldata/web-common/components/data-graphic/actions/outline";
  import { WithTween } from "@rilldata/web-common/components/data-graphic/functional-components";
  import WithGraphicContexts from "@rilldata/web-common/components/data-graphic/functional-components/WithGraphicContexts.svelte";
  import { justEnoughPrecision } from "@rilldata/web-common/lib/formatters";
  import { cubicOut } from "svelte/easing";
  import { fade } from "svelte/transition";
  export let point;
  export let xAccessor;
  export let yAccessor;
  export let location: "left" | "right" = "right";
  export let showText = true;
  export let showPoint = true;
  export let showReferenceLine = true;
  export let showDistanceFromZero = true;
  export let format = justEnoughPrecision;

  let lastAvailablePoint;

  /**
   * If the point is null, we want to use the last available point to
   * calculate the y position of the label. This is so that the label
   * doesn't jump around when the data is null.
   */
  $: if (point[yAccessor]) {
    lastAvailablePoint = { ...point };
  }

  function scaleFromOrigin(
    node,
    {
      delay = 0,
      duration = 1000,
      easing = cubicOut,
      start = 0,
      opacity = 0,
    } = {}
  ) {
    const style = getComputedStyle(node);
    const target_opacity = +style.opacity;
    const transform = style.transform === "none" ? "" : style.transform;

    const sd = 1 - start;
    const od = target_opacity * (1 - opacity);

    return {
      delay,
      duration,
      easing,
      css: (_t, u) => `
      transform-box: fill-box;
      transform-origin: 100% 100%;
			transform: ${transform} scale(${1 - sd * u});
			opacity: ${target_opacity - od * u}
		`,
    };
  }
</script>

<WithGraphicContexts let:xScale let:yScale let:config>
  {@const isNull = point[yAccessor] == null}
  {@const x = xScale(point[xAccessor])}
  {@const y = !isNull
    ? yScale(point[yAccessor])
    : lastAvailablePoint
    ? yScale(lastAvailablePoint[yAccessor])
    : (config.plotBottom - config.plotTop) / 2}
  <WithTween
    value={{ x, y, dy: point?.[yAccessor] || 0 }}
    tweenProps={{ duration: 50 }}
    let:output
  >
    {@const text = isNull
      ? "no data"
      : format
      ? format(point[yAccessor])
      : point[yAccessor]}
    {#if showReferenceLine}
      <line
        transition:fade|local={{ duration: 100 }}
        x1={output.x}
        x2={output.x}
        y1={config.plotTop}
        y2={config.plotBottom}
        stroke-width="1"
        class="stroke-gray-400"
        stroke-dasharray="2,1"
      />
    {/if}
    {#if showText}
      <text
        class:fill-gray-400={isNull}
        class:italic={isNull}
        use:outline
        x={output.x}
        y={output.y}
        text-anchor={location === "left" ? "end" : "start"}
        dx={8 * (location === "left" ? -1 : 1)}
        dy=".35em"
      >
        {text}
      </text>
    {/if}
    {#if !isNull && showDistanceFromZero}
      <line
        transition:fade|local={{ duration: 100 }}
        x1={output.x}
        x2={output.x}
        y1={yScale.range().at(0)}
        y2={output.y}
        stroke-width="4"
        class="stroke-blue-300"
      />
    {/if}
    {#if !isNull && showPoint}
      <circle
        transition:scaleFromOrigin|local
        cx={output.x}
        cy={output.y}
        r="3"
        fill="blue"
      />
    {/if}
  </WithTween>
</WithGraphicContexts>
