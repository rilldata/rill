<script lang="ts">
  import { PlusCircleIcon } from "lucide-svelte";
  import {
    ContextMenu,
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuTrigger,
  } from "@rilldata/web-common/components/context-menu";
  import type { CanvasComponentType } from "./components/types";
  import type { ComponentType, SvelteComponent } from "svelte";
  import ChartIcon from "./icons/ChartIcon.svelte";
  import TableIcon from "./icons/TableIcon.svelte";
  import TextIcon from "./icons/TextIcon.svelte";
  import BigNumberIcon from "./icons/BigNumberIcon.svelte";
  import { useDefaultMetrics } from "./selector";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { FileArtifact } from "../entity-management/file-artifact";
  import { getComponentRegistry } from "./components/util";
  import { parseDocument } from "yaml";
  import { findNextAvailablePosition } from "./util";

  export let fileArtifact: FileArtifact;

  $: ({
    saveLocalContent: updateComponentFile,
    editorContent,
    remoteContent,
    updateEditorContent,
  } = fileArtifact);
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

  async function addComponent(componentType: CanvasComponentType) {
    console.log("[CanvasWorkspace] adding component: ", componentType);

    const defaultMetrics = $metricsViewQuery?.data;
    if (!defaultMetrics) return;

    const newSpec = componentRegistry[componentType].newComponentSpec(
      defaultMetrics.metricsView,
      defaultMetrics.measure,
      defaultMetrics.dimension,
    );

    const { width, height } = componentRegistry[componentType].defaultSize;

    const parsedDocument = parseDocument(
      $editorContent ?? $remoteContent ?? "",
    );
    const docJson = parsedDocument.toJSON();
    const existingItems = docJson?.items || [];

    const [x, y] = findNextAvailablePosition(existingItems, width, height);

    const newComponent = {
      component: { [componentType]: newSpec },
      height,
      width,
      x,
      y,
    };

    const updatedItems = [...existingItems, newComponent];

    if (!docJson.items) {
      parsedDocument.set("items", updatedItems);
    } else {
      parsedDocument.set("items", updatedItems);
    }

    const newIndex = existingItems.length;
    updateEditorContent(parsedDocument.toString(), true);
    await updateComponentFile();
    scrollToComponent(newIndex);
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
</script>

<div class="size-full p-2 bg-white">
  <ContextMenu>
    <ContextMenuTrigger>
      <button
        type="button"
        class="flex flex-col items-center gap-2 p-8 rounded-[6px] border border-slate-200 w-full"
      >
        <PlusCircleIcon class="w-6 h-6 text-slate-500" />
        <span class="text-sm font-medium text-slate-500">Add a component</span>
      </button>
    </ContextMenuTrigger>
    <ContextMenuContent>
      {#each menuItems as item}
        <ContextMenuItem
          on:click={() => addComponent(item.id)}
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
</div>
