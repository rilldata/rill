<script lang="ts">
  import { dragTableCell } from "../drag-table-cell";

  import { createEventDispatcher } from "svelte";

  import type { HeaderPosition } from "../types";
  const dispatch = createEventDispatcher();
  export let header;
  export let position: HeaderPosition = "top";
  export let enableResize = true;

  let positionClasses;
  let offsetTop = false;
  $: {
    if (position === "top") {
      positionClasses = "absolute left-0 top-0";
    } else if (position === "left") {
      positionClasses = "absolute left-0 top-0 text-center font-bold";
      offsetTop = true;
    } else if (position === "top-left") {
      positionClasses = "sticky left-0 top-0 z-40  font-bold";
    }
  }

  function focus() {
    dispatch("focus");
  }

  function blur() {
    dispatch("blur");
  }
</script>

<div
  on:mouseover={focus}
  on:mouseleave={blur}
  on:focus={focus}
  on:blur={blur}
  style:transform="translate{position === "left" ? "Y" : "X"}({header.start}px)"
  style:width="{header.size}px"
  style:height="36px"
  class="{positionClasses}
   bg-white text-left border-b border-b-4 border-r border-r-1"
>
  <div
    class="
    text-ellipsis overflow-hidden whitespace-nowrap
  px-4
  border
  border-gray-200
  border-t-0
  border-l-0
  bg-gray-100
  {position === 'top' && 'py-2 text-left'}
  {position === 'left' && 'py-2'}
  {position === 'top-left' && 'py-2 text-center'}
"
  >
    <slot />
    {#if enableResize}
      <div
        use:dragTableCell
        on:resize
        on:dblclick={() => {
          dispatch("reset-column-width");
        }}
        class="absolute top-0 right-0 cursor-col-resize grid place-items-end"
        style:padding-right="1.25px"
        style:width="12px"
        style:height="32px"
      >
        <!-- <div
          style:width="2px"
          style:height="34px"
          class="border border-l border-gray-200"
        /> -->
      </div>
    {/if}
  </div>
</div>
