<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import MetricsIcon from "@rilldata/web-common/components/icons/Metrics.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import { MenuItem } from "@rilldata/web-common/components/menu";
  import { Divider } from "@rilldata/web-common/components/menu/index.js";
  import { useDashboardNames } from "@rilldata/web-common/features/dashboards/selectors";
  import { deleteFileArtifact } from "@rilldata/web-common/features/entity-management/actions";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import {
    FileArtifactsData,
    fileArtifactsStore,
  } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { SourceModelValidationStatus } from "@rilldata/web-common/features/metrics-views/errors.js";
  import { initBlankDashboardYAML } from "@rilldata/web-common/features/metrics-views/metrics-internal-store";
  import { appStore } from "@rilldata/web-common/layout/app-store";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import {
    createRuntimeServiceDeleteFileAndReconcile,
    createRuntimeServicePutFileAndReconcile,
    runtimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";
  import { invalidateAfterReconcile } from "@rilldata/web-common/runtime-client/invalidation";
  import { MetricsSourceSelectionError } from "@rilldata/web-local/lib/temp/errors/ErrorMessages.js";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";
  import { behaviourEvent } from "../../metrics/initMetrics";
  import { runtime } from "../../runtime-client/runtime-store";
  import AddAssetButton from "../entity-management/AddAssetButton.svelte";
  import RenameAssetModal from "../entity-management/RenameAssetModal.svelte";

  $: instanceId = $runtime.instanceId;

  $: dashboardNames = useDashboardNames(instanceId);

  const queryClient = useQueryClient();

  const createDashboard = createRuntimeServicePutFileAndReconcile();
  const deleteDashboard = createRuntimeServiceDeleteFileAndReconcile();

  let showMetricsDefs = true;

  let showRenameMetricsDefinitionModal = false;
  let renameMetricsDefName = null;

  async function getDashboardArtifact(
    instanceId: string,
    metricViewName: string
  ) {
    const filePath = getFilePathFromNameAndType(
      metricViewName,
      EntityType.MetricsDefinition
    );
    const resp = await runtimeServiceGetFile(instanceId, filePath);
    const metricYAMLString = resp.blob;
    fileArtifactsStore.setJSONRep(filePath, metricYAMLString);
  }

  const openRenameMetricsDefModal = (metricsDefName: string) => {
    showRenameMetricsDefinitionModal = true;
    renameMetricsDefName = metricsDefName;
  };

  const dispatchAddEmptyMetricsDef = async () => {
    if (!showMetricsDefs) {
      showMetricsDefs = true;
    }
    const newDashboardName = getName("dashboard", $dashboardNames.data);
    const filePath = getFilePathFromNameAndType(
      newDashboardName,
      EntityType.MetricsDefinition
    );
    const yaml = initBlankDashboardYAML(newDashboardName);
    const resp = await $createDashboard.mutateAsync({
      data: {
        instanceId,
        path: filePath,
        blob: yaml,
        create: true,
        createOnly: true,
        strict: false,
      },
    });
    fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

    goto(`/dashboard/${newDashboardName}`);
    return invalidateAfterReconcile(queryClient, instanceId, resp);
  };

  const editModel = async (dashboardName: string) => {
    await getDashboardArtifact(instanceId, dashboardName);

    const dashboardData = getDashboardData(
      $fileArtifactsStore.entities,
      dashboardName
    );
    const sourceModelName = dashboardData.jsonRepresentation.model;

    const previousActiveEntity = $appStore?.activeEntity?.type;
    goto(`/model/${sourceModelName}`);
    behaviourEvent.fireNavigationEvent(
      sourceModelName,
      BehaviourEventMedium.Menu,
      MetricsEventSpace.LeftPanel,
      EntityTypeToScreenMap[previousActiveEntity],
      MetricsEventScreenName.Model
    );
  };

  const editMetrics = (dashboardName: string) => {
    goto(`/dashboard/${dashboardName}/edit`);

    const previousActiveEntity = $appStore?.activeEntity?.type;
    behaviourEvent.fireNavigationEvent(
      dashboardName,
      BehaviourEventMedium.Menu,
      MetricsEventSpace.LeftPanel,
      EntityTypeToScreenMap[previousActiveEntity],
      MetricsEventScreenName.MetricsDefinition
    );
  };

  const deleteMetricsDef = async (dashboardName: string) => {
    await getDashboardArtifact(instanceId, dashboardName);

    const dashboardData = getDashboardData(
      $fileArtifactsStore.entities,
      dashboardName
    );
    await deleteFileArtifact(
      queryClient,
      instanceId,
      dashboardName,
      EntityType.MetricsDefinition,
      $deleteDashboard,
      $appStore.activeEntity,
      $dashboardNames.data
    );

    // redirect to model when metric is deleted
    const sourceModelName = dashboardData.jsonRepresentation.model;
    if ($appStore.activeEntity.name === dashboardName) {
      if (sourceModelName) {
        goto(`/model/${sourceModelName}`);

        behaviourEvent.fireNavigationEvent(
          sourceModelName,
          BehaviourEventMedium.Menu,
          MetricsEventSpace.LeftPanel,
          MetricsEventScreenName.MetricsDefinition,
          MetricsEventScreenName.Model
        );
      } else {
        goto("/");
      }
    }
  };

  const getDashboardData = (
    entities: Record<string, FileArtifactsData>,
    name: string
  ) => {
    const dashboardPath = getFilePathFromNameAndType(
      name,
      EntityType.MetricsDefinition
    );
    return entities[dashboardPath];
  };

  $: canAddDashboard = $featureFlags.readOnly === false;
</script>

<NavigationHeader bind:show={showMetricsDefs} toggleText="dashboards"
  >Dashboards</NavigationHeader
>

{#if showMetricsDefs && $dashboardNames.data}
  <div
    class="pb-3 justify-self-end"
    transition:slide={{ duration: LIST_SLIDE_DURATION }}
    id="assets-metrics-list"
  >
    {#each $dashboardNames.data as dashboardName (dashboardName)}
      {@const dashboardData = getDashboardData(
        $fileArtifactsStore.entities,
        dashboardName
      )}
      <NavigationEntry
        showContextMenu={!$featureFlags.readOnly}
        expandable={false}
        name={dashboardName}
        href={`/dashboard/${dashboardName}`}
        open={$page.url.pathname === `/dashboard/${dashboardName}` ||
          $page.url.pathname === `/dashboard/${dashboardName}/edit`}
      >
        <svelte:fragment slot="menu-items">
          {@const selectionError = MetricsSourceSelectionError(
            dashboardData?.errors
          )}
          {@const hasSourceError =
            selectionError !== SourceModelValidationStatus.OK &&
            selectionError !== ""}
          <MenuItem
            icon
            disabled={hasSourceError}
            on:select={() => editModel(dashboardName)}
          >
            <Model slot="icon" />
            Edit model
            <svelte:fragment slot="description">
              {#if hasSourceError}
                {selectionError}
              {/if}
            </svelte:fragment>
          </MenuItem>
          <MenuItem
            icon
            disabled={hasSourceError}
            on:select={() => editMetrics(dashboardName)}
          >
            <MetricsIcon slot="icon" />
            Edit metrics
          </MenuItem>
          <Divider />
          <MenuItem
            icon
            on:select={() => openRenameMetricsDefModal(dashboardName)}
          >
            <EditIcon slot="icon" />
            Rename...</MenuItem
          >
          <MenuItem icon on:select={() => deleteMetricsDef(dashboardName)}>
            <Cancel slot="icon" />
            Delete</MenuItem
          >
        </svelte:fragment>
      </NavigationEntry>
    {/each}
    {#if canAddDashboard}
      <AddAssetButton
        id="add-dashboard"
        label="Add dashboard"
        on:click={() => dispatchAddEmptyMetricsDef()}
      />
    {/if}
  </div>
  {#if showRenameMetricsDefinitionModal}
    <RenameAssetModal
      entityType={EntityType.MetricsDefinition}
      closeModal={() => (showRenameMetricsDefinitionModal = false)}
      currentAssetName={renameMetricsDefName}
    />
  {/if}
{/if}
