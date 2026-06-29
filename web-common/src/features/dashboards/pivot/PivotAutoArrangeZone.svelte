<script lang="ts">
  import Pivot from "@rilldata/web-common/components/icons/Pivot.svelte";
  import { modifierHeld } from "@rilldata/web-common/lib/modifier-key";
  import { slide } from "svelte/transition";
  import { dragDataStore } from "./DragList.svelte";
  import type { PivotChipData } from "./types";

  export let rows: PivotChipData[];
  export let columns: PivotChipData[];
  export let setRows: (items: PivotChipData[]) => void;
  export let setColumns: (items: PivotChipData[]) => void;

  $: dragData = $dragDataStore;
  // Auto-arrange only makes sense for mixed tags: a pure-dim or pure-measure
  // tag has a single natural target so the user can drop on Rows or Columns
  // directly.
  $: visible =
    !!dragData?.tagPayload &&
    dragData.tagPayload.dimensions.length > 0 &&
    dragData.tagPayload.measures.length > 0;

  let hover = false;

  function handleDrop(e: MouseEvent) {
    if (!dragData?.tagPayload) return;
    const replace = e.metaKey || e.ctrlKey;
    const { dimensions, measures } = dragData.tagPayload;
    if (replace) {
      // Both zones are set to the natural slice of the tag; cross-zone
      // cleanup happens implicitly since the zones don't overlap.
      setRows(dimensions);
      setColumns(measures);
    } else {
      if (dimensions.length > 0) {
        const existingRows = new Set(rows.map((c) => c.id));
        const dimAdds = dimensions.filter((d) => !existingRows.has(d.id));
        if (dimAdds.length > 0) setRows([...rows, ...dimAdds]);
      }
      if (measures.length > 0) {
        const existingCols = new Set(columns.map((c) => c.id));
        const measAdds = measures.filter((m) => !existingCols.has(m.id));
        if (measAdds.length > 0) setColumns([...columns, ...measAdds]);
      }
    }
    dragDataStore.set(null);
    hover = false;
  }
</script>

{#if visible && dragData?.tagPayload}
  <div class="header-row" transition:slide={{ duration: 160, axis: "y" }}>
    <span class="row-label">
      <Pivot size="16px" /> Auto
    </span>
    <div
      role="presentation"
      class="auto-arrange-zone"
      class:hover
      class:replace={$modifierHeld}
      onmouseenter={() => (hover = true)}
      onmouseleave={() => (hover = false)}
      onmouseup={handleDrop}
      aria-label={$modifierHeld
        ? "Drop here to replace rows and columns with this tag"
        : "Drop here to auto-arrange tag"}
    >
      {#if $modifierHeld}
        Drop to <strong>replace</strong>:
        <strong>{dragData.tagPayload.dimensions.length}</strong>
        {dragData.tagPayload.dimensions.length === 1 ? "dim" : "dims"}
        → rows,
        <strong>{dragData.tagPayload.measures.length}</strong>
        {dragData.tagPayload.measures.length === 1 ? "measure" : "measures"}
        → columns
      {:else}
        Drop here to split:
        <strong>{dragData.tagPayload.dimensions.length}</strong>
        {dragData.tagPayload.dimensions.length === 1 ? "dim" : "dims"}
        → rows,
        <strong>{dragData.tagPayload.measures.length}</strong>
        {dragData.tagPayload.measures.length === 1 ? "measure" : "measures"}
        → columns (<span class="kbd">⌘</span> + Drop to replace)
      {/if}
    </div>
  </div>
{/if}

<style lang="postcss">
  .header-row {
    @apply flex items-center gap-x-2 px-2;
  }

  .row-label {
    @apply w-20 flex items-center gap-x-1 flex-shrink-0 text-fg-secondary;
  }

  .auto-arrange-zone {
    @apply flex-1 flex items-center justify-center gap-x-1;
    @apply rounded-sm border border-dashed border-blue-400;
    @apply bg-blue-50/50 text-fg-secondary text-xs;
    @apply py-2 px-3;
  }

  .auto-arrange-zone.hover {
    @apply bg-blue-100 border-blue-500 text-fg-primary;
  }

  .auto-arrange-zone.replace {
    @apply border-amber-500 bg-amber-50/50;
  }

  .auto-arrange-zone.replace.hover {
    @apply bg-amber-100 border-amber-600;
  }

  .kbd {
    @apply inline-block px-1 py-px ml-1 rounded-sm border;
    @apply text-[10px] font-mono text-fg-secondary;
  }
</style>
