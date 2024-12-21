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

  let showGrid = true;

  // TODO: Remove later when we move to new tiling system
  const columns = 24;
  const gap = 1;

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

  async function handleComponentDelete(
    e: CustomEvent<{
      index: number;
    }>,
  ) {
    console.log("Canvas handleComponentDelete");
    if (!e.detail.index) return;
    await deleteComponent(e.detail.index);
  }

  async function deleteComponent(index: number) {
    console.log("Canvas deleteComponent");

    const parsedDocument = parseDocument(
      $editorContent ?? $remoteContent ?? "",
    );

    const items = parsedDocument.get("items") as any;
    if (!items) return;
    items.delete(index);
    updateEditorContent(parsedDocument.toString(), true);
    // updateLocalContent(parsedDocument.toString(), true);
    // FIXME: need to rerender gridstack after node removal
    if ($autoSave) await updateComponentFile();
  }

  async function handleComponentUpdate(
    e: CustomEvent<{
      index: number;
      x: number;
      y: number;
      w: number;
      h: number;
    }>,
  ) {
    console.log("handlePreviewUpdate: ", e.detail);

    const parsedDocument = parseDocument(
      $editorContent ?? $remoteContent ?? "",
    );
    const items = parsedDocument.get("items") as any;

    // if (!e.detail.index) {
    //   console.log("No index provided");
    // }

    const node = items.get(e.detail.index);

    // NOTE: V1CanvasItem uses width, height, x, y
    node.set("width", e.detail.w);
    node.set("height", e.detail.h);
    node.set("x", e.detail.x);
    node.set("y", e.detail.y);

    updateEditorContent(parsedDocument.toString(), true);

    if ($autoSave) await updateComponentFile();
  }
</script>

<CanvasDashboardPreview
  {items}
  bind:selectedIndex
  on:update={handleComponentUpdate}
  on:delete={handleComponentDelete}
/>

<svelte:window
  on:keydown={async (e) => {
    if (e.target !== document.body || selectedIndex === null) return;
    if (e.key === "Delete" || e.key === "Backspace") {
      await deleteComponent(selectedIndex);
    }
  }}
/>
