<script lang="ts">
  import { portal } from "@rilldata/web-common/lib/actions/portal";
  import { clamp } from "@rilldata/web-common/lib/clamp";
  import {
    type V1CanvasRow as APIV1CanvasRow,
    createQueryServiceResolveCanvas,
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
    getInitialHeight,
    mapGuard,
    MIN_HEIGHT,
    MIN_WIDTH,
    moveToRow,
    normalizeSizeArray,
    rowsGuard,
  } from "./layout-util";
  import RowDropZone from "./RowDropZone.svelte";
  import RowWrapper from "./RowWrapper.svelte";
  import { useDefaultMetrics } from "./selector";
  import { getCanvasStateManagers } from "./state-managers/state-managers";
  import { activeDivider, dropZone } from "./stores/ui-stores";

  const activelyEditing = writable(false);

  type V1CanvasRow = Omit<APIV1CanvasRow, "items"> & {
    items: (V1CanvasItem | null)[];
  };

  const ctx = getCanvasStateManagers();

  const {
    canvasEntity: {
      setSelectedComponent,
      spec: { canvasSpec, metricViewNames },
      name: canvasName,
    },
  } = ctx;

  export let fileArtifact: FileArtifact;
  export let openSidebar: () => void;

  let mousePosition = { x: 0, y: 0 };
  let initialMousePosition: { x: number; y: number } | null = null;
  let clientWidth: number;
  let selected: Set<string> = new Set();
  let offset = { x: 0, y: 0 };
  let resizeRow = -1;
  let resizeRowMinimum = MIN_HEIGHT;
  let initialHeight = 0;
  let dragItemInfo: DragItem | null = null;
  let resizeColumnInfo: {
    width: number;
    row: number;
    column: number;
    maxWidth: number;
    nextElementWidth: number;
  } | null = null;
  let timeout: ReturnType<typeof setTimeout> | null = null;
  let dragTimeout: ReturnType<typeof setTimeout> | null = null;
  let dragItemPosition = { top: 0, left: 0 };
  let dragItemDimensions = { width: 0, height: 0 };
  let spec = $canvasSpec ?? {
    rows: [],
    filtersEnabled: false,
    maxWidth: DEFAULT_DASHBOARD_WIDTH,
  };

  $: ({ instanceId } = $runtime);

  $: metricsViewQuery = useDefaultMetrics(instanceId, $metricViewNames?.[0]);

  $: ({ editorContent, updateEditorContent } = fileArtifact);
  $: contents = parseDocument($editorContent ?? "");

  $: if ($canvasSpec) {
    if (!get(activelyEditing)) {
      spec = structuredClone($canvasSpec ?? spec);
    }
  }

  $: activelyEditing.set(
    resizeRow !== -1 || !!dragItemInfo || !!resizeColumnInfo,
  );

  $: ({ rows = [], filtersEnabled, maxWidth: canvasMaxWidth } = spec);

  $: maxWidth = canvasMaxWidth || DEFAULT_DASHBOARD_WIDTH;

  $: specCanvasRows = structuredClone(rows) as V1CanvasRow[];

  $: rawYamlRows = rowsGuard(contents?.get("rows"));
  $: yamlCanvasRows = mapGuard(rawYamlRows);

  $: columnWidth = clientWidth / 12;

  $: mouseDelta = initialMousePosition
    ? calculateMouseDelta(initialMousePosition, mousePosition)
    : 0;

  $: dropZone.setMouseDelta(mouseDelta);

  $: resizeRowData = structuredClone(specCanvasRows?.[resizeRow]);

  $: resizeColumnData =
    resizeColumnInfo &&
    structuredClone(specCanvasRows?.[resizeColumnInfo.row]?.items);

  $: if (resizeRowData && initialMousePosition) {
    const diff = mousePosition.y - initialMousePosition.y;

    resizeRowData.height = Math.max(
      resizeRowMinimum,
      Math.floor(diff + initialHeight),
    );

    specCanvasRows[resizeRow] = resizeRowData;
  }

  $: defaultMetrics = $metricsViewQuery?.data;

  $: canvasResolverQuery = createQueryServiceResolveCanvas(
    instanceId,
    canvasName,
    {},
  );

  $: canvasData = $canvasResolverQuery.data;

  function onColumResizeStart(e: MouseEvent & { currentTarget: HTMLElement }) {
    initialMousePosition = mousePosition;
    const row = Number(e.currentTarget.getAttribute("data-row"));
    const column = Number(e.currentTarget.getAttribute("data-column"));
    const rowWidths =
      specCanvasRows[row]?.items?.map((el) => el?.width ?? 0) ?? [];

    const nextElementWidth = rowWidths[column + 1];

    const maxWidth = rowWidths.reduce((acc, el, i) => {
      if (i === column) {
        return acc;
      } else if (i === column + 1) {
        return acc - MIN_WIDTH;
      } else {
        return acc - el;
      }
    }, COLUMN_COUNT);

    if (!nextElementWidth) return;

    resizeColumnInfo = {
      width: Number(e.currentTarget.getAttribute("data-width")),
      row,
      column,
      nextElementWidth,
      maxWidth,
    };

    window.addEventListener("mousemove", onColumnResize);
    window.addEventListener("mouseup", onColumnResizeEnd);
  }

  function onColumnResize(e: MouseEvent) {
    if (!resizeColumnInfo || !resizeColumnData) return;

    const { row, column, width, maxWidth, nextElementWidth } = resizeColumnInfo;

    const layoutRow = resizeColumnData.map((el) => el?.width ?? 0) ?? [];

    const delta = e.clientX - (initialMousePosition?.x ?? 0);
    const columnDelta = Math.round(delta / columnWidth);

    const newValue = clamp(MIN_WIDTH, width + columnDelta, maxWidth);

    const clampedDelta = newValue - width;

    layoutRow[column] = newValue;

    layoutRow[column + 1] = nextElementWidth - clampedDelta;

    layoutRow.forEach((el, i) => {
      if (!resizeColumnData[i]) return;

      resizeColumnData[i].width = el;
    });

    if (!specCanvasRows[row]) return;

    specCanvasRows[row].items = resizeColumnData;
  }

  function onColumnResizeEnd() {
    window.removeEventListener("mousemove", onColumnResize);
    window.removeEventListener("mouseup", onColumnResizeEnd);

    if (!resizeColumnInfo || !resizeColumnData) return;
    const rowIndex = resizeColumnInfo.row;
    resizeColumnData.forEach((el, i) => {
      if (!el) return;
      contents.setIn(["rows", rowIndex, "items", i, "width"], el.width);
    });

    updateContents();
    resizeColumnInfo = null;
    document.body.style.cursor = "";
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

    initialMousePosition = mousePosition;

    const id = getId(metadata.position?.row, metadata.position?.column);
    const element = document.querySelector("#" + id);
    if (!element) return;

    const width = element.clientWidth;
    const height = element.clientHeight;

    const top = element.getBoundingClientRect().top;
    const left = element.getBoundingClientRect().left;
    dragItemDimensions = { width, height };

    offset = {
      x: left - mousePosition.x,
      y: top - mousePosition.y,
    };
  }

  $: if (dragItemInfo) {
    dragItemPosition = {
      top: mousePosition.y + offset.y,
      left: mousePosition.x + offset.x,
    };
  }

  function onDragEnd() {
    dragItemInfo = null;
  }

  function onRowResizeStart(
    rowIndex: number,

    types: (string | undefined)[],
  ) {
    initialMousePosition = mousePosition;
    resizeRow = rowIndex;
    initialHeight =
      document
        .querySelector(`#canvas-row-${resizeRow}`)
        ?.getBoundingClientRect().height ??
      Number(specCanvasRows[resizeRow]?.height ?? MIN_HEIGHT);
    resizeRowMinimum =
      types.reduce((acc, type) => {
        return Math.max(acc, getInitialHeight(type) ?? MIN_HEIGHT);
      }, 0) ?? MIN_HEIGHT;
  }

  function reset() {
    if (dragTimeout) {
      clearTimeout(dragTimeout);
    }
    if (resizeRow !== -1) {
      onRowResizeEnd();
    }

    if (dragItemInfo) {
      onDragEnd();
    }

    activeDivider.set(null);
    dropZone.clear();
  }

  function onRowResizeEnd() {
    const height = specCanvasRows[resizeRow]?.height;

    if (!height) return;

    try {
      contents.setIn(
        ["rows", resizeRow, "height"],
        Math.max(resizeRowMinimum, height) + "px",
      );
    } catch (e) {
      console.error(e);
    }

    initialMousePosition = null;
    resizeRow = -1;
    initialHeight = 0;

    updateContents();
  }

  function spreadEvenly(index: number) {
    const specRow = structuredClone(specCanvasRows[index]);
    const yamlRow = structuredClone(yamlCanvasRows[index]);
    if (!specRow?.items || !yamlRow?.items) return;

    const baseSize = COLUMN_COUNT / specRow.items.length;

    yamlRow.items.forEach((_, i) => {
      if (!specRow.items[i] || !yamlRow.items[i]) return;
      specRow.items[i].width = baseSize;
      yamlRow.items[i].width = baseSize;
    });

    updateAssets(specCanvasRows, yamlCanvasRows);
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

    contents.setIn(["rows"], yamlRows);

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
  }

  function updateContents() {
    updateEditorContent(contents.toString(), false, true);
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
</script>

<svelte:window
  on:mouseup={reset}
  on:mousemove={(e) => {
    mousePosition = { x: e.clientX, y: e.clientY };
  }}
  on:keydown={(e) => {
    if (e.key === "Backspace" && selected) {
      if (e.target === document.body) {
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
  {maxWidth}
  {filtersEnabled}
  onClick={resetSelection}
  showGrabCursor={!!dragItemInfo}
  bind:clientWidth
>
  {#each specCanvasRows as { items = [], height = MIN_HEIGHT, heightUnit = "px" }, rowIndex (rowIndex)}
    {@const widths = normalizeSizeArray(items?.map((el) => el?.width ?? 0))}
    {@const isSpreadEvenly = items?.every(
      (el) => el?.width === items?.[0]?.width,
    )}
    {@const types = items?.map(
      (item) =>
        canvasData?.resolvedComponents?.[item?.component ?? ""]?.component?.spec
          ?.renderer,
    )}
    <RowWrapper
      {maxWidth}
      {rowIndex}
      zIndex={50 - rowIndex * 2}
      height="{height}{heightUnit}"
      gridTemplate={widths.map((w) => `${w}fr`).join(" ")}
    >
      {#each items as item, columnIndex (columnIndex)}
        {@const width = widths[columnIndex]}
        {@const id = getId(rowIndex, columnIndex)}
        {@const type = types[columnIndex]}
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
            onMouseDown={onColumResizeStart}
            columnWidth={width}
            {rowIndex}
            dragging={!!dragItemInfo}
            resizeIndex={columnIndex}
            addIndex={columnIndex + 1}
            rowLength={items.length}
            {spreadEvenly}
            {addItems}
          />

          <DropZone
            column={columnIndex}
            row={rowIndex}
            maxColumns={items.length}
            allowDrop={!!dragItemInfo}
            {onDrop}
          />

          <CanvasComponent
            canvasItem={item}
            {id}
            editable
            ghost={dragItemInfo?.position?.row === rowIndex &&
              dragItemInfo?.position?.column === columnIndex}
            selected={selected.has(id)}
            allowPointerEvents={!resizeColumnInfo && resizeRow === -1}
            onMouseDown={(e) => {
              if (e.button !== 0) return;

              setSelectedComponent({ column: columnIndex, row: rowIndex });
              selected = new Set([id]);
              openSidebar();

              if (dragTimeout) clearTimeout(dragTimeout);

              dragTimeout = setTimeout(() => {
                handleDragStart({
                  position: { row: rowIndex, column: columnIndex },
                  type: type ?? "line_chart",
                });
              }, 100);
            }}
          />
        </ItemWrapper>
      {/each}

      <RowDropZone
        allowDrop={!!dragItemInfo}
        resizeIndex={rowIndex}
        dropIndex={rowIndex + 1}
        onRowResizeStart={() => {
          onRowResizeStart(rowIndex, types);
        }}
        {onDrop}
        addItem={(type) => {
          initializeRow(rowIndex + 1, type);
        }}
      />

      {#if rowIndex === 0}
        <RowDropZone
          allowDrop={!!dragItemInfo}
          dropIndex={0}
          {onDrop}
          addItem={(type) => {
            initializeRow(rowIndex, type);
          }}
        />
      {/if}
    </RowWrapper>
  {:else}
    <AddComponentDropdown
      componentForm
      onMouseEnter={() => {
        if (timeout) clearTimeout(timeout);
      }}
      onItemClick={(type) => {
        initializeRow(0, type);
      }}
    />
  {/each}
</CanvasDashboardWrapper>

{#if dragItemInfo && dragItemInfo.position}
  {@const item =
    specCanvasRows[dragItemInfo.position.row]?.items?.[
      dragItemInfo.position.column
    ]}
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
