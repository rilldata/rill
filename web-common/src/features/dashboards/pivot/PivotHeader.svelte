<script lang="ts">
  import Column from "@rilldata/web-common/components/icons/Column.svelte";
  import Pivot from "@rilldata/web-common/components/icons/Pivot.svelte";
  import Row from "@rilldata/web-common/components/icons/Row.svelte";
  import { splitPivotChips } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils.ts";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { modifierHeld } from "@rilldata/web-common/lib/modifier-key";
  import { slide } from "svelte/transition";
  import { getStateManagers } from "../state-managers/state-managers";
  import DragList, { dragDataStore } from "./DragList.svelte";
  import { lastNestState } from "./PivotToolbar.svelte";
  import { PivotChipType, type PivotChipData, type PivotState } from "./types";

  export let pivotState: PivotState;
  export let setRows: (items: PivotChipData[]) => void;
  export let setColumns: (items: PivotChipData[]) => void;

  const { exploreName } = getStateManagers();

  $: ({ rows, columns, tableMode } = pivotState);
  $: splitColumns = splitPivotChips(columns);
  $: fullColumns = splitColumns.dimension.concat(splitColumns.measure);
  $: isFlat = tableMode === "flat";

  $: dragData = $dragDataStore;
  // Auto-arrange only makes sense for mixed tags. A pure-dimension or
  // pure-measure tag has a single natural target — drop on Rows or
  // Columns directly.
  $: showAutoArrange =
    !isFlat &&
    !!dragData?.tagPayload &&
    dragData.tagPayload.dimensions.length > 0 &&
    dragData.tagPayload.measures.length > 0;

  let autoArrangeHover = false;

  function handleAutoArrangeDrop(e: MouseEvent) {
    if (!dragData?.tagPayload) return;
    const replace = e.metaKey || e.ctrlKey;
    const { dimensions, measures } = dragData.tagPayload;
    if (replace) {
      metricsExplorerStore.replacePivotRows($exploreName, dimensions);
      metricsExplorerStore.replacePivotColumns($exploreName, measures);
    } else {
      if (dimensions.length > 0) {
        metricsExplorerStore.addPivotFields(
          $exploreName,
          dimensions,
          "rows",
        );
      }
      if (measures.length > 0) {
        metricsExplorerStore.addPivotFields(
          $exploreName,
          measures,
          "columns",
        );
      }
    }
    dragDataStore.set(null);
    autoArrangeHover = false;
  }

  function updateColumn(items: PivotChipData[]) {
    // Reset lastNestState when columns are updated
    lastNestState.set(null);
    setColumns(items);
  }

  function updateRows(items: PivotChipData[]) {
    const filtered = items.filter(
      (item) => item.type !== PivotChipType.Measure,
    );
    setRows(filtered);
  }
</script>

<div class="header" transition:slide>
  {#if !isFlat}
    <div
      class="header-row"
      transition:slide={{
        duration: 200,
        axis: "y",
      }}
    >
      <span class="row-label">
        <Row size="16px" /> Rows
      </span>
      <DragList
        zone="rows"
        placeholder="Drag dimensions here"
        items={rows}
        onUpdate={updateRows}
      />
    </div>
  {/if}
  {#if showAutoArrange && dragData?.tagPayload}
    <div
      class="header-row"
      transition:slide={{ duration: 160, axis: "y" }}
    >
      <span class="row-label">
        <Pivot size="16px" /> Auto
      </span>
      <div
        role="presentation"
        class="auto-arrange-zone"
        class:hover={autoArrangeHover}
        class:replace={$modifierHeld}
        onmouseenter={() => (autoArrangeHover = true)}
        onmouseleave={() => (autoArrangeHover = false)}
        onmouseup={handleAutoArrangeDrop}
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
          → columns
          (<span class="kbd">⌘</span> + Drop to replace)
        {/if}
      </div>
    </div>
  {/if}
  <div class="header-row">
    <div class="row-label">
      <Column size="16px" /> Columns
    </div>

    <DragList
      zone="columns"
      {tableMode}
      items={isFlat ? columns : fullColumns}
      placeholder="Drag dimensions or measures here"
      onUpdate={updateColumn}
    />
  </div>
</div>

<style lang="postcss">
  .header {
    @apply flex flex-col border-b select-none;
    @apply bg-surface-background justify-center py-2 gap-y-2;
    @apply flex flex-col flex-none relative overflow-hidden;
    @apply z-0;
    transition-property: height;
    will-change: height;
    @apply select-none;
  }

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
