<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { portal } from "@rilldata/web-common/lib/actions/portal";
  import {
    type V1CanvasRow,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { onDestroy } from "svelte";
  import { get, writable } from "svelte/store";
  import { parseDocument } from "yaml";
  import ComponentError from "../components/ComponentError.svelte";
  import type { FileArtifact } from "../entity-management/file-artifact";
  import ReconcilingSpinner from "../entity-management/ReconcilingSpinner.svelte";
  import AddComponentDropdown from "./AddComponentDropdown.svelte";
  import CanvasComponent from "./CanvasComponent.svelte";
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import type { CanvasComponentType } from "./components/types";
  import EditableCanvasRow from "./EditableCanvasRow.svelte";
  import EditableCanvasTabGroup from "./EditableCanvasTabGroup.svelte";
  import ItemWrapper from "./ItemWrapper.svelte";
  import type { Transaction, YAMLRow } from "./layout-util";
  import {
    COLUMN_COUNT,
    DEFAULT_DASHBOARD_WIDTH,
    generateNewAssets,
    mapGuard,
    mousePosition,
    rowsGuard,
    tabNamePrefix,
    tabRowsPath,
    tabTargetFromPath,
    type EditTarget,
  } from "./layout-util";
  import RowWrapper from "./RowWrapper.svelte";
  import { useDefaultMetrics } from "./selector";
  import { getCanvasStore } from "./state-managers/state-managers";
  import { rowColFromPath } from "./stores/canvas-entity";
  import {
    addTab,
    addTabGroup,
    addTabGroupAt,
    convertRowToTabGroup,
    deleteTab,
    moveItemAcrossContainers,
    moveTab,
    renameTab,
    tabHasContent,
  } from "./stores/tab-edit";
  import { activeDivider, dropZone } from "./stores/ui-stores";

  const activelyEditing = writable(false);

  export let fileArtifact: FileArtifact;
  export let canvasName: string;
  export let openSidebar: () => void;

  const runtimeClient = useRuntimeClient();

  const MIN_DRAG_DISTANCE = 8;

  let initialMousePosition: { x: number; y: number } | null = null;
  let clientWidth: number;
  let offset = { x: 0, y: 0 };
  let dragComponent: BaseCanvasComponent | null = null;
  let pendingDragComponent: BaseCanvasComponent | null = null;
  let timeout: ReturnType<typeof setTimeout> | null = null;
  let dragTimeout: ReturnType<typeof setTimeout> | null = null;
  let dragItemPosition = { top: 0, left: 0 };
  let dragItemDimensions = { width: 0, height: 0 };
  let openSidebarAfterSelection = false;
  let pendingComponentDelete: string | undefined = undefined;
  let pendingTabDelete: { blockIndex: number; tabIndex: number } | undefined =
    undefined;

  $: ({
    canvasEntity: {
      setActiveTabInURL,
      setSelectedComponent,
      selectedComponent,
      componentsStore,
      processRows,
      specStore,
      unsubscribe,
      _rows,
      layout,
      firstLoad,
    },
  } = getCanvasStore(canvasName, instanceId));

  $: layoutRows = $_rows;
  $: blocks = $layout;

  $: ({ instanceId } = runtimeClient);

  $: components = $componentsStore;

  $: canvasData = $specStore.data;
  $: metricsViews = Object.entries(canvasData?.metricsViews ?? {});

  $: metricsViewQuery = useDefaultMetrics(
    runtimeClient,
    metricsViews?.[0]?.[0],
  );

  $: ({ editorContent, updateEditorContent } = fileArtifact);
  $: contents = parseDocument($editorContent ?? "");

  $: resolvedComponents = canvasData?.components;

  $: spec = canvasData?.canvas ?? {
    rows: [],
    filtersEnabled: false,
    maxWidth: DEFAULT_DASHBOARD_WIDTH,
  };

  $: if (canvasData?.canvas) {
    if (!get(activelyEditing)) {
      spec = structuredClone(canvasData?.canvas ?? spec);
    }
  }

  $: activelyEditing.set(!!$activeDivider || activelyDragging);

  $: ({ rows = [], filtersEnabled, maxWidth: canvasMaxWidth } = spec);

  $: maxWidth = canvasMaxWidth || DEFAULT_DASHBOARD_WIDTH;

  $: specCanvasRows = structuredClone(rows) as V1CanvasRow[];

  $: columnWidth = clientWidth / 12;

  $: mouseDelta = initialMousePosition
    ? calculateMouseDelta(initialMousePosition, $mousePosition)
    : 0;

  $: dropZone.setMouseDelta(mouseDelta);

  $: if (pendingDragComponent && mouseDelta >= MIN_DRAG_DISTANCE) {
    handleDragStart(pendingDragComponent);
    pendingDragComponent = null;
  }

  $: defaultMetrics = $metricsViewQuery?.data;

  $: activelyDragging = !!dragComponent;

  // Resolve the YAML rows path, name prefix, and current spec rows for an edit target.
  // undefined target => the top-level rows; a tab target => one tab's rows.
  function resolveTarget(target: EditTarget | undefined) {
    if (!target) {
      return { rowsPath: ["rows"], namePrefix: "", specRows: specCanvasRows };
    }
    return {
      rowsPath: tabRowsPath(target.blockIndex, target.tabIndex),
      namePrefix: tabNamePrefix(target.blockIndex, target.tabIndex),
      specRows:
        specCanvasRows[target.blockIndex]?.tabGroup?.tabs?.[target.tabIndex]
          ?.rows ?? [],
    };
  }

  function updateComponentWidths(
    rowIndex: number,
    newWidths: number[],
    target?: EditTarget,
  ) {
    const { rowsPath } = resolveTarget(target);
    newWidths.forEach((width, i) => {
      try {
        contents.setIn([...rowsPath, rowIndex, "items", i, "width"], width);
      } catch (e) {
        console.error(e);
      }
    });

    updateContents();
  }

  function getId(
    row: number | undefined,
    column: number | undefined,
    namePrefix = "",
  ) {
    return `${canvasName}--component-${namePrefix}${row ?? 0}-${column ?? 0}`;
  }

  function calculateMouseDelta(
    pos1: { x: number; y: number },
    pos2: { x: number; y: number },
  ) {
    return Math.sqrt((pos1.x - pos2.x) ** 2 + (pos1.y - pos2.y) ** 2);
  }

  function handleDragStart(component: BaseCanvasComponent) {
    dragComponent = component;

    const id = component.id;
    const element = document.querySelector("#" + id);
    if (!element) return;

    const width = element.clientWidth;
    const height = element.clientHeight;

    const top = element.getBoundingClientRect().top;
    const left = element.getBoundingClientRect().left;
    dragItemDimensions = { width, height };

    offset = {
      x: left - (initialMousePosition?.x ?? $mousePosition.x),
      y: top - (initialMousePosition?.y ?? $mousePosition.y),
    };
  }

  $: if (dragComponent) {
    dragItemPosition = {
      top: $mousePosition.y + offset.y,
      left: $mousePosition.x + offset.x,
    };
  }

  function onDragEnd() {
    dragComponent = null;
    // Safety net: remove portal ghost if Svelte 5's {#if} cleanup didn't
    document
      .querySelectorAll("#rill-portal .drag-container")
      .forEach((el) => el.remove());
  }

  function reset() {
    if (dragTimeout) {
      clearTimeout(dragTimeout);
    }

    pendingDragComponent = null;
    initialMousePosition = null;

    if (dragComponent) {
      onDragEnd();
    }

    if (openSidebarAfterSelection) {
      openSidebar();
      openSidebarAfterSelection = false;
    }

    dropZone.clear();
  }

  function updateRowHeight(
    newHeight: number,
    rowIndex: number,
    target?: EditTarget,
  ) {
    const { rowsPath } = resolveTarget(target);
    try {
      contents.setIn([...rowsPath, rowIndex, "height"], newHeight + "px");
    } catch (e) {
      console.error(e);
    }

    updateContents();
  }

  function spreadEvenly(index: number, target?: EditTarget) {
    const { rowsPath, specRows } = resolveTarget(target);
    const specRowsClone = structuredClone(specRows);
    const yamlRowsClone = structuredClone(
      mapGuard(rowsGuard(contents.getIn(rowsPath))),
    );
    const specRow = specRowsClone[index];
    const yamlRow = yamlRowsClone[index];
    if (!specRow?.items || !yamlRow?.items) return;

    const specItems = specRow.items;
    const yamlItems = yamlRow.items;
    const baseSize = COLUMN_COUNT / specItems.length;

    yamlItems.forEach((_, i) => {
      if (!specItems[i] || !yamlItems[i]) return;
      specItems[i].width = baseSize;
      yamlItems[i].width = baseSize;
    });

    updateAssets(specRowsClone, yamlRowsClone, undefined, target);
  }

  // Apply the result of an edit. specRows/yamlRows are the rows of the targeted
  // container (top-level or a single tab). For a tab target, the container's rows
  // are spliced back into the full spec before reprocessing so the whole tree stays valid.
  function updateAssets(
    specRows: V1CanvasRow[],
    yamlRows: YAMLRow[],
    newResolvedComponents?: Record<string, V1Resource>,
    target?: EditTarget,
  ) {
    const { rowsPath } = resolveTarget(target);

    if (!target) {
      specCanvasRows = specRows;
    } else {
      const full = structuredClone(specCanvasRows);
      const tabGroup = full[target.blockIndex]?.tabGroup;
      const tab = tabGroup?.tabs?.[target.tabIndex];
      if (tab) tab.rows = specRows;
      specCanvasRows = full;
    }

    if (newResolvedComponents) {
      processRows({
        // Merge with the existing resolved components so untouched containers still resolve.
        components: { ...resolvedComponents, ...newResolvedComponents },
        canvas: {
          rows: specCanvasRows,
        },
      });
    }

    try {
      contents.setIn(rowsPath, yamlRows);
    } catch (e) {
      console.error(e);
    }

    updateContents();
  }

  function performTransaction(
    transaction: Transaction,
    selectedId?: string,
    target?: EditTarget,
  ) {
    if (!defaultMetrics) return;

    const { namePrefix, specRows, rowsPath } = resolveTarget(target);
    const yamlRows = mapGuard(rowsGuard(contents.getIn(rowsPath)));

    try {
      const { newSpecRows, newYamlRows, newResolvedComponents, mover } =
        generateNewAssets({
          yamlRows,
          specRows,
          canvasName,
          namePrefix,
          defaultMetrics,
          resolvedComponents,
          transaction,
        });

      if (selectedId) {
        const idsAtPositions = specRows.map((row) => {
          return {
            items: row.items?.map((item) => {
              return item.component ?? "";
            }),
          };
        });

        const ids = mover(
          idsAtPositions,
          () => "new-item",
          (r) => r,
        );
        const rowIndex = ids.findIndex((r) =>
          r.items?.some((i) => i === selectedId),
        );
        const colIndex = ids[rowIndex]?.items?.indexOf(selectedId);

        if (colIndex !== undefined && colIndex !== -1 && rowIndex !== -1) {
          const newIdOfSelected = getId(rowIndex, colIndex, namePrefix);

          if (selectedId && newIdOfSelected)
            setSelectedComponent(newIdOfSelected);
        }
      }

      updateAssets(newSpecRows, newYamlRows, newResolvedComponents, target);
    } catch {
      // no-op
    }
  }

  function addItems(
    position: { row: number; column: number },
    items: CanvasComponentType[],
    target?: EditTarget,
  ) {
    if (!defaultMetrics) return;

    performTransaction(
      {
        operations: items.map((type, i) => ({
          type: "add",
          componentType: type,
          destination: {
            row: position.row,
            col: position.column + i,
          },
        })),
      },
      undefined,
      target,
    );

    const { namePrefix } = resolveTarget(target);
    const id = getId(position.row, position.column, namePrefix);

    setSelectedComponent(id);
    openSidebar();
  }

  function updateContents() {
    const newContent = contents.toString();
    if (newContent === $editorContent) {
      contents = parseDocument(newContent);
    } else {
      updateEditorContent(contents.toString(), false, true);
    }
  }

  // Tab group authoring. These mutate the YAML document directly (see stores/tab-edit)
  // and persist via updateContents; the spec round-trip rebuilds the layout.
  function addTabGroupAction() {
    addTabGroup(contents);
    updateContents();
    setTimeout(() => scrollToBottom(), 200);
  }

  function addTabGroupAtAction(index: number) {
    addTabGroupAt(contents, index);
    updateContents();
  }

  function addTabAction(blockIndex: number) {
    const newTabIndex = addTab(contents, blockIndex);
    if (newTabIndex >= 0) {
      const block = blocks[blockIndex];
      if (block?.kind === "tab-group") {
        block.group.activateWhenReady(newTabIndex);
      }
    }
    updateContents();
  }

  function renameTabAction(
    blockIndex: number,
    tabIndex: number,
    label: string,
  ) {
    renameTab(contents, blockIndex, tabIndex, label);
    updateContents();
  }

  function deleteTabAction(blockIndex: number, tabIndex: number) {
    // Confirm before deleting a tab that has content (deleting the last tab unwraps
    // the group, so layout is never lost, but the user should still confirm).
    if (tabHasContent(contents, blockIndex, tabIndex)) {
      pendingTabDelete = { blockIndex, tabIndex };
      return;
    }
    deleteTab(contents, blockIndex, tabIndex);
    updateContents();
  }

  function confirmDeleteTab() {
    if (!pendingTabDelete) return;
    deleteTab(contents, pendingTabDelete.blockIndex, pendingTabDelete.tabIndex);
    pendingTabDelete = undefined;
    updateContents();
  }

  function moveTabAction(
    blockIndex: number,
    tabIndex: number,
    direction: -1 | 1,
  ) {
    moveTab(contents, blockIndex, tabIndex, direction);
    updateContents();
  }

  function convertRowToTabGroupAction(rowIndex: number) {
    if (convertRowToTabGroup(contents, rowIndex)) updateContents();
  }

  function initializeRow(
    row: number,
    type: CanvasComponentType,
    target?: EditTarget,
  ) {
    if (!defaultMetrics) return;

    const { namePrefix } = resolveTarget(target);
    const id = getId(row, 0, namePrefix);

    performTransaction(
      {
        operations: [
          {
            type: "add",
            insertRow: true,
            componentType: type,
            destination: {
              row,
              col: 0,
            },
          },
        ],
      },
      undefined,
      target,
    );

    setSelectedComponent(id);
    openSidebar();
  }

  function sameTarget(a: EditTarget | undefined, b: EditTarget | undefined) {
    if (!a && !b) return true;
    if (!a || !b) return false;
    return a.blockIndex === b.blockIndex && a.tabIndex === b.tabIndex;
  }

  function onDrop(row: number, column: number | null, target?: EditTarget) {
    if (!$dropZone) return;
    dropZone.clear();

    if (!dragComponent) return;

    const sourceTarget = tabTargetFromPath(dragComponent.pathInYAML);
    const { row: fromRow, col: fromCol } = rowColFromPath(
      dragComponent.pathInYAML,
    );

    // Cross-container move (between top-level and a tab, or between tabs).
    if (!sameTarget(sourceTarget, target)) {
      moveComponentToTarget(dragComponent, target, { row, col: column });
      return;
    }

    if (row === fromRow && (column === fromCol || column === fromCol + 1)) {
      return;
    }

    performTransaction(
      {
        operations: [
          {
            type: "move",
            insertRow: column === null,
            source: {
              row: fromRow,
              col: fromCol,
            },
            destination: {
              row,
              col: column ?? 0,
            },
          },
        ],
      },
      dragComponent.id,
      target,
    );
  }

  function resetSelection() {
    setSelectedComponent(null);
  }

  function scrollToBottom() {
    const element = document.querySelector("#canvas-scroll-container");
    if (element) {
      element.scrollTop = element.scrollHeight;
    }
  }

  function deleteComponent(component: BaseCanvasComponent) {
    const { row, col } = rowColFromPath(component.pathInYAML);
    const target = tabTargetFromPath(component.pathInYAML);

    if (component.id === $selectedComponent) {
      resetSelection();
    }

    performTransaction(
      {
        operations: [
          {
            type: "delete",
            target: {
              row,
              col,
            },
          },
        ],
      },
      undefined,
      target,
    );
  }

  // Move a component out of its current container into another (top-level <-> tab, or
  // between tabs). Used for drag-across-tabs and dropping onto a tab strip. The YAML edit
  // and the optimistic spec are updated in lockstep, then the spec is reprocessed so the
  // moved component's path is refreshed immediately; skipping the reprocess leaves stale
  // paths that corrupt the document on the next edit.
  function moveComponentToTarget(
    component: BaseCanvasComponent,
    destTarget: EditTarget | undefined,
    destPos?: { row: number; col: number | null },
  ) {
    const sourceTarget = tabTargetFromPath(component.pathInYAML);
    const { row: fromRow, col: fromCol } = rowColFromPath(component.pathInYAML);

    const sourcePath = resolveTarget(sourceTarget).rowsPath;
    const destPath = resolveTarget(destTarget).rowsPath;

    const moved = moveItemAcrossContainers(
      contents,
      { rowsPath: sourcePath, row: fromRow, col: fromCol },
      { rowsPath: destPath, row: destPos?.row, col: destPos?.col ?? null },
    );
    if (!moved) return;

    // Mirror the same move in the optimistic spec (proto shape) and reprocess so grids and
    // component paths update now rather than only on the next reconcile.
    const full = structuredClone(specCanvasRows);
    const containerRows = (t: EditTarget | undefined) =>
      t ? full[t.blockIndex]?.tabGroup?.tabs?.[t.tabIndex]?.rows : full;
    const srcRows = containerRows(sourceTarget);
    const dstRows = containerRows(destTarget);
    const specItem = srcRows?.[fromRow]?.items?.[fromCol];

    if (srcRows && dstRows && specItem) {
      srcRows[fromRow].items?.splice(fromCol, 1);
      if (srcRows[fromRow].items?.length === 0) srcRows.splice(fromRow, 1);

      const destRow = destPos?.row;
      const destItems =
        destRow !== undefined ? dstRows[destRow]?.items : undefined;
      // Mirror moveItemAcrossContainers: join an existing row only if it has room (<4).
      if (destPos?.col != null && destItems && destItems.length < 4) {
        destItems.splice(Math.min(destPos.col, destItems.length), 0, specItem);
      } else {
        dstRows.splice(Math.min(destRow ?? dstRows.length, dstRows.length), 0, {
          items: [specItem],
        });
      }

      specCanvasRows = full;
      processRows({
        components: resolvedComponents,
        canvas: { rows: specCanvasRows },
      });
    }

    updateContents();
  }

  // Drop the currently-dragged component onto a tab (appends a new row in that tab).
  function dropComponentOnTab(blockIndex: number, tabIndex: number) {
    if (!dragComponent) return;
    moveComponentToTarget(dragComponent, { blockIndex, tabIndex });
  }

  function handleComponentMouseDown({
    event,
    id,
  }: {
    event: MouseEvent;
    id: string;
  }) {
    if (event.button !== 0) return;
    if (event.shiftKey) return;
    const component = componentsStore.getNonReactive(id);
    if (!component) return;

    initialMousePosition = $mousePosition;

    setSelectedComponent(id);

    if (dragTimeout) clearTimeout(dragTimeout);

    openSidebarAfterSelection = true;

    // Ensure cleanup on mouseup even if svelte:window handler fails
    window.addEventListener("mouseup", reset, { once: true });

    dragTimeout = setTimeout(() => {
      openSidebarAfterSelection = false;
      pendingDragComponent = component;
    }, 150);
  }

  function duplicateComponent(
    rowIndex: number,
    columnIndex: number,
    target?: EditTarget,
  ) {
    if (!defaultMetrics) return;

    performTransaction(
      {
        operations: [
          {
            type: "copy",
            insertRow: true,
            source: {
              row: rowIndex,
              col: columnIndex,
            },
            destination: {
              row: rowIndex + 1,
              col: 0,
            },
          },
        ],
      },
      undefined,
      target,
    );
  }

  onDestroy(() => {
    if (dragTimeout) {
      clearTimeout(dragTimeout);
    }

    if (timeout) {
      clearTimeout(timeout);
    }

    unsubscribe();
  });
</script>

<svelte:window
  onmouseup={reset}
  onkeydown={(e) => {
    const selected = $selectedComponent;
    if (!selected || e.key !== "Backspace") return;

    if (
      !(e.target instanceof HTMLElement) ||
      (e.target.tagName !== "INPUT" &&
        e.target.tagName !== "TEXTAREA" &&
        !e.target.isContentEditable)
    ) {
      pendingComponentDelete = selected;
    }
  }}
/>

<CanvasDashboardWrapper
  {canvasName}
  {maxWidth}
  {filtersEnabled}
  showGrabCursor={activelyDragging}
  onClick={resetSelection}
  builder
  bind:clientWidth
>
  {#each blocks as block, blockIndex (block.kind === "tab-group" ? `g-${block.group.name}` : `r-${block.rowIndex}`)}
    {#if block.kind === "tab-group"}
      <EditableCanvasTabGroup
        group={block.group}
        blockIndex={block.rowIndex}
        zIndex={blocks.length - blockIndex}
        isLastBlock={blockIndex === blocks.length - 1}
        {maxWidth}
        {columnWidth}
        {components}
        {dragComponent}
        {selectedComponent}
        hasValidMetrics={!!defaultMetrics}
        {onDrop}
        {addItems}
        {spreadEvenly}
        {initializeRow}
        {updateRowHeight}
        {updateComponentWidths}
        onComponentMouseDown={handleComponentMouseDown}
        onDuplicate={duplicateComponent}
        onDelete={({ component }) => deleteComponent(component)}
        onAddTab={addTabAction}
        onRenameTab={renameTabAction}
        onDeleteTab={deleteTabAction}
        onMoveTab={moveTabAction}
        onSelect={(tabName) => setActiveTabInURL(block.group.name, tabName)}
        onDropOnTab={dropComponentOnTab}
        onAddTabGroup={addTabGroupAtAction}
      />
    {:else if layoutRows[block.freeRowIndex]}
      {@const rowIndex = block.rowIndex}
      <EditableCanvasRow
        row={layoutRows[block.freeRowIndex]}
        {maxWidth}
        {rowIndex}
        {components}
        {columnWidth}
        {dragComponent}
        {selectedComponent}
        zIndex={blocks.length - blockIndex}
        {onDrop}
        {addItems}
        {spreadEvenly}
        {initializeRow}
        {updateRowHeight}
        {updateComponentWidths}
        onDelete={({ component }) => {
          deleteComponent(component);
        }}
        onDuplicate={({ columnIndex }) =>
          duplicateComponent(rowIndex, columnIndex)}
        onConvertToTabGroup={() => convertRowToTabGroupAction(rowIndex)}
        onAddTabGroup={addTabGroupAtAction}
        onComponentMouseDown={handleComponentMouseDown}
      />
    {/if}
  {:else}
    {#if $firstLoad}
      <div class="h-72 flex items-center justify-center">
        <ReconcilingSpinner />
      </div>
    {:else}
      <RowWrapper
        gridTemplate="12fr"
        zIndex={0}
        {maxWidth}
        id="add-component-row"
      >
        <ItemWrapper type="table" zIndex={0}>
          {#if defaultMetrics}
            <AddComponentDropdown
              componentForm
              onMouseEnter={() => {
                if (timeout) clearTimeout(timeout);
              }}
              onItemClick={(type) => {
                initializeRow(specCanvasRows.length, type);
              }}
              onAddTabGroup={addTabGroupAction}
            />
          {:else if canvasData}
            <ComponentError error="No valid metrics view in project" />
          {/if}
        </ItemWrapper>
      </RowWrapper>
    {/if}
  {/each}
</CanvasDashboardWrapper>

{#if specCanvasRows.length}
  <AddComponentDropdown
    floatingForm
    disabled={!defaultMetrics}
    onItemClick={(type) => {
      initializeRow(specCanvasRows.length, type);
      setTimeout(() => scrollToBottom(), 500);
    }}
    onAddTabGroup={addTabGroupAction}
  />
{/if}

{#if dragComponent}
  <div
    use:portal
    class="absolute pointer-events-none drag-container"
    style:z-index="1000"
    style:top="{dragItemPosition.top}px"
    style:left="{dragItemPosition.left}px"
    style:width="{dragItemDimensions.width}px"
    style:height="{dragItemDimensions.height}px"
  >
    <CanvasComponent
      component={dragComponent}
      allowPointerEvents={false}
      ghost
      selected
    />
  </div>
{/if}

{#if pendingComponentDelete !== undefined}
  <AlertDialog.Root
    open
    onOpenChange={(open) => {
      if (!open) pendingComponentDelete = undefined;
    }}
  >
    <AlertDialog.Content>
      <AlertDialog.Title>Delete widget?</AlertDialog.Title>

      <AlertDialog.Description>
        This widget and its configuration will be permanently removed.
      </AlertDialog.Description>

      <AlertDialog.Footer>
        <AlertDialog.Cancel>
          {#snippet child({ props })}
            <Button
              {...props}
              large
              type="secondary"
              onClick={() => {
                pendingComponentDelete = undefined;
              }}
            >
              Cancel
            </Button>
          {/snippet}
        </AlertDialog.Cancel>

        <AlertDialog.Action>
          {#snippet child({ props })}
            <Button
              {...props}
              large
              type="destructive"
              onClick={() => {
                if (!pendingComponentDelete) return;
                const component = componentsStore.getNonReactive(
                  pendingComponentDelete,
                );
                if (!component) return;
                deleteComponent(component);
                pendingComponentDelete = undefined;
              }}
            >
              Delete
            </Button>
          {/snippet}
        </AlertDialog.Action>
      </AlertDialog.Footer>
    </AlertDialog.Content>
  </AlertDialog.Root>
{/if}

{#if pendingTabDelete !== undefined}
  <AlertDialog.Root
    open
    onOpenChange={(open) => {
      if (!open) pendingTabDelete = undefined;
    }}
  >
    <AlertDialog.Content>
      <AlertDialog.Title>Delete tab?</AlertDialog.Title>

      <AlertDialog.Description>
        This tab and all of its widgets will be permanently removed.
      </AlertDialog.Description>

      <AlertDialog.Footer>
        <AlertDialog.Cancel>
          {#snippet child({ props })}
            <Button
              {...props}
              large
              type="secondary"
              onClick={() => {
                pendingTabDelete = undefined;
              }}
            >
              Cancel
            </Button>
          {/snippet}
        </AlertDialog.Cancel>

        <AlertDialog.Action>
          {#snippet child({ props })}
            <Button
              {...props}
              large
              type="destructive"
              onClick={confirmDeleteTab}
            >
              Delete
            </Button>
          {/snippet}
        </AlertDialog.Action>
      </AlertDialog.Footer>
    </AlertDialog.Content>
  </AlertDialog.Root>
{/if}

<style lang="postcss">
  .drag-container {
    container-type: inline-size;
    container-name: component-container;
  }
</style>
