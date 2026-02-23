<script lang="ts">
  import { afterNavigate } from "$app/navigation";
  import type { EditorView } from "@codemirror/view";
  import { customYAMLwithJSONandSQL } from "@rilldata/web-common/components/editor/presets/yamlWithJsonAndSql";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import FileWorkspaceHeader from "@rilldata/web-common/features/editor/FileWorkspaceHeader.svelte";
  import { getExtensionsForFile } from "@rilldata/web-common/features/editor/getExtensionsForFile";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { directoryState } from "@rilldata/web-common/features/file-explorer/directory-store";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import CanvasWorkspace from "@rilldata/web-common/features/workspaces/CanvasWorkspace.svelte";
  import ExploreWorkspace from "@rilldata/web-common/features/workspaces/ExploreWorkspace.svelte";
  import MetricsWorkspace from "@rilldata/web-common/features/workspaces/MetricsWorkspace.svelte";
  import ModelWorkspace from "@rilldata/web-common/features/workspaces/ModelWorkspace.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";

  const workspaces = new Map([
    [ResourceKind.Source, ModelWorkspace],
    [ResourceKind.Model, ModelWorkspace],
    [ResourceKind.MetricsView, MetricsWorkspace],
    [ResourceKind.Explore, ExploreWorkspace],
    [ResourceKind.Canvas, CanvasWorkspace],
    [null, null],
    [undefined, null],
  ]);

  export let fileArtifact: FileArtifact;

  let editor: EditorView;

  $: ({ instanceId } = $runtime);

  $: ({
    autoSave,
    hasUnsavedChanges,
    fileName,
    resourceName,
    inferredResourceKind,
    path,
    getResource,
    getParseError,
  } = fileArtifact);

  $: resourceKind = <ResourceKind | undefined>$resourceName?.kind;

  $: workspace = workspaces.get(resourceKind ?? $inferredResourceKind);

  $: resourceQuery = getResource(queryClient, instanceId);

  $: resource = $resourceQuery.data;

  $: extensions =
    resourceKind === ResourceKind.API
      ? [customYAMLwithJSONandSQL]
      : getExtensionsForFile(path);

  $: parseErrorQuery = getParseError(queryClient, instanceId);
  $: parseError = $parseErrorQuery;

  onMount(() => {
    expandDirectory(path);
  });

  afterNavigate(() => {
    expandDirectory(path);
  });

  function expandDirectory(filePath: string) {
    const directory = filePath.split("/").slice(0, -1).join("/");
    directoryState.expand(directory);
  }
</script>

<svelte:head>
  <title>Rill Developer | {fileName}</title>
</svelte:head>

<div class="flex h-full overflow-hidden">
  <div class="flex-1 overflow-hidden">
    {#if workspace}
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
        <WorkspaceEditorContainer slot="body" error={parseError?.message}>
          <Editor
            {fileArtifact}
            {extensions}
            bind:editor
            bind:autoSave={$autoSave}
          />
        </WorkspaceEditorContainer>
      </WorkspaceContainer>
    {/if}
  </div>
</div>
