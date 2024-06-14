<script lang="ts">
  import { afterNavigate } from "$app/navigation";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import FileWorkspaceHeader from "@rilldata/web-common/features/editor/FileWorkspaceHeader.svelte";
  import { getExtensionsForFile } from "@rilldata/web-common/features/editor/getExtensionsForFile";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { directoryState } from "@rilldata/web-common/features/file-explorer/directory-store";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import { onMount } from "svelte";
  import SourceModelPage from "../../[type=workspace]/[name]/+page.svelte";
  import ChartPage from "../../chart/[name]/+page.svelte";
  import CustomDashboardPage from "../../custom-dashboards/[name]/+page.svelte";
  import DashboardPage from "../../dashboard/[name]/edit/+page.svelte";
  import type { EditorView } from "@codemirror/view";

  const pages = new Map([
    [ResourceKind.Source, SourceModelPage],
    [ResourceKind.Model, SourceModelPage],
    [ResourceKind.MetricsView, DashboardPage],
    [ResourceKind.Component, ChartPage],
    [ResourceKind.Dashboard, CustomDashboardPage],
    [undefined, null],
  ]);

  export let data;

  let editor: EditorView;

  $: ({ filePath } = data);
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: ({ autoSave, hasUnsavedChanges, fileName, name } = fileArtifact);

  $: resourceKind = <ResourceKind | undefined>$name?.kind;

  $: page = pages.get(resourceKind);

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

{#if page}
  <svelte:component this={page} {data} />
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
