<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    runtimeServiceGetCatalogEntry,
    useRuntimeServiceDeleteFileAndReconcile,
    useRuntimeServicePutFileAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { SourceModelValidationStatus } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService.js";
  import { MetricsSourceSelectionError } from "@rilldata/web-local/common/errors/ErrorMessages.js";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import { getName } from "@rilldata/web-local/common/utils/incrementName";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import {
    FileArtifactsData,
    fileArtifactsStore,
  } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store.js";
  import { metricsTemplate } from "@rilldata/web-local/lib/application-state-stores/metrics-internal-store";
  import Model from "@rilldata/web-local/lib/components/icons/Model.svelte";
  import { Divider } from "@rilldata/web-local/lib/components/menu/index.js";
  import { deleteFileArtifact } from "@rilldata/web-local/lib/svelte-query/actions";
  import { useDashboardNames } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { slide } from "svelte/transition";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import Cancel from "../../icons/Cancel.svelte";
  import EditIcon from "../../icons/EditIcon.svelte";
  import { default as Explore } from "../../icons/Explore.svelte";
  import MetricsIcon from "../../icons/Metrics.svelte";
  import { MenuItem } from "../../menu";
  import MetricsDefinitionSummary from "../../metrics-definition/MetricsDefinitionSummary.svelte";
  import NavigationEntry from "../NavigationEntry.svelte";
  import NavigationHeader from "../NavigationHeader.svelte";
  import RenameAssetModal from "../RenameAssetModal.svelte";

  $: instanceId = $runtimeStore.instanceId;

  $: dashboardNames = useDashboardNames(instanceId);

  const queryClient = useQueryClient();

  const createDashboard = useRuntimeServicePutFileAndReconcile();
  const deleteDashboard = useRuntimeServiceDeleteFileAndReconcile();

  let showMetricsDefs = true;

  let showRenameMetricsDefinitionModal = false;
  let renameMetricsDefName = null;

  async function getModelFromDashboardName(
    instanceId: string,
    metricViewName: string
  ) {
    const metricViewCatalog = await runtimeServiceGetCatalogEntry(
      instanceId,
      metricViewName
    );
    const metricViewEntry = metricViewCatalog.entry?.metricsView;
    return metricViewEntry?.from;
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
    const resp = await $createDashboard.mutateAsync({
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
    return invalidateAfterReconcile(queryClient, instanceId, resp);
  };

  const editModel = async (dashboardName: string) => {
    const sourceModelName = await getModelFromDashboardName(
      instanceId,
      dashboardName
    );
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
    const sourceModelName = await getModelFromDashboardName(
      instanceId,
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

    if ($appStore.activeEntity.name === dashboardName) {
      if (sourceModelName) {
        goto(`/model/${sourceModelName}`);
      } else {
        goto("/");
      }
    }
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
