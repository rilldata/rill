<script lang="ts">
  import { goto } from "$app/navigation";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { EditorView } from "@codemirror/view";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import VisualConnector from "./VisualConnector.svelte";

  export let fileArtifact: FileArtifact;

  let editor: EditorView;

  const runtimeClient = useRuntimeClient();

  $: ({ hasUnsavedChanges, autoSave, path: filePath, fileName } = fileArtifact);

  $: workspace = workspaces.get(filePath);
  $: selectedView = workspace.view;


  $: allErrorsStore = fileArtifact.getAllErrors(queryClient);
  $: allErrors = $allErrorsStore;

  $: resourceQuery = fileArtifact.getResource(queryClient);
  $: resource = $resourceQuery.data;

  async function handleNameChange(newTitle: string) {
    const newRoute = await handleEntityRename(
      runtimeClient,
      newTitle,
      filePath,
      fileName,
    );
    if (newRoute) await goto(newRoute);
  }
</script>

<WorkspaceContainer inspector={false}>
  <WorkspaceHeader
    {filePath}
    {resource}
    resourceKind={ResourceKind.Connector}
    slot="header"
    titleInput={fileName}
    hasUnsavedChanges={$hasUnsavedChanges}
    onTitleChange={handleNameChange}
    showInspectorToggle={false}
    codeToggle
  />

  <svelte:fragment slot="body">
    {#if $selectedView === "viz"}
      {#key fileArtifact}
        <VisualConnector {fileArtifact} />
      {/key}
    {:else}
      <WorkspaceEditorContainer error={allErrors[0]?.message}>
        {#key getNameFromFile(filePath)}
          <Editor {fileArtifact} bind:editor bind:autoSave={$autoSave} />
        {/key}
      </WorkspaceEditorContainer>
    {/if}
  </svelte:fragment>
</WorkspaceContainer>
