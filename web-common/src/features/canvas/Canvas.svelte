<script lang="ts">
  import CanvasDashboardPreview from "@rilldata/web-common/features/canvas/CanvasDashboardPreview.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { Vector } from "@rilldata/web-common/features/canvas/types";
  import type { V1CanvasSpec } from "@rilldata/web-common/runtime-client";
  import { parseDocument } from "yaml";

  const { canvasStore, validSpecStore, fileArtifact } =
    getCanvasStateManagers();
  $: selectedIndex = $canvasStore?.selectedComponentIndex;

  let spec: V1CanvasSpec = {
    columns: 20, // TODO: to be removed
    gap: 4, // TODO: to be removed
    items: [],
  };
  $: ({
    saveLocalContent: updateComponentFile,
    autoSave,
    updateLocalContent,
    localContent,
    remoteContent,
  } = $fileArtifact);

  $: spec = structuredClone($validSpecStore ?? spec);

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

    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    const items = parsedDocument.get("items") as any;
    if (!items) return;
    items.delete(index);
    updateLocalContent(parsedDocument.toString(), true);
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

    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
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

    updateLocalContent(parsedDocument.toString(), true);

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
