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
  import {
    getQueryServiceResolveCanvasQueryKey,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import AlertDialogGuardedConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-guarded-confirmation.svelte";
  import {
    createAdminServiceDeletePersonalFile,
    getAdminServiceListPersonalFilesQueryKey,
  } from "@rilldata/web-admin/client";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { goto } from "$app/navigation";
  import { getCanvasModeStore } from "@rilldata/web-admin/features/personal-files/canvas/mode-utils.ts";

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

  let mode = $derived(getCanvasModeStore(organization, project, name));

  let resourceQuery = $derived(
    getPersonalFilteredResourceByName(runtimeClient, name),
  );
  let { data } = $derived($resourceQuery);
  let displayName = $derived(data?.canvas?.spec?.displayName ?? name);

  let showDeleteConfirmation = $state(false);
  const deleteDashboardMutation = createAdminServiceDeletePersonalFile();
  async function deleteDashboard() {
    await $deleteDashboardMutation.mutateAsync({
      org: organization,
      project,
      name,
    });

    // Invalidate resources and personal files queries
    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(
        runtimeClient.instanceId,
        {},
      ),
    });
    await queryClient.invalidateQueries({
      queryKey: getAdminServiceListPersonalFilesQueryKey(organization, project),
      type: "all",
    });

    eventBus.emit("notification", {
      type: "success",
      message: `Dashboard ${displayName} delete successfully`,
    });
    await goto(`/${organization}/${project}`);
  }

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

<AlertDialogGuardedConfirmation
  bind:open={showDeleteConfirmation}
  title="Delete dashboard?"
  description={`The dashboard "${displayName}" will be permanently deleted. This action cannot be undone.`}
  confirmText={`delete ${displayName}`}
  confirmButtonText="Delete"
  confirmButtonType="destructive"
  loading={$deleteDashboardMutation.isPending}
  error={$deleteDashboardMutation.error?.message}
  onConfirm={deleteDashboard}
>
  <div class="hidden"></div>
</AlertDialogGuardedConfirmation>
