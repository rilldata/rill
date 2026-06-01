<script lang="ts">
  import CanvasInitialization from "@rilldata/web-common/features/canvas/CanvasInitialization.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import { WorkspaceContainer } from "@rilldata/web-common/layout/workspace";
  import ReconcileWarningPanel from "@rilldata/web-common/features/entity-management/ReconcileWarningPanel.svelte";
  import VisualCanvasEditing from "@rilldata/web-common/features/canvas/inspector/VisualCanvasEditing.svelte";
  import SaveDefaultsButton from "@rilldata/web-common/features/canvas/components/SaveDefaultsButton.svelte";
  import CanvasLoadingState from "@rilldata/web-common/features/canvas/CanvasLoadingState.svelte";
  import CanvasBuilder from "@rilldata/web-common/features/canvas/CanvasBuilder.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";
  import { createRootCauseErrorQuery } from "@rilldata/web-common/features/entity-management/error-utils.ts";

  export let fileArtifact: FileArtifact;

  const runtimeClient = useRuntimeClient();

  let canvasName: string;

  $: ({
    autoSave,
    path: filePath,
    getResource,
    remoteContent,
    saveState: { saving },
  } = fileArtifact);

  $: resourceQuery = getResource(queryClient);

  $: ({ data } = $resourceQuery);

  $: canvasName = getNameFromFile(filePath);

  // Parse error for the editor gutter and banner
  $: parseErrorQuery = fileArtifact.getParseError(queryClient);
  $: parseError = $parseErrorQuery;

  $: reconcileError = data?.meta?.reconcileError;
  $: rootCauseQuery = createRootCauseErrorQuery(
    runtimeClient,
    data,
    reconcileError,
  );
  $: rootCauseReconcileError = reconcileError
    ? ($rootCauseQuery?.data ?? reconcileError)
    : undefined;
</script>

<CanvasInitialization
  {canvasName}
  instanceId={runtimeClient.instanceId}
  allowUnvalidatedSpec={true}
  let:ready
  let:isReconciling
  let:isLoading
>
  <WorkspaceContainer>
    <div class="flex justify-between" slot="header">
      {#if ready}
        <SaveDefaultsButton
          {canvasName}
          instanceId={runtimeClient.instanceId}
          saving={$saving}
        />
      {/if}
    </div>

    <svelte:fragment slot="body">
      <div class="flex flex-col h-full">
        <div class="flex-1 min-h-0">
          <WorkspaceEditorContainer
            resource={data}
            {parseError}
            remoteContent={$remoteContent}
            {filePath}
            showError={ready}
          >
            <CanvasLoadingState
              {ready}
              {isReconciling}
              {isLoading}
              errorMessage={rootCauseReconcileError}
              {filePath}
            >
              <CanvasBuilder
                {canvasName}
                {fileArtifact}
                openSidebar={() => {}}
              />
            </CanvasLoadingState>
          </WorkspaceEditorContainer>
        </div>
        <ReconcileWarningPanel {fileArtifact} />
      </div>
    </svelte:fragment>
    <svelte:fragment slot="inspector">
      {#if ready}
        <VisualCanvasEditing {canvasName} {fileArtifact} autoSave={$autoSave} />
      {/if}
    </svelte:fragment>
  </WorkspaceContainer>
</CanvasInitialization>
