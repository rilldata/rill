<script lang="ts">
  import { afterNavigate } from "$app/navigation";
  import type { EditorView } from "@codemirror/view";
  import { customYAMLwithJSONandSQL } from "@rilldata/web-common/components/editor/presets/yamlWithJsonAndSql";
  import { GeneratingMessage } from "@rilldata/web-common/components/generating-message";
  import { generatingCanvasFilePath } from "@rilldata/web-common/features/canvas/ai-generation/generateCanvas";
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
  import ThemeWorkspace from "@rilldata/web-common/features/workspaces/ThemeWorkspace.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
  import { onMount } from "svelte";
  import { getReadonlyNotice } from "@rilldata/web-common/features/entity-management/actions/protected-files.ts";

  const workspaces = new Map([
    [ResourceKind.Source, ModelWorkspace],
    [ResourceKind.Model, ModelWorkspace],
    [ResourceKind.MetricsView, MetricsWorkspace],
    [ResourceKind.Explore, ExploreWorkspace],
    [ResourceKind.Canvas, CanvasWorkspace],
    [ResourceKind.Theme, ThemeWorkspace],
    [null, null],
    [undefined, null],
  ]);

  let { fileArtifact }: { fileArtifact: FileArtifact } = $props();

  // Needed to get the correct type
  let editor: EditorView | null = $state(null);

  let {
    autoSave,
    hasUnsavedChanges,
    fileName,
    resourceName,
    inferredResourceKind,
    path,
    managed,
    getResource,
    getParseError,
    remoteContent,
  } = $derived(fileArtifact);
  let notice = $derived(getReadonlyNotice(path));

  let resourceKind = $derived($resourceName?.kind as ResourceKind | undefined);

  let WorkspaceComponent = $derived(
    workspaces.get(resourceKind ?? $inferredResourceKind),
  );

  let resourceQuery = $derived(getResource(queryClient));

  let resource = $derived($resourceQuery.data);
  let extensions = $derived(
    resourceKind === ResourceKind.API
      ? [customYAMLwithJSONandSQL]
      : getExtensionsForFile(path),
  );

  let parseErrorQuery = $derived(getParseError(queryClient));
  let parseError = $derived($parseErrorQuery);

  let isGeneratingThisFile = $derived($generatingCanvasFilePath === path);

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
    {#if isGeneratingThisFile}
      <GeneratingMessage title="Generating your Canvas dashboard..." />
    {:else if WorkspaceComponent}
      <WorkspaceComponent {fileArtifact} />
    {:else}
      <WorkspaceContainer inspector={false}>
        <FileWorkspaceHeader
          slot="header"
          {fileArtifact}
          {resource}
          resourceKind={resourceKind ?? $inferredResourceKind ?? undefined}
          hasUnsavedChanges={$hasUnsavedChanges}
        />
        <svelte:fragment slot="body">
          {#if managed && notice}
            <div class="flex flex-col size-full items-center justify-center">
              {@render notice()}
            </div>
          {:else}
            <WorkspaceEditorContainer
              {resource}
              {parseError}
              remoteContent={$remoteContent}
            >
              <Editor
                {fileArtifact}
                {extensions}
                bind:editor
                bind:autoSave={$autoSave}
              />
            </WorkspaceEditorContainer>
          {/if}
        </svelte:fragment>
      </WorkspaceContainer>
    {/if}
  </div>
</div>
