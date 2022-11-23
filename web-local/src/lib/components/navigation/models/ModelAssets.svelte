<script lang="ts">
  import { page } from "$app/stores";
  import {
    useRuntimeServiceListFiles,
    useRuntimeServicePutFileAndMigrate,
  } from "@rilldata/web-common/runtime-client";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import { createModel } from "@rilldata/web-local/lib/components/navigation/models/createModel";
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";
  import { getName } from "../../../../common/utils/incrementName";
  import { runtimeStore } from "../../../application-state-stores/application-store";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "../../../application-state-stores/model-stores";
  import ColumnProfile from "../../column-profile/ColumnProfile.svelte";
  import ModelIcon from "../../icons/Model.svelte";
  import NavigationEntry from "../NavigationEntry.svelte";
  import NavigationHeader from "../NavigationHeader.svelte";
  import RenameAssetModal from "../RenameAssetModal.svelte";
  import ModelMenuItems from "./ModelMenuItems.svelte";
  import ModelTooltip from "./ModelTooltip.svelte";

  $: getFiles = useRuntimeServiceListFiles($runtimeStore.repoId);
  $: modelNames = $getFiles?.data?.paths
    ?.filter((path) => path.includes("models/"))
    .map((path) => path.replace("/models/", "").replace(".sql", ""));

  const createModelMutation = useRuntimeServicePutFileAndMigrate();

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let showModels = true;

  async function handleAddModel() {
    await createModel(
      $runtimeStore,
      getName("model", modelNames),
      $createModelMutation
    );
    // if the models are not visible in the assets list, show them.
    if (!showModels) {
      showModels = true;
    }
  }

  /** rename the model */
  let showRenameModelModal = false;
  let renameModelName = null;
  const openRenameModelModal = (modelName: string) => {
    showRenameModelModal = true;
    renameModelName = modelName;
  };
</script>

<NavigationHeader
  bind:show={showModels}
  tooltipText="create a new model"
  on:add={handleAddModel}
  contextButtonID={"create-model-button"}
>
  <ModelIcon size="16px" /> Models
</NavigationHeader>

{#if showModels}
  <div
    class="pb-6 justify-self-end"
    transition:slide={{ duration: LIST_SLIDE_DURATION }}
    id="assets-model-list"
  >
    {#if modelNames && $persistentModelStore?.entities && $derivedModelStore?.entities}
      {#each modelNames as modelName (modelName)}
        {@const persistentModel = $persistentModelStore.entities.find(
          (model) => model["name"] === modelName
        )}
        {@const derivedModel = $derivedModelStore.entities.find(
          (model) => model["id"] === persistentModel?.id
        )}
        <NavigationEntry
          name={modelName}
          href={`/model/${modelName}`}
          open={$page.url.pathname === `/model/${modelName}`}
        >
          <svelte:fragment slot="more">
            <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
              <ColumnProfile
                indentLevel={1}
                cardinality={derivedModel?.cardinality ?? 0}
                profile={derivedModel?.profile ?? []}
                head={derivedModel?.preview ?? []}
                entityId={persistentModel?.id}
              />
            </div>
          </svelte:fragment>

          <svelte:fragment slot="tooltip-content">
            <ModelTooltip {modelName} />
          </svelte:fragment>

          <svelte:fragment slot="menu-items" let:toggleMenu>
            <ModelMenuItems
              {modelName}
              {toggleMenu}
              on:rename-asset={() => {
                openRenameModelModal(modelName);
              }}
            />
          </svelte:fragment>
        </NavigationEntry>
      {/each}
    {/if}
  </div>
{/if}

{#if showRenameModelModal}
  <RenameAssetModal
    entityType={EntityType.Model}
    closeModal={() => (showRenameModelModal = false)}
    currentAssetName={renameModelName.replace(".sql", "")}
  />
{/if}
