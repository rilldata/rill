<script lang="ts">
  import { goto } from "$app/navigation";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { SourceModelValidationStatus } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import { MetricsSourceSelectionError } from "@rilldata/web-local/common/errors/ErrorMessages";
  import { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/common/metrics-service/MetricsTypes";
  import { getNextEntityId } from "@rilldata/web-local/common/utils/getNextEntityId";
  import { waitUntil } from "@rilldata/web-local/common/utils/waitUtils";
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
  import CollapsibleSectionTitle from "../../CollapsibleSectionTitle.svelte";
  import CollapsibleTableSummary from "../../column-profile/CollapsibleTableSummary.svelte";
  import ContextButton from "../../column-profile/ContextButton.svelte";
  import AddIcon from "../../icons/Add.svelte";
  import Cancel from "../../icons/Cancel.svelte";
  import EditIcon from "../../icons/EditIcon.svelte";
  import { default as Explore } from "../../icons/Explore.svelte";
  import MetricsIcon from "../../icons/Metrics.svelte";
  import Model from "../../icons/Model.svelte";
  import { Divider, MenuItem } from "../../menu";
  import MetricsDefinitionSummary from "../../metrics-definition/MetricsDefinitionSummary.svelte";
  import RenameEntityModal from "../RenameEntityModal.svelte";

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

  const dispatchAddEmptyMetricsDef = async () => {
    if (!showMetricsDefs) {
      showMetricsDefs = true;
    }
    await store.dispatch(createMetricsDefsAndFocusApi());
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

  const dispatchSetMetricsDefActive = (id: string) => {
    goto(`/dashboard/${id}`);

    const previousActiveEntity = $appStore?.activeEntity?.type;
    navigationEvent.fireEvent(
      id,
      BehaviourEventMedium.AssetName,
      MetricsEventSpace.LeftPanel,
      EntityTypeToScreenMap[previousActiveEntity],
      MetricsEventScreenName.Dashboard
    );
  };

  const deleteMetricsDef = (id: string) => {
    const nextMetricsDefId = getNextEntityId($metricsDefinitions, id);

    if (nextMetricsDefId) {
      goto(`/dashboard/${nextMetricsDefId}`);
    } else {
      goto("/");
    }

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
    id="assets-metrics-list"
  >
    {#each $metricsDefinitions as metricsDef (metricsDef.id)}
      <CollapsibleTableSummary
        entityType={EntityType.MetricsDefinition}
        name={metricsDef.metricDefLabel ?? ""}
        active={$appStore?.activeEntity?.id === metricsDef.id}
        showRows={false}
        on:select={() => dispatchSetMetricsDefActive(metricsDef.id)}
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
          <Divider />
          <MenuItem
            icon
            on:select={() =>
              openRenameMetricsDefModal(
                metricsDef.id,
                metricsDef.metricDefLabel
              )}
          >
            <EditIcon slot="icon" />
            rename...</MenuItem
          >
          <MenuItem icon on:select={() => deleteMetricsDef(metricsDef.id)}>
            <Cancel slot="icon" />
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
