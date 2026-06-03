<script lang="ts">
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact.ts";
  import VirtualCanvasEditor from "@rilldata/web-admin/features/personal-files/canvas/VirtualCanvasEditor.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { getPersonalFilteredResourceByName } from "@rilldata/web-admin/features/personal-files/selectors.ts";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { sessionStorageStore } from "@rilldata/web-common/lib/store-utils/session-storage.ts";
  import { page } from "$app/state";
  import type { VirtualFileIo } from "@rilldata/web-admin/features/personal-files/virtual-file-io.ts";
  import { onMount } from "svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { getQueryServiceResolveCanvasQueryKey } from "@rilldata/web-common/runtime-client";

  let {
    fileArtifact,
    fileIo,
    name,
  }: {
    fileArtifact: FileArtifact;
    fileIo: VirtualFileIo;
    name: string;
  } = $props();

  const runtimeClient = useRuntimeClient();

  let { organization, project } = $derived(page.params);

  let mode = $derived(
    sessionStorageStore(`app:rill:${organization}:${project}:${name}`, "view"),
  );

  let resourceQuery = $derived(
    getPersonalFilteredResourceByName(runtimeClient, name),
  );
  let { data } = $derived($resourceQuery);
  let displayName = $derived(data?.canvas?.spec?.displayName ?? name);

  function toggleMode() {
    mode.set($mode === "edit" ? "view" : "edit");
  }

  onMount(() =>
    fileIo.on("write", (event) => {
      if (event.name === name && event.kind === ResourceKind.Canvas) {
        void queryClient.invalidateQueries({
          queryKey: getQueryServiceResolveCanvasQueryKey(
            runtimeClient.instanceId,
            { canvas: name },
          ),
        });
      }
    }),
  );
</script>

{#if $mode === "edit"}
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
