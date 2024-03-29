<script lang="ts">
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import WorkspaceTableContainer from "@rilldata/web-common/layout/workspace/WorkspaceTableContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import { ConnectedPreviewTable } from "@rilldata/web-common/components/preview-table";
  import { resourceIsLoading } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createRuntimeServiceGetFile } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import { EntityType } from "../../entity-management/types";
  import SourceEditor from "../editor/SourceEditor.svelte";
  import ErrorPane from "../errors/ErrorPane.svelte";
  import { useIsSourceUnsaved, useSource } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;

  const queryClient = useQueryClient();
  $: sourceStore = useSourceStore(sourceName);

  $: filePath = getFilePathFromNameAndType(sourceName, EntityType.Table);
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  $: file = createRuntimeServiceGetFile($runtime.instanceId, filePath, {
    query: {
      // this will ensure that any changes done outside our app is pulled in.
      refetchOnWindowFocus: true,
    },
  });

  $: yaml = $file.data?.blob || "";

  $: allErrors = fileArtifact.getAllErrors(queryClient, $runtime.instanceId);

  $: sourceQuery = useSource($runtime.instanceId, sourceName);

  $: isSourceUnsavedQuery = useIsSourceUnsaved(
    $runtime.instanceId,
    sourceName,
    $sourceStore.clientYAML,
  );
  $: isSourceUnsaved = $isSourceUnsavedQuery.data;
</script>

<div class="editor-pane h-full overflow-hidden w-full flex flex-col">
  <WorkspaceEditorContainer>
    <SourceEditor {sourceName} {yaml} />
  </WorkspaceEditorContainer>

  <WorkspaceTableContainer fade={isSourceUnsaved}>
    {#if !$allErrors?.length}
      <ConnectedPreviewTable
        objectName={$sourceQuery?.data?.source?.state?.table}
        loading={resourceIsLoading($sourceQuery?.data)}
      />
    {:else if $allErrors[0].message}
      <ErrorPane {sourceName} errorMessage={$allErrors[0].message} />
    {/if}
  </WorkspaceTableContainer>
</div>
