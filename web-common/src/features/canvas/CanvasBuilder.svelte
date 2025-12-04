<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { portal } from "@rilldata/web-common/lib/actions/portal";
  import {
    type V1CanvasRow,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onDestroy } from "svelte";
  import { get, writable } from "svelte/store";
  import { parseDocument } from "yaml";
  import ComponentError from "../components/ComponentError.svelte";
  import type { FileArtifact } from "../entity-management/file-artifact";
  import AddComponentDropdown from "./AddComponentDropdown.svelte";
  import CanvasComponent from "./CanvasComponent.svelte";
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import type { CanvasComponentType } from "./components/types";
  import EditableCanvasRow from "./EditableCanvasRow.svelte";
  import ItemWrapper from "./ItemWrapper.svelte";
  import type { Transaction, YAMLRow } from "./layout-util";
  import {
    COLUMN_COUNT,
    DEFAULT_DASHBOARD_WIDTH,
    generateNewAssets,
    mapGuard,
    mousePosition,
    rowsGuard,
  } from "./layout-util";
  import RowWrapper from "./RowWrapper.svelte";
  import { useDefaultMetrics } from "./selector";
  import { getCanvasStore } from "./state-managers/state-managers";
  import { activeDivider, dropZone } from "./stores/ui-stores";
  import ReconcilingSpinner from "../entity-management/ReconcilingSpinner.svelte";

  const activelyEditing = writable(false);

  export let fileArtifact: FileArtifact;
  export let canvasName: string;
  export let openSidebar: () => void;

  let initialMousePosition: { x: number; y: number } | null = null;
  let clientWidth: number;
  let offset = { x: 0, y: 0 };
  let dragComponent: BaseCanvasComponent | null = null;
  let timeout: ReturnType<typeof setTimeout> | null = null;
  let dragTimeout: ReturnType<typeof setTimeout> | null = null;
  let dragItemPosition = { top: 0, left: 0 };
  let dragItemDimensions = { width: 0, height: 0 };
  let openSidebarAfterSelection = false;
  let pendingComponentDelete: string | undefined = undefined;

  $: ({
    canvasEntity: {
      setSelectedComponent,
      selectedComponent,
      components,
      processRows,
      specStore,
      unsubscribe,
      _rows,
      firstLoad,
    },
  } = getCanvasStore(canvasName, instanceId));

  $: layoutRows = $_rows;

  $: ({ instanceId } = $runtime);

  $: canvasData = $specStore.data;
  $: metricsViews = Object.entries(canvasData?.metricsViews ?? {});

  $: metricsViewQuery = useDefaultMetrics(instanceId, metricsViews?.[0]?.[0]);

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

  $: rawYamlRows = rowsGuard(contents?.get("rows"));
  $: yamlCanvasRows = mapGuard(rawYamlRows);

  $: columnWidth = clientWidth / 12;

  $: mouseDelta = initialMousePosition
    ? calculateMouseDelta(initialMousePosition, $mousePosition)
    : 0;

  $: dropZone.setMouseDelta(mouseDelta);

  $: defaultMetrics = $metricsViewQuery?.data;

  $: activelyDragging = !!dragComponent;

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
    return `${canvasName}--component-${row ?? 0}-${column ?? 0}`;
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
  }

  function reset() {
    if (dragTimeout) {
      clearTimeout(dragTimeout);
    }

    if (dragComponent) {
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
      if (!specRow?.items?.[i] || !yamlRow.items[i]) return;
      specRow.items[i].width = baseSize;
      yamlRow.items[i].width = baseSize;
    });

    updateAssets(specRowsClone, yamlRowsClone);
  }

  function updateAssets(
    specRows: V1CanvasRow[],
    yamlRows: YAMLRow[],
    resolvedComponents?: Record<string, V1Resource>,
  ) {
    specCanvasRows = specRows;

    if (resolvedComponents) {
      processRows({
        components: resolvedComponents,

        canvas: {
          rows: specRows,
        },
      });
    }

    try {
      contents.setIn(["rows"], yamlRows);
    } catch (e) {
      console.error(e);
    }

    updateContents();
  }

  function performTransaction(transaction: Transaction, selectedId?: string) {
    if (!defaultMetrics) return;

    try {
      const { newSpecRows, newYamlRows, newResolvedComponents, mover } =
        generateNewAssets({
          yamlRows: yamlCanvasRows,
          specRows: specCanvasRows,
          canvasName,
          defaultMetrics,
          resolvedComponents,
          transaction,
        });

      if (selectedId) {
        const idsAtPositions = specCanvasRows.map((row) => {
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
          const newIdOfSelected = getId(rowIndex, colIndex);

          if (selectedId && newIdOfSelected)
            setSelectedComponent(newIdOfSelected);
        }
      }

      updateAssets(newSpecRows, newYamlRows, newResolvedComponents);
    } catch {
      // no-op
    }
  }

  function addItems(
    position: { row: number; column: number },
    items: CanvasComponentType[],
  ) {
    performTransaction({
      operations: items.map((type, i) => ({
        type: "add",
        componentType: type,
        destination: {
          row: position.row,
          col: position.column + i,
        },
      })),
    });

    const id = getId(position.row, position.column);

    setSelectedComponent(id);
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

    const id = getId(row, 0);

    performTransaction({
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
    });

    setSelectedComponent(id);
  }

  function onDrop(row: number, column: number | null) {
    if (!$dropZone) return;
    dropZone.clear();

    if (dragComponent) {
      const [, fromRow, , fromCol] = dragComponent.pathInYAML;
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
      );
    }
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
    const [, row, , col] = component.pathInYAML;

    if (component.id === $selectedComponent) {
      resetSelection();
    }

    performTransaction({
      operations: [
        {
          type: "delete",
          target: {
            row,
            col,
          },
        },
      ],
    });
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
  on:mouseup={reset}
  on:keydown={(e) => {
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
  bind:clientWidth
>
  {#each layoutRows as row, rowIndex (rowIndex)}
    <EditableCanvasRow
      {row}
      {maxWidth}
      {rowIndex}
      {components}
      {columnWidth}
      {dragComponent}
      {selectedComponent}
      zIndex={layoutRows.length - rowIndex}
      {onDrop}
      {addItems}
      {spreadEvenly}
      {initializeRow}
      {updateRowHeight}
      {updateComponentWidths}
      onDelete={({ component }) => {
        deleteComponent(component);
      }}
      onDuplicate={({ columnIndex }) => {
        if (!defaultMetrics) return;

        performTransaction({
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
        });
      }}
      onComponentMouseDown={({ event, id }) => {
        if (event.button !== 0) return;
        const component = components.get(id);
        if (!component) return;

        initialMousePosition = $mousePosition;

        setSelectedComponent(id);

        if (dragTimeout) clearTimeout(dragTimeout);

        openSidebarAfterSelection = true;

        dragTimeout = setTimeout(() => {
          openSidebarAfterSelection = false;
          handleDragStart(component);
        }, 150);
      }}
    />
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
    onItemClick={(type) => {
      initializeRow(specCanvasRows.length, type);
      setTimeout(() => scrollToBottom(), 500);
    }}
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
        Are you sure you want to delete this widget? This action cannot be
        undone.
      </AlertDialog.Description>

      <AlertDialog.Footer>
        <AlertDialog.Cancel asChild let:builder>
          <Button
            large
            builders={[builder]}
            type="secondary"
            onClick={() => {
              pendingComponentDelete = undefined;
            }}
          >
            Cancel
          </Button>
        </AlertDialog.Cancel>

        <AlertDialog.Action asChild let:builder>
          <Button
            large
            builders={[builder]}
            type="primary"
            danger
            onClick={() => {
              if (!pendingComponentDelete) return;
              const component = components.get(pendingComponentDelete);
              if (!component) return;
              deleteComponent(component);
              pendingComponentDelete = undefined;
            }}
          >
            Delete
          </Button>
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
