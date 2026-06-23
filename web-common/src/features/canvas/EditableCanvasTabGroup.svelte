<script lang="ts">
  import type { Writable } from "svelte/store";
  import ComponentError from "../components/ComponentError.svelte";
  import AddComponentDropdown from "./AddComponentDropdown.svelte";
  import CanvasTabStrip from "./CanvasTabStrip.svelte";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import type { CanvasComponentType } from "./components/types";
  import EditableCanvasRow from "./EditableCanvasRow.svelte";
  import ItemWrapper from "./ItemWrapper.svelte";
  import type { EditTarget } from "./layout-util";
  import RowDropZone from "./RowDropZone.svelte";
  import RowWrapper from "./RowWrapper.svelte";
  import type { TabGroup } from "./stores/tab-group";

  export let group: TabGroup;
  export let blockIndex: number;
  export let maxWidth: number;
  export let columnWidth: number;
  // Block-level stacking order, matching the value top-level rows receive. The region
  // establishes its own stacking context (isolation: isolate) so the tab's rows and their
  // drop zones are hit-tested as a self-contained layer; without this, neighboring blocks
  // overlap the tab's drop zones and dragging components within a tab silently fails.
  export let zIndex: number;
  // Whether this is the last block on the canvas. When true, an explicit "add outside the
  // tab" affordance is shown after the group so authors can append free rows below it.
  export let isLastBlock: boolean = false;
  export let components: Map<string, BaseCanvasComponent>;
  export let dragComponent: BaseCanvasComponent | null;
  export let selectedComponent: Writable<string | null>;
  export let hasValidMetrics: boolean;

  // Builder primitives, all target-aware.
  export let onDrop: (
    row: number,
    column: number | null,
    target?: EditTarget,
  ) => void;
  export let addItems: (
    position: { row: number; column: number },
    items: CanvasComponentType[],
    target?: EditTarget,
  ) => void;
  export let spreadEvenly: (index: number, target?: EditTarget) => void;
  export let initializeRow: (
    row: number,
    type: CanvasComponentType,
    target?: EditTarget,
  ) => void;
  export let updateRowHeight: (
    newHeight: number,
    index: number,
    target?: EditTarget,
  ) => void;
  export let updateComponentWidths: (
    index: number,
    newWidths: number[],
    target?: EditTarget,
  ) => void;
  export let onComponentMouseDown: (params: {
    id: string;
    event: MouseEvent;
  }) => void;
  export let onDuplicate: (
    rowIndex: number,
    columnIndex: number,
    target: EditTarget,
  ) => void;
  export let onDelete: (params: { component: BaseCanvasComponent }) => void;

  // Tab management.
  export let onAddTab: (blockIndex: number) => void;
  export let onRenameTab: (
    blockIndex: number,
    tabIndex: number,
    label: string,
  ) => void;
  export let onDeleteTab: (blockIndex: number, tabIndex: number) => void;
  export let onMoveTab: (
    blockIndex: number,
    tabIndex: number,
    direction: -1 | 1,
  ) => void;
  // Drop a dragged component onto a tab (cross-container move).
  export let onDropOnTab: (blockIndex: number, tabIndex: number) => void;
  // Insert a new tab group at a given top-level index (for the "add outside" affordance).
  export let onAddTabGroup: (index: number) => void;

  $: tabs = group.tabs;
  $: activeTabIndex = group.activeTabIndex;
  $: activeTab = $tabs[$activeTabIndex];
  $: grid = activeTab?.grid;
  $: target = { blockIndex, tabIndex: $activeTabIndex } satisfies EditTarget;
  $: idPrefix = `${group.name}-${activeTab?.name ?? ""}-`;
</script>

<RowWrapper
  gridTemplate="12fr"
  {zIndex}
  {maxWidth}
  id={`tab-group-row-${group.name}`}
