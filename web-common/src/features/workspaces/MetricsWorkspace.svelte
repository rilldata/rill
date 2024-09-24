<script lang="ts">
  import { goto } from "$app/navigation";
  import LocalAvatarButton from "@rilldata/web-common/features/authentication/LocalAvatarButton.svelte";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import DeployDashboardCta from "@rilldata/web-common/features/dashboards/workspace/DeployDashboardCTA.svelte";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import {
    ResourceKind,
    resourceIsLoading,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import MetricsInspector from "@rilldata/web-common/features/metrics-views/MetricsInspector.svelte";
  import PreviewButton from "@rilldata/web-common/features/metrics-views/workspace/PreviewButton.svelte";
  import MetricsEditor from "@rilldata/web-common/features/metrics-views/workspace/editor/MetricsEditor.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import WorkspaceEditorContainer from "../../layout/workspace/WorkspaceEditorContainer.svelte";
  import {
    useIsModelingSupportedForDefaultOlapDriver,
    useIsModelingSupportedForOlapDriver,
  } from "../connectors/olap/selectors";

  const TOOLTIP_CTA = "Fix this error to enable your dashboard.";

  export let fileArtifact: FileArtifact;

  let previewStatus: string[] = [];

  $: ({ instanceId } = $runtime);
  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    remoteContent,
    fileName,
  } = fileArtifact);

  $: metricsViewName = getNameFromFile(filePath);

  $: initLocalUserPreferenceStore(metricsViewName);

  $: allErrorsQuery = fileArtifact.getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;
  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: ({ data: resourceData, isFetching } = $resourceQuery);
  $: isResourceLoading = resourceIsLoading(resourceData);

  $: connector = resourceData?.metricsView?.state?.validSpec?.connector ?? "";
  $: database = resourceData?.metricsView?.state?.validSpec?.database ?? "";
  $: databaseSchema =
    resourceData?.metricsView?.state?.validSpec?.databaseSchema ?? "";
  $: table = resourceData?.metricsView?.state?.validSpec?.table ?? "";

  $: isModelingSupportedForDefaultOlapDriver =
    useIsModelingSupportedForDefaultOlapDriver(instanceId);
  $: isModelingSupportedForOlapDriver = useIsModelingSupportedForOlapDriver(
    instanceId,
    connector,
  );
  $: isModelingSupported = connector
    ? $isModelingSupportedForOlapDriver
    : $isModelingSupportedForDefaultOlapDriver;

  $: previewDisabled =
    !$remoteContent?.length ||
    !!allErrors?.length ||
    isResourceLoading ||
    isFetching;

  $: if (!$remoteContent?.length) {
    previewStatus = [
      "Your metrics definition is empty. Get started by trying one of the options in the editor.",
    ];
  } else if (allErrors?.length && allErrors[0].message) {
    // content & errors
    previewStatus = [allErrors[0].message, TOOLTIP_CTA];
  } else {
    // preview is available
    previewStatus = ["Explore your metrics dashboard"];
  }

  async function onChangeCallback(
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) {
    const newRoute = await handleEntityRename(
      instanceId,
      e.currentTarget,
      filePath,
      fileName,
      fileArtifacts.getNamesForKind(ResourceKind.MetricsView),
    );
    if (newRoute) await goto(newRoute);
  }
</script>

<WorkspaceContainer inspector={isModelingSupported}>
  <WorkspaceHeader
    hasUnsavedChanges={$hasUnsavedChanges}
    on:change={onChangeCallback}
    showInspectorToggle={isModelingSupported}
    slot="header"
    titleInput={fileName}
  >
    <div class="flex gap-x-2" slot="cta">
      <PreviewButton
        dashboardName={metricsViewName}
        disabled={previewDisabled}
        status={previewStatus}
      />
      <DeployDashboardCta />
      <LocalAvatarButton />
    </div>
  </WorkspaceHeader>

  <WorkspaceEditorContainer slot="body">
    <MetricsEditor
      bind:autoSave={$autoSave}
      {fileArtifact}
      {filePath}
      {allErrors}
      {metricsViewName}
    />
  </WorkspaceEditorContainer>

  <MetricsInspector
    {filePath}
    {connector}
    {database}
    {databaseSchema}
    {table}
    slot="inspector"
  />
</WorkspaceContainer>
