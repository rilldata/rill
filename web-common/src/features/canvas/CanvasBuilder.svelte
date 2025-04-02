<script lang="ts">
  import { portal } from "@rilldata/web-common/lib/actions/portal";
  import {
    type V1CanvasRow as APIV1CanvasRow,
    type V1CanvasItem,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { get, writable } from "svelte/store";
  import { parseDocument } from "yaml";
  import type { FileArtifact } from "../entity-management/file-artifact";
  import AddComponentDropdown from "./AddComponentDropdown.svelte";
  import CanvasComponent from "./CanvasComponent.svelte";
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import DropZone from "./components/DropZone.svelte";
  import type { CanvasComponentType } from "./components/types";
  import ElementDivider from "./ElementDivider.svelte";
  import ItemWrapper from "./ItemWrapper.svelte";
  import type { DragItem, YAMLRow } from "./layout-util";
  import {
    COLUMN_COUNT,
    DEFAULT_DASHBOARD_WIDTH,
    mapGuard,
    moveToRow,
    rowsGuard,
    mousePosition,
  } from "./layout-util";
  import { activeDivider } from "./stores/ui-stores";

  import RowWrapper from "./RowWrapper.svelte";
  import { useDefaultMetrics, type CanvasResponse } from "./selector";
  import { getCanvasStore } from "./state-managers/state-managers";
  import { dropZone } from "./stores/ui-stores";
  import ComponentError from "./components/ComponentError.svelte";
  import EditableCanvasRow from "./EditableCanvasRow.svelte";

  const activelyEditing = writable(false);

  type V1CanvasRow = Omit<APIV1CanvasRow, "items"> & {
    items: (V1CanvasItem | null)[];
  };

  export let fileArtifact: FileArtifact;
  export let canvasData: CanvasResponse | undefined;
  export let canvasName: string;
  export let openSidebar: () => void;

  $: ({
    canvasEntity: { setSelectedComponent },
  } = getCanvasStore(canvasName));

  let initialMousePosition: { x: number; y: number } | null = null;
  let clientWidth: number;
  let selected: Set<string> = new Set();
  let offset = { x: 0, y: 0 };

  let dragItemInfo: DragItem | null = null;
  let timeout: ReturnType<typeof setTimeout> | null = null;
  let dragTimeout: ReturnType<typeof setTimeout> | null = null;
  let dragItemPosition = { top: 0, left: 0 };
  let dragItemDimensions = { width: 0, height: 0 };
  let spec = canvasData?.canvas ?? {
    rows: [],
    filtersEnabled: false,
    maxWidth: DEFAULT_DASHBOARD_WIDTH,
  };
  let openSidebarAfterSelection = false;

  $: ({ instanceId } = $runtime);
  $: metricsViews = Object.entries(canvasData?.metricsViews ?? {});

  $: metricsViewQuery = useDefaultMetrics(instanceId, metricsViews?.[0]?.[0]);

  $: ({ editorContent, updateEditorContent } = fileArtifact);
  $: contents = parseDocument($editorContent ?? "");

  $: if (canvasData?.canvas) {
    if (!get(activelyEditing)) {
      spec = structuredClone(canvasData?.canvas ?? spec);
    }
  }

  $: activelyEditing.set(!!$activeDivider || !!dragItemInfo);

  $: ({ rows = [], filtersEnabled, maxWidth: canvasMaxWidth } = spec);

  $: maxWidth = canvasMaxWidth || DEFAULT_DASHBOARD_WIDTH;

  $: specCanvasRows = structuredClone(rows) as V1CanvasRow[];

  $: rawYamlRows = rowsGuard(contents?.get("rows"));
  $: yamlCanvasRows = mapGuard(rawYamlRows);

  $: columnWidth = clientWidth / 12;

  $: mouseDelta = initialMousePosition
    ? calculateMouseDelta(initialMousePosition, $mousePosition)
    : 0;

  $: dropZone.setMouseDelta(mouseDelta);

  $: defaultMetrics = $metricsViewQuery?.data;

  function updateComponentWidths(rowIndex: number, newWidths: number[]) {
    newWidths.forEach((width, i) => {
      try {
        contents.setIn(["rows", rowIndex, "items", i, "width"], width);
      } catch (e) {
        console.error(e);
      }
    });

    updateContents();
  }

  function getId(row: number | undefined, column: number | undefined) {
    return `component-${row ?? 0}-${column ?? 0}`;
  }

  function calculateMouseDelta(
    pos1: { x: number; y: number },
    pos2: { x: number; y: number },
  ) {
    return Math.sqrt((pos1.x - pos2.x) ** 2 + (pos1.y - pos2.y) ** 2);
  }

  function handleDragStart(metadata: DragItem) {
    dragItemInfo = metadata;

    const id = getId(metadata.position?.row, metadata.position?.column);
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

  $: if (dragItemInfo) {
    dragItemPosition = {
      top: $mousePosition.y + offset.y,
      left: $mousePosition.x + offset.x,
    };
  }

  function onDragEnd() {
    dragItemInfo = null;
  }

  function reset() {
    if (dragTimeout) {
      clearTimeout(dragTimeout);
    }

    if (dragItemInfo) {
      onDragEnd();
    }

    if (openSidebarAfterSelection) {
      openSidebar();
      openSidebarAfterSelection = false;
    }

    dropZone.clear();
  }

  function updateRowHeight(newHeight: number, rowIndex: number) {
    try {
      contents.setIn(["rows", rowIndex, "height"], newHeight + "px");
    } catch (e) {
      console.error(e);
    }

    updateContents();
  }

  function spreadEvenly(index: number) {
    const specRowsClone = structuredClone(specCanvasRows);
    const yamlRowsClone = structuredClone(yamlCanvasRows);
    const specRow = specRowsClone[index];
    const yamlRow = yamlRowsClone[index];
    if (!specRow?.items || !yamlRow?.items) return;

    const baseSize = COLUMN_COUNT / specRow.items.length;

    yamlRow.items.forEach((_, i) => {
      if (!specRow.items[i] || !yamlRow.items[i]) return;
      specRow.items[i].width = baseSize;
      yamlRow.items[i].width = baseSize;
    });

    updateAssets(specRowsClone, yamlRowsClone);
  }

  function updateAssets(
    specRows: V1CanvasRow[],
    yamlRows: YAMLRow[],
    clearSelection = true,
  ) {
    if (clearSelection) {
      selected = new Set();
      setSelectedComponent(null);
    }

    specCanvasRows = specRows;

    try {
      contents.setIn(["rows"], yamlRows);
    } catch (e) {
      console.error(e);
    }

    updateContents();
  }

  function dropItemsInExistingRow(
    items: DragItem[],
    row: number,
    column: number,
  ) {
    const newYamlRows = moveToRow(yamlCanvasRows, items, { row, column });
    const newSpecRows = moveToRow(specCanvasRows, items, { row, column });

    updateAssets(newSpecRows, newYamlRows);
  }

  function moveToNewRow(items: DragItem[], row: number) {
    const newSpecRows = moveToRow(specCanvasRows, items, { row });
    const newYamlRows = moveToRow(yamlCanvasRows, items, { row });

    updateAssets(newSpecRows, newYamlRows);
  }

  function removeItems(items: { position: { row: number; column: number } }[]) {
    const newSpecRows = moveToRow(specCanvasRows, items);
    const newYamlRows = moveToRow(yamlCanvasRows, items);

    updateAssets(newSpecRows, newYamlRows);
  }

  function addItems(
    position: { row: number; column: number },
    items: CanvasComponentType[],
  ) {
    if (!defaultMetrics) return;
    const newYamlRows = moveToRow(
      yamlCanvasRows,
      items.map((type) => ({ type })),
      position,

      defaultMetrics,
    );

    const newSpecRows = moveToRow(
      specCanvasRows,
      items.map((type) => ({ type })),
      position,
    );

    updateAssets(newSpecRows, newYamlRows);

    const id = getId(position.row, position.column);

    selected = new Set([id]);

    setSelectedComponent({ column: position.column, row: position.row });
  }

  function updateContents() {
    const newContent = contents.toString();
    if (newContent === $editorContent) {
      contents = parseDocument(newContent);
    } else {
      updateEditorContent(contents.toString(), false, true);
    }
  }

  function initializeRow(row: number, type: CanvasComponentType) {
    if (!defaultMetrics) return;

    const newYamlRows = moveToRow(
      yamlCanvasRows,
      [{ type }],
      { row },
      defaultMetrics,
    );
    const newSpecRows = moveToRow(specCanvasRows, [{ type }], { row });

    updateAssets(newSpecRows, newYamlRows);

    const id = getId(row, 0);

    selected = new Set([id]);

    setSelectedComponent({ column: 0, row });
  }

  function onDrop(row: number, column: number | null) {
    if (!$dropZone) return;
    dropZone.clear();

    if (dragItemInfo) {
      if (
        row === dragItemInfo.position?.row &&
        (column === dragItemInfo.position.column ||
          column === dragItemInfo.position?.column + 1)
      ) {
        return;
      }
      if (column === null) {
        moveToNewRow([dragItemInfo], row);
      } else {
        dropItemsInExistingRow([dragItemInfo], row, column);
      }
    }
  }

  function resetSelection() {
    setSelectedComponent(null);
    selected = new Set();
  }

  function scrollToBottom() {
    const element = document.querySelector("#canvas-scroll-container");
    if (element) {
      element.scrollTop = element.scrollHeight;
    }
  }
</script>

<svelte:window
  on:mouseup={reset}
  on:keydown={(e) => {
    if (e.key === "Backspace" && selected) {
      if (
        !(e.target instanceof HTMLElement) ||
        (e.target.tagName !== "INPUT" &&
          e.target.tagName !== "TEXTAREA" &&
          !e.target.isContentEditable)
      ) {
        removeItems(
          Array.from(selected).map((id) => {
            const [row, column] = id.split("-").slice(1).map(Number);
            return { position: { row, column } };
          }),
        );
      }
    }
  }}
/>

<CanvasDashboardWrapper
  {canvasName}
  {maxWidth}
  {filtersEnabled}
  onClick={resetSelection}
  showGrabCursor={!!dragItemInfo}
  bind:clientWidth
>
  {#each specCanvasRows as row, rowIndex (rowIndex)}
    <EditableCanvasRow
      {row}
      {canvasData}
      {maxWidth}
      {rowIndex}
      movingWidget={!!dragItemInfo}
      zIndex={50 - rowIndex * 2}
      {columnWidth}
      {updateComponentWidths}
      {onDrop}
      {initializeRow}
      {updateRowHeight}
      let:widths
      let:isSpreadEvenly
      let:types
      let:items
      let:onColumnResizeStart
    >
      {#each items as item, columnIndex (columnIndex)}
        {@const id = getId(rowIndex, columnIndex)}
        {@const type = types[columnIndex]}
        {@const componentResource =
          canvasData?.components?.[item?.component ?? ""]}
        <ItemWrapper {type} zIndex={4 - columnIndex}>
          {#if columnIndex === 0}
            <ElementDivider
              {rowIndex}
              resizeIndex={-1}
              addIndex={columnIndex}
              rowLength={items.length}
              dragging={!!dragItemInfo}
              {isSpreadEvenly}
              {spreadEvenly}
              {addItems}
            />
          {/if}

          <ElementDivider
            {isSpreadEvenly}
            columnWidth={widths[columnIndex]}
            {rowIndex}
            dragging={!!dragItemInfo}
            resizeIndex={columnIndex}
            addIndex={columnIndex + 1}
            rowLength={items.length}
            {spreadEvenly}
            {addItems}
            {onColumnResizeStart}
          />

          <DropZone
            column={columnIndex}
            row={rowIndex}
            maxColumns={items.length}
            allowDrop={!!dragItemInfo}
            {onDrop}
          />

          <CanvasComponent
            {canvasName}
            {componentResource}
            canvasItem={item}
            {id}
            editable
            ghost={dragItemInfo?.position?.row === rowIndex &&
              dragItemInfo?.position?.column === columnIndex}
            selected={selected.has(id)}
            allowPointerEvents={!$activeDivider}
            onMouseDown={(e) => {
              if (e.button !== 0) return;
              e.preventDefault();

              initialMousePosition = $mousePosition;

              setSelectedComponent({ column: columnIndex, row: rowIndex });
              selected = new Set([id]);

              if (dragTimeout) clearTimeout(dragTimeout);

              openSidebarAfterSelection = true;

              dragTimeout = setTimeout(() => {
                openSidebarAfterSelection = false;
                handleDragStart({
                  position: { row: rowIndex, column: columnIndex },
                  type: type ?? "line_chart",
                });
              }, 150);
            }}
            onDuplicate={() => {
              if (!defaultMetrics) return;

              const newYamlRows = moveToRow(
                yamlCanvasRows,
                [{ position: { row: rowIndex, column: columnIndex } }],
                { row: rowIndex + 1, copy: true },
              );
              const newSpecRows = moveToRow(
                specCanvasRows,
                [{ position: { row: rowIndex, column: columnIndex } }],
                { row: rowIndex + 1, copy: true },
              );

              updateAssets(newSpecRows, newYamlRows);
            }}
            onDelete={() =>
              removeItems([
                { position: { row: rowIndex, column: columnIndex } },
              ])}
          />
        </ItemWrapper>
      {/each}
    </EditableCanvasRow>
  {:else}
    <RowWrapper
      gridTemplate="12fr"
      zIndex={0}
      {maxWidth}
      rowIndex={specCanvasRows.length}
    >
      <ItemWrapper zIndex={0}>
        {#if defaultMetrics}
          <AddComponentDropdown
            componentForm
            onMouseEnter={() => {
              if (timeout) clearTimeout(timeout);
            }}
            onItemClick={(type) => {
              initializeRow(specCanvasRows.length, type);
            }}
          />
        {:else if canvasData}
          <ComponentError error="No valid metrics view in project" />
        {/if}
      </ItemWrapper>
    </RowWrapper>
  {/each}
</CanvasDashboardWrapper>

{#if specCanvasRows.length}
  <AddComponentDropdown
    floatingForm
    onItemClick={(type) => {
      initializeRow(specCanvasRows.length, type);
      setTimeout(() => scrollToBottom(), 500);
    }}
  />
{/if}

{#if dragItemInfo && dragItemInfo.position}
  {@const item =
    specCanvasRows[dragItemInfo.position.row]?.items?.[
      dragItemInfo.position.column
    ]}
  {@const componentResource = canvasData?.components?.[item?.component ?? ""]}
  {#if item}
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
        {canvasName}
        {componentResource}
        id="canvas-drag-item"
        canvasItem={item}
        allowPointerEvents={false}
        ghost
        selected
      />
    </div>
  {/if}
{/if}

<style lang="postcss">
  .drag-container {
    container-type: inline-size;
    container-name: component-container;
  }
</style>
