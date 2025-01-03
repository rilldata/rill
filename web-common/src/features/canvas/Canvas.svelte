<script lang="ts">
  import CanvasDashboardPreview from "@rilldata/web-common/features/canvas/CanvasDashboardPreview.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import type {
    V1CanvasItem,
    V1CanvasSpec,
  } from "@rilldata/web-common/runtime-client";
  import { parseDocument } from "yaml";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import type { Vector } from "./types";
  import * as defaults from "./constants";

  export let fileArtifact: FileArtifact;

  const ctx = getCanvasStateManagers();

  const {
    canvasEntity: {
      selectedComponentIndex: selectedIndex,
      spec: { canvasSpec },
    },
  } = ctx;

  let showGrid = true;

  // Open inspector when a canvas item is selected
  $: workspaceLayout = workspaces.get(fileArtifact.path);
  $: if ($selectedIndex !== null && $selectedIndex !== undefined) {
    workspaceLayout.inspector.open();
  }

  let spec: V1CanvasSpec = {
    items: [],
    filtersEnabled: true,
  };

  $: ({ saveLocalContent, updateEditorContent, editorContent } = fileArtifact);

  $: spec = structuredClone($canvasSpec ?? spec);

  $: ({ items = [], filtersEnabled } = spec);

  async function handleDelete(
    e: CustomEvent<{
      index: number;
    }>,
  ) {
    if (e.detail.index === undefined || e.detail.index === null) return;
    await deleteComponent(e.detail.index);
  }

  function redistributeRowColumns(row: { items: V1CanvasItem[] }) {
    const totalItems = row.items.length;
    if (totalItems === 0) return;

    // Total available columns
    const totalColumns = defaults.COLUMN_COUNT; // Should be 12

    // Distribute columns evenly among items
    const baseWidth = Math.floor(totalColumns / totalItems); // Minimum width per item
    const extraColumns = totalColumns % totalItems; // Remainder columns to distribute

    // Sort items by x position to maintain visual order
    const sortedItems = [...row.items].sort((a, b) => (a.x ?? 0) - (b.x ?? 0));

    // Update widths and x positions
    let currentX = 0;
    sortedItems.forEach((item, index) => {
      // Assign base width to each item
      item.width = baseWidth;

      // Distribute any leftover columns to the first few items
      if (index < extraColumns) {
        item.width += 1;
      }

      // Update the x position
      item.x = currentX;

      // Increment currentX for the next item
      currentX += item.width;
    });

    return sortedItems;
  }

  async function deleteComponent(index: number) {
    // Validate input
    if (index === undefined || index === null) {
      console.error("[Canvas] Invalid index for deletion:", index);
      return;
    }

    // Get item to delete
    const itemToDelete = items[index];
    if (!itemToDelete) return;

    // Create updated items array and redistribute row items
    const updatedItems = [...items.slice(0, index), ...items.slice(index + 1)];
    const affectedRowIndex = itemToDelete.y ?? 0;
    const rowItems = updatedItems.filter(
      (item) => (item.y ?? 0) === affectedRowIndex,
    );

    if (rowItems.length > 0) {
      redistributeRowColumns({ items: rowItems });
    }

    // Update document
    const parsedDocument = parseDocument(
      $editorContent ?? $remoteContent ?? "",
    );
    const rawItems = parsedDocument.get("items") as any;

    // Remove deleted item
    rawItems.delete(index);

    // Update positions and dimensions of remaining items
    updatedItems.forEach((item, idx) => {
      const node = rawItems.get(idx);
      if (!node) return;

      const updates = {
        width: item.width,
        height: item.height,
        x: item.x,
        y: item.y,
      };

      Object.entries(updates).forEach(([key, value]) => node.set(key, value));
    });

    // Save changes
    updateEditorContent(parsedDocument.toString(), true);
    items = updatedItems;
    $canvasStore.setSelectedComponentIndex(null);

    if ($autoSave) {
      await updateComponentFile();
    }
  }

  async function handleUpdate(
    e: CustomEvent<{
      index: number;
      position: Vector;
      dimensions: Vector;
      items: V1CanvasItem[];
    }>,
  ) {
    console.log("[Canvas] Handling update:", {
      index: e.detail.index,
      position: e.detail.position,
      dimensions: e.detail.dimensions,
    });

    const parsedDocument = parseDocument($editorContent ?? "");
    const rawItems = parsedDocument.get("items") as any;

    e.detail.items.forEach((item, idx) => {
      const node = rawItems.get(idx);
      if (node) {
        node.set("width", item.width);
        node.set("height", item.height);
        node.set("x", item.x);
        node.set("y", item.y);
      }
    });

    updateEditorContent(parsedDocument.toString(), true);
    await saveLocalContent();
  }
</script>

<CanvasDashboardPreview
  {items}
  {showGrid}
  showFilterBar={filtersEnabled}
  selectedIndex={$selectedIndex}
  on:update={handleUpdate}
  on:delete={handleDelete}
/>

<svelte:window
  on:keydown={async (e) => {
    if (e.target !== document.body || $selectedIndex === null) return;
    if (e.key === "Delete" || e.key === "Backspace") {
      console.log("[Canvas] Fired `delete` key");
      await deleteComponent($selectedIndex);
    }
  }}
/>
