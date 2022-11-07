<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import {
    MetricsDefinitionEntity,
    SourceModelValidationStatus,
  } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import { MetricsSourceSelectionError } from "@rilldata/web-local/common/errors/ErrorMessages";
  import { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/common/metrics-service/MetricsTypes";
  import { waitUntil } from "@rilldata/web-local/common/utils/waitUtils";
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
  import CollapsibleSectionTitle from "../../CollapsibleSectionTitle.svelte";
  import ContextButton from "../../column-profile/ContextButton.svelte";
  import AddIcon from "../../icons/Add.svelte";
  import Cancel from "../../icons/Cancel.svelte";
  import EditIcon from "../../icons/EditIcon.svelte";
  import { default as Explore } from "../../icons/Explore.svelte";
  import MetricsIcon from "../../icons/Metrics.svelte";
  import Model from "../../icons/Model.svelte";
  import { Divider, MenuItem } from "../../menu";
  import MetricsDefinitionSummary from "../../metrics-definition/MetricsDefinitionSummary.svelte";
  import RenameAssetModal from "../RenameAssetModal.svelte";

  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import NavigationEntry from "../NavigationEntry.svelte";

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

  const deleteMetricsDef = (metricsDef: MetricsDefinitionEntity) => {
    const sourceModelId = metricsDef.sourceModelId;

    notificationStore.send({
      message: `Dashboard "${metricsDef.metricDefLabel}" deleted`,
    });
    store.dispatch(deleteMetricsDefsApi(metricsDef.id));

    if (
      ($applicationStore.activeEntity.type === EntityType.MetricsDefinition ||
        $applicationStore.activeEntity.type === EntityType.MetricsExplorer) &&
      $applicationStore.activeEntity.id === metricsDef.id
    ) {
      if (sourceModelId) {
        goto(`/model/${sourceModelId}`);
      } else {
        goto("/");
      }
    }
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
  class="pl-4 pb-3 pr-3 grid justify-between"
  style="grid-template-columns: auto max-content;"
  out:slide={{ duration: LIST_SLIDE_DURATION }}
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
    width={24}
    height={24}
    rounded
    on:click={dispatchAddEmptyMetricsDef}
  >
    <AddIcon />
  </ContextButton>
</div>
{#if showMetricsDefs && $metricsDefinitions}
  <div
    class="pb-6 justify-self-end"
    transition:slide={{ duration: LIST_SLIDE_DURATION }}
    id="assets-metrics-list"
  >
    {#each $metricsDefinitions as metricsDef (metricsDef.id)}
      <NavigationEntry
        notExpandable={true}
        name={metricsDef.metricDefLabel}
        href={`/dashboard/${metricsDef.id}`}
        open={$page.url.pathname === `/dashboard/${metricsDef.id}`}
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
          <MenuItem icon on:select={() => deleteMetricsDef(metricsDef)}>
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
