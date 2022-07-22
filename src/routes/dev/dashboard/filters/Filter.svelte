<script lang="ts">
  import Cancel from "$lib/components/icons/Cancel.svelte";
  import { cubicOut } from "svelte/easing";
  import { slide } from "svelte/transition";

  export let collapseDirection: "vertical" | "horizontal" = undefined;

  /** a custom action for Chip collapsing */
  function fold(
    node,
    {
      delay = 0,
      duration = 300,
      easing = cubicOut,
      rightOffset = 0,
      leftOffset = 8,
    }
  ) {
    const style = getComputedStyle(node);
    const width = parseFloat(style.width);
    const opacity = +style.opacity;
    const paddingLeft = parseFloat(style.paddingLeft);
    const paddingRight = parseFloat(style.paddingRight);
    const marginLeft = parseFloat(style.marginTop);
    const marginRight = parseFloat(style.marginBottom);
    const borderLeftWidth = parseFloat(style.borderLeftWidth);
    const borderRightWidth = parseFloat(style.borderRightWidth);

    return {
      delay,
      duration,
      easing,
      css: (t) => `
        outline: 1px solid orange;
        overflow: hidden;
        white-space: nowrap;
        opacity: ${Math.min(t * 20, 1) * opacity};
        width: ${t * width}px;
        padding-right: ${t * paddingRight}px;
        padding-left: ${t * paddingLeft}px;
        margin-left: ${t * marginLeft - (1 - t) * leftOffset}px;
        margin-right: ${t * marginRight - (1 - t) * rightOffset}px;
        border-left-width: ${t * borderLeftWidth}px;
        border-right-width: ${t * borderRightWidth}px;
      `,
    };
  }
  let duration = 300;
  let action;
  let params = { duration: duration };
  if (collapseDirection === "vertical") {
    action = slide;
  } else if (collapseDirection === "horizontal") {
    action = fold;
  } else {
    action = slide;
    params = { duration: 0 };
  }

  let overflowClass = "";
  /** special styling to prevent ellipses on transition in and transition out */
  setTimeout(() => {
    overflowClass = "text-ellipsis";
  }, duration);
</script>

<div
  class="filter grid grid-flow-col items-center w-max gap-x-1 bg-gray-200 p-1 pr-2 pt-1 pb-1 rounded"
  style:grid-template-columns="max-content auto"
  style:max-width="200px"
  in:action|local={params}
  out:action|local={params}
>
  <button
    style:width="16px"
    style:height="16px"
    class="hover:bg-gray-300 w-4 grid place-items-center rounded-sm"
    on:click><Cancel size="16px" /></button
  >
  <div class=" {overflowClass} overflow-hidden whitespace-nowrap">
    <slot />
  </div>
</div>
