<script lang="ts">
  import { page } from "$app/stores";
  import ColumnProfile from "@rilldata/web-common/components/column-profile/ColumnProfile.svelte";
  import RenameAssetModal from "@rilldata/web-common/features/entity-management/RenameAssetModal.svelte";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useRuntimeServicePutFileAndReconcile } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../../layout/config";
  import NavigationEntry from "../../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../../layout/navigation/NavigationHeader.svelte";
  import { getName } from "../../entity-management/name-utils";
  import { createModel } from "../createModel";
  import { useModelNames } from "../selectors";
  import ModelMenuItems from "./ModelMenuItems.svelte";
  import ModelTooltip from "./ModelTooltip.svelte";

  $: modelNames = useModelNames($runtimeStore.instanceId);

  const queryClient = useQueryClient();

  const createModelMutation = useRuntimeServicePutFileAndReconcile();

  let showModels = true;

  async function handleAddModel() {
    await createModel(
      queryClient,
      $runtimeStore.instanceId,
      getName("model", $modelNames.data),
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
  contextButtonID={"create-model-button"}
  on:add={handleAddModel}
  tooltipText="Create a new model"
>
  Models
</NavigationHeader>

{#if showModels}
  <div
    class="pb-3 justify-self-end"
    transition:slide={{ duration: LIST_SLIDE_DURATION }}
    id="assets-model-list"
  >
    {#if $modelNames?.data}
      {#each $modelNames.data as modelName (modelName)}
        <NavigationEntry
          name={modelName}
          href={`/model/${modelName}`}
          open={$page.url.pathname === `/model/${modelName}`}
        >
          <svelte:fragment slot="more">
            <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
              <ColumnProfile indentLevel={1} objectName={modelName} />
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
