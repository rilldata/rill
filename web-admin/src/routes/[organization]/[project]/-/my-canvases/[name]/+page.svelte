<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { V1PersonalVirtualFileType } from "@rilldata/web-admin/client";
  import PersonalCanvasWorkspace from "@rilldata/web-admin/features/personal-canvases/PersonalCanvasWorkspace.svelte";
  import { VirtualFilePersistence } from "@rilldata/web-admin/features/personal-canvases/VirtualFilePersistence";
  import { Button } from "@rilldata/web-common/components/button";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let data: {
    canvasName: string;
    displayName: string;
    yaml: string;
    mode: "view" | "edit";
  };

  const runtimeClient = useRuntimeClient();

  $: ({
    params: { organization, project },
  } = $page);
  $: ({ canvasName, displayName, mode } = data);

  let persistence: VirtualFilePersistence | undefined;

  // Recreate the persistence layer whenever the route changes. The workspace components
  // observe persistence.editorContent / .remoteContent for live YAML sync.
  $: if (organization && project && canvasName) {
    persistence = new VirtualFilePersistence(runtimeClient, {
      org: organization,
      project,
      type: V1PersonalVirtualFileType.PERSONAL_VIRTUAL_FILE_TYPE_CANVAS,
      name: canvasName,
      displayName,
    });
    persistence.editorContent.set(data.yaml);
    persistence.remoteContent.set(data.yaml);
  }

  async function enterEditMode() {
    const url = new URL($page.url);
    url.searchParams.set("mode", "edit");
    await goto(url.toString(), { replaceState: false, keepFocus: true });
  }
</script>

{#if mode === "edit"}
  {#if persistence}
    <PersonalCanvasWorkspace
      {persistence}
      {canvasName}
      initialDisplayName={displayName}
    />
  {/if}
{:else}
  <div class="flex flex-col h-full overflow-hidden">
    <div class="flex items-center justify-between px-4 py-2 border-b">
      <div class="flex items-center gap-2">
        <h1 class="text-lg font-medium">{displayName}</h1>
        <span class="text-xs text-gray-500">
          Personal — only you can see this
        </span>
      </div>
      <Button type="primary" onClick={enterEditMode}>Edit</Button>
    </div>
    {#key `${runtimeClient.instanceId}::${canvasName}`}
      <div class="flex-1 min-h-0">
        <CanvasProvider
          {canvasName}
          instanceId={runtimeClient.instanceId}
          projectId={project}
          showBanner
        >
          <CanvasDashboardEmbed {canvasName} />
        </CanvasProvider>
      </div>
    {/key}
  </div>
{/if}
