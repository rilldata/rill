<script lang="ts">
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { SourceModelValidationStatus } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import { MetricsSourceSelectionError } from "$common/errors/ErrorMessages";
  import { waitUntil } from "$common/utils/waitUtils";
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
  import { default as Explore } from "$lib/components/icons/Explore.svelte";
  import MetricsIcon from "$lib/components/icons/Metrics.svelte";
  import Model from "$lib/components/icons/Model.svelte";
  import Divider from "$lib/components/menu/Divider.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";
  import MetricsDefinitionSummary from "$lib/components/metrics-definition/MetricsDefinitionSummary.svelte";
  import {
    createMetricsDefsAndFocusApi,
    deleteMetricsDefsApi,
    fetchManyMetricsDefsApi,
    validateSelectedSources,
  } from "$lib/redux-store/metrics-definition/metrics-definition-apis";
  import { getAllMetricsDefinitionsReadable } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import { store } from "$lib/redux-store/store-root";
  import { getContext, onMount } from "svelte";
  import { slide } from "svelte/transition";

  const metricsDefinitions = getAllMetricsDefinitionsReadable();
  const appStore = getContext("rill:app:store") as ApplicationStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let showMetricsDefs = true;
  const dispatchAddEmptyMetricsDef = () => {
    if (!showMetricsDefs) {
      showMetricsDefs = true;
    }
    store.dispatch(createMetricsDefsAndFocusApi());
  };

  const dispatchSetMetricsDefActive = (id: string) => {
    dataModelerService.dispatch("setActiveAsset", [
      EntityType.MetricsDefinition,
      id,
    ]);
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
        store.dispatch(
          validateSelectedSources({
            id: metricsDefinition.id,
            derivedModelState: $derivedModelStore,
          })
        )
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
    tooltipText={"metrics"}
    bind:active={showMetricsDefs}
  >
    <h4 class="flex flex-row items-center gap-x-2">
      <MetricsIcon size="16px" /> Metrics
    </h4>
  </CollapsibleSectionTitle>
  <ContextButton
    id={"create-model-button"}
    tooltipText="create a new model"
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
              dataModelerService.dispatch("setActiveAsset", [
                EntityType.Model,
                metricsDef.sourceModelId,
              ]);
            }}
          >
            <svelte:fragment slot="icon">
              <Model />
            </svelte:fragment>
            see model for metrics
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
              dataModelerService.dispatch("setActiveAsset", [
                EntityType.MetricsExplorer,
                metricsDef.id,
              ]);
            }}
          >
            <svelte:fragment slot="icon">
              <Explore />
            </svelte:fragment>
            go to dashboard
          </MenuItem>
          <Divider />
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
{/if}
