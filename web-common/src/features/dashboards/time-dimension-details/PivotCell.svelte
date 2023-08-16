<script lang="ts">
  import type { VirtualItem } from "@tanstack/svelte-virtual";

  export let fixed = false;
  export let element = "td";
  export let rowIdx: number;
  export let item: VirtualItem;
  export let rowHeight: number | undefined = undefined;
  export let renderCell: (rowIdx: number, colIdx: number) => any;
  let _class = "";
  export { _class as class };

  let style = "";
  $: {
    style = `width: ${item.size}px; `;
    if (rowHeight) style += `height: ${rowHeight}px; `;
    if (fixed) {
      style += ` position: sticky; left: ${item.start}px; z-index: 2;`;
    }
  }
</script>

<svelte:element this={element} class={`p-0 ${_class}`} {style}>
  <svelte:component this={renderCell(rowIdx, item.index)} />
</svelte:element>
