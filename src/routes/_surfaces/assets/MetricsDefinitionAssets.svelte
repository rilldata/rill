<script lang="ts">
  import { slide } from "svelte/transition";
  import ModelIcon from "$lib/components/icons/Code.svelte";
  import AddIcon from "$lib/components/icons/Add.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";
  import { store, reduxReadable } from "$lib/redux-store/store-root";
  import CollapsibleMetricsDefinitionSummary from "$lib/components/metrics-definition/CollapsibleMetricsDefinitionSummary.svelte";
  import { onMount } from "svelte";
  import {
    createMetricsDefsApi,
    fetchManyMetricsDefsApi,
    manyMetricsDefsSelector,
  } from "$lib/redux-store/metrics-definition-slice";

  $: metricsDefinitions = manyMetricsDefsSelector($reduxReadable);

  let showMetricsDefs = true;
  const dispatch_addEmptyMetricsDef = () => {
    if (!showMetricsDefs) {
      showMetricsDefs = true;
    }
    store.dispatch(createMetricsDefsApi());
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
    on:click={dispatch_addEmptyMetricsDef}
  >
    <AddIcon />
  </ContextButton>
</div>
{#if showMetricsDefs && metricsDefinitions}
  <div
    class="pb-6 justify-self-end"
    transition:slide={{ duration: 200 }}
    id="assets-model-list"
  >
    {#each metricsDefinitions as { id } (id)}
      <CollapsibleMetricsDefinitionSummary metricsDefId={id} indentLevel={1} />
    {/each}
  </div>
{/if}
