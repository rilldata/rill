<script lang="ts">
  import CanvasDashboardPreview from "./CanvasDashboardPreview.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import type { V1CanvasSpec } from "@rilldata/web-common/runtime-client";
  import { parseDocument } from "yaml";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import { useDefaultMetrics } from "./selector";
  import { getComponentRegistry } from "./components/util";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { findNextAvailablePosition } from "./util";
  import type { CanvasComponentType } from "./components/types";
  import BlankCanvas from "./BlankCanvas.svelte";
  import CanvasFilters from "@rilldata/web-common/features/canvas/filters/CanvasFilters.svelte";
  import {
    ContextMenu,
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuTrigger,
  } from "@rilldata/web-common/components/context-menu";
  import ChartIcon from "./icons/ChartIcon.svelte";
  import TableIcon from "./icons/TableIcon.svelte";
  import TextIcon from "./icons/TextIcon.svelte";
  import BigNumberIcon from "./icons/BigNumberIcon.svelte";

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

  const menuItems: {
    id: CanvasComponentType;
    label: string;
    icon: ComponentType<SvelteComponent>;
  }[] = [
    { id: "bar_chart", label: "Chart", icon: ChartIcon },
    { id: "table", label: "Table", icon: TableIcon },
    { id: "markdown", label: "Text", icon: TextIcon },
    { id: "kpi", label: "KPI", icon: BigNumberIcon },
    { id: "image", label: "Image", icon: ChartIcon },
  ];

  async function deleteComponent(index: number) {
    if (index === undefined || index === null) {
      console.error("[Canvas] Invalid index for deletion:", index);
      return;
    }

    const itemToDelete = items[index];
    if (!itemToDelete) return;

    const updatedItems = [...items.slice(0, index), ...items.slice(index + 1)];

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

  function handleUpdate(
    e: CustomEvent<{
      index: number;
      x: number;
      y: number;
      w: number;
      h: number;
    }>,
  ) {
    const parsedDocument = parseDocument($editorContent ?? "");
    const items = parsedDocument.get("items") as any;

    if (!items) {
      console.warn("[Canvas] No items found in document");
      return;
    }

    const node = items.get(e.detail.index);
    if (node) {
      node.set("width", e.detail.w);
      node.set("height", e.detail.h);
      node.set("x", e.detail.x);
      node.set("y", e.detail.y);
    }

    updateEditorContent(parsedDocument.toString(), false, true);
  }

  function addComponent(componentType: CanvasComponentType) {
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

    updateEditorContent(parsedDocument.toString(), false, true);
    canvasEntity.setSelectedComponentIndex(itemsToPosition.length);
    scrollToComponent(itemsToPosition.length);
  }

  function scrollToComponent(index: number) {
    setTimeout(() => {
      const component = document.querySelector(`[data-index="${index}"]`);
      if (component) {
        component.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }, 100);
  }

  async function handleAdd(e: CustomEvent<{ type: CanvasComponentType }>) {
    await addComponent(e.detail.type);
  }

  function handleContextMenu(
    event: CustomEvent<{ originalEvent: MouseEvent }>,
  ) {
    const target = event.detail.originalEvent.target as HTMLElement;
    const gridStackEl = target.closest(".grid-stack-item");

    // Prevent context menu if clicking on a grid item
    if (gridStackEl) {
      event.detail.originalEvent.preventDefault();
      event.detail.originalEvent.stopPropagation();
      event.preventDefault();
      event.stopPropagation();
      return false;
    }
  }
</script>

{#if filtersEnabled}
  <div
    id="header"
    class="border-b w-fit min-w-full flex flex-col bg-slate-50 slide"
  >
    <CanvasFilters />
  </div>
{/if}

<ContextMenu>
  <ContextMenuTrigger
    class="h-full w-full block"
    on:contextmenu={handleContextMenu}
  >
    {#if items.length === 0}
      <BlankCanvas on:add={handleAdd} />
    {:else}
      <CanvasDashboardPreview
        {items}
        {spec}
        activeIndex={$selectedIndex}
        on:update={handleUpdate}
        on:delete={handleDelete}
      />
    {/if}
  </ContextMenuTrigger>
  <ContextMenuContent>
    {#each menuItems as item}
      <ContextMenuItem
        on:click={() =>
          handleAdd(new CustomEvent("add", { detail: { type: item.id } }))}
        class="text-gray-700 text-xs"
      >
        <div class="flex flex-row gap-x-2">
          <svelte:component this={item.icon} />
          <span class="text-gray-700 text-xs font-normal">{item.label}</span>
        </div>
      </ContextMenuItem>
    {/each}
  </ContextMenuContent>
</ContextMenu>

<svelte:window
  on:keydown={async (e) => {
    if (e.target !== document.body || $selectedIndex === null) return;
    if (e.key === "Delete" || e.key === "Backspace") {
      await deleteComponent($selectedIndex);
    }
  }}
/>
