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
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import ReconcilingSpinner from "../entity-management/ReconcilingSpinner.svelte";
  import AddComponentDropdown from "./AddComponentDropdown.svelte";
  import BlockDeleteDialog from "./BlockDeleteDialog.svelte";
  import CanvasComponent from "./CanvasComponent.svelte";
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import EditableCanvasContainer from "./EditableCanvasContainer.svelte";
  import EditableCanvasRow from "./EditableCanvasRow.svelte";
  import EditableCanvasTabGroup from "./EditableCanvasTabGroup.svelte";
  import ItemWrapper from "./ItemWrapper.svelte";
  import type { Transaction, YAMLRow } from "./layout-util";
  import {
    COLUMN_COUNT,
    DEFAULT_DASHBOARD_WIDTH,
    canvasItemInstanceId,
    componentNameFromInstanceId,
    containerRowsPath,
    generateNewAssets,
    isBlockSpecRow,
    mapGuard,
    mousePosition,
    positionKeyPrefix,
    rowsGuard,
    rowsPathOf,
    sameTarget,
    specRowsAt,
    tabRowsPath,
    targetOfComponent,
    type AddableItem,
    type EditTarget,
  } from "./layout-util";
  import {
    ResourceKind,
    useFilteredResources,
  } from "../entity-management/resource-selectors";
  import RowWrapper from "./RowWrapper.svelte";
  import { useDefaultMetrics } from "./selector";
  import { getCanvasStore } from "./state-managers/state-managers";
  import {
    addBlockAt,
    blockHasContent,
    deleteBlock,
    duplicateBlock,
    moveBlock,
    moveItemAcrossContainers,
    nameBlocks,
    unwrapRows,
    wrapRow,
    type BlockKind,
  } from "./stores/block-edit";
  import { rowColFromPath } from "./stores/canvas-entity";
  import { layoutBlockKey } from "./stores/layout-blocks";
  import {
    addTab,
    deleteTab,
    duplicateTab,
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
  // A block ghost follows the pointer like a component ghost, but only shows the block's
  // header and a slice of the body: dragging a full-height copy of a many-row block would
  // cover the drop zones the author is aiming for.
  const MAX_BLOCK_GHOST_HEIGHT = 200;

  let initialMousePosition: { x: number; y: number } | null = null;
  let clientWidth: number = 0;
  let offset = { x: 0, y: 0 };
  let dragComponent: BaseCanvasComponent | null = null;
  let pendingDragComponent: BaseCanvasComponent | null = null;
  // Whole-block drags: the top-level index of the tab group or container being moved.
  let dragBlockIndex: number | null = null;
  let pendingDragBlockIndex: number | null = null;
  let timeout: ReturnType<typeof setTimeout> | null = null;
  let dragTimeout: ReturnType<typeof setTimeout> | null = null;
  let dragItemPosition = { top: 0, left: 0 };
  let dragItemDimensions = { width: 0, height: 0 };
  let openSidebarAfterSelection = false;
  let pendingComponentDelete: string | undefined = undefined;
  let pendingTabDelete: { blockIndex: number; tabIndex: number } | undefined =
    undefined;
  // Top-level index of a container waiting for its delete to be confirmed.
  let pendingBlockDelete: number | undefined = undefined;

  $: ({ canvasEntity } = getCanvasStore(canvasName, instanceId));
  $: ({
    setActiveTabInURL,
    setSelectedComponent,
    setSelectedBlock,
    selectedComponent,
    selectedBlock,
    componentsStore,
    processRows,
    specStore,
    _rows,
    layout,
    firstLoad,
  } = canvasEntity);

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

  // Project component resources by name, used when adding component references.
  $: componentResourcesQuery = useFilteredResources(
    runtimeClient,
    ResourceKind.Component,
  );
  $: componentResourceMap = Object.fromEntries(
    ($componentResourcesQuery?.data ?? [])
      .filter((res) => !!res.meta?.name?.name)
      .map((res) => [res.meta!.name!.name as string, res]),
  );

  $: ({ editorContent, updateEditorContent } = fileArtifact);
  // Blocks are named before any edit, so an edit that shifts rows keeps their keys (see nameBlocks).
  $: contents = nameBlocks(parseDocument($editorContent ?? ""));

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

  $: specCanvasRows = structuredClone(rows);

  $: columnWidth = clientWidth / 12;

  $: mouseDelta = initialMousePosition
    ? calculateMouseDelta(initialMousePosition, $mousePosition)
    : 0;

  $: dropZone.setMouseDelta(mouseDelta);

  $: if (pendingDragComponent && mouseDelta >= MIN_DRAG_DISTANCE) {
    handleDragStart(pendingDragComponent);
    pendingDragComponent = null;
  }

  $: if (pendingDragBlockIndex !== null && mouseDelta >= MIN_DRAG_DISTANCE) {
    handleBlockDragStart(pendingDragBlockIndex);
    pendingDragBlockIndex = null;
  }

  $: defaultMetrics = $metricsViewQuery?.data;

  $: activelyDragging = !!dragComponent || dragBlockIndex !== null;

  // The block being dragged, for its ghost (undefined when no block is dragged), and the
  // tabs of a dragged tab group, whose labels the ghost shows.
  $: dragBlock = dragBlockIndex !== null ? blocks[dragBlockIndex] : undefined;
  $: dragGroupTabs =
    dragBlock?.kind === "tab-group" ? dragBlock.group.tabs : null;

  // Resolve the YAML rows path, name prefix, and current spec rows for an edit target.
  // undefined target => the top-level rows; otherwise the rows of one tab or container.
  function resolveTarget(target: EditTarget | undefined) {
    const rowsPath = rowsPathOf(target);
    return {
      rowsPath,
      namePrefix: positionKeyPrefix(rowsPath),
      specRows: specRowsAt(specCanvasRows, rowsPath) ?? [],
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
    // getElementById, not querySelector: instance ids of referenced components
    // contain "::", which is not valid in a CSS selector.
    const element = document.getElementById(id);
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

  function handleBlockDragStart(blockIndex: number) {
    const element = document.getElementById(`block-region-${blockIndex}`);
    if (!element) return;

    dragBlockIndex = blockIndex;
    setSelectedComponent(null);

    const { top, left, width, height } = element.getBoundingClientRect();
    dragItemDimensions = {
      width,
      height: Math.min(height, MAX_BLOCK_GHOST_HEIGHT),
    };

    offset = {
      x: left - (initialMousePosition?.x ?? $mousePosition.x),
      y: top - (initialMousePosition?.y ?? $mousePosition.y),
    };
  }

  $: if (dragComponent || dragBlockIndex !== null) {
    dragItemPosition = {
      top: $mousePosition.y + offset.y,
      left: $mousePosition.x + offset.x,
    };
  }

  function onDragEnd() {
    dragComponent = null;
    dragBlockIndex = null;
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
    pendingDragBlockIndex = null;
    initialMousePosition = null;

    if (dragComponent || dragBlockIndex !== null) {
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

  // Apply the result of an edit. specRows/yamlRows are the rows of the targeted rows
  // list (the top level, a tab or a container). For a nested target, the list's rows are
  // spliced back into the full spec before reprocessing so the whole tree stays valid.
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
      const rows = specRowsAt(full, rowsPath, true);
      rows?.splice(0, rows.length, ...specRows);
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
          componentResources: componentResourceMap,
        });

      if (selectedId) {
        // Selection ids are per-position instance ids (which for items that
        // reference an external component differ from the item's component name).
        // Block rows get no items, so the empty-row cleanup keeps them (see generateNewAssets).
        const idsAtPositions = specRows.map((row, rowIdx) => {
          return {
            items: isBlockSpecRow(row)
              ? undefined
              : row.items?.map((item, colIdx) =>
                  canvasItemInstanceId(item, rowIdx, colIdx, namePrefix),
                ),
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
          const newIdOfSelected = selectedId.includes("::")
            ? `${componentNameFromInstanceId(selectedId)}::${namePrefix}${rowIndex}-${colIndex}`
            : getId(rowIndex, colIndex, namePrefix);

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
    items: AddableItem[],
    target?: EditTarget,
  ) {
    if (!defaultMetrics) return;

    performTransaction(
      {
        operations: items.map((item, i) => ({
          type: "add",
          ...(typeof item === "string"
            ? { componentType: item }
            : {
                componentType: "component_ref" as const,
                componentName: item.componentName,
              }),
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
    const id = getInstanceId(
      items[0],
      position.row,
      position.column,
      namePrefix,
    );

    setSelectedComponent(id);
    openSidebar();
  }

  // The client-side instance id an added item will get: the generated resource name
  // for inline components, or the positional reference id for component references.
  function getInstanceId(
    item: AddableItem | undefined,
    row: number,
    column: number,
    namePrefix = "",
  ) {
    if (item && typeof item !== "string") {
      return `${item.componentName}::${namePrefix}${row}-${column}`;
    }
    return getId(row, column, namePrefix);
  }

  function updateContents() {
    const newContent = contents.toString();
    if (newContent === $editorContent) {
      contents = parseDocument(newContent);
    } else {
      updateEditorContent(contents.toString(), false, true);
    }
  }

  // Block and tab authoring. These mutate the YAML document directly (see stores/block-edit
  // and stores/tab-edit) and persist via updateContents; the spec round-trip rebuilds the layout.
  function addBlockAtAction(kind: BlockKind, index: number) {
    addBlockAt(contents, index, kind);
    updateContents();
  }

  function appendBlockAction(kind: BlockKind) {
    addBlockAtAction(kind, specCanvasRows.length);
    setTimeout(() => scrollToBottom(), 200);
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

  function duplicateTabAction(blockIndex: number, tabIndex: number) {
    const newTabIndex = duplicateTab(contents, blockIndex, tabIndex);
    if (newTabIndex >= 0) {
      const block = blocks[blockIndex];
      if (block?.kind === "tab-group") {
        block.group.activateWhenReady(newTabIndex);
      }
    }
    updateContents();
  }

  function selectBlock(name: string) {
    setSelectedBlock(name);
    openSidebar();
  }

  // Duplicate a block and select the copy.
  function duplicateBlockAt(blockIndex: number) {
    const copy = duplicateBlock(contents, blockIndex);
    if (!copy) return;
    setSelectedBlock(copy);
    updateContents();
    openSidebar();
  }

  function wrapRowAt(rowIndex: number, kind: BlockKind) {
    if (wrapRow(contents, rowIndex, kind)) updateContents();
  }

  // Replace a container with its rows, which keep its condition (see unwrapRows).
  function ungroupContainerAt(blockIndex: number) {
    if (unwrapRows(contents, blockIndex, containerRowsPath(blockIndex)) < 0) {
      return;
    }
    setSelectedBlock(null);
    updateContents();
  }

  // Delete a block, asking first when it has content.
  function requestDeleteBlock(blockIndex: number) {
    if (blockHasContent(contents, blockIndex)) {
      pendingBlockDelete = blockIndex;
      return;
    }
    deleteBlockAt(blockIndex);
  }

  function deleteBlockAt(blockIndex: number) {
    if (!deleteBlock(contents, blockIndex)) return;
    setSelectedBlock(null);
    updateContents();
  }

  // Drop the dragged block into the top-level slot `to` (a row drop zone index).
  function dropBlock(from: number, to: number) {
    const block = blocks[from];
    if (!block || block.kind === "row") return;

    const newIndex = moveBlock(contents, from, to);
    if (newIndex < 0) return;

    // Mirror the move in the optimistic spec and reprocess, so the canvas reorders now
    // rather than on the next reconcile (component names are re-derived then). The spec's
    // block names are the ones in the document, so each block keeps its instance.
    const full = structuredClone(specCanvasRows);
    const [moved] = full.splice(from, 1);
    full.splice(newIndex, 0, moved);
    specCanvasRows = full;
    processRows({
      components: resolvedComponents,
      canvas: { rows: specCanvasRows },
    });

    updateContents();
  }

  function handleBlockMouseDown(blockIndex: number, event: MouseEvent) {
    if (event.button !== 0) return;
    if (event.shiftKey) return;

    initialMousePosition = $mousePosition;

    if (dragTimeout) clearTimeout(dragTimeout);

    // Ensure cleanup on mouseup even if svelte:window handler fails
    window.addEventListener("mouseup", reset, { once: true });

    dragTimeout = setTimeout(() => {
      pendingDragBlockIndex = blockIndex;
    }, 150);
  }

  function initializeRow(row: number, item: AddableItem, target?: EditTarget) {
    if (!defaultMetrics) return;

    const { namePrefix } = resolveTarget(target);
    const id = getInstanceId(item, row, 0, namePrefix);

    performTransaction(
      {
        operations: [
          {
            type: "add",
            insertRow: true,
            ...(typeof item === "string"
              ? { componentType: item }
              : {
                  componentType: "component_ref" as const,
                  componentName: item.componentName,
                }),
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

  function onDrop(row: number, column: number | null, target?: EditTarget) {
    if (!$dropZone) return;
    dropZone.clear();

    // A dragged block only lands in top-level row slots (never inside a block or a row).
    if (dragBlockIndex !== null) {
      if (!target && column === null) dropBlock(dragBlockIndex, row);
      return;
    }

    if (!dragComponent) return;

    const sourceTarget = targetOfComponent(dragComponent.pathInYAML);
    const { row: fromRow, col: fromCol } = rowColFromPath(
      dragComponent.pathInYAML,
    );

    // Cross-container move (between the top level, tabs and containers).
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
    setSelectedBlock(null);
  }

  function scrollToBottom() {
    const element = document.querySelector("#canvas-scroll-container");
    if (element) {
      element.scrollTop = element.scrollHeight;
    }
  }

  function deleteComponent(component: BaseCanvasComponent) {
    const { row, col } = rowColFromPath(component.pathInYAML);
    const target = targetOfComponent(component.pathInYAML);

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

  // Move a component out of its current rows list into another (between the top level, tabs
  // and containers). Used for drags across blocks and dropping onto a tab strip. The YAML edit
  // and the optimistic spec are updated in lockstep, then the spec is reprocessed so the
  // moved component's path is refreshed immediately; skipping the reprocess leaves stale
  // paths that corrupt the document on the next edit.
  function moveComponentToTarget(
    component: BaseCanvasComponent,
    destTarget: EditTarget | undefined,
    destPos?: { row: number; col: number | null },
  ) {
    const sourceTarget = targetOfComponent(component.pathInYAML);
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
    const srcRows = specRowsAt(full, rowsPathOf(sourceTarget));
    // An empty block's rows are missing from the spec (proto JSON omits empty lists).
    const dstRows = specRowsAt(full, rowsPathOf(destTarget), true);
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
    const target = { rowsPath: tabRowsPath(blockIndex, tabIndex) };
    // Releasing over the label of the tab the component already lives in has no positional
    // intent, so skip it: otherwise the move would pull the item from its row and re-append
    // it as a new last row, reshaping the tab for a harmless release.
    if (sameTarget(targetOfComponent(dragComponent.pathInYAML), target)) return;
    moveComponentToTarget(dragComponent, target);
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
  {#each blocks as block, blockIndex (layoutBlockKey(block))}
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
        {dragBlockIndex}
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
        onAddBlock={addBlockAtAction}
        onSelectBlock={selectBlock}
        onBlockMouseDown={handleBlockMouseDown}
        onDuplicateBlock={duplicateBlockAt}
        onAddTab={addTabAction}
        onRenameTab={renameTabAction}
        onDeleteTab={deleteTabAction}
        onMoveTab={moveTabAction}
        onDuplicateTab={duplicateTabAction}
        onSelect={(tabName) => setActiveTabInURL(block.group.name, tabName)}
        onDropOnTab={dropComponentOnTab}
      />
    {:else if block.kind === "container"}
      <EditableCanvasContainer
        container={block.container}
        canvas={canvasEntity}
        selected={$selectedBlock === block.container.name}
        blockIndex={block.rowIndex}
        zIndex={blocks.length - blockIndex}
        isLastBlock={blockIndex === blocks.length - 1}
        {maxWidth}
        {columnWidth}
        {components}
        {dragComponent}
        {dragBlockIndex}
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
        onAddBlock={addBlockAtAction}
        onSelectBlock={selectBlock}
        onBlockMouseDown={handleBlockMouseDown}
        onDuplicateBlock={duplicateBlockAt}
        onUngroup={ungroupContainerAt}
        onDeleteBlock={requestDeleteBlock}
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
        blockDragging={dragBlockIndex !== null}
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
        onWrapRow={(kind) => wrapRowAt(rowIndex, kind)}
        onAddBlock={addBlockAtAction}
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
              onAddComponentRef={(componentName) => {
                initializeRow(specCanvasRows.length, { componentName });
              }}
              onAddBlock={appendBlockAction}
            />
          {:else if canvasData}
            <ComponentError error={m.canvas_no_valid_metrics_view()} />
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
    onAddComponentRef={(componentName) => {
      initializeRow(specCanvasRows.length, { componentName });
      setTimeout(() => scrollToBottom(), 500);
    }}
    onAddBlock={appendBlockAction}
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

{#if dragBlock && dragBlock.kind !== "row"}
  <div
    use:portal
    class="absolute pointer-events-none drag-container"
    style:z-index="1000"
    style:top="{dragItemPosition.top}px"
    style:left="{dragItemPosition.left}px"
    style:width="{dragItemDimensions.width}px"
    style:height="{dragItemDimensions.height}px"
  >
    <div class="block-ghost" class:dashed={dragBlock.kind === "container"}>
      <div class="block-ghost-header">
        {#if dragGroupTabs}
          {#each $dragGroupTabs ?? [] as tab (tab.name)}
            <span>{tab.displayName}</span>
          {/each}
        {:else}
          <span>{m.canvas_container()}</span>
        {/if}
      </div>
    </div>
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
      <AlertDialog.Title>{m.canvas_delete_widget()}</AlertDialog.Title>

      <AlertDialog.Description>
        {m.canvas_delete_widget_description()}
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
              {m.canvas_cancel()}
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
              {m.canvas_delete()}
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
      <AlertDialog.Title>{m.canvas_delete_tab()}</AlertDialog.Title>

      <AlertDialog.Description>
        {m.canvas_delete_tab_description()}
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
              {m.common_cancel()}
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
              {m.common_delete()}
            </Button>
          {/snippet}
        </AlertDialog.Action>
      </AlertDialog.Footer>
    </AlertDialog.Content>
  </AlertDialog.Root>
{/if}

{#if pendingBlockDelete !== undefined}
  <BlockDeleteDialog
    kind="container"
    onConfirm={() => {
      if (pendingBlockDelete !== undefined) deleteBlockAt(pendingBlockDelete);
      pendingBlockDelete = undefined;
    }}
    onCancel={() => (pendingBlockDelete = undefined)}
  />
{/if}

<style lang="postcss">
  .drag-container {
    container-type: inline-size;
    container-name: component-container;
  }

  /* Ghost of a block being dragged: its header (a tab group's labels, or the container's
     name) over a faded body. */
  .block-ghost {
    @apply size-full overflow-hidden rounded-md border border-primary-400 px-3 py-2;
    @apply bg-surface-subtle/80 opacity-80 shadow-md;
  }

  .block-ghost.dashed {
    @apply border-dashed;
  }

  .block-ghost-header {
    @apply flex h-9 items-end gap-x-4 border-b border-gray-200 pb-2;
    @apply text-sm font-medium text-fg-secondary;
  }
</style>
