<script lang="ts">
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import StaticCanvasRow from "@rilldata/web-common/features/canvas/StaticCanvasRow.svelte";
  import CanvasPdfExportHeader from "./CanvasPdfExportHeader.svelte";
  import CanvasPdfExportTabGroup from "./CanvasPdfExportTabGroup.svelte";

  // A dedicated, read-only render of the whole canvas used solely as the PDF
  // capture target. It reuses the live dashboard's block/component rendering off
  // the same canvas store, but renders every component eagerly (lazy={false}) so
  // the capture is complete without touching the live dashboard's lazy-load
  // state. Tab groups are flattened for print: every exported tab (all tabs, or
  // only each group's active one — see exportAllTabs) renders as a label band
  // followed by its rows, stacked in strip order.
  // See captureCanvasBlocks, which reads from #canvas-pdf-export-view.
  export let canvasName: string;
  export let instanceId: string;
  // Render width in px; matches the on-screen content width for fidelity.
  export let width: number;

  $: ({
    canvasEntity: { componentsStore, _rows, layout, exportAllTabs },
  } = getCanvasStore(canvasName, instanceId));

  $: components = $componentsStore;
  $: rows = $_rows;
  $: blocks = $layout;
</script>

<div
  id="canvas-pdf-export-view"
  data-canvas-name={canvasName}
  data-instance-id={instanceId}
  class="flex flex-col items-center bg-surface-background"
  style:width="{width}px"
>
  <CanvasPdfExportHeader {canvasName} {instanceId} maxWidth={width} />

  <div
    class="row-container w-full h-fit flex flex-col items-center relative"
    style:width="{width}px"
  >
    {#each blocks as block (block.kind === "tab-group" ? `g-${block.group.name}` : `r-${block.rowIndex}`)}
      {#if block.kind === "tab-group"}
        <CanvasPdfExportTabGroup
          group={block.group}
          {components}
          maxWidth={width}
          allTabs={$exportAllTabs}
        />
      {:else if rows[block.freeRowIndex]}
        <StaticCanvasRow
          row={rows[block.freeRowIndex]}
          rowIndex={block.rowIndex}
          {components}
          maxWidth={width}
          navigationEnabled={false}
          lazy={false}
        />
      {/if}
    {/each}
  </div>
</div>

<style>
  .row-container {
    container-type: inline-size;
    container-name: canvas-container;
  }

  /* Scrollbars on overflowing components (tables, legends) would be rasterized
     into the capture. Applied per element via `*` because html-to-image inlines
     each element's computed styles into its clone: computed scrollbar-width
     survives the cloning, whereas ::-webkit-scrollbar stylesheet rules do not. */
  #canvas-pdf-export-view :global(*) {
    scrollbar-width: none;
  }
</style>
