<script lang="ts">
  import { getContext } from "svelte";
  import { dragTableCell } from "../drag-table-cell";
  import type { HeaderPosition, VirtualizedTableConfig } from "../types";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click";

  const config: VirtualizedTableConfig = getContext("config");
  const isDimensionTable = config.table === "DimensionTable";

  export let header;
  export let position: HeaderPosition = "top";
  export let enableResize = true;
  export let borderRight = false;
  export let bgClass = "";
  export let onClick: undefined | (() => void) = undefined;
  export let onShiftClick: undefined | (() => void) = undefined;
  export let onResetColumnWidth: () => void = () => {};
  export let onBlur: () => void = () => {};
  export let onFocus: () => void = () => {};
  export let onResize: (size: number) => void = () => {};

  let isResizing = false;
  let resizeSuppressTimeout;

  function suppressClickAfterResize() {
    isResizing = true;
    clearTimeout(resizeSuppressTimeout);
    resizeSuppressTimeout = setTimeout(() => {
      isResizing = false;
    }, 100);
  }

  let positionClasses: string;
  $: {
    if (position === "top") {
      positionClasses = "absolute left-0 top-0";
    } else if (position === "left") {
      positionClasses = "absolute left-0 top-0 text-center font-semibold";
    } else if (position === "top-left") {
      positionClasses = "sticky left-0 top-0 z-40 font-bold";
    }
  }

  $: borderClassesOuterDiv =
    (borderRight ? "border-r " : "") +
    (isDimensionTable
      ? position === "left"
        ? ""
        : "border-b"
      : "border-b border-b-1 border-r border-r-1 border border-t-0 border-l-0 bg-gray-100");

  const borderClassesInnerDiv = isDimensionTable ? "" : "whitespace-nowrap";
</script>

<button
  on:mouseover={onFocus}
  on:mouseleave={onBlur}
  on:focus={focus}
  on:blur={blur}
  on:click={(e) => {
    if (isResizing) {
      e.stopPropagation();
      return;
    }
    modified({ shift: onShiftClick, click: onClick })(e);
  }}
  style:transform="translate{position === 'left' ? 'Y' : 'X'}({header.start}px)"
  style:padding-right={position === "left" ? "0px" : "10px"}
  style:width="{position === 'top-left' ? header.size + 1 : header.size}px"
  style:height="{position === 'left'
    ? config.rowHeight
    : config.columnHeaderHeight}px"
  class="block {positionClasses}
   {bgClass} text-left {borderClassesOuterDiv}"
>
  <div
    class="
    ui-copy
    text-ellipsis overflow-hidden
    {isDimensionTable ? '' : 'px-4'}
    {borderClassesInnerDiv}
    {position === 'top' && `text-left`}
    {position === 'top-left' &&
      `${isDimensionTable ? 'font-normal' : 'text-center'}`}
    "
  >
    <slot />
    {#if enableResize}
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <div
        role="columnheader"
        tabindex="0"
        use:dragTableCell
        on:resize={(e) => {
          onResize(e.detail);
        }}
        on:resizeend={suppressClickAfterResize}
        on:dblclick={onResetColumnWidth}
        on:click|stopPropagation
        class="absolute top-0 right-0 cursor-col-resize grid place-items-end"
        style:padding-right="1.25px"
        style:width="12px"
        style:height="32px"
      >
        <!-- <div
          style:width="2px"
          style:height="34px"
          class="border border-l"
        /> -->
      </div>
    {/if}
  </div>
</button>
