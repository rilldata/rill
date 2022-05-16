<script lang="ts">
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";
  import { flip } from "svelte/animate";

  import type { ApplicationStore } from "$lib/application-state-stores/application-store";

  import Portal from "$lib/components/Portal.svelte";

  import ParquetIcon from "$lib/components/icons/Parquet.svelte";
  import ModelIcon from "$lib/components/icons/Code.svelte";
  import AddIcon from "$lib/components/icons/Add.svelte";
  import CollapsibleTableSummary from "$lib/components/column-profile/CollapsibleTableSummary.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";

  import { drag } from "$lib/drag";
  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "$lib/application-state-stores/table-stores";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "$lib/application-state-stores/model-stores";
  import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

  import {
    assetVisibilityTween,
    assetsVisible,
    layout,
  } from "$lib/application-state-stores/layout-store";
  import { onManualSourceUpload, onSourceDrop } from "$lib/util/file-upload";

  export let fileUploadElement: HTMLElement;

  const store = getContext("rill:app:store") as ApplicationStore;
  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let showTables = true;
</script>

<div
  class="pl-4 pb-3 pr-4 pt-5 grid justify-between"
  style="grid-template-columns: auto max-content;"
  on:drop|preventDefault|stopPropagation={onSourceDrop}
  on:drag|preventDefault|stopPropagation
  on:dragenter|preventDefault|stopPropagation
  on:dragover|preventDefault|stopPropagation
  on:dragleave|preventDefault|stopPropagation
>
  <CollapsibleSectionTitle tooltipText={"tables"} bind:active={showTables}>
    <h4 class="flex flex-row items-center gap-x-2">
      <ParquetIcon size="16px" /> Tables
    </h4>
  </CollapsibleSectionTitle>

  <ContextButton
    id={"create-table-button"}
    tooltipText="import csv or parquet file into a table"
    on:click={/**
     * Manual file upload
     * ------------------
     * clicks on the fileUploadElement above, which is hidden from the user.
     *
     */
    async () => {
      fileUploadElement.click();
    }}
  >
    <AddIcon />
  </ContextButton>
</div>
{#if showTables}
  <div
    class="pb-6"
    transition:slide|local={{ duration: 200 }}
    on:drop={onSourceDrop}
    on:drag|preventDefault|stopPropagation
    on:dragenter|preventDefault|stopPropagation
    on:dragover|preventDefault|stopPropagation
    on:dragleave|preventDefault|stopPropagation
  >
    {#if $persistentTableStore?.entities && $derivedTableStore?.entities}
      <!-- TODO: fix the object property access back to t.id from t["id"] once svelte fixes it -->
      {#each $persistentTableStore.entities as { path, tableName, id } (id)}
        {@const derivedTable = $derivedTableStore.entities.find(
          (t) => t["id"] === id
        )}
        <div animate:flip={{ duration: 200 }} out:slide={{ duration: 200 }}>
          <CollapsibleTableSummary
            indentLevel={1}
            icon={ParquetIcon}
            name={tableName}
            cardinality={derivedTable?.cardinality ?? 0}
            profile={derivedTable?.profile ?? []}
            head={derivedTable?.preview ?? []}
            {path}
            sizeInBytes={derivedTable?.sizeInBytes ?? 0}
            on:delete={() => {
              dataModelerService.dispatch("dropTable", [tableName]);
            }}
          />
        </div>
      {/each}
    {/if}
  </div>
{/if}
