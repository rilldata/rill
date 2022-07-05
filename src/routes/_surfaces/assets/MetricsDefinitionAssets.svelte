<script lang="ts">
  import { slide } from "svelte/transition";
  import ModelIcon from "$lib/components/icons/Code.svelte";
  import AddIcon from "$lib/components/icons/Add.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";
  import { store } from "$lib/redux-store/store-root";
  import { getContext, onMount } from "svelte";
  import {
    createMetricsDefsApi,
    deleteMetricsDefsApi,
    fetchManyMetricsDefsApi,
  } from "$lib/redux-store/metrics-definition/metrics-definition-apis";
  import { getAllMetricsDefinitionsReadable } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import CollapsibleTableSummary from "$lib/components/column-profile/CollapsibleTableSummary.svelte";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import MetricsDefinitionSummary from "$lib/components/metrics-definition/MetricsDefinitionSummary.svelte";
  import {
    ApplicationStore,
    dataModelerService,
  } from "$lib/application-state-stores/application-store";
  import ExpandCaret from "$lib/components/icons/ExpandCaret.svelte";

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
    tooltipText={"metrics definitions"}
    bind:active={showMetricsDefs}
  >
    <h4 class="flex flex-row items-center gap-x-2">
      <ModelIcon size="16px" /> Metrics Definitions
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
        emphasizeTitle={$appStore?.activeEntity?.id === id}
        showRows={false}
        on:select={() => dispatchSetMetricsDefActive(id)}
        on:delete={() => dispatchDeleteMetricsDef(id)}
      >
        <svelte:fragment slot="summary" let:containerWidth>
          <MetricsDefinitionSummary indentLevel={1} {containerWidth} />
        </svelte:fragment>
        <span class="self-center" slot="header-buttons">
          <ContextButton
            {id}
            tooltipText="expand"
            location="left"
            on:click={() => {
              dataModelerService.dispatch("setActiveAsset", [
                EntityType.MetricsLeaderboard,
                id,
              ]);
            }}><ExpandCaret /></ContextButton
          >
        </span>
      </CollapsibleTableSummary>
    {/each}
  </div>
{/if}
