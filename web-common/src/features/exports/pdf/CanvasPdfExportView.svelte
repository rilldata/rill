<script lang="ts">
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import StaticCanvasRow from "@rilldata/web-common/features/canvas/StaticCanvasRow.svelte";
  import CanvasPdfExportHeader from "./CanvasPdfExportHeader.svelte";

  // A dedicated, read-only render of the whole canvas used solely as the PDF
  // capture target. It reuses the live dashboard's row/component rendering off
  // the same canvas store, but renders every component eagerly (lazy={false}) so
  // the capture is complete without touching the live dashboard's lazy-load
  // state. See captureCanvasBlocks, which reads from #canvas-pdf-export-view.
  export let canvasName: string;
  export let instanceId: string;
  // Render width in px; matches the on-screen content width for fidelity.
  export let width: number;

  $: ({
    canvasEntity: { componentsStore, _rows },
  } = getCanvasStore(canvasName, instanceId));

  $: components = $componentsStore;
  $: rows = $_rows;
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
    {#each rows as row, rowIndex (rowIndex)}
      <StaticCanvasRow
        {row}
        {rowIndex}
        {components}
        maxWidth={width}
        navigationEnabled={false}
        lazy={false}
      />
    {/each}
  </div>
</div>

<style>
  .row-container {
    container-type: inline-size;
    container-name: canvas-container;
  }
</style>
