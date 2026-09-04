<script lang="ts">
  import { goto } from "$app/navigation";
  import EvalRunnerPanel from "@rilldata/web-common/features/ai-evals/runner/EvalRunnerPanel.svelte";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import { getExtensionsForFile } from "@rilldata/web-common/features/editor/getExtensionsForFile";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/actions/ui-actions.ts";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import Inspector from "@rilldata/web-common/layout/workspace/Inspector.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let fileArtifact: FileArtifact;

  const runtimeClient = useRuntimeClient();

  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    fileName,
    remoteContent,
  } = fileArtifact);

  $: resourceQuery = fileArtifact.getResource(queryClient);
  $: resource = $resourceQuery.data;

  $: parseErrorQuery = fileArtifact.getParseError(queryClient);
  $: parseError = $parseErrorQuery;

  async function onChangeCallback(newTitle: string) {
    const newRoute = await handleEntityRename(
      runtimeClient,
      newTitle,
      filePath,
      fileName,
    );
    if (newRoute) await goto(newRoute);
  }
</script>

<WorkspaceContainer>
  <WorkspaceHeader
    {resource}
    hasUnsavedChanges={$hasUnsavedChanges}
    onTitleChange={onChangeCallback}
    slot="header"
    titleInput={fileName}
    {filePath}
    codeToggle={false}
    resourceKind={ResourceKind.AIEval}
  />

  <svelte:fragment slot="body">
    <WorkspaceEditorContainer
      {resource}
      {parseError}
      remoteContent={$remoteContent}
      {filePath}
    >
      <Editor
        {fileArtifact}
        extensions={getExtensionsForFile(filePath)}
        bind:autoSave={$autoSave}
      />
    </WorkspaceEditorContainer>
  </svelte:fragment>

  <svelte:fragment slot="inspector">
    <Inspector {filePath}>
      <EvalRunnerPanel {fileArtifact} />
    </Inspector>
  </svelte:fragment>
</WorkspaceContainer>
