<script lang="ts">
  import { parseDocument, Scalar, YAMLMap, YAMLSeq } from "yaml";
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
  import { PlusCircleIcon } from "lucide-svelte";

  const MIN_WIDTH = 3;
  const MINIMUM_MOVEMENT = 12;
  const baseLayoutArrays = [[], [12], [6, 6], [4, 4, 4], [3, 3, 3, 3]];

  const ctx = getCanvasStateManagers();

  const componentRegistry = getComponentRegistry();

  const {
    canvasEntity: {
      selectedComponentIndex: selectedIndex,
      spec: { canvasSpec },
    },
  } = ctx;

  const { canvasEntity } = getCanvasStateManagers();

  let spec: V1CanvasSpec = {
    items: [],
    filtersEnabled: true,
  };

  $: ({ instanceId } = $runtime);

  $: spec = structuredClone($canvasSpec ?? spec);

  $: ({ items: canvasItems = [], filtersEnabled } = spec);

  export let maxWidth: number = 1200;
  export let fileArtifact: FileArtifact;

  let mousePosition = { x: 0, y: 0 };
  let initialMousePosition: { x: number; y: number } | null = null;
  let clientWidth: number;
  let dragIndex = -1;
  let selected: Set<string> = new Set();
  let dragName = "";
  let dragColumn = -1;
  let clone: HTMLElement;
  let offset = { x: 0, y: 0 };
  let resizeRow = -1;
  let initialHeight = 0;
  // {row}-{order}
  let hoveredDropZone: string | null = null;
  let rowHover: number | null = null;
  let resizeInfo: { width: number; row: number; column: number } | null = null;

  $: ({ editorContent, updateEditorContent } = fileArtifact);

  $: fileContent = parseDocument($editorContent ?? "");

  $: metricsViewQuery = useDefaultMetrics(instanceId);

  $: contents = fileContent.get("layout") as YAMLMap;
  $: rawYamlItems = fileContent.get("items") as YAMLSeq;

  $: yamlRows = contents?.get("rows") as YAMLSeq<Scalar>;
  $: yamlLayout = contents?.get("layout") as YAMLSeq<Scalar>;
  $: yamlHeights = contents?.get("heights") as YAMLSeq<Scalar>;

  $: components = rawYamlItems?.items ?? [];
  $: heightStrings = yamlHeights?.items ?? [];
  $: layoutStrings = yamlLayout?.items ?? [];
  $: rowStrings = yamlRows?.items ?? [];

  $: noLayoutComponents =
    heightStrings.length === 0 ||
    layoutStrings.length === 0 ||
    rowStrings.length === 0;

  $: heights = heightStrings.map((el) => Number(el.toString().trim()));
  $: layoutArrays = layoutStrings.map((el) =>
    el
      .toString()
      .split(",")
      .map((el) => Number(el.trim())),
  );
  $: rowArrays = rowStrings.map((el) =>
    el
      .toString()
      .split(",")
      .map((el) => el.trim()),
  );

  $: columnWidth = clientWidth / 12;

  $: mouseDelta = initialMousePosition
    ? calculateMouseDelta(initialMousePosition, mousePosition)
    : 0;
  $: passedThreshold = mouseDelta > MINIMUM_MOVEMENT;

  $: if (resizeRow !== -1 && initialMousePosition) {
    const diff = mousePosition.y - initialMousePosition.y;

    heights[resizeRow] = (diff + initialHeight) / 12;
  }

  function onColumResizeStart(e: MouseEvent & { currentTarget: HTMLElement }) {
    initialMousePosition = mousePosition;
    resizeInfo = {
      width: Number(e.currentTarget.getAttribute("data-width")),
      row: Number(e.currentTarget.getAttribute("data-row")),
      column: Number(e.currentTarget.getAttribute("data-column")),
    };

    window.addEventListener("mousemove", onColumnResize);
    window.addEventListener("mouseup", onColumnResizeEnd);
  }

  function onColumnResize(e: MouseEvent) {
    if (!resizeInfo) return;

    const { row, column, width } = resizeInfo;
    const layoutRow = layoutArrays[row];

    const delta = e.clientX - (initialMousePosition?.x ?? 0);
    const columnDelta = Math.round(delta / columnWidth);

    const maxWidth = 12 - MIN_WIDTH * (layoutRow.length - 1);

    const newValue = clamp(MIN_WIDTH, width + columnDelta, maxWidth);

    const leftOver = 12 - newValue;

    const newRow = layoutRow.map((_, i) => {
      if (i === column) {
        return newValue;
      } else {
        return leftOver / (layoutRow.length - 1);
      }
    });

    layoutArrays[row] = newRow;
  }

  function onColumnResizeEnd() {
    window.removeEventListener("mousemove", onColumnResize);
    window.removeEventListener("mouseup", onColumnResizeEnd);

    contents.set(
      "layout",
      layoutArrays.map((el) => el.join(", ")),
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

  function handleDragStart(metadata: {
    name: string;
    row: number;
    column: number;
  }) {
    dragName = metadata.name;
    dragIndex = metadata.row;
    dragColumn = metadata.column;

    initialMousePosition = mousePosition;

    const id = getId(metadata.row, metadata.column);
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
    dragIndex = -1;

    dragName = "";

    dragColumn = -1;
  }

  function onRowResizeStart(e: MouseEvent & { currentTarget: HTMLElement }) {
    initialMousePosition = mousePosition;
    resizeRow = Number(e.currentTarget.getAttribute("data-row"));
    initialHeight = (heights[resizeRow] ?? 12) * 12;
  }

  function reset() {
    if (resizeRow !== -1) {
      onRowResizeEnd();
    }
  }

  function onRowResizeEnd() {
    heights[resizeRow] = Math.round(heights[resizeRow]);

    initialMousePosition = null;
    resizeRow = -1;
    initialHeight = 0;

    console.log({ contents });

    contents.set("heights", heights);
    updateContents();
  }

  function spreadEvenly(index: number) {
    const layoutClone = structuredClone(layoutArrays);

    layoutClone[index] = baseLayoutArrays[layoutClone[index].length];

    contents.set(
      "layout",
      layoutClone.map((el) => el.join(", ")),
    );
    updateContents();
  }

  function moveItem(
    item: { name: string; row: number; column: number },
    rowIndex: number,
    column: number,
  ) {
    console.log("move item");
    console.log({ rowIndex, column, item });
    // Dropped in the same position
    if (
      rowIndex === item.row &&
      (column === item.column || column === item.column + 1)
    )
      return;

    console.log("move item 2");

    selected = new Set();

    const rowsClone = structuredClone(rowArrays);
    const layoutClone = structuredClone(layoutArrays);
    const heightsClone = structuredClone(heights);

    // Blank out existing item
    rowsClone[item.row][item.column] = "";

    if (rowsClone[item.row].length === 1) {
      heightsClone.splice(item.row, 1);
    }

    // Add item to new position
    rowsClone[rowIndex].splice(column, 0, item.name);

    const touchedRows = new Set([item.row, rowIndex]);

    const filteredRows = rowsClone.map((el) => el.filter(Boolean));

    const newLayout = layoutClone.map((el, i) => {
      if (touchedRows.has(i)) {
        return baseLayoutArrays[filteredRows[i].length];
      }
      return el;
    });

    contents.set("heights", heightsClone);

    contents.set(
      "rows",
      filteredRows.filter((items) => items.length).map((el) => el.join(", ")),
    );

    contents.set(
      "layout",
      newLayout.filter((items) => items.length).map((el) => el.join(", ")),
    );

    updateContents();
  }

  function addRow(
    items: {
      name: string;
      type?: CanvasComponentType;
      position: { row: number; order: number };
    }[],
    newRowIndex?: number,
  ) {
    if (!contents) {
      contents = new YAMLMap();
      fileContent.set("layout", contents);
    }
    const rowsClone = structuredClone(rowArrays);
    const layoutClone = structuredClone(layoutArrays);
    const heightsClone = structuredClone(heights);
    const componentsClone = [...components];

    const newRow = items.map((el) => el.name);
    const newLayout = baseLayoutArrays[newRow.length];

    let added = false;

    const touchedRows = new Set(items.map((el) => el.position.row));

    const originalRowHeights: number[] = [];

    // Blank out moved items
    items.forEach((item) => {
      if (item.position.row !== -1 && item.position.order !== -1) {
        rowsClone[item.position.row][item.position.order] = "";
      } else if (item.type) {
        added = true;
        const newComponent = createComponent(item.type);
        componentsClone.push(newComponent);
      }

      originalRowHeights.push(heightsClone[item.position.row] ?? 12);
      // layoutClone[item.position.row][item.position.order] = "";
    });

    // Add new row
    if (newRowIndex !== undefined) {
      rowsClone.splice(newRowIndex, 0, newRow);
      layoutClone.splice(newRowIndex, 0, newLayout);

      heightsClone.splice(newRowIndex, 0, average(originalRowHeights) || 12);
    } else {
      rowsClone.push(newRow);
      layoutClone.push(newLayout);
      heightsClone.push(average(originalRowHeights) || 12);
    }

    const updatedTouchedIndexes = new Set(
      Array.from(touchedRows).map((originalIndex) =>
        newRowIndex === undefined
          ? originalIndex
          : originalIndex >= newRowIndex
            ? originalIndex + 1
            : originalIndex,
      ),
    );

    // Cleanup
    const filteredRows = rowsClone.map((row) => row.filter(Boolean));

    const finalLayout = layoutClone
      .map((row, index) => {
        if (updatedTouchedIndexes.has(index)) {
          return baseLayoutArrays[filteredRows[index].length];
        }
        return row;
      })
      .filter((el) => el.length);

    const finalRows = filteredRows.filter((row) => row.length);

    selected = new Set();

    contents.set(
      "rows",
      finalRows.map((el) => el.join(", ")),
    );
    contents.set(
      "layout",
      finalLayout.map((el) => el.join(", ")),
    );

    if (added) {
      console.log(componentsClone);
      fileContent.set("items", componentsClone);
    }

    contents.set("heights", heightsClone);

    updateContents();
  }

  function average(arr: number[]) {
    return arr.reduce((acc, el) => acc + el, 0) / arr.length;
  }

  function removeItems(items: { row: number; order: number }[]) {
    selected = new Set();
    const rowsClone = structuredClone(rowArrays);
    const layoutClone = structuredClone(layoutArrays);
    const heightsClone = structuredClone(heights);

    const touchedRows = new Set(items.map((el) => el.row));

    items.forEach((item) => {
      rowsClone[item.row][item.order] = "";
    });

    const newRows = rowsClone.map((el) => el.filter(Boolean));

    const newLayout = layoutClone.map((el, i) => {
      if (touchedRows.has(i)) {
        return baseLayoutArrays[newRows[i].length];
      }
      return el;
    });

    const newHeights = heightsClone.filter((_, i) => {
      if (newRows[i].length) return true;
    });

    contents.set("heights", newHeights);

    contents.set(
      "rows",
      newRows.filter((row) => row.length).map((el) => el.join(", ")),
    );
    contents.set(
      "layout",
      newLayout.filter((row) => row.length).map((el) => el.join(", ")),
    );

    updateContents();
  }

  function addItem(
    rowIndex: number,
    columnIndex: number,
    itemType: CanvasComponentType,
  ) {
    const rowsClone = structuredClone(rowArrays);
    const layoutClone = structuredClone(layoutArrays);
    const componentsClone = [...components];

    const newRowArray = rowsClone[rowIndex];

    // Adding item by referencing its index for now
    newRowArray.splice(columnIndex, 0, canvasItems.length);

    rowsClone[rowIndex] = newRowArray;
    layoutClone[rowIndex] = baseLayoutArrays[newRowArray.length];

    const newComponent = createComponent(itemType);
    fileContent.set("items", [...componentsClone, newComponent]);

    contents.set(
      "rows",
      rowsClone.map((el) => el.join(", ")),
    );
    contents.set(
      "layout",
      layoutClone.map((el) => el.join(", ")),
    );

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

    // const parsedDocument = parseDocument($editorContent ?? "");
    // const items = parsedDocument.get("items") as any;

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
</script>

<svelte:window
  on:mouseup={reset}
  on:mousemove={(e) => {
    mousePosition = { x: e.clientX, y: e.clientY };
  }}
  on:keydown={(e) => {
    if (e.key === "Backspace" && selected) {
      removeItems(
        Array.from(selected).map((id) => {
          const [row, order] = id.split("-").slice(1).map(Number);
          return { row, order };
        }),
      );
    }
  }}
/>

<div
  class="size-full overflow-hidden overflow-y-auto pb-48 pt-8 px-8 bg-white select-none"
>
  {#if noLayoutComponents}
    <AddComponentDropdown
      componentForm
      onMouseEnter={() => {
        if (timeout) clearTimeout(timeout);
      }}
      onItemClick={(type) => {
        addRow(
          [
            {
              name: canvasItems.length,
              type,
              position: { row: -1, order: -1 },
            },
          ],
          0,
        );
      }}
    />
  {/if}

  <div
    class="w-full h-fit flex flex-col row-container max-w-[{maxWidth}px] relative"
    bind:clientWidth
  >
    {#each rowArrays as itemNames, rowIndex (rowIndex)}
      <div
        role="presentation"
        class="size-full grid min-h-fit row relative"
        style:z-index={50 - rowIndex * 2}
        style:--row-height="{(heights[rowIndex] ?? 128) * 12}px"
        style:grid-template-columns={layoutArrays[rowIndex]
          .map((el) => `${el}fr`)
          .join(" ")}
      >
        {#each itemNames as name, columnIndex (columnIndex)}
          {@const id = getId(rowIndex, columnIndex)}
          {@const item = canvasItems[Number(name)]}

          <div
            class="p-2 relative pointer-events-none size-full container"
            style:min-height="{(heights[rowIndex] ?? 128) * 12}px"
            style:height="{(heights[rowIndex] ?? 128) * 12}px"
          >
            <div
              style:height="calc(100% - 16px)"
              class="absolute top-2 right-0 w-full z-20 pointer-events-none flex items-center justify-center"
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
                  rowLength={itemNames.length}
                  {spreadEvenly}
                  {addItem}
                />
              {/if}

              <ElementDivider
                onMouseDown={onColumResizeStart}
                activelyResizing={resizeInfo?.row === rowIndex &&
                  resizeInfo?.column === columnIndex}
                hoveringOnDropZone={passedThreshold &&
                  hoveredDropZone === `${rowIndex}-${columnIndex + 1}`}
                columnWidth={layoutArrays[rowIndex][columnIndex]}
                {rowIndex}
                resizeIndex={columnIndex}
                addIndex={columnIndex + 1}
                rowLength={itemNames.length}
                {spreadEvenly}
                {addItem}
              />

              <div
                class:pointer-events-auto={dragName}
                style:height="calc(100% - 80px)"
                class="w-1/2"
                role="presentation"
                on:mouseenter={() => {
                  hoveredDropZone = `${rowIndex}-${columnIndex}`;
                }}
                on:mouseleave={() => {
                  hoveredDropZone = null;
                }}
                on:mouseup={() => {
                  if (!passedThreshold) return;
                  if (dragIndex > -1) {
                    moveItem(
                      { name: dragName, row: dragIndex, column: dragColumn },
                      rowIndex,
                      columnIndex,
                    );
                    dragName = "";
                  }
                }}
              />

              <div
                class:pointer-events-auto={dragName}
                style:height="calc(100% - 80px)"
                class="h-full w-1/2"
                role="presentation"
                on:mouseenter={() => {
                  hoveredDropZone = `${rowIndex}-${columnIndex + 1}`;
                }}
                on:mouseleave={() => {
                  hoveredDropZone = null;
                }}
                on:mouseup={() => {
                  if (!passedThreshold) return;
                  if (dragIndex > -1) {
                    moveItem(
                      { name: dragName, row: dragIndex, column: dragColumn },
                      rowIndex,
                      columnIndex + 1,
                    );
                    dragName = "";
                  }
                }}
              />
            </div>

            <button
              {id}
              class:selected={selected.has(id)}
              class:opacity-20={dragIndex === rowIndex &&
                dragColumn === columnIndex}
              class:pointer-events-none={resizeInfo}
              class:pointer-events-auto={!resizeInfo}
              class="card w-full z-0 p-4 h-full relative bg-white overflow-hidden rounded-sm border flex items-center justify-center"
              on:mousedown={(e) => {
                if (e.shiftKey) {
                  selected.add(id);
                  selected = selected;
                  return;
                }

                selected = new Set([id]);

                handleDragStart({
                  name,
                  row: rowIndex,
                  column: columnIndex,
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
                {name}
                <!-- <div class="element h-fit min-h-fit">
                    {#each { length: 4 } as _, i (i)}
                      <div
                        class="size-full border-r border-b min-h-48 text-2xl grid place-content-center"
                      >
                        {Math.round(Math.random() * 1000)}
                      </div>
                    {/each}
                  </div> -->
              {/if}
            </button>
          </div>
        {/each}

        <div
          role="presentation"
          class:pointer-events-none={dragIndex === -1}
          class="absolute bottom-0 w-full z-10left-0 h-20 translate-y-1/2 flex items-center justify-center px-2"
          on:mouseenter={() => {
            rowHover = rowIndex;
          }}
          on:mouseleave={() => {
            timeout = setTimeout(() => (rowHover = null), 150);
          }}
          on:mouseup={() => {
            if (!passedThreshold) return;
            if (dragName) {
              addRow(
                [
                  {
                    name: dragName,
                    position: { row: dragIndex, order: dragColumn },
                  },
                ],
                rowIndex + 1,
              );
              dragName = "";
            }
          }}
        >
          <div
            class:!flex={resizeRow === -1 && !dragName && rowHover === rowIndex}
            class="hidden pointer-events-auto shadow-sm absolute right-2 w-fit z-[50] bg-white translate-x-full border rounded-sm"
          >
            <AddComponentDropdown
              onMouseEnter={() => {
                if (timeout) clearTimeout(timeout);
                rowHover = rowIndex;
              }}
              onItemClick={(type) => {
                rowHover = null;

                addRow(
                  [
                    {
                      name: canvasItems.length,
                      type,
                      position: { row: -1, order: -1 },
                    },
                  ],
                  rowIndex + 1,
                );
              }}
            />
          </div>

          <button
            data-row={rowIndex}
            class:cursor-row-resize={dragIndex === -1}
            class="w-full h-3 group z-50 flex items-center justify-center pointer-events-auto"
            on:mousedown={onRowResizeStart}
          >
            <span
              class:bg-primary-300={rowHover === rowIndex &&
                (!dragName || passedThreshold)}
              class="w-full h-[3px] group-hover:bg-primary-300"
            />
          </button>
        </div>
      </div>
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
