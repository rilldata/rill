<script lang="ts">
  import { page } from "$app/stores";
  import { flip } from "svelte/animate";
  import ColumnProfile from "@rilldata/web-common/components/column-profile/ColumnProfile.svelte";
  import RenameAssetModal from "@rilldata/web-common/features/entity-management/RenameAssetModal.svelte";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import { useSourceFileNames } from "@rilldata/web-common/features/sources/selectors";
  import { slide } from "svelte/transition";
  import { appScreen } from "../../../layout/app-store";
  import { LIST_SLIDE_DURATION as duration } from "../../../layout/config";
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

  let showTables = true;
  let showRenameTableModal = false;
  let renameTableName: null | string = null;

  $: sourceNames = useSourceFileNames($runtime.instanceId);
  $: modelNames = useModelFileNames($runtime.instanceId);

  $: hasNoAssets = $sourceNames.data?.length === 0;

  async function openShowAddSourceModal() {
    addSourceModal.open();

    await behaviourEvent?.fireSourceTriggerEvent(
      BehaviourEventAction.SourceAdd,
      BehaviourEventMedium.Button,
      $appScreen.type,
      MetricsEventSpace.LeftPanel,
    );
  }

  async function queryHandler(tableName: string) {
    await createModelFromSource(
      $runtime.instanceId,
      $modelNames?.data ?? [],
      tableName,
      tableName,
    );
    // TODO: fire telemetry
  }

  function openRenameTableModal(tableName: string) {
    showRenameTableModal = true;
    renameTableName = tableName;
  }
</script>

<NavigationHeader bind:show={showTables} toggleText="sources">
  Sources
</NavigationHeader>

{#if showTables}
  <ol class="pb-3" transition:slide={{ duration }}>
    {#if $sourceNames?.data}
      {#each $sourceNames.data as sourceName (sourceName)}
        <li
          animate:flip={{ duration: 200 }}
          out:slide|global={{ duration }}
          aria-label={sourceName}
        >
          <NavigationEntry
            expandable
            name={sourceName}
            href={`/source/${sourceName}`}
            open={$page.url.pathname === `/source/${sourceName}`}
            on:command-click={() => queryHandler(sourceName)}
          >
            <div slot="more" transition:slide={{ duration }}>
              <ColumnProfile indentLevel={1} objectName={sourceName} />
            </div>

            <SourceTooltip slot="tooltip-content" {sourceName} connector="" />

            <SourceMenuItems
              slot="menu-items"
              {sourceName}
              on:rename-asset={() => {
                openRenameTableModal(sourceName);
              }}
            />
          </NavigationEntry>
        </li>
      {/each}
    {/if}
    <AddAssetButton
      id="add-table"
      label="Add source"
      bold={hasNoAssets}
      on:click={openShowAddSourceModal}
    />
  </ol>
{/if}

{#if showRenameTableModal && renameTableName !== null}
  <RenameAssetModal
    entityType={EntityType.Table}
    closeModal={() => (showRenameTableModal = false)}
    currentAssetName={renameTableName}
  />
{/if}
