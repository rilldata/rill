<script lang="ts" module>
  import type { LocalizedString } from "@inlang/paraglide-js";
  import type { Writable } from "svelte/store";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import type { AddableItem, EditTarget } from "./layout-util";
  import type { BlockKind } from "./stores/block-edit";

  // What the builder gives every block in the editor: shared drag and selection state, the
  // primitives that edit rows (each takes the edit target of the rows it acts on), and the
  // block-level actions.
  export type BlockBuilderProps = {
    // Top-level index of the block in the canvas rows.
    blockIndex: number;
    maxWidth: number;
    columnWidth: number;
    // Block-level stacking order, matching the value top-level rows receive.
    zIndex: number;
    // Whether this is the last block on the canvas. When true, an explicit "add below"
    // affordance is shown after the block so authors can append free rows below it.
    isLastBlock?: boolean;
    components: Map<string, BaseCanvasComponent>;
    dragComponent: BaseCanvasComponent | null;
    // Top-level index of the block currently being dragged, or null when none is. While a
    // block is dragged, the row drop zones around this block become landing slots (except
    // the dragged block's own two, which would be no-ops) and the dragged block renders as
    // a ghost.
    dragBlockIndex?: number | null;
    selectedComponent: Writable<string | null>;
    hasValidMetrics: boolean;
    onDrop: (row: number, column: number | null, target?: EditTarget) => void;
    addItems: (
      position: { row: number; column: number },
      items: AddableItem[],
      target?: EditTarget,
    ) => void;
    spreadEvenly: (index: number, target?: EditTarget) => void;
    initializeRow: (
      row: number,
      type: AddableItem,
      target?: EditTarget,
    ) => void;
    updateRowHeight: (
      newHeight: number,
      index: number,
      target?: EditTarget,
    ) => void;
    updateComponentWidths: (
      index: number,
      newWidths: number[],
      target?: EditTarget,
    ) => void;
    onComponentMouseDown: (params: { id: string; event: MouseEvent }) => void;
    onDuplicate: (
      rowIndex: number,
      columnIndex: number,
      target: EditTarget,
    ) => void;
    onDelete: (params: { component: BaseCanvasComponent }) => void;
    // Insert a new block of the kind at a top-level index.
    onAddBlock: (kind: BlockKind, index: number) => void;
    // Select a block by name for editing (opens the block inspector panel).
    onSelectBlock: (name: string) => void;
    // Start dragging the block at the top-level index to another position on the canvas.
    onBlockMouseDown: (blockIndex: number, event: MouseEvent) => void;
    // Duplicate the block at the top-level index (the copy is inserted right after it).
    onDuplicateBlock: (blockIndex: number) => void;
  };
</script>

<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { Snippet } from "svelte";
  import ComponentError from "../components/ComponentError.svelte";
  import AddComponentDropdown from "./AddComponentDropdown.svelte";
  import EditableCanvasRow from "./EditableCanvasRow.svelte";
  import ItemWrapper from "./ItemWrapper.svelte";
  import RowDropZone from "./RowDropZone.svelte";
  import RowWrapper from "./RowWrapper.svelte";
  import type { Grid } from "./stores/grid";
  import { dropZone } from "./stores/ui-stores";

  // The editable body of a block, shared by tab groups and containers: a frame around one
  // rows list wired to its edit target, with the block's header on top (the tab strip, or
  // the container's chip), an empty-state drop target, the top-level drop zones around the
  // block and, after the last block, an affordance to add free rows below it.
  let {
    variant,
    grid,
    target,
    idPrefix,
    zoneScope,
    rowKey,
    emptyLabel,
    afterLabel,
    header,
    selected = false,
    onSelect,
    blockIndex,
    maxWidth,
    columnWidth,
    zIndex,
    isLastBlock = false,
    components,
    dragComponent,
    dragBlockIndex = null,
    selectedComponent,
    hasValidMetrics,
    onDrop,
    addItems,
    spreadEvenly,
    initializeRow,
    updateRowHeight,
    updateComponentWidths,
    onComponentMouseDown,
    onDuplicate,
    onDelete,
    onAddBlock,
  }: Omit<
    BlockBuilderProps,
    "onSelectBlock" | "onBlockMouseDown" | "onDuplicateBlock"
  > & {
    // "solid" is the tab group's bordered, padded region, which viewers see too. "dashed" is
    // the container's editor-only frame: an outline that takes no space, so the rows inside
    // keep the widths and heights viewers get.
    variant: "solid" | "dashed";
    // The rows to edit (the active tab's, or the container's).
    grid: Grid | undefined;
    target: EditTarget;
    // Row DOM ids are built from indices, not user-controlled names: names allow spaces
    // and punctuation, which break the `#${id}` querySelector used for height resizing.
    idPrefix: string;
    // Scope of the drop zones of the rows, so they don't collide with other blocks' zones.
    zoneScope: string;
    // Prefix of the rows' keys, so switching to another rows list remounts the rows.
    rowKey: string;
    // Label of the add button shown while the rows list is empty.
    emptyLabel: LocalizedString;
    // Label of the add button shown below the block when it is the last one.
    afterLabel: LocalizedString;
    header: Snippet;
    // Whether the block is selected for editing (styles the dashed frame).
    selected?: boolean;
    // Called on a click on the frame itself, outside the header and the widgets.
    onSelect?: () => void;
  } = $props();

  const blockDragging = $derived(dragBlockIndex !== null);
  const isDragSource = $derived(dragBlockIndex === blockIndex);
  const allowBlockDrop = $derived(blockDragging && !isDragSource);
  // A widget dragged over one of the drop zones of the rows inside highlights the frame.
  const isDropTarget = $derived(
    !!dragComponent &&
      !!$dropZone &&
      ($dropZone === zoneScope || $dropZone.startsWith(`${zoneScope}::`)),
  );
  const rows = $derived(grid);
  const rowCount = $derived($rows?.length ?? 0);
