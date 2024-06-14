<script lang="ts">
  import { beforeNavigate, goto } from "$app/navigation";
  import { page } from "$app/stores";
  import WorkspaceError from "@rilldata/web-common/components/WorkspaceError.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { useIsModelingSupportedForCurrentOlapDriver as canModel } from "@rilldata/web-common/features/connectors/olap/selectors";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import DeployDashboardCta from "@rilldata/web-common/features/dashboards/workspace/DeployDashboardCTA.svelte";
  import {
    getFileAPIPathFromNameAndType,
    getNameFromFile,
  } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import {
    ResourceKind,
    resourceIsLoading,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import PreviewButton from "@rilldata/web-common/features/metrics-views/workspace/PreviewButton.svelte";
  import MetricsEditor from "@rilldata/web-common/features/metrics-views/workspace/editor/MetricsEditor.svelte";
  import MetricsInspector from "@rilldata/web-common/features/metrics-views/workspace/inspector/MetricsInspector.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";

  const { readOnly } = featureFlags;
  const TOOLTIP_CTA = "Fix this error to enable your dashboard.";

  export let data: { fileArtifact?: FileArtifact } = {};

  let fileNotFound = false;
  let showDeployModal = false;
  let previewStatus: string[] = [];

  $: fileArtifact = data?.fileArtifact ?? getLegacyFileArtifact();

  $: metricViewName = getNameFromFile(filePath);

  $: ({ instanceId } = $runtime);
  $: initLocalUserPreferenceStore(metricViewName);
  $: isModelingSupportedQuery = canModel(instanceId);
  $: isModelingSupported = $isModelingSupportedQuery;

  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    remoteContent,
    fileName,
  } = fileArtifact);

  $: allErrorsQuery = fileArtifact.getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;
  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: ({ data: resourceData, isFetching } = $resourceQuery);
  $: isResourceLoading = resourceIsLoading(resourceData);

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

  onMount(() => {
    if ($readOnly) {
      fileNotFound = true;
    }
  });

  beforeNavigate(() => {
    fileNotFound = false;
  });

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

  function getLegacyFileArtifact() {
    const metricViewName = $page.params.name;
    const filePath = getFileAPIPathFromNameAndType(
      metricViewName,
      EntityType.MetricsDefinition,
    );
    return fileArtifacts.getFileArtifact(filePath);
  }
</script>

<svelte:head>
  <title>Rill Developer | {fileName}</title>
</svelte:head>

<svelte:window on:focus={fileArtifact.refetch} />

{#if fileNotFound}
  <WorkspaceError message="File not found." />
{:else}
  <WorkspaceContainer inspector={isModelingSupported}>
    <WorkspaceHeader
      slot="header"
      showInspectorToggle={isModelingSupported}
      titleInput={fileName}
      hasUnsavedChanges={$hasUnsavedChanges}
      on:change={onChangeCallback}
    >
      <div slot="cta" class="flex gap-x-2">
        <Tooltip distance={8}>
          <Button on:click={() => (showDeployModal = true)} type="secondary">
            Deploy
          </Button>
          <TooltipContent slot="tooltip-content">
            Deploy this dashboard to Rill Cloud
          </TooltipContent>
        </Tooltip>
        <PreviewButton
          dashboardName={metricViewName}
          status={previewStatus}
          disabled={previewDisabled}
        />
      </div>
    </WorkspaceHeader>

    <MetricsEditor
      slot="body"
      bind:autoSave={$autoSave}
      {fileArtifact}
      {filePath}
      {allErrors}
      {metricViewName}
    />

    <MetricsInspector yaml={$remoteContent ?? ""} slot="inspector" />
  </WorkspaceContainer>
{/if}

<DeployDashboardCta
  on:close={() => (showDeployModal = false)}
  open={showDeployModal}
/>
