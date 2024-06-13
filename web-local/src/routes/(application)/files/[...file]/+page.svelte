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

  export let data;

  $: ({ filePath } = data);

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: ({ autoSave, hasUnsavedChanges, fileName, name } = fileArtifact);

  $: resourceKind = $name?.kind;

  $: isSource = resourceKind === ResourceKind.Source;
  $: isModel = resourceKind === ResourceKind.Model;
  $: isDashboard = resourceKind === ResourceKind.MetricsView;
  $: isChart = resourceKind === ResourceKind.Component;
  $: isCustomDashboard = resourceKind === ResourceKind.Dashboard;
  $: isOther =
    !isSource && !isModel && !isDashboard && !isChart && !isCustomDashboard;

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

{#if isSource || isModel}
  <SourceModelPage data={{ fileArtifact }} />
{:else if isDashboard}
  <DashboardPage data={{ fileArtifact }} />
{:else if isChart}
  {#key filePath}
    <ChartPage data={{ fileArtifact }} />
  {/key}
{:else if isCustomDashboard}
  <CustomDashboardPage data={{ fileArtifact }} />
{:else if isOther}
  <WorkspaceContainer inspector={false}>
    <FileWorkspaceHeader
      {filePath}
      hasUnsavedChanges={$hasUnsavedChanges}
      slot="header"
    />
    <div
      slot="body"
      class="editor-pane size-full overflow-hidden flex flex-col"
    >
      <WorkspaceEditorContainer>
        <Editor
          {fileArtifact}
          extensions={getExtensionsForFile(filePath)}
          bind:autoSave={$autoSave}
        />
      </WorkspaceEditorContainer>
    </div>
  </WorkspaceContainer>
{/if}
