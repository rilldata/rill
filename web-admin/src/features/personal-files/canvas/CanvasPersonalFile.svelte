<script lang="ts">
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import VirtualCanvasEditor from "@rilldata/web-admin/features/personal-files/canvas/VirtualCanvasEditor.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { getPersonalFilteredResourceByName } from "@rilldata/web-admin/features/personal-files/selectors.ts";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { VirtualFileIo } from "@rilldata/web-admin/features/personal-files/virtual-file-io.ts";
  import { onMount } from "svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { getQueryServiceResolveCanvasQueryKey } from "@rilldata/web-common/runtime-client";
  import { goto } from "$app/navigation";
  import { getCanvasModeStore } from "@rilldata/web-admin/features/personal-files/canvas/mode-utils.ts";
  import DeletePersonalCanvasDialog from "@rilldata/web-admin/features/personal-files/canvas/DeletePersonalCanvasDialog.svelte";
  import { personalCanvasesHomeHref } from "@rilldata/web-admin/features/personal-files/personal-file-client.ts";

  let {
    fileArtifact,
    fileIo,
    organization,
    project,
    name,
  }: {
    fileArtifact: FileArtifact;
    fileIo: VirtualFileIo;
    organization: string;
    project: string;
    name: string;
  } = $props();

  const runtimeClient = useRuntimeClient();

  let mode = $derived(getCanvasModeStore(organization, project, name));

  let resourceQuery = $derived(
    getPersonalFilteredResourceByName(runtimeClient, name),
  );
  let { data } = $derived($resourceQuery);
  let displayName = $derived(data?.canvas?.spec?.displayName ?? name);

  let showDeleteConfirmation = $state(false);

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
  <VirtualCanvasEditor
    {fileArtifact}
    {name}
    onPreview={toggleMode}
    onDelete={() => (showDeleteConfirmation = true)}
  />
{:else}
  <div class="flex flex-col h-full overflow-hidden">
    <div class="flex items-center justify-between px-4 py-2 border-b">
      <div class="flex items-center gap-2">
        <h1 class="text-lg font-medium">{displayName}</h1>
        <span class="text-xs text-secondary-foreground">
          {m.personal_files_canvas_visibility()}
        </span>
      </div>
      <Button type="primary" onClick={toggleMode}>{m.canvas_edit()}</Button>
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

<DeletePersonalCanvasDialog
  bind:open={showDeleteConfirmation}
  org={organization}
  {project}
  {name}
  {displayName}
  onDeleted={() => goto(personalCanvasesHomeHref(organization, project))}
/>
