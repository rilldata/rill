<script lang="ts">
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import WorkspaceTableContainer from "@rilldata/web-common/layout/workspace/WorkspaceTableContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import { ConnectedPreviewTable } from "@rilldata/web-common/components/preview-table";
  import { resourceIsLoading } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createRuntimeServiceGetFile } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import SourceEditor from "../editor/SourceEditor.svelte";
  import ErrorPane from "../errors/ErrorPane.svelte";
  import { useIsSourceUnsaved } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let filePath: string;

  const queryClient = useQueryClient();
  const sourceStore = useSourceStore(filePath);

  $: file = createRuntimeServiceGetFile($runtime.instanceId, filePath, {
    query: {
      // this will ensure that any changes done outside our app is pulled in.
      refetchOnWindowFocus: true,
    },
  });

  $: yaml = $file.data?.blob || "";

  $: allErrors = fileArtifactsStore.getAllErrorsForFile(
    queryClient,
    $runtime.instanceId,
    filePath,
  );

  $: sourceQuery = fileArtifactsStore.getResourceForFile(
    queryClient,
    $runtime.instanceId,
    filePath,
  );

  $: isSourceUnsavedQuery = useIsSourceUnsaved(
    $runtime.instanceId,
    filePath,
    $sourceStore.clientYAML,
  );
  $: isSourceUnsaved = $isSourceUnsavedQuery.data;
</script>

<div class="editor-pane h-full overflow-hidden w-full flex flex-col">
  <WorkspaceEditorContainer>
    <SourceEditor {filePath} {yaml} />
  </WorkspaceEditorContainer>

  <WorkspaceTableContainer fade={isSourceUnsaved}>
    {#if !$allErrors?.length}
      <ConnectedPreviewTable
        objectName={$sourceQuery?.data?.source?.state?.table}
        loading={resourceIsLoading($sourceQuery?.data)}
      />
    {:else if $allErrors[0].message}
      <ErrorPane {filePath} errorMessage={$allErrors[0].message} />
    {/if}
  </WorkspaceTableContainer>
</div>
