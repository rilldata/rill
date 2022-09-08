<script lang="ts">
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { SourceModelValidationStatus } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import { MetricsSourceSelectionError } from "$common/errors/ErrorMessages";
  import { waitUntil } from "$common/utils/waitUtils";
  import { BehaviourEventMedium } from "$common/metrics-service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "$common/metrics-service/MetricsTypes";
  import {
    ApplicationStore,
    dataModelerService,
  } from "$lib/application-state-stores/application-store";
  import type { DerivedModelStore } from "$lib/application-state-stores/model-stores";
  import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";
  import CollapsibleTableSummary from "$lib/components/column-profile/CollapsibleTableSummary.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import AddIcon from "$lib/components/icons/Add.svelte";
  import Cancel from "$lib/components/icons/Cancel.svelte";
  import EditIcon from "$lib/components/icons/EditIcon.svelte";
  import { default as Explore } from "$lib/components/icons/Explore.svelte";
  import MetricsIcon from "$lib/components/icons/Metrics.svelte";
  import Model from "$lib/components/icons/Model.svelte";
  import { Divider, MenuItem } from "$lib/components/menu";
  import MetricsDefinitionSummary from "$lib/components/metrics-definition/MetricsDefinitionSummary.svelte";
  import RenameEntityModal from "$lib/components/modal/RenameEntityModal.svelte";
  import {
    createMetricsDefsAndFocusApi,
    deleteMetricsDefsApi,
    fetchManyMetricsDefsApi,
    validateSelectedSources,
  } from "$lib/redux-store/metrics-definition/metrics-definition-apis";
  import { getAllMetricsDefinitionsReadable } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import { store } from "$lib/redux-store/store-root";
  import { navigationEvent } from "$lib/metrics/initMetrics";
  import { getContext, onMount } from "svelte";
  import { slide } from "svelte/transition";

  const metricsDefinitions = getAllMetricsDefinitionsReadable();
  const appStore = getContext("rill:app:store") as ApplicationStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

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

  const dispatchAddEmptyMetricsDef = () => {
    if (!showMetricsDefs) {
      showMetricsDefs = true;
    }
    store.dispatch(createMetricsDefsAndFocusApi());
  };

  const editModel = (sourceModelId: string) => {
    const previousActiveEntity = $appStore?.activeEntity?.type;

    dataModelerService.dispatch("setActiveAsset", [
      EntityType.Model,
      sourceModelId,
    ]);

    navigationEvent.fireEvent(
      sourceModelId,
      BehaviourEventMedium.Menu,
      MetricsEventSpace.LeftPanel,
      EntityTypeToScreenMap[previousActiveEntity],
      MetricsEventScreenName.Model
    );
  };

  const editMetrics = (metricsId: string) => {
    const previousActiveEntity = $appStore?.activeEntity?.type;

    dataModelerService.dispatch("setActiveAsset", [
      EntityType.MetricsDefinition,
      metricsId,
    ]);

    navigationEvent.fireEvent(
      metricsId,
      BehaviourEventMedium.Menu,
      MetricsEventSpace.LeftPanel,
      EntityTypeToScreenMap[previousActiveEntity],
      MetricsEventScreenName.MetricsDefinition
    );
  };

  const dispatchSetMetricsDefActive = (id: string) => {
    const previousActiveEntity = $appStore?.activeEntity?.type;

    dataModelerService.dispatch("setActiveAsset", [
      EntityType.MetricsExplorer,
      id,
    ]);

    navigationEvent.fireEvent(
      id,
      BehaviourEventMedium.AssetName,
      MetricsEventSpace.LeftPanel,
      EntityTypeToScreenMap[previousActiveEntity],
      MetricsEventScreenName.Dashboard
    );
  };

  const dispatchDeleteMetricsDef = (id: string) => {
    store.dispatch(deleteMetricsDefsApi(id));
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

<div
  class="pl-4 pb-3 pr-4 grid justify-between"
  style="grid-template-columns: auto max-content;"
  out:slide={{ duration: 200 }}
>
  <CollapsibleSectionTitle
    tooltipText={"dashboards"}
    bind:active={showMetricsDefs}
  >
    <h4 class="flex flex-row items-center gap-x-2">
      <Explore size="16px" /> Dashboards
    </h4>
  </CollapsibleSectionTitle>
  <ContextButton
    id={"create-dashboard-button"}
    tooltipText="create a new dashboard"
    on:click={dispatchAddEmptyMetricsDef}
  >
    <AddIcon />
  </ContextButton>
</div>
{#if showMetricsDefs && $metricsDefinitions}
  <div
    class="pb-6 justify-self-end"
    transition:slide={{ duration: 200 }}
    id="assets-model-list"
  >
    {#each $metricsDefinitions as metricsDef (metricsDef.id)}
      <CollapsibleTableSummary
        entityType={EntityType.MetricsDefinition}
        name={metricsDef.metricDefLabel ?? ""}
        active={$appStore?.activeEntity?.id === metricsDef.id}
        showRows={false}
        on:select={() => dispatchSetMetricsDefActive(metricsDef.id)}
        on:delete={() => dispatchDeleteMetricsDef(metricsDef.id)}
        notExpandable={true}
      >
        <svelte:fragment slot="summary" let:containerWidth>
          <MetricsDefinitionSummary indentLevel={1} {containerWidth} />
        </svelte:fragment>

        <svelte:fragment slot="menu-items">
          {@const selectionError = MetricsSourceSelectionError(metricsDef)}
          {@const hasSourceError =
            selectionError !== SourceModelValidationStatus.OK &&
            selectionError !== ""}
          <MenuItem
            icon
            disabled={hasSourceError}
            on:select={() => {
              editModel(metricsDef.sourceModelId);
            }}
          >
            <svelte:fragment slot="icon">
              <Model />
            </svelte:fragment>
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
            on:select={() => {
              editMetrics(metricsDef.id);
            }}
          >
            <svelte:fragment slot="icon">
              <MetricsIcon />
            </svelte:fragment>
            edit metrics
          </MenuItem>
          <Divider />
          <MenuItem
            icon
            on:select={() =>
              openRenameMetricsDefModal(
                metricsDef.id,
                metricsDef.metricDefLabel
              )}
          >
            <svelte:fragment slot="icon">
              <EditIcon />
            </svelte:fragment>
            rename...</MenuItem
          >
          <MenuItem
            icon
            on:select={() => dispatchDeleteMetricsDef(metricsDef.id)}
          >
            <svelte:fragment slot="icon">
              <Cancel />
            </svelte:fragment>
            delete</MenuItem
          >
        </svelte:fragment>
      </CollapsibleTableSummary>
    {/each}
  </div>
  {#if showRenameMetricsDefinitionModal}
    <RenameEntityModal
      entityType={EntityType.MetricsDefinition}
      closeModal={() => (showRenameMetricsDefinitionModal = false)}
      entityId={renameMetricsDefId}
      currentEntityName={renameMetricsDefName}
    />
  {/if}
{/if}
