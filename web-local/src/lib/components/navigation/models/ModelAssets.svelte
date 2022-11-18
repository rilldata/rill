<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    getRuntimeServiceListFilesQueryKey,
    useRuntimeServiceListFiles,
    useRuntimeServicePutFileAndMigrate,
  } from "@rilldata/web-common/runtime-client";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";
  import { getNextModelName } from "../../../../common/utils/incrementName";
  import { runtimeStore } from "../../../application-state-stores/application-store";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "../../../application-state-stores/model-stores";
  import { queryClient } from "../../../svelte-query/globalQueryClient";
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

  const createModel = useRuntimeServicePutFileAndMigrate();

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let showModels = true;

  async function handleAddModel() {
    const newModelName = getNextModelName(modelNames);
    $createModel.mutate(
      {
        data: {
          repoId: $runtimeStore.repoId,
          instanceId: $runtimeStore.instanceId,
          path: `models/${newModelName}.sql`,
          blob: ``,
          create: true,
          createOnly: true,
          strict: true,
        },
      },
      {
        onSuccess: () => {
          goto(`/model/${newModelName}`);
          queryClient.invalidateQueries(
            getRuntimeServiceListFilesQueryKey($runtimeStore.repoId)
          );
          // if the models are not visible in the assets list, show them.
          if (!showModels) {
            showModels = true;
          }
        },
      }
    );
  }

  /** rename the model */
  let showRenameModelModal = false;
  let renameModelID = null;
  let renameModelName = null;
  const openRenameModelModal = (modelID: string, modelName: string) => {
    showRenameModelModal = true;
    renameModelID = modelID;
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
                openRenameModelModal(persistentModel?.id, modelName);
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
    entityId={renameModelID}
    currentAssetName={renameModelName.replace(".sql", "")}
  />
{/if}
