<script lang="ts">
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";

  // import type { ApplicationStore } from "$lib/application-state-stores/application-store";

  import ModelIcon from "$lib/components/icons/Code.svelte";
  import AddIcon from "$lib/components/icons/Add.svelte";
  import CollapsibleTableSummary from "$lib/components/column-profile/CollapsibleTableSummary.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";

  import { addEmptyMetricsDef } from "$lib/redux-store/metrics-definition/metrics-definition-slice";
  import { store, reduxReadable } from "$lib/redux-store/store-root";
  import CollapsibleMetricsDefinitionSummary from "$lib/components/metrics-definition/CollapsibleMetricsDefinitionSummary.svelte";

  // import { dataModelerService } from "$lib/application-state-stores/application-store";

  // import type {
  //   DerivedModelStore,
  //   PersistentModelStore,
  // } from "$lib/application-state-stores/model-stores";
  // import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  // import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

  // const store = getContext("rill:app:store") as ApplicationStore;
  // const persistentModelStore = getContext(
  //   "rill:app:persistent-model-store"
  // ) as PersistentModelStore;
  // const derivedModelStore = getContext(
  //   "rill:app:derived-model-store"
  // ) as DerivedModelStore;

  // let
  $: metricsDefIds =
    $reduxReadable !== undefined ? $reduxReadable.metricsDefinition.ids : [];

  // let metricsDefIds2 = derived(reduxReadable, (r) => r.metricsDefinition.ids);

  // $: console.log($metricsDefIds2);
  $: console.log(metricsDefIds);

  let showMetricsDefs = true;
  const dispatch_addEmptyMetricsDef = () => {
    if (!showMetricsDefs) {
      showMetricsDefs = true;
    }
    store.dispatch(addEmptyMetricsDef());
  };
  // const dispatch_addEmptyMetricsDef = () => undefined;
  // async function addModel() {
  //   // create the new model.
  //   let response = await dataModelerService.dispatch("addModel", [{}]);
  //   // change the active asset to the new model.
  //   dataModelerService.dispatch("setActiveAsset", [
  //     EntityType.Model,
  //     response.id,
  //   ]);
  //   // if the models are not visible in the assets list, show them.
  //   if (!showMetricsDefs) {
  //     showMetricsDefs = true;
  //   }
  // }

  // // type Coll

  // let persistentModelEntities: PersistentModelEntity[] = [];
  // $: activeEntityID = $store?.activeEntity?.id;
  // $: persistentModelEntities =
  //   ($persistentModelStore && $persistentModelStore.entities) || [];

  // $: availableModels = persistentModelEntities.map((query) => {
  //   let derivedModel = $derivedModelStore.entities.find(
  //     (model) => model.id === query.id
  //   );

  //   return {
  //     id: query.id,
  //     tableSummaryProps: {
  //       name: query.name,
  //       cardinality: derivedModel?.cardinality ?? 0,
  //       profile: derivedModel?.profile ?? [],
  //       head: derivedModel?.preview ?? [],
  //       sizeInBytes: derivedModel?.sizeInBytes ?? 0,
  //       emphasizeTitle: query?.id === activeEntityID,
  //     },
  //   };
  // });
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
{#if showMetricsDefs}
  <div
    class="pb-6 justify-self-end"
    transition:slide={{ duration: 200 }}
    id="assets-model-list"
  >
    {#each metricsDefIds as id}
      <CollapsibleMetricsDefinitionSummary metricsDefId={id} />
      <!-- {id}{"   "} -->
    {/each}

    <!-- {#each availableModels as { id, tableSummaryProps }, i (id)}
      <CollapsibleTableSummary
        on:select={() => {
          dataModelerService.dispatch("setActiveAsset", [EntityType.Model, id]);
        }}
        on:delete={() => {
          dataModelerService.dispatch("deleteModel", [id]);
        }}
        indentLevel={1}
        {...tableSummaryProps}
      />
    {/each} -->
  </div>
{/if}
