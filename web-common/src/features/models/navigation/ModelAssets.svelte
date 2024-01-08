<script lang="ts">
  import {_} from "svelte-i18n";
  import { page } from "$app/stores";
  import ColumnProfile from "@rilldata/web-common/components/column-profile/ColumnProfile.svelte";
  import RenameAssetModal from "@rilldata/web-common/features/entity-management/RenameAssetModal.svelte";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import { useSourceFileNames } from "@rilldata/web-common/features/sources/selectors";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../../layout/config";
  import NavigationEntry from "../../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../../layout/navigation/NavigationHeader.svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import AddAssetButton from "../../entity-management/AddAssetButton.svelte";
  import { getName } from "../../entity-management/name-utils";
  import { createModel } from "../createModel";
  import ModelMenuItems from "./ModelMenuItems.svelte";
  import ModelTooltip from "./ModelTooltip.svelte";

  $: sourceNames = useSourceFileNames($runtime.instanceId);
  $: useModelNames = useModelFileNames($runtime.instanceId);

  $: modelNames = $useModelNames?.data ?? [];
  let showModels = true;

  async function handleAddModel() {
    await createModel($runtime.instanceId, getName("model", modelNames));
    // if the models are not visible in the assets list, show them.
    if (!showModels) {
      showModels = true;
    }
  }

  /** rename the model */
  let showRenameModelModal = false;
  let renameModelName: string | null = null;
  const openRenameModelModal = (modelName: string) => {
    showRenameModelModal = true;
    renameModelName = modelName;
  };

  $: hasSourceButNoModels =
    $sourceNames?.data?.length !== undefined &&
    $sourceNames?.data?.length > 0 &&
    modelNames.length === 0;
</script>

<NavigationHeader bind:show={showModels} toggleText="models"
  >{$_('models')}</NavigationHeader
>

{#if showModels}
  <div
    class="pb-3 justify-self-end"
    transition:slide|global={{ duration: LIST_SLIDE_DURATION }}
    id="assets-model-list"
  >
    {#each modelNames as modelName (modelName)}
      <NavigationEntry
        name={modelName}
        href={`/model/${modelName}`}
        open={$page.url.pathname === `/model/${modelName}`}
      >
        <svelte:fragment slot="more">
          <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
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

    <AddAssetButton
      id="create-model-button"
      label={$_('add-model')}
      bold={hasSourceButNoModels}
      on:click={handleAddModel}
    />
  </div>
{/if}

{#if showRenameModelModal && renameModelName !== null}
  <RenameAssetModal
    entityType={EntityType.Model}
    closeModal={() => (showRenameModelModal = false)}
    currentAssetName={renameModelName.replace(".sql", "")}
  />
{/if}
