<script lang="ts">
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import { getExtensionsForFile } from "@rilldata/web-common/features/editor/getExtensionsForFile";
  import { mapParseErrorsToLines } from "@rilldata/web-common/features/metrics-views/errors";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import RunReportButton from "./RunReportButton.svelte";

  export let fileArtifact: FileArtifact;

  $: ({ instanceId } = $runtime);
  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    fileName,
    remoteContent,
    resourceName,
  } = fileArtifact);

  $: reportName = $resourceName?.name ?? getNameFromFile(filePath);

  $: allErrorsQuery = fileArtifact.getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;
  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: resource = $resourceQuery.data;

  $: extensions = getExtensionsForFile(filePath);
  $: errors = mapParseErrorsToLines(allErrors, $remoteContent ?? "");
  $: mainError = errors?.at(0);
</script>

<WorkspaceContainer inspector={false}>
  <WorkspaceHeader
    slot="header"
    {filePath}
    {resource}
    resourceKind={ResourceKind.Report}
    titleInput={fileName}
    hasUnsavedChanges={$hasUnsavedChanges}
  >
    <div slot="cta">
      <RunReportButton {reportName} {instanceId} />
    </div>
  </WorkspaceHeader>

  <WorkspaceEditorContainer slot="body" error={mainError}>
    <Editor
      {fileArtifact}
      {extensions}
      bind:autoSave={$autoSave}
    />
  </WorkspaceEditorContainer>
</WorkspaceContainer>

