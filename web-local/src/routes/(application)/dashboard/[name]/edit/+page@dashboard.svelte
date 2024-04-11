<script lang="ts">
  import { page } from "$app/stores";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";
  import { CATALOG_ENTRY_NOT_FOUND } from "../../../../../lib/errors/messages";
  import DeployDashboardCta from "@rilldata/web-common/features/dashboards/workspace/DeployDashboardCTA.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import MetricsInspector from "@rilldata/web-common/features/metrics-views/workspace/inspector/MetricsInspector.svelte";
  import MetricsEditor from "@rilldata/web-common/features/metrics-views/workspace/editor/MetricsEditor.svelte";
  import { useIsModelingSupportedForCurrentOlapDriver as canModel } from "@rilldata/web-common/features/tables/selectors";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { goto } from "$app/navigation";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import PreviewButton from "@rilldata/web-common/features/metrics-views/workspace/PreviewButton.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  const { readOnly } = featureFlags;
  const TOOLTIP_CTA = "Fix this error to enable your dashboard.";

  export let data;

  let showDeployModal = false;
  let previewStatus: string[] = [];

  onMount(() => {
    if ($readOnly) {
      throw error(404, "Page not found");
    }
  });

  $: instanceId = data.instanceId;

  $: metricViewName = $page.params.name;

  $: initLocalUserPreferenceStore(metricViewName);
  $: isModelingSupportedQuery = canModel(instanceId);
  $: isModelingSupported = $isModelingSupportedQuery.data;

  $: filePath = getFileAPIPathFromNameAndType(
    metricViewName,
    EntityType.MetricsDefinition,
  );

  $: fileQuery = createRuntimeServiceGetFile(instanceId, filePath, {
    query: {
      onError: (err) => {
        if (err.response?.data?.message.includes(CATALOG_ENTRY_NOT_FOUND)) {
          throw error(404, "Dashboard not found");
        }

        throw error(err.response?.status || 500, err.message);
      },
      // this will ensure that any changes done outside our app is pulled in.
      refetchOnWindowFocus: true,
    },
  });

  $: yaml = $fileQuery.data?.blob || "";

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: allErrorsQuery = fileArtifact.getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;

  $: dashboardQuery = fileArtifact.getResource(queryClient, instanceId);
  $: dashboard = $dashboardQuery;

  $: previewDisbaled = !yaml?.length || !!allErrors?.length;

  $: if (!yaml?.length) {
    previewStatus = [
      "Your metrics definition is empty. Get started by trying one of the options in the editor.",
    ];
  }
  // content & errors
  else if (allErrors?.length && allErrors[0].message) {
    previewStatus = [allErrors[0].message, TOOLTIP_CTA];
  }
  // preview is available
  else {
    previewStatus = ["Explore your metrics dashboard"];
  }

  const onChangeCallback = async (
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) => {
    const newRoute = await handleEntityRename(
      instanceId,
      e.currentTarget,
      filePath,
      EntityType.MetricsDefinition,
    );
    if (newRoute) await goto(newRoute + "/edit");
  };
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

<WorkspaceContainer inspector={isModelingSupported}>
  <WorkspaceHeader
    slot="header"
    showInspectorToggle={isModelingSupported}
    titleInput={metricViewName}
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

  <MetricsEditor
    slot="body"
    {yaml}
    {filePath}
    {allErrors}
    {dashboard}
    {metricViewName}
  />
  <MetricsInspector {filePath} slot="inspector" />
</WorkspaceContainer>

<DeployDashboardCta
  on:close={() => (showDeployModal = false)}
  open={showDeployModal}
/>
