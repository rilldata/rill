<script lang="ts">
  import CanvasDashboardPreview from "@rilldata/web-common/features/canvas/CanvasDashboardPreview.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { Vector } from "@rilldata/web-common/features/canvas/types";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import type {
    V1CanvasItem,
    V1CanvasSpec,
  } from "@rilldata/web-common/runtime-client";
  import { parseDocument } from "yaml";

  export let fileArtifact: FileArtifact;

  const { canvasStore, validSpecStore } = getCanvasStateManagers();
  $: selectedIndex = $canvasStore?.selectedComponentIndex;

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
    if (!e.detail.index) return;
    await deleteComponent(e.detail.index);
  }

  async function deleteComponent(index: number) {
    console.log("[Canvas] deleting component: ", index);
    const parsedDocument = parseDocument(
      $editorContent ?? $remoteContent ?? "",
    );

    const items = parsedDocument.get("items") as any;
    if (!items) return;
    items.delete(index);
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
  on:update={handlePreviewUpdate}
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
