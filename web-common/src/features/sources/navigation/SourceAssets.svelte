<script lang="ts">
  import { page } from "$app/stores";
  import ColumnProfile from "@rilldata/web-common/components/column-profile/ColumnProfile.svelte";
  import RenameAssetModal from "@rilldata/web-common/features/entity-management/RenameAssetModal.svelte";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import { useSourceFileNames } from "@rilldata/web-common/features/sources/selectors";
  import { flip } from "svelte/animate";
  import { slide } from "svelte/transition";
  import { appScreen } from "../../../layout/app-store";
  import { LIST_SLIDE_DURATION } from "../../../layout/config";
  import NavigationEntry from "../../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../../layout/navigation/NavigationHeader.svelte";
  import { behaviourEvent } from "../../../metrics/initMetrics";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "../../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
  import { runtime } from "../../../runtime-client/runtime-store";
  import AddAssetButton from "../../entity-management/AddAssetButton.svelte";
  import { EntityType } from "../../entity-management/types";
  import { createModelFromSource } from "../createModel";
  import SourceMenuItems from "./SourceMenuItems.svelte";
  import SourceTooltip from "./SourceTooltip.svelte";
  import { addSourceModal } from "../modal/add-source-visibility";

  $: sourceNames = useSourceFileNames($runtime.instanceId);
  $: modelNames = useModelFileNames($runtime.instanceId);

  let showTables = true;

  async function openShowAddSourceModal() {
    addSourceModal.open();

    await behaviourEvent?.fireSourceTriggerEvent(
      BehaviourEventAction.SourceAdd,
      BehaviourEventMedium.Button,
      $appScreen,
      MetricsEventSpace.LeftPanel,
    );
  }

  const queryHandler = async (tableName: string) => {
    await createModelFromSource(
      $runtime.instanceId,
      $modelNames?.data ?? [],
      tableName,
      tableName,
    );
    // TODO: fire telemetry
  };

  let showRenameTableModal = false;
  let renameTableName: null | string = null;
  const openRenameTableModal = (tableName: string) => {
    showRenameTableModal = true;
    renameTableName = tableName;
  };

  $: hasNoAssets = $sourceNames.data?.length === 0;
</script>

<NavigationHeader bind:show={showTables} toggleText="sources"
  >Sources</NavigationHeader
>

{#if showTables}
  <div class="pb-3" transition:slide={{ duration: LIST_SLIDE_DURATION }}>
    {#if $sourceNames?.data}
      {#each $sourceNames.data as sourceName (sourceName)}
        <div
          animate:flip={{ duration: 200 }}
          out:slide|global={{ duration: LIST_SLIDE_DURATION }}
        >
          <NavigationEntry
            name={sourceName}
            href={`/source/${sourceName}`}
            open={$page.url.pathname === `/source/${sourceName}`}
            immediatelyNavigate={false}
            on:command-click={() => queryHandler(sourceName)}
          >
            <svelte:fragment slot="more">
              <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
                <ColumnProfile indentLevel={1} objectName={sourceName} />
              </div>
            </svelte:fragment>

            <svelte:fragment slot="tooltip-content">
              <SourceTooltip {sourceName} connector="" />
            </svelte:fragment>

            <svelte:fragment slot="menu-items" let:toggleMenu>
              <SourceMenuItems
                {sourceName}
                {toggleMenu}
                on:rename-asset={() => {
                  openRenameTableModal(sourceName);
                }}
              />
            </svelte:fragment>
          </NavigationEntry>
        </div>
      {/each}
    {/if}
    <AddAssetButton
      id="add-table"
      label="Add source"
      bold={hasNoAssets}
      on:click={openShowAddSourceModal}
    />
  </div>
{/if}

{#if showRenameTableModal && renameTableName !== null}
  <RenameAssetModal
    entityType={EntityType.Table}
    closeModal={() => (showRenameTableModal = false)}
    currentAssetName={renameTableName}
  />
{/if}