</script>

{#snippet region()}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div
    id={`block-region-${blockIndex}`}
    role="presentation"
    class={[
      "block-region",
      variant,
      {
        selected,
        "drop-target": isDropTarget,
        "opacity-20": isDragSource,
      },
    ]}
    onclick={(e) => {
      if (e.target === e.currentTarget) onSelect?.();
    }}
  >
    {@render header()}

    {#if grid}
      {#each $rows ?? [] as row, rowIndex (`${rowKey}-${rowIndex}`)}
        <EditableCanvasRow
          {row}
          {maxWidth}
          {rowIndex}
          {idPrefix}
          {zoneScope}
          {components}
          {columnWidth}
          {dragComponent}
          {blockDragging}
          {selectedComponent}
          zIndex={rowCount - rowIndex + 1}
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

      {#if rowCount === 0}
        <RowWrapper
          gridTemplate="12fr"
          zIndex={0}
          {maxWidth}
          id={`block-add-${blockIndex}`}
        >
          <ItemWrapper fitContent zIndex={0}>
            {#if hasValidMetrics}
              <!-- The empty body doubles as a drop target so a dragged widget can be
                   dropped straight into an empty block (which has no rows/drop zones yet). -->
              <div
                role="presentation"
                class="w-full"
                onmouseenter={() => {
                  if (dragComponent) dropZone.set(zoneScope);
                }}
                onmouseleave={() => {
                  if (dragComponent) dropZone.clear();
                }}
                onmouseup={() => {
                  if (dragComponent) onDrop(0, null, target);
                }}
              >
                <!-- While dragging, let the drop pass through to the wrapper instead of
                     opening the add menu. -->
                <div class={{ "pointer-events-none": !!dragComponent }}>
                  <AddComponentDropdown
                    componentForm
                    label={emptyLabel}
                    onItemClick={(type) => initializeRow(0, type, target)}
                    onAddComponentRef={(componentName) =>
                      initializeRow(0, { componentName }, target)}
                  />
                </div>
              </div>
            {:else}
              <ComponentError error={m.canvas_no_valid_metrics_view()} />
            {/if}
          </ItemWrapper>
        </RowWrapper>
      {/if}
    {/if}
  </div>
{/snippet}

<RowWrapper
  gridTemplate="12fr"
  {zIndex}
  {maxWidth}
  id={`block-row-${blockIndex}`}
>
  {#if variant === "solid"}
    <ItemWrapper fitContent zIndex={0}>
      {@render region()}
    </ItemWrapper>
  {:else}
    {@render region()}
  {/if}

  <RowDropZone
    allowDrop={!!dragComponent || allowBlockDrop}
    dropIndex={blockIndex}
    position="top"
    {onDrop}
    addItem={(type) => initializeRow(blockIndex, type)}
    onAddBlock={(kind) => onAddBlock(kind, blockIndex)}
  />

  <RowDropZone
    allowDrop={!!dragComponent || allowBlockDrop}
    dropIndex={blockIndex + 1}
    position="bottom"
    {onDrop}
    addItem={(type) => initializeRow(blockIndex + 1, type)}
    onAddBlock={(kind) => onAddBlock(kind, blockIndex + 1)}
  />
</RowWrapper>

{#if isLastBlock && hasValidMetrics}
  <!-- Add a widget (or another block) on the free canvas, after this block. -->
  <RowWrapper
    gridTemplate="12fr"
    zIndex={0}
    {maxWidth}
    id={`after-block-${blockIndex}`}
  >
    <ItemWrapper type="table" zIndex={0}>
      <AddComponentDropdown
        componentForm
        label={afterLabel}
        onItemClick={(type) => initializeRow(blockIndex + 1, type)}
        onAddComponentRef={(componentName) =>
          initializeRow(blockIndex + 1, { componentName })}
        onAddBlock={(kind) => onAddBlock(kind, blockIndex + 1)}
      />
    </ItemWrapper>
  </RowWrapper>
{/if}

<style lang="postcss">
  /* A self-contained stacking context for the block: keeps its rows and their drop zones
     hit-testable as one layer. Without it, neighboring blocks overlap the block's drop
     zones and dragging components within the block silently fails. */
  .block-region {
    @apply pointer-events-auto relative flex w-full min-w-0 flex-col items-center;
    isolation: isolate;
  }

  /* A tab group: a visible boundary showing which widgets belong to the tab versus the
     free canvas. View mode draws the same region. */
  .block-region.solid {
    @apply gap-y-2.5 rounded-md border border-gray-200 bg-surface-subtle/40 px-3 py-2;
  }

  /* A container: viewers see its rows as free rows, so the frame is an outline, which
     takes no space. The header band holds the chip, and the bottom margin keeps the drop
     slot for a new last row apart from the slot below the container. */
  .block-region.dashed {
    @apply rounded-md pb-2;
    outline: 1px dashed var(--color-gray-300);
    outline-offset: -1px;
  }

  .block-region.dashed:hover {
    outline-color: var(--color-gray-400);
  }

  .block-region.dashed.selected,
  .block-region.dashed.drop-target {
    outline: 1.5px dashed var(--color-primary-400);
  }
</style>
