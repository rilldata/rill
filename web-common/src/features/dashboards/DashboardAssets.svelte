<script lang="ts">
  import {_} from "svelte-i18n";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import MetricsIcon from "@rilldata/web-common/components/icons/Metrics.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import { MenuItem } from "@rilldata/web-common/components/menu";
  import { Divider } from "@rilldata/web-common/components/menu/index.js";
  import { useDashboardFileNames } from "@rilldata/web-common/features/dashboards/selectors";
  import { deleteFileArtifact } from "@rilldata/web-common/features/entity-management/actions";
  import {
    getFileAPIPathFromNameAndType,
    getFilePathFromNameAndType,
  } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import {
    FileArtifactsData,
    fileArtifactsStore,
  } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { SourceModelValidationStatus } from "@rilldata/web-common/features/metrics-views/errors.js";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import { useSourceFileNames } from "@rilldata/web-common/features/sources/selectors";
  import { appScreen } from "@rilldata/web-common/layout/app-store";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import {
    createRuntimeServicePutFile,
    runtimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";
  import { MetricsSourceSelectionError } from "@rilldata/web-local/lib/temp/errors/ErrorMessages.js";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";
  import { behaviourEvent } from "../../metrics/initMetrics";
  import { runtime } from "../../runtime-client/runtime-store";
  import AddAssetButton from "../entity-management/AddAssetButton.svelte";
  import RenameAssetModal from "../entity-management/RenameAssetModal.svelte";

  $: instanceId = $runtime.instanceId;

  $: sourceNames = useSourceFileNames(instanceId);
  $: modelNames = useModelFileNames(instanceId);
  $: dashboardNames = useDashboardFileNames(instanceId);

  const createDashboard = createRuntimeServicePutFile();

  let showMetricsDefs = true;

  let showRenameMetricsDefinitionModal = false;
  let renameMetricsDefName: string | null = null;

  async function getDashboardArtifact(
    instanceId: string,
    metricViewName: string,
  ) {
    const filePath = getFilePathFromNameAndType(
      metricViewName,
      EntityType.MetricsDefinition,
    );
    const resp = await runtimeServiceGetFile(instanceId, filePath);
    const metricYAMLString = resp.blob as string;
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
    const newDashboardName = getName("dashboard", $dashboardNames?.data ?? []);
    await $createDashboard.mutateAsync({
      instanceId,
      path: getFileAPIPathFromNameAndType(
        newDashboardName,
        EntityType.MetricsDefinition,
      ),
      data: {
        blob: "",
        create: true,
        createOnly: true,
      },
    });

    goto(`/dashboard/${newDashboardName}`);
  };

  const editModel = async (dashboardName: string) => {
    await getDashboardArtifact(instanceId, dashboardName);

    const dashboardData = getDashboardData(
      $fileArtifactsStore.entities,
      dashboardName,
    );
    const sourceModelName = dashboardData.jsonRepresentation?.model as string;

    const previousActiveEntity = $appScreen?.type;
    goto(`/model/${sourceModelName}`);
    behaviourEvent.fireNavigationEvent(
      sourceModelName,
      BehaviourEventMedium.Menu,
      MetricsEventSpace.LeftPanel,
      previousActiveEntity,
      MetricsEventScreenName.Model,
    );
  };

  const editMetrics = (dashboardName: string) => {
    goto(`/dashboard/${dashboardName}/edit`);

    const previousActiveEntity = $appScreen?.type;
    behaviourEvent.fireNavigationEvent(
      dashboardName,
      BehaviourEventMedium.Menu,
      MetricsEventSpace.LeftPanel,
      previousActiveEntity,
      MetricsEventScreenName.MetricsDefinition,
    );
  };

  const deleteMetricsDef = async (dashboardName: string) => {
    await getDashboardArtifact(instanceId, dashboardName);

    const dashboardData = getDashboardData(
      $fileArtifactsStore.entities,
      dashboardName,
    );
    await deleteFileArtifact(
      instanceId,
      dashboardName,
      EntityType.MetricsDefinition,
      $dashboardNames?.data ?? [],
    );

    // redirect to model when metric is deleted
    const sourceModelName = dashboardData?.jsonRepresentation?.model as string;
    if ($appScreen?.name === dashboardName) {
      if (sourceModelName) {
        goto(`/model/${sourceModelName}`);

        behaviourEvent.fireNavigationEvent(
          sourceModelName,
          BehaviourEventMedium.Menu,
          MetricsEventSpace.LeftPanel,
          MetricsEventScreenName.MetricsDefinition,
          MetricsEventScreenName.Model,
        );
      } else {
        goto("/");
      }
    }
  };

  const getDashboardData = (
    entities: Record<string, FileArtifactsData>,
    name: string,
  ) => {
    const dashboardPath = getFilePathFromNameAndType(
      name,
      EntityType.MetricsDefinition,
    );
    return entities[dashboardPath];
  };

  $: canAddDashboard = $featureFlags.readOnly === false;

  $: hasSourceAndModelButNoDashboards =
    $sourceNames?.data &&
    $modelNames?.data &&
    $sourceNames?.data?.length > 0 &&
    $modelNames?.data?.length > 0 &&
    $dashboardNames?.data?.length === 0;
</script>

<NavigationHeader bind:show={showMetricsDefs} toggleText="dashboards"
  >{$_('dashboards')}</NavigationHeader
>

{#if showMetricsDefs && $dashboardNames.data}
  <div
    class="pb-3 justify-self-end"
    transition:slide|global={{ duration: LIST_SLIDE_DURATION }}
    id="assets-metrics-list"
  >
    {#each $dashboardNames.data as dashboardName (dashboardName)}
      {@const dashboardData = getDashboardData(
        $fileArtifactsStore.entities,
        dashboardName,
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
            dashboardData?.errors,
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
            {$_('edit-metrics')}
          </MenuItem>
          <Divider />
          <MenuItem
            icon
            on:select={() => openRenameMetricsDefModal(dashboardName)}
          >
            <EditIcon slot="icon" />
            {$_('rename')}</MenuItem
          >
          <MenuItem icon on:select={() => deleteMetricsDef(dashboardName)}>
            <Cancel slot="icon" />
            {$_('delete')}</MenuItem
          >
        </svelte:fragment>
      </NavigationEntry>
    {/each}
    {#if canAddDashboard}
      <AddAssetButton
        id="add-dashboard"
        label={$_('add-dashboard')}
        bold={hasSourceAndModelButNoDashboards ?? false}
        on:click={() => dispatchAddEmptyMetricsDef()}
      />
    {/if}
  </div>
  {#if showRenameMetricsDefinitionModal && renameMetricsDefName}
    <RenameAssetModal
      entityType={EntityType.MetricsDefinition}
      closeModal={() => (showRenameMetricsDefinitionModal = false)}
      currentAssetName={renameMetricsDefName}
    />
  {/if}
{/if}
