<script lang="ts">
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact.ts";
  import VirtualCanvasEditor from "@rilldata/web-admin/features/personal-files/canvas/VirtualCanvasEditor.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { getPersonalFilteredResourceByName } from "@rilldata/web-admin/features/personal-files/selectors.ts";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  let {
    fileArtifact,
    name,
  }: {
    fileArtifact: FileArtifact;
    name: string;
  } = $props();

  const runtimeClient = useRuntimeClient();

  let mode = $state<"edit" | "view">("view");

  let resourceQuery = $derived(
    getPersonalFilteredResourceByName(runtimeClient, name),
  );
  let { data } = $derived($resourceQuery);
  let displayName = $derived(data?.canvas?.spec?.displayName ?? name);

  function toggleMode() {
    mode = mode === "edit" ? "view" : "edit";
  }
</script>

{#if mode === "edit"}
  <VirtualCanvasEditor {fileArtifact} {name} onPreview={toggleMode} />
{:else}
  <div class="flex flex-col h-full overflow-hidden">
    <div class="flex items-center justify-between px-4 py-2 border-b">
      <div class="flex items-center gap-2">
        <h1 class="text-lg font-medium">{displayName}</h1>
        <span class="text-xs text-secondary-foreground">
          Personal — only you can see this
        </span>
      </div>
      <Button type="primary" onClick={toggleMode}>Edit</Button>
    </div>
    {#key `${runtimeClient.instanceId}::${name}`}
      <div class="flex-1 min-h-0">
        <CanvasProvider
          canvasName={name}
          instanceId={runtimeClient.instanceId}
          projectId={runtimeClient.instanceId}
          showBanner
        >
          <CanvasDashboardEmbed canvasName={name} />
        </CanvasProvider>
      </div>
    {/key}
  </div>
{/if}
