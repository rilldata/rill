<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import type { PersistentModelEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  import { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/common/metrics-service/MetricsTypes";
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";
  import {
    ApplicationStore,
    dataModelerService,
  } from "../../../application-state-stores/application-store";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "../../../application-state-stores/model-stores";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import CollapsibleSectionTitle from "../../CollapsibleSectionTitle.svelte";
  import ColumnProfileNavEntry from "../../column-profile/ColumnProfileNavEntry.svelte";
  import ContextButton from "../../column-profile/ContextButton.svelte";
  import AddIcon from "../../icons/Add.svelte";
  import ModelIcon from "../../icons/Model.svelte";
  import NavigationEntry from "../NavigationEntry.svelte";
  import ModelMenuItems from "./ModelMenuItems.svelte";
  import ModelTooltip from "./ModelTooltip.svelte";

  const store = getContext("rill:app:store") as ApplicationStore;
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;
  const applicationStore = getContext("rill:app:store") as ApplicationStore;

  let showModels = true;

  let showRenameModelModal = false;
  let renameModelID = null;
  let renameModelName = null;

  const viewModel = (id: string) => {
    goto(`/model/${id}`);

    if (id != activeEntityID) {
      const previousActiveEntity = $store?.activeEntity?.type;
      navigationEvent.fireEvent(
        id,
        BehaviourEventMedium.AssetName,
        MetricsEventSpace.LeftPanel,
        EntityTypeToScreenMap[previousActiveEntity],
        MetricsEventScreenName.Model
      );
    }
  };

  async function addModel() {
    let response = await dataModelerService.dispatch("addModel", [{}]);
    goto(`/model/${response.id}`);
    // if the models are not visible in the assets list, show them.
    if (!showModels) {
      x;
      showModels = true;
    }
  }

  // type Coll

  let persistentModelEntities: PersistentModelEntity[] = [];
  $: activeEntityID = $store?.activeEntity?.id;
  $: persistentModelEntities =
    ($persistentModelStore && $persistentModelStore.entities) || [];

  $: availableModels = persistentModelEntities.map((query) => {
    let derivedModel = $derivedModelStore.entities.find(
      (model) => model.id === query.id
    );

    return {
      id: query.id,
      modelName: query.name,
      tableSummaryProps: {
        name: query.name,
        cardinality: derivedModel?.cardinality ?? 0,
        profile: derivedModel?.profile ?? [],
        head: derivedModel?.preview ?? [],
        sizeInBytes: derivedModel?.sizeInBytes ?? 0,
        active: query?.id === activeEntityID,
      },
    };
  });
</script>

<div
  class="pl-4 pb-3 pr-3 grid justify-between"
  style="grid-template-columns: auto max-content;"
  out:slide={{ duration: 200 }}
>
  <CollapsibleSectionTitle tooltipText={"models"} bind:active={showModels}>
    <h4 class="flex flex-row items-center gap-x-2">
      <ModelIcon size="16px" /> Models
    </h4>
  </CollapsibleSectionTitle>
  <ContextButton
    id={"create-model-button"}
    tooltipText="create a new model"
    on:click={addModel}
    width={24}
    height={24}
    rounded
  >
    <AddIcon />
  </ContextButton>
</div>
{#if showModels}
  <div
    class="pb-6 justify-self-end"
    transition:slide={{ duration: 200 }}
    id="assets-model-list"
  >
    {#each availableModels as { id, modelName, tableSummaryProps }, i (id)}
      {@const derivedModel = $derivedModelStore.entities.find(
        (t) => t["id"] === id
      )}
      <NavigationEntry
        name={modelName}
        href={`/model/${id}`}
        open={$page.url.pathname === `/model/${id}`}
      >
        <svelte:fragment slot="more">
          <ColumnProfileNavEntry
            indentLevel={1}
            cardinality={tableSummaryProps.cardinality}
            profile={tableSummaryProps.profile}
            head={tableSummaryProps.head}
            entityId={id}
          />
        </svelte:fragment>

        <svelte:fragment slot="tooltip-content">
          <ModelTooltip {modelName} />
        </svelte:fragment>

        <svelte:fragment slot="menu-items">
          <ModelMenuItems modelID={derivedModel.id} />
        </svelte:fragment>
      </NavigationEntry>
    {/each}
  </div>
{/if}
