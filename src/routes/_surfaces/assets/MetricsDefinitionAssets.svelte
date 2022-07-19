<script lang="ts">
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import {
    ApplicationStore,
    dataModelerService,
  } from "$lib/application-state-stores/application-store";
  import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";
  import CollapsibleTableSummary from "$lib/components/column-profile/CollapsibleTableSummary.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import AddIcon from "$lib/components/icons/Add.svelte";
  import ExploreIcon from "$lib/components/icons/Explore.svelte";
  import MetricsIcon from "$lib/components/icons/Metrics.svelte";
  import MetricsDefinitionSummary from "$lib/components/metrics-definition/MetricsDefinitionSummary.svelte";
  import {
    createMetricsDefsApi,
    deleteMetricsDefsApi,
    fetchManyMetricsDefsApi,
  } from "$lib/redux-store/metrics-definition/metrics-definition-apis";
  import { getAllMetricsDefinitionsReadable } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import { store } from "$lib/redux-store/store-root";
  import { getContext, onMount } from "svelte";
  import { slide } from "svelte/transition";

  const metricsDefinitions = getAllMetricsDefinitionsReadable();
  const appStore = getContext("rill:app:store") as ApplicationStore;

  let showMetricsDefs = true;
  const dispatchAddEmptyMetricsDef = () => {
    if (!showMetricsDefs) {
      showMetricsDefs = true;
    }
    store.dispatch(createMetricsDefsApi());
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
    store.dispatch(fetchManyMetricsDefsApi());
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
    {#each $metricsDefinitions as { id, metricDefLabel } (id)}
      <CollapsibleTableSummary
        entityType={EntityType.MetricsDefinition}
        name={metricDefLabel ?? ""}
        active={$appStore?.activeEntity?.id === id}
        showRows={false}
        on:select={() => dispatchSetMetricsDefActive(id)}
        on:delete={() => dispatchDeleteMetricsDef(id)}
        notExpandable={true}
      >
        <svelte:fragment slot="summary" let:containerWidth>
          <MetricsDefinitionSummary indentLevel={1} {containerWidth} />
        </svelte:fragment>
        <span class="self-center" slot="header-buttons">
          <ContextButton
            {id}
            tooltipText="explore metrics"
            location="left"
            on:click={() => {
              dataModelerService.dispatch("setActiveAsset", [
                EntityType.MetricsExplore,
                id,
              ]);
            }}><ExploreIcon /></ContextButton
          >
        </span>
      </CollapsibleTableSummary>
    {/each}
  </div>
{/if}
