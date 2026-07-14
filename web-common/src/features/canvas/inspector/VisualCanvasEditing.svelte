<script lang="ts">
  import { replaceState } from "$app/navigation";
  import ComponentsEditor from "@rilldata/web-common/features/canvas/inspector/ComponentsEditor.svelte";
  import PageEditor from "@rilldata/web-common/features/canvas/inspector/PageEditor.svelte";
  import TabGroupEditor from "@rilldata/web-common/features/canvas/inspector/TabGroupEditor.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { Inspector } from "@rilldata/web-common/layout/workspace";
  import { parseDocument } from "yaml";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let fileArtifact: FileArtifact;
  export let autoSave: boolean;
  export let canvasName: string;

  const runtimeClient = useRuntimeClient();

  $: ({ instanceId } = runtimeClient);

  $: ({
    canvasEntity: {
      selectedComponent,
      componentsStore,
      selectedTabGroup,
      setSelectedTabGroup,
      layout,
    },
  } = getCanvasStore(canvasName, instanceId));

  // Resolve the selected tab group to its layout block (for the active group + block index).
  $: tabGroupBlock =
    $selectedTabGroup != null
      ? $layout.find(
          (block) =>
            block.kind === "tab-group" &&
            block.group.name === $selectedTabGroup,
        )
      : undefined;

  $: ({ editorContent, updateEditorContent, saveLocalContent, path } =
    fileArtifact);

  $: parsedDocument = parseDocument($editorContent ?? "");

  $: components = $componentsStore;
  $: component = components.get($selectedComponent ?? "");

  $: isCustomChartComponent = component?.type === "custom_chart";

  async function updateProperties(
    newRecord: Record<string, unknown>,
    removeProperties?: Array<string | string[]>,
  ) {
    Object.entries(newRecord).forEach(([property, value]) => {
      if (!value) {
        parsedDocument.delete(property);
      } else {
        parsedDocument.set(property, value);
      }
    });

    if (removeProperties) {
      removeProperties.forEach((prop) => {
        try {
          if (Array.isArray(prop)) {
            parsedDocument.deleteIn(prop);
          } else {
            parsedDocument.delete(prop);
          }
        } catch (e) {
          console.warn(e);
        }
      });
    }

    killState();
    updateEditorContent(parsedDocument.toString(), false, autoSave);
    await saveLocalContent();
  }

  function killState() {
    replaceState(window.location.origin + window.location.pathname, {});
  }
</script>

<Inspector
  minWidth={320}
  filePath={path}
  maxWidth={isCustomChartComponent ? 600 : 420}
>
  {#if component}
    <ComponentsEditor {component} />
  {:else if tabGroupBlock && tabGroupBlock.kind === "tab-group"}
    <TabGroupEditor
      group={tabGroupBlock.group}
      blockIndex={tabGroupBlock.rowIndex}
      {fileArtifact}
      {autoSave}
      onClose={() => setSelectedTabGroup(null)}
      onRename={(name) => setSelectedTabGroup(name)}
    />
  {:else}
    <PageEditor {canvasName} {fileArtifact} {updateProperties} />
  {/if}
</Inspector>
