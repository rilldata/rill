<script lang="ts">
  import { page } from "$app/stores";
  import ColumnProfile from "@rilldata/web-common/components/column-profile/ColumnProfile.svelte";
  import RenameAssetModal from "@rilldata/web-common/features/entity-management/RenameAssetModal.svelte";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import { useSourceFileNames } from "@rilldata/web-common/features/sources/selectors";
  import { slide } from "svelte/transition";
  import { flip } from "svelte/animate";
  import { LIST_SLIDE_DURATION as duration } from "../../../layout/config";
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

<NavigationHeader bind:show={showModels} toggleText="models">
  Models
</NavigationHeader>

{#if showModels}
  <ol
    class="pb-3 justify-self-end"
    transition:slide|global={{ duration }}
    id="assets-model-list"
  >
    {#each modelNames as modelName (modelName)}
      <li
        animate:flip={{ duration }}
        out:slide|global={{ duration }}
        aria-label={modelName}
      >
        <NavigationEntry
          expandable
          name={modelName}
          href={`/model/${modelName}`}
          open={$page.url.pathname === `/model/${modelName}`}
        >
          <div transition:slide={{ duration }} slot="more">
            <ColumnProfile indentLevel={1} objectName={modelName} />
          </div>

          <ModelTooltip slot="tooltip-content" {modelName} />

          <ModelMenuItems
            slot="menu-items"
            {modelName}
            on:rename-asset={() => {
              openRenameModelModal(modelName);
            }}
          />
        </NavigationEntry>
      </li>
    {/each}

    <AddAssetButton
      id="create-model-button"
      label="Add model"
      bold={hasSourceButNoModels}
      on:click={handleAddModel}
    />
  </ol>
{/if}

{#if showRenameModelModal && renameModelName !== null}
  <RenameAssetModal
    entityType={EntityType.Model}
    closeModal={() => (showRenameModelModal = false)}
    currentAssetName={renameModelName.replace(".sql", "")}
  />
{/if}
