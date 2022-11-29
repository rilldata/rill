<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    getRuntimeServiceListFilesQueryKey,
    useRuntimeServiceDeleteFileAndReconcile,
    useRuntimeServicePutFileAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { SourceModelValidationStatus } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService.js";
  import { MetricsSourceSelectionError } from "@rilldata/web-local/common/errors/ErrorMessages.js";
  import { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/common/metrics-service/MetricsTypes";
  import { getName } from "@rilldata/web-local/common/utils/incrementName";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import {
    fileArtifactsStore,
    FileArtifactsData,
  } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store.js";
  import { metricsTemplate } from "@rilldata/web-local/lib/application-state-stores/metrics-internal-store";
  import { getFileFromName } from "@rilldata/web-local/lib/util/entity-mappers.js";
  import Model from "@rilldata/web-local/lib/components/icons/Model.svelte";
  import { Divider } from "@rilldata/web-local/lib/components/menu/index.js";
  import { deleteFileArtifact } from "@rilldata/web-local/lib/svelte-query/actions";
  import { useDashboardNames } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";
  import type { ApplicationStore } from "../../../application-state-stores/application-store";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import { queryClient } from "../../../svelte-query/globalQueryClient";
  import Cancel from "../../icons/Cancel.svelte";
  import { default as Explore } from "../../icons/Explore.svelte";
  import { MenuItem } from "../../menu";
  import MetricsDefinitionSummary from "../../metrics-definition/MetricsDefinitionSummary.svelte";
  import NavigationEntry from "../NavigationEntry.svelte";
  import NavigationHeader from "../NavigationHeader.svelte";
  import RenameAssetModal from "../RenameAssetModal.svelte";
  import MetricsIcon from "../../icons/Metrics.svelte";
  import EditIcon from "../../icons/EditIcon.svelte";

  $: instanceId = $runtimeStore.instanceId;

  $: dashboardNames = useDashboardNames(instanceId);

  const createDashboard = useRuntimeServicePutFileAndReconcile();
  const deleteDashboard = useRuntimeServiceDeleteFileAndReconcile();

  const appStore = getContext("rill:app:store") as ApplicationStore;
  const applicationStore = getContext("rill:app:store") as ApplicationStore;

  let showMetricsDefs = true;

  let showRenameMetricsDefinitionModal = false;
  let renameMetricsDefName = null;

  const openRenameMetricsDefModal = (metricsDefName: string) => {
    showRenameMetricsDefinitionModal = true;
    renameMetricsDefName = metricsDefName;
  };

  const dispatchAddEmptyMetricsDef = async () => {
    if (!showMetricsDefs) {
      showMetricsDefs = true;
    }
    const newDashboardName = getName("dashboard", $dashboardNames.data);
    await $createDashboard.mutateAsync({
      data: {
        instanceId,
        path: `dashboards/${newDashboardName}.yaml`,
        blob: metricsTemplate,
        create: true,
        createOnly: true,
        strict: false,
      },
    });
    goto(`/dashboard/${newDashboardName}`);
    queryClient.invalidateQueries(
      getRuntimeServiceListFilesQueryKey(instanceId)
    );
  };

  const editModel = (sourceModelName: string) => {
    goto(`/model/${sourceModelName}`);

    const previousActiveEntity = $appStore?.activeEntity?.type;
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
    await deleteFileArtifact(
      instanceId,
      dashboardName,
      EntityType.MetricsDefinition,
      $deleteDashboard,
      $applicationStore.activeEntity,
      $dashboardNames.data
    );
  };

  const getDashboardData = (
    entities: Record<string, FileArtifactsData>,
    name: string
  ) => {
    return entities[name];
  };
</script>

<NavigationHeader
  bind:show={showMetricsDefs}
  on:add={dispatchAddEmptyMetricsDef}
  tooltipText="create a new dashboard"
>
  <Explore size="16px" /> Dashboards
</NavigationHeader>

{#if showMetricsDefs && $dashboardNames.data}
  <div
    class="pb-6 justify-self-end"
    transition:slide={{ duration: LIST_SLIDE_DURATION }}
    id="assets-metrics-list"
  >
    {#each $dashboardNames.data as dashboardName (dashboardName)}
      {@const dashboardData = getDashboardData(
        $fileArtifactsStore.entities,
        dashboardName
      )}
      <NavigationEntry
        notExpandable={true}
        name={dashboardName}
        href={`/dashboard/${dashboardName}`}
        open={$page.url.pathname === `/dashboard/${dashboardName}` ||
          $page.url.pathname === `/dashboard/${dashboardName}/edit`}
      >
        <svelte:fragment slot="summary" let:containerWidth>
          <MetricsDefinitionSummary indentLevel={1} {containerWidth} />
        </svelte:fragment>

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
            edit model
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
            edit metrics
          </MenuItem>
          <Divider />
          <MenuItem
            icon
            on:select={() => openRenameMetricsDefModal(dashboardName)}
          >
            <EditIcon slot="icon" />
            rename...</MenuItem
          >
          <MenuItem icon on:select={() => deleteMetricsDef(dashboardName)}>
            <Cancel slot="icon" />
            delete</MenuItem
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
