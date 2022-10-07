<script lang="ts">
  import { goto } from "$app/navigation";
  import type { DerivedModelEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type { PersistentModelEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  import { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/common/metrics-service/MetricsTypes";
  import { getNextEntityId } from "@rilldata/web-local/common/utils/getNextEntityId";
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";
  import {
    ApplicationStore,
    dataModelerService,
  } from "../../application-state-stores/application-store";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "../../application-state-stores/model-stores";
  import { navigationEvent } from "../../metrics/initMetrics";
  import { deleteModelApi } from "../../redux-store/model/model-apis";
  import { autoCreateMetricsDefinitionForModel } from "../../redux-store/source/source-apis";
  import {
    derivedProfileEntityHasTimestampColumn,
    selectTimestampColumnFromProfileEntity,
  } from "../../redux-store/source/source-selectors";
  import CollapsibleSectionTitle from "../CollapsibleSectionTitle.svelte";
  import CollapsibleTableSummary from "../column-profile/CollapsibleTableSummary.svelte";
  import ColumnProfileNavEntry from "../column-profile/ColumnProfileNavEntry.svelte";
  import ContextButton from "../column-profile/ContextButton.svelte";
  import AddIcon from "../icons/Add.svelte";
  import Cancel from "../icons/Cancel.svelte";
  import EditIcon from "../icons/EditIcon.svelte";
  import Explore from "../icons/Explore.svelte";
  import ModelIcon from "../icons/Model.svelte";
  import { Divider, MenuItem } from "../menu";
  import RenameEntityModal from "../modal/RenameEntityModal.svelte";

  const store = getContext("rill:app:store") as ApplicationStore;
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

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

  const deleteModel = (id: string) => {
    const nextModelId = getNextEntityId($persistentModelStore.entities, id);

    if (nextModelId) {
      goto(`/model/${nextModelId}`);
    } else {
      goto("/");
    }

    deleteModelApi(id);
  };

  const quickStartMetrics = (derivedModel: DerivedModelEntity) => {
    const previousActiveEntity = $store?.activeEntity?.type;

    autoCreateMetricsDefinitionForModel(
      $persistentModelStore.entities.find(
        (model) => model.id === derivedModel.id
      ).tableName,
      derivedModel.id,
      selectTimestampColumnFromProfileEntity(derivedModel)[0].name
    ).then((createdMetricsId) => {
      navigationEvent.fireEvent(
        createdMetricsId,
        BehaviourEventMedium.Menu,
        MetricsEventSpace.LeftPanel,
        EntityTypeToScreenMap[previousActiveEntity],
        MetricsEventScreenName.Dashboard
      );
    });
  };

  const openRenameModelModal = (modelID: string, modelName: string) => {
    showRenameModelModal = true;
    renameModelID = modelID;
    renameModelName = modelName;
  };

  async function addModel() {
    let response = await dataModelerService.dispatch("addModel", [{}]);
    goto(`/model/${response.id}`);
    // if the models are not visible in the assets list, show them.
    if (!showModels) {
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
  class="pl-4 pb-3 pr-4 grid justify-between"
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
      <CollapsibleTableSummary
        entityType={EntityType.Model}
        on:select={() => viewModel(id)}
        cardinality={tableSummaryProps.cardinality}
        name={tableSummaryProps.name.split(".")[0]}
        sizeInBytes={tableSummaryProps.sizeInBytes}
        active={tableSummaryProps.active}
      >
        <svelte:fragment slot="summary" let:containerWidth>
          <ColumnProfileNavEntry
            indentLevel={1}
            {containerWidth}
            cardinality={tableSummaryProps.cardinality}
            profile={tableSummaryProps.profile}
            head={tableSummaryProps.head}
            entityId={id}
          />
        </svelte:fragment>
        <svelte:fragment slot="menu-items">
          <MenuItem
            disabled={!derivedProfileEntityHasTimestampColumn(derivedModel)}
            icon
            on:select={() => quickStartMetrics(derivedModel)}
          >
            <svelte:fragment slot="icon"><Explore /></svelte:fragment>
            autogenerate dashboard
            <svelte:fragment slot="description">
              {#if !derivedProfileEntityHasTimestampColumn(derivedModel)}
                requires a timestamp column
              {/if}
            </svelte:fragment>
          </MenuItem>
          <Divider />
          <MenuItem icon on:select={() => openRenameModelModal(id, modelName)}>
            <svelte:fragment slot="icon">
              <EditIcon />
            </svelte:fragment>
            rename...</MenuItem
          >
          <MenuItem icon on:select={() => deleteModel(id)}>
            <svelte:fragment slot="icon">
              <Cancel />
            </svelte:fragment>
            delete</MenuItem
          >
        </svelte:fragment>
      </CollapsibleTableSummary>
    {/each}
  </div>
  {#if showRenameModelModal}
    <RenameEntityModal
      entityType={EntityType.Model}
      closeModal={() => (showRenameModelModal = false)}
      entityId={renameModelID}
      currentEntityName={renameModelName.replace(".sql", "")}
    />
  {/if}
{/if}