>
  <ItemWrapper fitContent zIndex={0}>
    <div class="tab-group-region">
      <CanvasTabStrip
        {group}
        {maxWidth}
        editable
        {dragComponent}
        onAddTab={() => onAddTab(blockIndex)}
        onRenameTab={(tabIndex, label) =>
          onRenameTab(blockIndex, tabIndex, label)}
        onDeleteTab={(tabIndex) => onDeleteTab(blockIndex, tabIndex)}
        onMoveTab={(tabIndex, direction) =>
          onMoveTab(blockIndex, tabIndex, direction)}
        onDropOnTab={(tabIndex) => onDropOnTab(blockIndex, tabIndex)}
      />

      {#if activeTab && grid}
        {#each $grid as row, rowIndex (rowIndex)}
          <EditableCanvasRow
            {row}
            {maxWidth}
            {rowIndex}
            {idPrefix}
            {components}
            {columnWidth}
            {dragComponent}
            {selectedComponent}
            zIndex={$grid.length - rowIndex + 1}
            onDrop={(r, c) => onDrop(r, c, target)}
            addItems={(pos, items) => addItems(pos, items, target)}
            spreadEvenly={(index) => spreadEvenly(index, target)}
            initializeRow={(r, type) => initializeRow(r, type, target)}
            updateRowHeight={(h, index) => updateRowHeight(h, index, target)}
            updateComponentWidths={(index, widths) =>
              updateComponentWidths(index, widths, target)}
            {onComponentMouseDown}
            onDuplicate={({ columnIndex }) =>
              onDuplicate(rowIndex, columnIndex, target)}
            {onDelete}
          />
        {/each}

        <!-- Add a widget at the end of this tab. Doubles as the empty-tab state. -->
        <RowWrapper
          gridTemplate="12fr"
          zIndex={0}
          {maxWidth}
          id={`tab-add-${group.name}`}
        >
          <ItemWrapper fitContent zIndex={0}>
            {#if hasValidMetrics}
              <AddComponentDropdown
                componentForm
                label="Add widget to tab"
                onItemClick={(type) =>
                  initializeRow($grid.length, type, target)}
              />
            {:else}
              <ComponentError error="No valid metrics view in project" />
            {/if}
          </ItemWrapper>
        </RowWrapper>
      {/if}
    </div>
  </ItemWrapper>

  <RowDropZone
    allowDrop={!!dragComponent}
    dropIndex={blockIndex}
    position="top"
    {onDrop}
    addItem={(type) => {
      initializeRow(blockIndex, type);
    }}
    onAddTabGroup={() => onAddTabGroup(blockIndex)}
  />

  <RowDropZone
    allowDrop={!!dragComponent}
    dropIndex={blockIndex + 1}
    position="bottom"
    {onDrop}
    addItem={(type) => {
      initializeRow(blockIndex + 1, type);
    }}
    onAddTabGroup={() => onAddTabGroup(blockIndex + 1)}
  />
</RowWrapper>

{#if isLastBlock && hasValidMetrics}
  <!-- Add a widget (or another tab group) on the free canvas, after this tab group. -->
  <RowWrapper
    gridTemplate="12fr"
    zIndex={0}
    {maxWidth}
    id={`after-tab-${group.name}`}
  >
    <ItemWrapper type="table" zIndex={0}>
      <AddComponentDropdown
        componentForm
        label="Add widget below tabs"
        onItemClick={(type) => initializeRow(blockIndex + 1, type)}
        onAddTabGroup={() => onAddTabGroup(blockIndex + 1)}
      />
    </ItemWrapper>
  </RowWrapper>
{/if}

<style lang="postcss">
  /* A self-contained stacking context for the tab group: keeps the tab's rows and
     their drop zones hit-testable as one layer, and gives authors a visible boundary
     showing which widgets belong to the tab versus the free canvas. */
  .tab-group-region {
    @apply pointer-events-auto relative flex w-full flex-col items-center;
    @apply rounded-md border border-gray-200 bg-surface-subtle/40 px-3 py-2;
    isolation: isolate;
  }
</style>
