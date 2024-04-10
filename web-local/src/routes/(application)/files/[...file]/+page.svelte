<script lang="ts">
  import { page } from "$app/stores";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFile,
  } from "@rilldata/web-common/runtime-client";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import FileWorkspaceHeader from "@rilldata/web-common/features/editor/FileWorkspaceHeader.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import SourceModelPage from "../../[type=workspace]/[name]/+page.svelte";
  import ChartPage from "../../chart/[name]/+page.svelte";
  import DashboardPage from "../../dashboard/[name]/edit/+page.svelte";
  import CustomDashboardPage from "../../custom-dashboard/[name]/+page.svelte";

  $: filePath = $page.params.file;
  $: fileQuery = createRuntimeServiceGetFile($runtime.instanceId, filePath);
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: name = fileArtifact.name;
  $: resourceKind = $name?.kind;

  $: isSource = resourceKind === ResourceKind.Source;
  $: isModel = resourceKind === ResourceKind.Model;
  $: isDashboard = resourceKind === ResourceKind.MetricsView;
  $: isChart = resourceKind === ResourceKind.Chart;
  $: isCustomDashboard = resourceKind === ResourceKind.Dashboard;
  $: isUnknown =
    !isSource && !isModel && !isDashboard && !isChart && !isCustomDashboard;

  $: console.log($name, resourceKind);

  // TODO: optimistically update the get file cache
  const putFile = createRuntimeServicePutFile();

  function handleFileUpdate(content: string) {
    if ($fileQuery.data?.blob === content) return;
    return $putFile.mutateAsync({
      instanceId: $runtime.instanceId,
      data: {
        blob: content,
      },
      path: filePath,
    });
  }
</script>

<!-- on:write={(evt) => $putFile.mutate(evt.detail.blob)} -->
{#if isSource || isModel}
  <SourceModelPage data={{ fileArtifact }} />
{:else if isDashboard}
  <DashboardPage data={{ fileArtifact }} />
{:else if isChart}
  <ChartPage data={{ fileArtifact }} />
{:else if isCustomDashboard}
  <CustomDashboardPage data={{ fileArtifact }} />
{:else if isUnknown}
  <WorkspaceContainer>
    <FileWorkspaceHeader filePath={$page.params.file} slot="header" />
    <Editor
      content={$fileQuery.data?.blob ?? ""}
      on:write={({ detail: { content } }) => handleFileUpdate(content)}
      slot="body"
    />
  </WorkspaceContainer>
{/if}
