<script lang="ts">
  import { afterNavigate } from "$app/navigation";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import FileWorkspaceHeader from "@rilldata/web-common/features/editor/FileWorkspaceHeader.svelte";
  import { getExtensionsForFile } from "@rilldata/web-common/features/editor/getExtensionsForFile";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { directoryState } from "@rilldata/web-common/features/file-explorer/directory-store";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import { onMount } from "svelte";
  import type { EditorView } from "@codemirror/view";
  import SourceWorkspace from "@rilldata/web-common/features/workspaces/SourceWorkspace.svelte";
  import ModelWorkspace from "@rilldata/web-common/features/workspaces/ModelWorkspace.svelte";
  import MetricsWorkspace from "@rilldata/web-common/features/workspaces/MetricsWorkspace.svelte";
  import ChartWorkspace from "@rilldata/web-common/features/workspaces/ChartWorkspace.svelte";
  import CustomDashboardWorkspace from "@rilldata/web-common/features/workspaces/CustomDashboardWorkspace.svelte";

  const workspaces = new Map([
    [ResourceKind.Source, SourceWorkspace],
    [ResourceKind.Model, ModelWorkspace],
    [ResourceKind.MetricsView, MetricsWorkspace],
    [ResourceKind.Component, ChartWorkspace],
    [ResourceKind.Dashboard, CustomDashboardWorkspace],
    [undefined, null],
  ]);

  export let data;

  let editor: EditorView;

  $: ({ filePath, fileArtifact } = data);
  $: ({ autoSave, hasUnsavedChanges, fileName, name } = fileArtifact);

  $: resourceKind = <ResourceKind | undefined>$name?.kind;

  $: workspace = workspaces.get(resourceKind);

  onMount(() => {
    expandDirectory(filePath);
    // TODO: Focus on the code editor
  });

  afterNavigate(() => {
    expandDirectory(filePath);
    // TODO: Focus on the code editor
  });

  // TODO: move this logic into the DirectoryState
  // TODO: expand all directories in the path, not just the last one
  function expandDirectory(filePath: string) {
    const directory = filePath.split("/").slice(0, -1).join("/");
    directoryState.expand(directory);
  }
</script>

<svelte:head>
  <title>Rill Developer | {fileName}</title>
</svelte:head>

{#if workspace}
  <svelte:component this={workspace} {fileArtifact} />
{:else}
  <WorkspaceContainer inspector={false}>
    <FileWorkspaceHeader
      slot="header"
      {filePath}
      hasUnsavedChanges={$hasUnsavedChanges}
    />
    <WorkspaceEditorContainer slot="body">
      <Editor
        {fileArtifact}
        extensions={getExtensionsForFile(filePath)}
        bind:editor
        bind:autoSave={$autoSave}
      />
    </WorkspaceEditorContainer>
  </WorkspaceContainer>
{/if}
