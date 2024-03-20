<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import MetricsIcon from "@rilldata/web-common/components/icons/Metrics.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import {
    useDashboardFileNames,
    useDashboards,
  } from "@rilldata/web-common/features/dashboards/selectors";
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
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import NavigationMenuSeparator from "@rilldata/web-common/layout/navigation/NavigationMenuSeparator.svelte";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import type { V1ReconcileError } from "@rilldata/web-common/runtime-client";
  import {
    createRuntimeServicePutFile,
    runtimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";
  import { flip } from "svelte/animate";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";
  import { behaviourEvent } from "../../metrics/initMetrics";
  import { runtime } from "../../runtime-client/runtime-store";
  import AddAssetButton from "../entity-management/AddAssetButton.svelte";
  import RenameAssetModal from "../entity-management/RenameAssetModal.svelte";
  import { ResourceKind } from "../entity-management/resource-selectors";

  $: instanceId = $runtime.instanceId;

  $: sourceNames = useSourceFileNames(instanceId);
  $: modelNames = useModelFileNames(instanceId);
  $: dashboardNames = useDashboardFileNames(instanceId);
  $: dashboards = useDashboards(instanceId);

  const MetricsSourceSelectionError = (
    errors: Array<V1ReconcileError> | undefined,
  ): string => {
    return (
      errors?.find((error) => error?.propertyPath?.length === 0)?.message ?? ""
    );
  };

  const createDashboard = createRuntimeServicePutFile();

  const { readOnly } = featureFlags;

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

    await goto(`/dashboard/${newDashboardName}`);
  };

  /**
   * Get the name of a dashboard's underlying model (if any).
   * Note that not all dashboards have an underlying model.
   * Some dashboards are underpinned by a source/table.
   */
  function getModelForDashboard(dashboardName: string) {
    const dashboard = $dashboards.data?.filter(
      (dashboard) => dashboard.meta?.name?.name === dashboardName,
    )[0];
    const modelRef = dashboard?.meta?.refs?.filter(
      (ref) => ref?.kind === ResourceKind.Model,
    )[0];
    if (!modelRef) return "";
    return modelRef?.name;
  }

  const editModel = async (modelName: string) => {
    const previousActiveEntity = $appScreen?.type;
    await goto(`/model/${modelName}`);
    await behaviourEvent.fireNavigationEvent(
      modelName,
      BehaviourEventMedium.Menu,
      MetricsEventSpace.LeftPanel,
      previousActiveEntity,
      MetricsEventScreenName.Model,
    );
  };

  const editMetrics = async (dashboardName: string) => {
    await goto(`/dashboard/${dashboardName}/edit`);

    const previousActiveEntity = $appScreen?.type;
    await behaviourEvent.fireNavigationEvent(
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
        await goto(`/model/${sourceModelName}`);

        await behaviourEvent.fireNavigationEvent(
          sourceModelName,
          BehaviourEventMedium.Menu,
          MetricsEventSpace.LeftPanel,
          MetricsEventScreenName.MetricsDefinition,
          MetricsEventScreenName.Model,
        );
      } else {
        await goto("/");
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

  $: canAddDashboard = $readOnly === false;

  $: hasSourceAndModelButNoDashboards =
    $sourceNames?.data &&
    $modelNames?.data &&
    $sourceNames?.data?.length > 0 &&
    $modelNames?.data?.length > 0 &&
    $dashboardNames?.data?.length === 0;
</script>

<div class="h-fit flex flex-col">
  <NavigationHeader bind:show={showMetricsDefs}>Dashboards</NavigationHeader>
  {#if showMetricsDefs}
    <ol transition:slide={{ duration }} id="assets-metrics-list">
      {#if $dashboardNames.data}
        {#each $dashboardNames.data as dashboardName (dashboardName)}
          {@const dashboardData = getDashboardData(
            $fileArtifactsStore.entities,
            dashboardName,
          )}
          {@const modelForDashboard = getModelForDashboard(dashboardName)}
          <li animate:flip={{ duration }} aria-label={dashboardName}>
            <NavigationEntry
              showContextMenu={!$readOnly}
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
                {#if modelForDashboard}
                  <NavigationMenuItem
                    disabled={hasSourceError}
                    on:click={() => editModel(modelForDashboard)}
                  >
                    <Model slot="icon" />
                    Edit model
                    <svelte:fragment slot="description">
                      {#if hasSourceError}
                        {selectionError}
                      {/if}
                    </svelte:fragment>
                  </NavigationMenuItem>
                {/if}
                <NavigationMenuItem
                  disabled={hasSourceError}
                  on:click={() => editMetrics(dashboardName)}
                >
                  <MetricsIcon slot="icon" />
                  Edit metrics
                </NavigationMenuItem>
                <NavigationMenuSeparator />
                <NavigationMenuItem
                  on:click={() => openRenameMetricsDefModal(dashboardName)}
                >
                  <EditIcon slot="icon" />
                  Rename...
                </NavigationMenuItem>
                <NavigationMenuItem
                  on:click={() => deleteMetricsDef(dashboardName)}
                >
                  <Cancel slot="icon" />
                  Delete
                </NavigationMenuItem>
              </svelte:fragment>
            </NavigationEntry>
          </li>
        {/each}
        {#if canAddDashboard}
          <AddAssetButton
            id="add-dashboard"
            label="Add dashboard"
            bold={hasSourceAndModelButNoDashboards ?? false}
            on:click={() => dispatchAddEmptyMetricsDef()}
          />
        {/if}
      {/if}
    </ol>
  {/if}
</div>

{#if showRenameMetricsDefinitionModal && renameMetricsDefName}
  <RenameAssetModal
    entityType={EntityType.MetricsDefinition}
    closeModal={() => (showRenameMetricsDefinitionModal = false)}
    currentAssetName={renameMetricsDefName}
  />
{/if}
