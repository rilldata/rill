<script lang="ts">
  import { afterNavigate } from "$app/navigation";
  import type { EditorView } from "@codemirror/view";
  import { customYAMLwithJSONandSQL } from "@rilldata/web-common/components/editor/presets/yamlWithJsonAndSql";
  import { GeneratingMessage } from "@rilldata/web-common/components/generating-message";
  import { generatingCanvas } from "@rilldata/web-common/features/canvas/ai-generation/generateCanvas";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import FileWorkspaceHeader from "@rilldata/web-common/features/editor/FileWorkspaceHeader.svelte";
  import { getExtensionsForFile } from "@rilldata/web-common/features/editor/getExtensionsForFile";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { directoryState } from "@rilldata/web-common/features/file-explorer/directory-store";
  import CanvasWorkspace from "@rilldata/web-common/features/workspaces/CanvasWorkspace.svelte";
  import ExploreWorkspace from "@rilldata/web-common/features/workspaces/ExploreWorkspace.svelte";
  import MetricsWorkspace from "@rilldata/web-common/features/workspaces/MetricsWorkspace.svelte";
  import ModelWorkspace from "@rilldata/web-common/features/workspaces/ModelWorkspace.svelte";
  import { editorMode } from "@rilldata/web-common/layout/editor-mode-store";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
  import { onMount } from "svelte";
  import type { PageData } from "./$types";

  const VISUAL_KINDS = new Set<ResourceKind>([
    ResourceKind.MetricsView,
    ResourceKind.Explore,
    ResourceKind.Canvas,
  ]);

  const workspaces = new Map([
    [ResourceKind.Source, ModelWorkspace],
    [ResourceKind.Model, ModelWorkspace],
    [ResourceKind.MetricsView, MetricsWorkspace],
    [ResourceKind.Explore, ExploreWorkspace],
    [ResourceKind.Canvas, CanvasWorkspace],
    [null, null],
    [undefined, null],
  ]);

  export let data: PageData;

  let editor: EditorView;

  $: ({ fileArtifact } = data);
  $: ({
    autoSave,
    hasUnsavedChanges,
    fileName,
    resourceName,
    inferredResourceKind,
    path,
    getResource,
    remoteContent,
  } = fileArtifact);

  $: resourceKind = <ResourceKind | undefined>$resourceName?.kind;

  $: effectiveKind = resourceKind ?? $inferredResourceKind;

  $: workspace = workspaces.get(effectiveKind);

  // Auto-promote to code mode when the user navigates to a non-visual-editable
  // file (e.g. a Source or Model) while the global editor mode is "visual".
  $: if (
    $editorMode === "visual" &&
    effectiveKind !== undefined &&
    (effectiveKind === null || !VISUAL_KINDS.has(effectiveKind))
  ) {
    editorMode.set("code");
  }

  $: resourceQuery = getResource(queryClient);

  $: resource = $resourceQuery.data;

  $: extensions =
    resourceKind === ResourceKind.API
      ? [customYAMLwithJSONandSQL]
      : getExtensionsForFile(path);

  $: parseErrorStore = fileArtifact.getParseError(queryClient);
  $: parseError = $parseErrorStore;

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

{#if $generatingCanvas}
  <GeneratingMessage title="Generating your Canvas dashboard..." />
{:else if workspace}
  <svelte:component this={workspace} {fileArtifact} />
{:else}
  <WorkspaceContainer inspector={false}>
    <FileWorkspaceHeader
      slot="header"
      {resource}
      resourceKind={resourceKind ?? $inferredResourceKind ?? undefined}
      filePath={path}
      hasUnsavedChanges={$hasUnsavedChanges}
    />
    <WorkspaceEditorContainer
      slot="body"
      {resource}
      {parseError}
      remoteContent={$remoteContent}
      filePath={path}
    >
      <Editor
        {fileArtifact}
        {extensions}
        bind:editor
        bind:autoSave={$autoSave}
      />
    </WorkspaceEditorContainer>
  </WorkspaceContainer>
{/if}
