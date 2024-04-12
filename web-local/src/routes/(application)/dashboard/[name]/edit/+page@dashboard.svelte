<script lang="ts">
  import { beforeNavigate, goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import DeployDashboardCta from "@rilldata/web-common/features/dashboards/workspace/DeployDashboardCTA.svelte";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import PreviewButton from "@rilldata/web-common/features/metrics-views/workspace/PreviewButton.svelte";
  import MetricsEditor from "@rilldata/web-common/features/metrics-views/workspace/editor/MetricsEditor.svelte";
  import MetricsInspector from "@rilldata/web-common/features/metrics-views/workspace/inspector/MetricsInspector.svelte";
  import { splitFolderAndName } from "@rilldata/web-common/features/sources/extract-file-name";
  import { useIsModelingSupportedForCurrentOlapDriver as canModel } from "@rilldata/web-common/features/tables/selectors";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";

  const { readOnly } = featureFlags;
  const TOOLTIP_CTA = "Fix this error to enable your dashboard.";

  export let data: { fileArtifact?: FileArtifact } = {};

  let filePath: string;
  let fileArtifact: FileArtifact;
  let metricViewName: string;
  let fileNotFound = false;
  let showDeployModal = false;
  let previewStatus: string[] = [];

  onMount(() => {
    if ($readOnly) {
      fileNotFound = true;
    }
  });

  $: if (data.fileArtifact) {
    fileArtifact = data.fileArtifact;
    filePath = fileArtifact.path;
    metricViewName = fileArtifact.getEntityName();
  } else {
    fileArtifact = fileArtifacts.getFileArtifact(filePath);
    metricViewName = $page.params.name;
    filePath = getFileAPIPathFromNameAndType(
      metricViewName,
      EntityType.MetricsDefinition,
    );
  }
  $: [, fileName] = splitFolderAndName(filePath);

  $: instanceId = $runtime.instanceId;
  $: initLocalUserPreferenceStore(metricViewName);
  $: isModelingSupportedQuery = canModel(instanceId);
  $: isModelingSupported = $isModelingSupportedQuery.data;

  $: fileQuery = createRuntimeServiceGetFile(instanceId, filePath, {
    query: {
      onError: () => (fileNotFound = true),
      // this will ensure that any changes done outside our app is pulled in.
      refetchOnWindowFocus: true,
      keepPreviousData: true,
    },
  });

  let yaml = "";
  $: yaml = $fileQuery.isFetching ? yaml : $fileQuery.data?.blob || "";

  $: allErrorsQuery = fileArtifact.getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;

  $: previewDisbaled = !yaml.length || !!allErrors?.length;

  $: if (!yaml?.length) {
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
      metricViewName,
    );
    if (newRoute) await goto(newRoute);
  }
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

{#if fileNotFound}
  <div class="size-full grid place-content-center">
    <div class="flex flex-col items-center gap-y-2">
      <AlertCircleOutline size="40px" />
      <h1>Page not found</h1>
    </div>
  </div>
{:else}
  <WorkspaceContainer inspector={isModelingSupported}>
    <WorkspaceHeader
      slot="header"
      showInspectorToggle={isModelingSupported}
      titleInput={fileName}
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
          {metricViewName}
          status={previewStatus}
          disabled={previewDisbaled}
        />
      </div>
    </WorkspaceHeader>

    <MetricsEditor slot="body" {yaml} {filePath} {allErrors} {metricViewName} />

    <MetricsInspector {filePath} slot="inspector" />
  </WorkspaceContainer>
{/if}

<DeployDashboardCta
  on:close={() => (showDeployModal = false)}
  open={showDeployModal}
/>
