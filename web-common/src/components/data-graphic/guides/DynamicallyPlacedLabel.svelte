<script lang="ts">
  import { getContext, onDestroy, onMount } from "svelte";
  import { tweened } from "svelte/motion";
  import { outline } from "../../../components/data-graphic/actions/outline";
  import { contexts } from "../../../components/data-graphic/constants";
  import type {
    ScaleStore,
    SimpleConfigurationStore,
  } from "../../../components/data-graphic/state/types";

  const config = getContext(contexts.config) as SimpleConfigurationStore;
  const xScale = getContext(contexts.scale("x")) as ScaleStore;
  const yScale = getContext(contexts.scale("y")) as ScaleStore;

  export let x: number = undefined;
  export let y: number = undefined;
  export let rx: number = undefined;
  export let ry: number = undefined;
  export let dy: string | number = undefined;

  export let color: string = undefined;
  export let colorClass: string = undefined;
  export let location: "left" | "right" = "right";
  export let buffer = 8;

  let element;

  let elementWidth = 0;
  let elementHeight = 0;
  let elementX = 0;
  let elementY = 0;
  let xOffset = tweened(buffer, { duration: 0 });

  function update() {
    let bb = element.getBBox();
    elementWidth = bb.width;
    elementHeight = bb.height;
    elementX = bb.x;
    elementY = bb.y;
    if (location === "right" && elementX + elementWidth > $config.plotRight) {
      xOffset.set(-elementWidth - buffer);
    } else {
      xOffset.set(buffer);
    }
  }

  let resize;
  let mutation;

  onMount(() => {
    // resize if element updates.
    resize = new ResizeObserver(() => {
      if (element) update();
    });
    // reposition if element DOM parameters change.
    mutation = new MutationObserver(() => {
      update();
    });
    mutation.observe(element, {
      attributes: true,
      childList: true,
    });
    resize.observe(element);
    update();
  });

  onDestroy(() => {
    resize.unobserve(element);
    mutation.disconnect();
  });

  $: trueX = rx || $xScale(x);
  $: trueY = ry || $yScale(y);
</script>

<g transform="translate({$xOffset} 0)">
  <text
    use:outline
    class={colorClass}
    style={color ? `color:${color}` : undefined}
    bind:this={element}
    x={trueX}
    y={trueY}
    {dy}
  >
    <slot />
  </text>
</g>
