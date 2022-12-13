<script lang="ts">
  import { dragTableCell } from "../drag-table-cell";

  import { createEventDispatcher, getContext } from "svelte";

  import type { HeaderPosition, VirtualizedTableConfig } from "../types";

  const config: VirtualizedTableConfig = getContext("config");
  const isDimensionTable = config.table === "DimensionTable";

  const dispatch = createEventDispatcher();
  export let header;
  export let position: HeaderPosition = "top";
  export let enableResize = true;
  export let borderRight = false;

  let positionClasses;
  let offsetTop = false;
  $: {
    if (position === "top") {
      positionClasses = "absolute left-0 top-0";
    } else if (position === "left") {
      positionClasses = "absolute left-0 top-0 text-center font-semibold";
      offsetTop = true;
    } else if (position === "top-left") {
      positionClasses = "sticky left-0 top-0 z-40 ui-copy-strong";
    }
  }

  $: borderClassesOuterDiv =
    (borderRight ? "border-r " : "") +
    (isDimensionTable
      ? position === "left"
        ? ""
        : "border-b"
      : "border-b border-b-4 border-r border-r-1");

  const borderClassesInnerDiv = isDimensionTable
    ? ""
    : "whitespace-nowrap border border-gray-200 border-t-0 border-l-0 bg-gray-100";

  const paddingVerticalTop = config.columnHeaderHeight <= 28 ? "py-1" : "py-2";
  const paddingVerticalLeft = config.rowHeight <= 28 ? "" : "py-2";

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
  on:click
  style:transform="translate{position === "left" ? "Y" : "X"}({header.start}px)"
  style:width="{header.size}px"
  style:height="{position === "left"
    ? config.rowHeight
    : config.columnHeaderHeight}px"
  class="{positionClasses}
   surface text-left {borderClassesOuterDiv}"
>
  <div
    class="
    ui-copy
    text-ellipsis overflow-hidden
    {isDimensionTable ? (position === 'left' ? '' : 'px-1') : 'px-4'}
    {borderClassesInnerDiv}
    {position === 'top' && `${paddingVerticalTop} text-left`}
    {position === 'left' && paddingVerticalLeft}
    {position === 'top-left' &&
      `${paddingVerticalTop} ${
        isDimensionTable ? 'font-normal' : 'text-center'
      }`}
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
