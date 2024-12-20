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
    columns: 20,
    gap: 4,
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

  $: ({ items = [], columns = 20, gap = 4, variables = [] } = spec);

  async function handleDeleteEvent(
    e: CustomEvent<{
      index: number;
    }>,
  ) {
    if (!e.detail.index) return;
    await deleteComponent(e.detail.index);
  }

  async function deleteComponent(index: number) {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");

    const items = parsedDocument.get("items") as any;
    if (!items) return;
    items.delete(index);
    updateLocalContent(parsedDocument.toString(), true);
    if ($autoSave) await updateComponentFile();
  }

  async function handlePreviewUpdate(
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

    if (!e.detail.index) return;

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
  {gap}
  {items}
  {columns}
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
