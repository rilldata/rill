<script lang="ts">
  import CanvasDashboardPreview from "@rilldata/web-common/features/canvas/CanvasDashboardPreview.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { Vector } from "@rilldata/web-common/features/canvas/types";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import type { V1CanvasSpec } from "@rilldata/web-common/runtime-client";
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

  async function handleDeleteEvent(
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

  async function handlePreviewUpdate(
    e: CustomEvent<{
      index: number;
      position: Vector;
      dimensions: Vector;
    }>,
  ) {
    const parsedDocument = parseDocument(
      $editorContent ?? $remoteContent ?? "",
    );
    const items = parsedDocument.get("items") as any;

    const node = items.get(e.detail.index);

    node.set("width", e.detail.dimensions[0]);
    node.set("height", e.detail.dimensions[1]);
    node.set("x", e.detail.position[0]);
    node.set("y", e.detail.position[1]);

    updateEditorContent(parsedDocument.toString(), true);

    if ($autoSave) await updateComponentFile();
  }
</script>

<CanvasDashboardPreview
  {items}
  bind:selectedIndex
  on:update={handlePreviewUpdate}
  on:delete={handleDeleteEvent}
/>

<svelte:window
  on:keydown={async (e) => {
    if (e.target !== document.body || selectedIndex === null) return;
    if (e.key === "Delete" || e.key === "Backspace") {
      await deleteComponent(selectedIndex);
    }
  }}
/>
