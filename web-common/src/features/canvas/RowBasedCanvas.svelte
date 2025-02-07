<script lang="ts">
  import { parseDocument, YAMLMap, YAMLSeq } from "yaml";
  import { clamp } from "@rilldata/web-common/lib/clamp";
  import ElementDivider from "./ElementDivider.svelte";
  import AddComponentDropdown from "./AddComponentDropdown.svelte";
  import type { FileArtifact } from "../entity-management/file-artifact";
  import { getCanvasStateManagers } from "./state-managers/state-managers";
  import type { V1CanvasSpec } from "@rilldata/web-common/runtime-client";
  import PreviewElement from "./PreviewElement.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { CanvasComponentType } from "./components/types";
  import { getComponentRegistry } from "./components/util";
  import { findNextAvailablePosition } from "./util";
  import { useDefaultMetrics } from "./selector";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import CanvasFilters from "./filters/CanvasFilters.svelte";
  import DropZone from "./components/DropZone.svelte";

  const initialHeights: Record<CanvasComponentType, number> = {
    line_chart: 350,
    bar_chart: 400,
    area_chart: 400,
    stacked_bar: 400,
    markdown: 160,
    kpi: 200,
    image: 420,
    table: 400,
  };

  function getInitalHeight(id: string | undefined) {
    return initialHeights[id as CanvasComponentType] ?? MIN_HEIGHT;
  }

  const COLUMN_COUNT = 12;
  const MIN_HEIGHT = 160;
  const MIN_WIDTH = 3;
  const MINIMUM_MOVEMENT = 12;
  const baseLayoutArrays = [
    [],
    [COLUMN_COUNT],
    [6, 6],
    [4, 4, 4],
    [3, 3, 3, 3],
  ];

  const ctx = getCanvasStateManagers();

  const componentRegistry = getComponentRegistry();

  const {
    canvasEntity: {
      setSelectedComponentIndex,
      spec: { canvasSpec },
    },
  } = ctx;

  const { canvasEntity } = getCanvasStateManagers();

  let spec: V1CanvasSpec = {
    items: [],
    filtersEnabled: true,
  };

  export let maxWidth: number = 1200;
  export let fileArtifact: FileArtifact;

  let mousePosition = { x: 0, y: 0 };
  let initialMousePosition: { x: number; y: number } | null = null;
  let clientWidth: number;
  let selected: Set<string> = new Set();
  let clone: HTMLElement;
  let offset = { x: 0, y: 0 };
  let resizeRow = -1;
  let initialHeight = 0;
  // {row}-{order}
  let hoveredDropZone: string | null = null;
  let rowHover: number | null = null;
  let dragItemInfo: DragItem | null = null;
  let resizeInfo: {
    width: number;
    row: number;
    column: number;
    maxWidth: number;
    nextElementWidth: number;
  } | null = null;

  type DragItem = {
    name: string | number;
    row: number;
    order: number;
    height?: number;
    type?: CanvasComponentType;
  };

  type Row = {
    items: (string | number | null)[];
    height: number;
    layout: number[];
  };

  type YAMLRow = {
    items: string;
    height: string;
    layout: string;
  };

  $: ({ instanceId } = $runtime);

  $: spec = structuredClone($canvasSpec ?? spec);

  $: ({ items: canvasItems = [], filtersEnabled } = spec);

  $: ({ editorContent, updateEditorContent } = fileArtifact);

  $: fileContent = parseDocument($editorContent ?? "");

  $: metricsViewQuery = useDefaultMetrics(instanceId);

  $: contents = fileContent.get("layout") as YAMLMap;
  $: rawYamlItems = fileContent.get("items") as YAMLSeq;

  $: yamlRows = rowsGuard(contents?.get("rows"));

  $: rowMaps = mapGuard(yamlRows.items);

  $: components = rawYamlItems?.items ?? [];

  $: columnWidth = clientWidth / 12;

  $: mouseDelta = initialMousePosition
    ? calculateMouseDelta(initialMousePosition, mousePosition)
    : 0;
  $: passedThreshold = mouseDelta > MINIMUM_MOVEMENT;

  $: if (resizeRow !== -1 && initialMousePosition) {
    const diff = mousePosition.y - initialMousePosition.y;

    rowMaps[resizeRow].height = Math.max(
      MIN_HEIGHT,
      Math.floor(diff + initialHeight),
    );
  }

  function rowsGuard(value: unknown): YAMLSeq {
    if (!value || !(value instanceof YAMLSeq)) {
      return new YAMLSeq();
    } else {
      return value as YAMLSeq;
    }
  }

  function mapGuard(value: unknown[]): Array<Row> {
    return value.map((el) => {
      if (el instanceof YAMLMap) {
        const jsonObject = el.toJSON() as Partial<YAMLRow>;

        return {
          items:
            jsonObject?.items
              ?.toString()
              .split(",")
              ?.map((el) => Number(el.trim())) ?? [],
          height: jsonObject?.height ? Number(jsonObject?.height) : MIN_HEIGHT,
          layout:
            jsonObject?.layout
              ?.toString()
              .split(",")
              ?.map((el) => Number(el.trim())) ?? [],
        };
      } else {
        return {
          items: [],
          height: MIN_HEIGHT,
          layout: [],
        };
      }
    });
  }

  function onColumResizeStart(e: MouseEvent & { currentTarget: HTMLElement }) {
    initialMousePosition = mousePosition;
    const row = Number(e.currentTarget.getAttribute("data-row"));
    const column = Number(e.currentTarget.getAttribute("data-column"));

    const currentLayout = rowMaps[row].layout;
    const nextElementWidth = currentLayout[column + 1];

    const maxWidth = currentLayout.reduce((acc, el, i) => {
      if (i === column) {
        return acc;
      } else if (i === column + 1) {
        return acc - MIN_WIDTH;
      } else {
        return acc - el;
      }
    }, COLUMN_COUNT);

    if (!nextElementWidth) return;

    resizeInfo = {
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
    if (!resizeInfo) return;

    const { row, column, width, maxWidth, nextElementWidth } = resizeInfo;
    const layoutRow = [...rowMaps[row].layout];

    const delta = e.clientX - (initialMousePosition?.x ?? 0);
    const columnDelta = Math.round(delta / columnWidth);

    const newValue = clamp(MIN_WIDTH, width + columnDelta, maxWidth);

    const clampedDelta = newValue - width;

    layoutRow[column] = newValue;

    layoutRow[column + 1] = nextElementWidth - clampedDelta;

    rowMaps[row].layout = layoutRow;
  }

  function onColumnResizeEnd() {
    window.removeEventListener("mousemove", onColumnResize);
    window.removeEventListener("mouseup", onColumnResizeEnd);

    if (!resizeInfo) return;
    contents.setIn(
      ["rows", resizeInfo.row, "layout"],
      rowMaps[resizeInfo.row].layout.join(", "),
    );

    updateContents();
    resizeInfo = null;
  }

  function getId(row: number, column: number) {
    return `component-${row}-${column}`;
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

    const id = getId(metadata.row, metadata.order);
    const element = document.querySelector("#" + id);
    if (!element) return;

    // duplicate element
    clone = element.cloneNode(true) as HTMLElement;

    const width = element.clientWidth;
    const height = element.clientHeight;

    const top = element.getBoundingClientRect().top;
    const left = element.getBoundingClientRect().left;

    offset = {
      x: left - mousePosition.x,
      y: top - mousePosition.y,
    };

    document.body.appendChild(clone);

    clone.style.position = "absolute";
    clone.style.top = top + "px";
    clone.style.left = left + "px";
    clone.style.width = width + 2 + "px";
    clone.style.height = height + 2 + "px";
    clone.classList.add("outline", "outline-primary-300");
    clone.style.opacity = "0.8";
    clone.style.pointerEvents = "none";
    clone.style.zIndex = "1000";
    clone.classList.add("shadow-md");

    window.addEventListener("mousemove", onDrag);
    window.addEventListener("mouseup", onDragEnd);
  }

  function onDrag(e: MouseEvent) {
    clone.style.top = e.clientY + offset.y + "px";
    clone.style.left = e.clientX + offset.x + "px";
  }

  function onDragEnd() {
    // selected = new Set();
    window.removeEventListener("mousemove", onDrag);
    window.removeEventListener("mouseup", onDragEnd);
    clone.remove();
    dragItemInfo = null;
    rowHover = null;
  }

  function onRowResizeStart(e: MouseEvent & { currentTarget: HTMLElement }) {
    initialMousePosition = mousePosition;
    resizeRow = Number(e.currentTarget.getAttribute("data-row"));
    initialHeight = rowMaps[resizeRow].height;
  }

  function reset() {
    if (resizeRow !== -1) {
      onRowResizeEnd();
    }
  }

  function onRowResizeEnd() {
    const height = rowMaps[resizeRow]?.height;

    try {
      fileContent.setIn(["layout", "rows", resizeRow, "height"], height);
    } catch (e) {
      console.error(e);
    }

    initialMousePosition = null;
    resizeRow = -1;
    initialHeight = 0;

    updateContents();
  }

  function spreadEvenly(index: number) {
    contents.setIn(
      ["rows", index, "layout"],
      baseLayoutArrays[rowMaps[index].items.length].join(", "),
    );
    updateContents();
  }

  function dropItemsInExistingRow(
    items: DragItem[],
    rowIndex: number,
    column: number,
  ) {
    if (!contents) {
      contents = new YAMLMap();
      fileContent.set("layout", contents);
    }
    const rowsClone = structuredClone(rowMaps) as (Row | null)[];
    const destinationRow = rowsClone[rowIndex];
    const touchedRows = new Set(items.map((el) => el.row));

    if (!destinationRow) return;

    items.forEach((item) => {
      const row = rowsClone[item.row];
      if (!row) return;
      row.items[item.order] = null;
    });

    destinationRow.items.splice(column, 0, ...items.map((el) => el.name));
    if (destinationRow.items.filter(itemExists).length > 4) {
      return;
    }
    destinationRow.layout = baseLayoutArrays[destinationRow.items.length];

    touchedRows.forEach((rowIndex) => {
      const row = rowsClone[rowIndex];
      if (!row) return;
      const validItemsLeft = row.items.filter(itemExists);

      if (!validItemsLeft.length) {
        rowsClone[rowIndex] = null;
      } else {
        row.items = validItemsLeft;
        row.layout = baseLayoutArrays[validItemsLeft.length];
      }
    });

    const filtered = rowsClone.filter((row) => row !== null);

    const yamlSequence = new YAMLSeq();

    filtered.forEach((row) => {
      const map = new YAMLMap();
      map.set("items", row.items.join(", "));
      map.set("layout", row.layout.join(", "));
      map.set("height", row.height);

      return yamlSequence.add(map);
    });

    selected = new Set();

    fileContent.setIn(["layout", "rows"], yamlSequence);

    updateContents();
  }

  function initializeRow(
    row: number,
    items: {
      name: string | number;
      type: CanvasComponentType;
    }[],
  ) {
    const newComponents: Array<ReturnType<typeof createComponent>> = [];
    const componentsClone = [...components];
    const newRow: (string | number)[] = [];

    let newHeight = MIN_HEIGHT;

    items.forEach((item, i) => {
      newComponents.push(createComponent(item.type));
      newRow.push(canvasItems.length + i);
      newHeight = Math.max(newHeight, getInitalHeight(item.type));
    });

    const newLayout = baseLayoutArrays[newRow.length];

    const newYamlRow = new YAMLMap();
    newYamlRow.set("items", newRow.join(", "));
    newYamlRow.set("layout", newLayout.join(", "));
    newYamlRow.set("height", newHeight);

    const yamlItems = [...yamlRows.items];
    yamlItems.splice(row, 0, newYamlRow);

    fileContent.setIn(["layout", "rows"], yamlItems);
    fileContent.set("items", [...componentsClone, ...newComponents]);

    updateContents();
  }

  function moveToNewRow(items: DragItem[], rowIndex: number) {
    if (!contents) {
      contents = new YAMLMap();
      fileContent.set("layout", contents);
    }
    const rowsClone = structuredClone(rowMaps) as (Row | null)[];
    const newRowItems: (string | number)[] = [];
    const touchedRows = new Set(items.map((el) => el.row));

    let newHeight = MIN_HEIGHT;

    items.forEach((item) => {
      newRowItems.push(item.name);
      const row = rowsClone[item.row];
      if (!row) return;
      row.items[item.order] = null;
      newHeight = Math.max(
        newHeight,
        item?.height ?? getInitalHeight(item.type),
      );
    });

    const newLayout = baseLayoutArrays[newRowItems.length];

    touchedRows.forEach((rowIndex) => {
      const row = rowsClone[rowIndex];
      if (!row) return;
      const validItemsLeft = row.items.filter(itemExists);

      if (!validItemsLeft.length) {
        rowsClone[rowIndex] = null;
      } else {
        row.items = validItemsLeft;
        row.layout = baseLayoutArrays[validItemsLeft.length];
      }
    });

    rowsClone.splice(rowIndex, 0, {
      items: newRowItems,
      height: newHeight,
      layout: newLayout,
    });

    const filtered = rowsClone.filter((row) => row !== null);

    const yamlSequence = new YAMLSeq();

    filtered.forEach((row) => {
      const map = new YAMLMap();
      map.set("items", row.items.join(", "));
      map.set("layout", row.layout.join(", "));
      map.set("height", row.height);

      return yamlSequence.add(map);
    });

    selected = new Set();

    fileContent.setIn(["layout", "rows"], yamlSequence);

    updateContents();
  }

  function itemExists(name: string | number) {
    return name !== undefined && name !== "" && name !== null;
  }

  function removeItems(items: { row: number; order: number }[]) {
    selected = new Set();
    const rowsClone = structuredClone(rowMaps);
    const touchedRows = new Set(items.map((el) => el.row));
    const deletedRows: number[] = [];

    items.forEach((item) => {
      rowsClone[item.row].items[item.order] = null;
    });

    touchedRows.forEach((row) => {
      const filtered = rowsClone[row].items.filter(itemExists);
      if (!filtered.length) {
        deletedRows.push(row);
      } else {
        contents.setIn(["rows", row, "items"], filtered.join(", "));
        contents.setIn(
          ["rows", row, "layout"],
          baseLayoutArrays[filtered.length].join(", "),
        );
      }
    });

    deletedRows.forEach((row) => {
      contents.deleteIn(["rows", row]);
    });

    updateContents();
  }

  function addItems(
    items: {
      type: CanvasComponentType;
      position: { row: number; order: number };
    }[],
  ) {
    const rowsClone = structuredClone(rowMaps);
    const newComponents: Array<ReturnType<typeof createComponent>> = [];
    const componentsClone = [...components];
    const touchedRows = new Set(items.map((el) => el.position.row));

    items.forEach((item, i) => {
      newComponents.push(createComponent(item.type));

      rowsClone[item.position.row].items.splice(
        item.position.order,
        0,
        canvasItems.length + i,
      );
    });

    touchedRows.forEach((rowIndex) => {
      const newRowArray = rowsClone[rowIndex].items;
      const newLayoutArray = baseLayoutArrays[newRowArray.length];

      contents.setIn(["rows", rowIndex, "items"], newRowArray.join(", "));
      contents.setIn(["rows", rowIndex, "layout"], newLayoutArray.join(", "));
    });

    fileContent.set("items", [...componentsClone, ...newComponents]);

    updateContents();
  }

  function updateContents() {
    updateEditorContent(fileContent.toString(), false, true);
  }

  function createComponent(componentType: CanvasComponentType) {
    const defaultMetrics = $metricsViewQuery?.data;
    if (!defaultMetrics) return;

    const newSpec = componentRegistry[componentType].newComponentSpec(
      defaultMetrics.metricsView,
      defaultMetrics.measure,
      defaultMetrics.dimension,
    );

    const { width, height } = componentRegistry[componentType].defaultSize;

    const itemsToPosition =
      spec?.items?.map((item) => ({
        x: item.x ?? 0,
        y: item.y ?? 0,
        width: item.width ?? 0,
        height: item.height ?? 0,
      })) ?? [];

    const [x, y] = findNextAvailablePosition(itemsToPosition, width, height);

    return {
      component: { [componentType]: newSpec },
      height,
      width,
      x,
      y,
    };
  }

  let timeout: ReturnType<typeof setTimeout> | null = null;

  function resetDropZone() {
    hoveredDropZone = null;
  }

  function setDropZone(id: string) {
    hoveredDropZone = id;
  }

  function onDrop(row: number, column: number) {
    if (!passedThreshold) return;
    if (dragItemInfo) {
      dropItemsInExistingRow([dragItemInfo], row, column);
      dragItemInfo = null;
    }
  }
</script>

<svelte:window
  on:mouseup={reset}
  on:mousemove={(e) => {
    mousePosition = { x: e.clientX, y: e.clientY };
  }}
  on:keydown={(e) => {
    console.log(e);
    if (e.key === "Backspace" && selected) {
      e.preventDefault();
      removeItems(
        Array.from(selected).map((id) => {
          const [row, order] = id.split("-").slice(1).map(Number);
          return { row, order };
        }),
      );
    }
  }}
/>

{#if filtersEnabled}
  <header class="bg-background border-b py-4 px-2 w-full">
    <CanvasFilters />
  </header>
{/if}

<div
  class="size-full overflow-hidden overflow-y-auto pb-48 pt-8 px-8 flex flex-col items-center bg-white select-none"
>
  <div
    style:max-width="{maxWidth}px"
    class="w-full h-fit flex flex-col items-center row-container relative"
    bind:clientWidth
  >
    {#each rowMaps as { items, height, layout }, rowIndex (rowIndex)}
      <section
        role="presentation"
        class="size-full grid min-h-fit row relative"
        style:z-index={50 - rowIndex * 2}
        style:--row-height="{height}px"
        style:grid-template-columns={layout.map((el) => `${el}fr`).join(" ")}
      >
        {#each items as itemIndex, columnIndex (columnIndex)}
          {@const id = getId(rowIndex, columnIndex)}
          {@const item = canvasItems[Number(itemIndex)]}

          <div
            class="p-2 relative pointer-events-none size-full container"
            style:min-height="{height}px"
            style:height="{height}px"
          >
            {#if columnIndex === 0}
              <ElementDivider
                left
                activelyResizing={false}
                hoveringOnDropZone={passedThreshold &&
                  hoveredDropZone === `${rowIndex}-${columnIndex}`}
                {rowIndex}
                resizeIndex={-1}
                addIndex={columnIndex}
                rowLength={items.length}
                {spreadEvenly}
                {addItems}
              />
            {/if}

            <ElementDivider
              onMouseDown={onColumResizeStart}
              activelyResizing={resizeInfo?.row === rowIndex &&
                resizeInfo?.column === columnIndex}
              hoveringOnDropZone={passedThreshold &&
                hoveredDropZone === `${rowIndex}-${columnIndex + 1}`}
              columnWidth={layout[columnIndex]}
              {rowIndex}
              resizeIndex={columnIndex}
              addIndex={columnIndex + 1}
              rowLength={items.length}
              {spreadEvenly}
              {addItems}
            />

            <DropZone
              column={columnIndex}
              row={rowIndex}
              allowDrop={!!dragItemInfo && passedThreshold}
              onHover={setDropZone}
              onMouseLeave={resetDropZone}
              {onDrop}
            />

            <article
              role="presentation"
              {id}
              class:selected={selected.has(id)}
              class:opacity-20={dragItemInfo?.row === rowIndex &&
                dragItemInfo.order === columnIndex}
              class:pointer-events-none={resizeInfo}
              class:pointer-events-auto={!resizeInfo}
              class="card w-full cursor-pointer z-0 p-4 h-full relative bg-white overflow-hidden rounded-sm border flex items-center justify-center"
              on:mousedown={(e) => {
                if (e.shiftKey) {
                  selected.add(id);
                  selected = selected;

                  return;
                }
                setSelectedComponentIndex(Number(itemIndex));

                selected = new Set([id]);

                if (itemIndex === null) return;

                handleDragStart({
                  name: itemIndex,
                  row: rowIndex,
                  order: columnIndex,
                  type: "kpi",
                  height: height,
                });
              }}
            >
              {#if item}
                <PreviewElement
                  i={columnIndex}
                  {instanceId}
                  selected={false}
                  component={item}
                />
              {:else}
                <LoadingSpinner size="36px" />

                <!-- <div class="element h-fit min-h-fit">
                    {#each { length: 4 } as _, i (i)}
                      <div
                        class="size-full border-r border-b min-h-48 text-2xl grid place-content-center"
                      >
                        {Math.round(Math.random() * 1000)}
                      </div>
                    {/each}
                  </div> -->
                <!-- {:else}
                <FileWarning size="24px" /> -->
              {/if}
            </article>
          </div>
        {/each}

        <div
          role="presentation"
          class:pointer-events-none={!dragItemInfo}
          class="absolute bottom-0 w-full z-10 left-0 h-20 translate-y-1/2 flex items-center justify-center px-2"
          on:mouseenter={() => {
            rowHover = rowIndex;
          }}
          on:mouseleave={() => {
            timeout = setTimeout(() => (rowHover = null), 150);
          }}
          on:mouseup={() => {
            if (!passedThreshold) return;
            if (dragItemInfo) {
              moveToNewRow([dragItemInfo], rowIndex + 1);

              dragItemInfo = null;
            }
          }}
        >
          <div
            class:!flex={resizeRow === -1 &&
              !dragItemInfo &&
              rowHover === rowIndex}
            class="hidden pointer-events-auto shadow-sm absolute right-2 w-fit z-[50] bg-white translate-x-full border rounded-sm"
          >
            <AddComponentDropdown
              onMouseEnter={() => {
                if (timeout) clearTimeout(timeout);
                rowHover = rowIndex;
              }}
              onItemClick={(type) => {
                rowHover = null;

                initializeRow(rowIndex + 1, [
                  {
                    name: canvasItems.length,
                    type,
                  },
                ]);
              }}
            />
          </div>

          <button
            data-row={rowIndex}
            class:cursor-row-resize={!dragItemInfo}
            class="w-full h-3 group z-50 flex items-center justify-center pointer-events-auto"
            on:mousedown={onRowResizeStart}
          >
            <span
              class:bg-primary-300={rowHover === rowIndex &&
                (!dragItemInfo || passedThreshold)}
              class="w-full h-[3px] group-hover:bg-primary-300"
            />
          </button>
        </div>
      </section>
    {:else}
      <AddComponentDropdown
        componentForm
        onMouseEnter={() => {
          if (timeout) clearTimeout(timeout);
        }}
        onItemClick={(type) => {
          initializeRow(0, [
            {
              name: canvasItems.length,
              type,
            },
          ]);
        }}
      />
    {/each}
  </div>
</div>

<style lang="postcss">
  .card {
    @apply shadow-sm;
  }

  .card:hover {
    @apply shadow-md;
  }

  .container {
    container-type: inline-size;
    container-name: container;
  }

  .selected {
    @apply outline outline-primary-300;
  }

  .row-container {
    container-type: inline-size;
    container-name: row-container;
  }

  .row {
    /* height: var(--row-height); */
    grid-auto-rows: max-content;
  }

  @container row-container (inline-size < 600px) {
    .row {
      grid-template-columns: repeat(1, 1fr) !important;
      /* grid-auto-rows: max-content; */
    }
  }

  .element {
    @apply size-full grid;
    /* container-type: inline-size; */
    grid-template-columns: repeat(4, 1fr);
  }

  @container container (inline-size < 600px) {
    .element {
      grid-template-columns: repeat(2, 1fr);
    }
  }

  @container container (inline-size < 400px) {
    .element {
      grid-template-columns: repeat(1, 1fr);
    }
  }
</style>
