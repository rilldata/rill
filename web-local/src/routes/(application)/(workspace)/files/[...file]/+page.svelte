<script lang="ts">
  import { afterNavigate } from "$app/navigation";
  import type { EditorView } from "@codemirror/view";
  import { customYAMLwithJSONandSQL } from "@rilldata/web-common/components/editor/presets/yamlWithJsonAndSql";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import FileWorkspaceHeader from "@rilldata/web-common/features/editor/FileWorkspaceHeader.svelte";
  import { getExtensionsForFile } from "@rilldata/web-common/features/editor/getExtensionsForFile";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { directoryState } from "@rilldata/web-common/features/file-explorer/directory-store";
  import CanvasDashboardWorkspace from "@rilldata/web-common/features/workspaces/CanvasDashboardWorkspace.svelte";
  import ComponentWorkspace from "@rilldata/web-common/features/workspaces/ComponentWorkspace.svelte";
  import ExploreWorkspace from "@rilldata/web-common/features/workspaces/ExploreWorkspace.svelte";
  import MetricsWorkspace from "@rilldata/web-common/features/workspaces/MetricsWorkspace.svelte";
  import ModelWorkspace from "@rilldata/web-common/features/workspaces/ModelWorkspace.svelte";
  import SourceWorkspace from "@rilldata/web-common/features/workspaces/SourceWorkspace.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import { onMount } from "svelte";

  const workspaces = new Map([
    [ResourceKind.Source, SourceWorkspace],
    [ResourceKind.Model, ModelWorkspace],
    [ResourceKind.MetricsView, MetricsWorkspace],
    [ResourceKind.Explore, ExploreWorkspace],
    [ResourceKind.Component, ComponentWorkspace],
    [ResourceKind.Canvas, CanvasDashboardWorkspace],
    [null, null],
    [undefined, null],
  ]);

  export let data;

  let editor: EditorView;

  $: ({ fileArtifact } = data);
  $: ({
    autoSave,
    hasUnsavedChanges,
    fileName,
    resourceName,
    inferredResourceKind,
    path,
  } = fileArtifact);

  $: resourceKind = <ResourceKind | undefined>$resourceName?.kind;

  $: console.log(
    { resourceKind },
    $inferredResourceKind,
    resourceKind,
    $resourceName,
  );

  $: workspace = workspaces.get(resourceKind ?? $inferredResourceKind);

  $: extensions =
    resourceKind === ResourceKind.API
      ? [customYAMLwithJSONandSQL]
      : getExtensionsForFile(path);

  onMount(() => {
    expandDirectory(path);
    // TODO: Focus on the code editor
  });

  afterNavigate(() => {
    expandDirectory(path);
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
      resourceKind={resourceKind ?? $inferredResourceKind ?? undefined}
      filePath={path}
      hasUnsavedChanges={$hasUnsavedChanges}
    />
    <WorkspaceEditorContainer slot="body">
      <Editor
        {fileArtifact}
        {extensions}
        bind:editor
        bind:autoSave={$autoSave}
      />
    </WorkspaceEditorContainer>
  </WorkspaceContainer>
{/if}
