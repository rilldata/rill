<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import MetricsIcon from "@rilldata/web-common/components/icons/Metrics.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import { MenuItem } from "@rilldata/web-common/components/menu";
  import { Divider } from "@rilldata/web-common/components/menu/index.js";
  import { deleteFileArtifact } from "@rilldata/web-common/features/entity-management/actions";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import {
    FileArtifactsData,
    fileArtifactsStore,
  } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import {
    runtimeServiceGetFile,
    useRuntimeServiceDeleteFileAndReconcile,
    useRuntimeServicePutFileAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { initBlankDashboardYAML } from "@rilldata/web-local/lib/application-state-stores/metrics-internal-store";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import { useDashboardNames } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { MetricsSourceSelectionError } from "@rilldata/web-local/lib/temp/errors/ErrorMessages.js";
  import { SourceModelValidationStatus } from "@rilldata/web-local/lib/temp/metrics.js";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { slide } from "svelte/transition";
  import RenameAssetModal from "../../../../../../web-common/src/features/entity-management/RenameAssetModal.svelte";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import NavigationEntry from "../NavigationEntry.svelte";
  import NavigationHeader from "../NavigationHeader.svelte";

  $: instanceId = $runtimeStore.instanceId;

  $: dashboardNames = useDashboardNames(instanceId);

  const queryClient = useQueryClient();

  const createDashboard = useRuntimeServicePutFileAndReconcile();
  const deleteDashboard = useRuntimeServiceDeleteFileAndReconcile();

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
    navigationEvent.fireEvent(
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
    navigationEvent.fireEvent(
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

        navigationEvent.fireEvent(
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

  $: canAddDashboard = $runtimeStore.readOnly === false;
</script>

<NavigationHeader
  bind:show={showMetricsDefs}
  on:add={dispatchAddEmptyMetricsDef}
  tooltipText="Create a new dashboard"
  toggleText="dashboards"
  canAddAsset={canAddDashboard}
>
  Dashboards
</NavigationHeader>

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
  </div>
  {#if showRenameMetricsDefinitionModal}
    <RenameAssetModal
      entityType={EntityType.MetricsDefinition}
      closeModal={() => (showRenameMetricsDefinitionModal = false)}
      currentAssetName={renameMetricsDefName}
    />
  {/if}
{/if}
