<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    getRuntimeServiceListFilesQueryKey,
    useRuntimeServiceDeleteFileAndMigrate,
    useRuntimeServiceListFiles,
    useRuntimeServicePutFileAndMigrate,
    useRuntimeServiceRenameFileAndMigrate,
  } from "@rilldata/web-common/runtime-client";
  import { queryClient } from "../../../svelte-query/globalQueryClient";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type { MetricsDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import { MetricsSourceSelectionError } from "@rilldata/web-local/common/errors/ErrorMessages";
  import { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/common/metrics-service/MetricsTypes";
  import { waitUntil } from "@rilldata/web-local/common/utils/waitUtils";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import notificationStore from "@rilldata/web-local/lib/components/notifications";
  import { getContext, onMount } from "svelte";
  import { slide } from "svelte/transition";
  import type { ApplicationStore } from "../../../application-state-stores/application-store";
  import type { DerivedModelStore } from "../../../application-state-stores/model-stores";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import {
    createMetricsDefsAndFocusApi,
    deleteMetricsDefsApi,
    fetchManyMetricsDefsApi,
    validateSelectedSources,
  } from "../../../redux-store/metrics-definition/metrics-definition-apis";
  import { getAllMetricsDefinitionsReadable } from "../../../redux-store/metrics-definition/metrics-definition-readables";
  import { store } from "../../../redux-store/store-root";
  import Cancel from "../../icons/Cancel.svelte";
  import EditIcon from "../../icons/EditIcon.svelte";
  import { default as Explore } from "../../icons/Explore.svelte";
  import MetricsIcon from "../../icons/Metrics.svelte";
  import Model from "../../icons/Model.svelte";
  import { Divider, MenuItem } from "../../menu";
  import MetricsDefinitionSummary from "../../metrics-definition/MetricsDefinitionSummary.svelte";
  import NavigationEntry from "../NavigationEntry.svelte";
  import NavigationHeader from "../NavigationHeader.svelte";
  import RenameAssetModal from "../RenameAssetModal.svelte";
  import { metricsTemplate } from "./metricsUtils";
  import { getName } from "@rilldata/web-local/common/utils/incrementName";

  $: repoId = $runtimeStore.repoId;
  $: instanceId = $runtimeStore.instanceId;

  $: getFiles = useRuntimeServiceListFiles($runtimeStore.repoId);

  $: dashboardNames = $getFiles?.data?.paths
    ?.filter((path) => path.includes("dashboards/"))
    .map((path) => path.replace("/dashboards/", "").replace(".yaml", ""));

  const createDashboard = useRuntimeServicePutFileAndMigrate();
  const deleteDashboard = useRuntimeServiceDeleteFileAndMigrate();
  const renameDashboard = useRuntimeServiceRenameFileAndMigrate();

  const metricsDefinitions = getAllMetricsDefinitionsReadable();
  const appStore = getContext("rill:app:store") as ApplicationStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;
  const applicationStore = getContext("rill:app:store") as ApplicationStore;

  let showMetricsDefs = true;

  let showRenameMetricsDefinitionModal = false;
  let renameMetricsDefId = null;
  let renameMetricsDefName = null;

  const openRenameMetricsDefModal = (
    metricsDefId: string,
    metricsDefName: string
  ) => {
    showRenameMetricsDefinitionModal = true;
    renameMetricsDefId = metricsDefId;
    renameMetricsDefName = metricsDefName;
  };

  const dispatchAddEmptyMetricsDef = async () => {
    if (!showMetricsDefs) {
      showMetricsDefs = true;
    }
    const newDashboardName = getName("dashboard", dashboardNames);
    const yaml = metricsTemplate;
    $createDashboard.mutate(
      {
        data: {
          repoId,
          instanceId,
          path: `dashboards/${newDashboardName}.yaml`,
          blob: yaml,
          create: true,
          createOnly: true,
          strict: false,
        },
      },
      {
        onSuccess: async () => {
          goto(`/dashboard/${newDashboardName}`);
          queryClient.invalidateQueries(
            getRuntimeServiceListFilesQueryKey($runtimeStore.repoId)
          );
        },
      }
    );
  };

  const editModel = (sourceModelId: string) => {
    goto(`/model/${sourceModelId}`);

    const previousActiveEntity = $appStore?.activeEntity?.type;
    navigationEvent.fireEvent(
      sourceModelId,
      BehaviourEventMedium.Menu,
      MetricsEventSpace.LeftPanel,
      EntityTypeToScreenMap[previousActiveEntity],
      MetricsEventScreenName.Model
    );
  };

  const editMetrics = (metricsId: string) => {
    goto(`/dashboard/${metricsId}/edit`);

    const previousActiveEntity = $appStore?.activeEntity?.type;
    navigationEvent.fireEvent(
      metricsId,
      BehaviourEventMedium.Menu,
      MetricsEventSpace.LeftPanel,
      EntityTypeToScreenMap[previousActiveEntity],
      MetricsEventScreenName.MetricsDefinition
    );
  };

  const deleteMetricsDef = async (dashboardName: string) => {
    const sourceModelName = undefined; //TODO fix later with model integration

    await $deleteDashboard.mutate(
      {
        data: {
          repoId: $runtimeStore.repoId,
          instanceId: $runtimeStore.instanceId,
          path: `dashboards/${dashboardName}.yaml`,
        },
      },
      {
        onSuccess: () => {
          notificationStore.send({
            message: `Dashboard "${dashboardName}" deleted`,
          });
          if (
            ($applicationStore.activeEntity.type ===
              EntityType.MetricsDefinition ||
              $applicationStore.activeEntity.type ===
                EntityType.MetricsExplorer) &&
            $applicationStore.activeEntity.id === dashboardName
          ) {
            if (sourceModelName) {
              goto(`/model/${sourceModelName}`);
            } else {
              goto("/");
            }
          }
          queryClient.invalidateQueries(
            getRuntimeServiceListFilesQueryKey($runtimeStore.repoId)
          );
        },
      }
    );
  };

  onMount(() => {
    // TODO: once we have everything in redux store we can easily move this to its own async thunk
    store.dispatch(fetchManyMetricsDefsApi()).then(async () => {
      await waitUntil(() => {
        return !!$derivedModelStore;
      }, -1);
      $metricsDefinitions.forEach((metricsDefinition) =>
        store.dispatch(validateSelectedSources(metricsDefinition.id))
      );
    });
  });
</script>

<NavigationHeader
  bind:show={showMetricsDefs}
  tooltipText="create a new dashboard"
  on:add={dispatchAddEmptyMetricsDef}
>
  <Explore size="16px" /> Dashboards
</NavigationHeader>

{#if showMetricsDefs && dashboardNames}
  <div
    class="pb-6 justify-self-end"
    transition:slide={{ duration: LIST_SLIDE_DURATION }}
    id="assets-metrics-list"
  >
    {#each dashboardNames as dashboardName (dashboardName)}
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
          <!-- {@const selectionError = MetricsSourceSelectionError(metricsDef)}
          {@const hasSourceError =
            selectionError !== SourceModelValidationStatus.OK &&
            selectionError !== ""} -->
          <!-- <MenuItem
            icon
            disabled={hasSourceError}
            on:select={() => editModel(metricsDef.sourceModelId)}
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
            on:select={() => editMetrics(metricsDef.id)}
          >
            <MetricsIcon slot="icon" />
            edit metrics
          </MenuItem>
          <Divider /> -->
          <!-- <MenuItem
            icon
            on:select={() =>
              openRenameMetricsDefModal(
                metricsDef.id,
                metricsDef.metricDefLabel
              )}
          >
            <EditIcon slot="icon" />
            rename...</MenuItem
          > -->
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
      entityId={renameMetricsDefId}
      currentAssetName={renameMetricsDefName}
    />
  {/if}
{/if}
