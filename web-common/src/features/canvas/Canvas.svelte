<script lang="ts">
  import CanvasDashboardPreview from "./CanvasDashboardPreview.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import type {
    V1CanvasItem,
    V1CanvasSpec,
  } from "@rilldata/web-common/runtime-client";
  import { parseDocument } from "yaml";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import type { Vector } from "./types";
  import BlankCanvas from "./BlankCanvas.svelte";
  import { useDefaultMetrics } from "./selector";
  import { getComponentRegistry } from "./components/util";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { findNextAvailablePosition } from "./util";
  import type { CanvasComponentType } from "./components/types";

  export let fileArtifact: FileArtifact;

  const ctx = getCanvasStateManagers();

  const {
    canvasEntity: {
      selectedComponentIndex: selectedIndex,
      spec: { canvasSpec },
    },
  } = ctx;

  const { canvasEntity } = getCanvasStateManagers();

  $: workspaceLayout = workspaces.get(fileArtifact.path);
  // Open inspector when a canvas item is selected
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

  $: ({ instanceId } = $runtime);

  $: metricsViewQuery = useDefaultMetrics(instanceId);

  const componentRegistry = getComponentRegistry();

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

    // Update document
    const parsedDocument = parseDocument($editorContent ?? "");
    const rawItems = parsedDocument.get("items") as any;

    rawItems.delete(index);

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

    updateEditorContent(parsedDocument.toString(), true);
    items = updatedItems;
    canvasEntity.setSelectedComponentIndex(null);
    await saveLocalContent();
  }

  async function handleDelete(
    e: CustomEvent<{
      index: number;
    }>,
  ) {
    if (e.detail.index === undefined || e.detail.index === null) return;
    await deleteComponent(e.detail.index);
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
    const items = parsedDocument.get("items") as any;

    if (!items) {
      console.warn("[Canvas] No items found in document");
      return;
    }

    e.detail.items.forEach((item, idx) => {
      const node = items.get(idx);
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

  async function addComponent(componentType: CanvasComponentType) {
    const defaultMetrics = $metricsViewQuery?.data;
    if (!defaultMetrics) return;

    const newSpec = componentRegistry[componentType].newComponentSpec(
      defaultMetrics.metricsView,
      defaultMetrics.measure,
      defaultMetrics.dimension,
    );

    const { width, height } = componentRegistry[componentType].defaultSize;

    const parsedDocument = parseDocument($editorContent ?? "");
    const items = parsedDocument.get("items") as any;

    const itemsToPosition =
      spec?.items?.map((item) => ({
        x: item.x ?? 0,
        y: item.y ?? 0,
        width: item.width ?? 0,
        height: item.height ?? 0,
      })) ?? [];

    const [x, y] = findNextAvailablePosition(itemsToPosition, width, height);

    const newComponent = {
      component: { [componentType]: newSpec },
      height,
      width,
      x,
      y,
    };

    if (!items) {
      parsedDocument.set("items", [newComponent]);
    } else {
      items.add(newComponent);
    }
    updateEditorContent(parsedDocument.toString(), true);
    await saveLocalContent();
    canvasEntity.setSelectedComponentIndex(itemsToPosition.length);
    scrollToComponent(itemsToPosition.length);
  }

  function scrollToComponent(index: number) {
    setTimeout(() => {
      const component = document.querySelector(
        `[data-component-index="${index}"]`,
      );
      if (component) {
        component.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }, 100);
  }

  async function handleAdd(e: CustomEvent<{ type: CanvasComponentType }>) {
    await addComponent(e.detail.type);
  }
</script>

{#if items.length > 0}
  <CanvasDashboardPreview
    {items}
    showFilterBar={filtersEnabled}
    selectedIndex={$selectedIndex}
    on:update={handleUpdate}
    on:delete={handleDelete}
  />
{:else}
  <BlankCanvas />
{/if}

<svelte:window
  on:keydown={async (e) => {
    if (e.target !== document.body || $selectedIndex === null) return;
    if (e.key === "Delete" || e.key === "Backspace") {
      console.log("[Canvas] Fired `delete` key");
      await deleteComponent($selectedIndex);
    }
  }}
/>
