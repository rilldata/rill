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
  import { groupItemsByRow } from "./util";
  import type { Vector } from "./types";
  import { convertToGridItems, sortItemsByPosition, compactGrid } from "./util";

  export let fileArtifact: FileArtifact;

  const { canvasStore, validSpecStore } = getCanvasStateManagers();
  $: selectedIndex = $canvasStore?.selectedComponentIndex;

  // Open inspector when a canvas item is selected
  $: workspaceLayout = workspaces.get(fileArtifact.path);
  $: if (selectedIndex !== null && selectedIndex !== undefined) {
    workspaceLayout.inspector.open();
  }

  let spec: V1CanvasSpec = {
    items: [],
  };

  $: ({
    saveLocalContent: updateComponentFile,
    autoSave,
    updateEditorContent,
    editorContent,
    remoteContent,
  } = fileArtifact);

  $: spec = structuredClone($validSpecStore?.data ?? spec);

  $: ({ items = [] } = spec);

  async function handleDelete(
    e: CustomEvent<{
      index: number;
    }>,
  ) {
    if (e.detail.index === undefined || e.detail.index === null) return;
    await deleteComponent(e.detail.index);
  }

  async function deleteComponent(index: number) {
    console.log("[Canvas] deleteComponent: ", index);
    const parsedDocument = parseDocument(
      $editorContent ?? $remoteContent ?? "",
    );

    const docItems = parsedDocument.get("items") as any;
    if (!docItems) return;

    // Remove the item
    docItems.delete(index);

    // Process remaining items
    const remainingItems = convertToGridItems(docItems.items);
    const sortedItems = sortItemsByPosition(remainingItems);
    compactGrid(sortedItems);

    // Save changes
    updateEditorContent(parsedDocument.toString(), true);
    if ($autoSave) await updateComponentFile();
  }

  async function handleUpdate(event: CustomEvent) {
    const { index, position, dimensions, items } = event.detail;
    console.log("[Canvas] Handling update:", {
      index,
      position,
      dimensions,
      items,
    });

    const parsedDocument = parseDocument(
      $editorContent ?? $remoteContent ?? "",
    );
    const docItems = parsedDocument.get("items") as any;

    if (!docItems) return;

    const node = docItems.get(index);
    if (!node) return;

    node.set("x", position[0]);
    node.set("y", position[1]);
    node.set("width", dimensions[0]);
    node.set("height", dimensions[1]);

    updateEditorContent(parsedDocument.toString(), true);
    if ($autoSave) await updateComponentFile();
  }
</script>

<CanvasDashboardPreview
  {items}
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
