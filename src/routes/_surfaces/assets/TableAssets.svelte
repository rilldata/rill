<script lang="ts">
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";
  import { flip } from "svelte/animate";

  import ParquetIcon from "$lib/components/icons/Parquet.svelte";
  import AddIcon from "$lib/components/icons/Add.svelte";
  import CollapsibleTableSummary from "$lib/components/column-profile/CollapsibleTableSummary.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";

  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "$lib/application-state-stores/table-stores";
  import type { PersistentModelStore } from "$lib/application-state-stores/model-stores";
  import notificationStore from "$lib/components/notifications/";

  import { uploadFilesWithDialog } from "$lib/util/file-upload";
  import Modal from "$lib/components/modal/Modal.svelte";
  import ModalAction from "$lib/components/modal/ModalAction.svelte";
  import ModalActions from "$lib/components/modal/ModalActions.svelte";
  import ModalContent from "$lib/components/modal/ModalContent.svelte";
  import ModalTitle from "$lib/components/modal/ModalTitle.svelte";
  import Input from "$lib/components/Input.svelte";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;

  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;

  let showTables = true;
  let showRenameTableDialog = false;
  let renameTableID = null;
  let renameTableCurrentName = null;
  let renameTableNewName = null;
  let formValidationError = null;
  const onSubmitRenameForm = (tableID: string, newName: string) => {
    if (!newName || newName.length === 0) {
      formValidationError = "source name cannot be empty";
      return;
    }
    if (newName === renameTableCurrentName) {
      formValidationError = "new name must be different from current name";
      return;
    }
    dataModelerService.dispatch("updateTableName", [tableID, newName]);
    showRenameTableDialog = false;
  };
</script>

<div
  class="pl-4 pb-3 pr-4 pt-5 grid justify-between"
  style="grid-template-columns: auto max-content;"
>
  <CollapsibleSectionTitle tooltipText={"sources"} bind:active={showTables}>
    <h4 class="flex flex-row items-center gap-x-2">
      <ParquetIcon size="16px" /> Sources
    </h4>
  </CollapsibleSectionTitle>

  <ContextButton
    id={"create-table-button"}
    tooltipText="import csv or parquet file as a source"
    on:click={uploadFilesWithDialog}
  >
    <AddIcon />
  </ContextButton>
</div>
{#if showTables}
  <div class="pb-6" transition:slide|local={{ duration: 200 }}>
    {#if $persistentTableStore?.entities && $derivedTableStore?.entities}
      <!-- TODO: fix the object property access back to t.id from t["id"] once svelte fixes it -->
      {#each $persistentTableStore.entities as { tableName, id } (id)}
        {@const derivedTable = $derivedTableStore.entities.find(
          (t) => t["id"] === id
        )}
        <div animate:flip={{ duration: 200 }} out:slide={{ duration: 200 }}>
          <CollapsibleTableSummary
            entityType={EntityType.Table}
            indentLevel={1}
            name={tableName}
            cardinality={derivedTable?.cardinality ?? 0}
            profile={derivedTable?.profile ?? []}
            head={derivedTable?.preview ?? []}
            sizeInBytes={derivedTable?.sizeInBytes ?? 0}
            on:rename={() => {
              showRenameTableDialog = true;
              renameTableCurrentName = tableName;
              renameTableID = id;
              renameTableNewName = null;
              formValidationError = null;
            }}
            on:query={async () => {
              // check existing models to avoid a name conflict
              const existingNames = $persistentModelStore?.entities
                .filter((model) => model.name.includes(`query_${tableName}`))
                .map((model) => model.tableName)
                .sort();
              const nextName =
                existingNames.length === 0
                  ? `query_${tableName}`
                  : `query_${tableName}_${existingNames.length + 1}`;

              const response = await dataModelerService.dispatch("addModel", [
                {
                  name: nextName,
                  query: `select * from ${tableName}`,
                },
              ]);

              // change the active asset to the new model
              await dataModelerService.dispatch("setActiveAsset", [
                EntityType.Model,
                response.id,
              ]);

              notificationStore.send({
                message: `queried ${tableName} in workspace`,
              });
            }}
            on:delete={() => {
              dataModelerService.dispatch("dropTable", [tableName]);
            }}
          />
        </div>
      {/each}
    {/if}
  </div>
  <Modal
    open={showRenameTableDialog}
    onBackdropClick={() => (showRenameTableDialog = false)}
  >
    <ModalTitle>
      rename <span class="text-gray-500 italic">{renameTableCurrentName}</span>
    </ModalTitle>
    <ModalContent>
      <form
        on:submit|preventDefault={() =>
          onSubmitRenameForm(renameTableID, renameTableNewName)}
      >
        <Input
          id="source-name"
          label="source name"
          bind:value={renameTableNewName}
          error={formValidationError}
        />
      </form>
    </ModalContent>
    <ModalActions>
      <ModalAction onClick={() => (showRenameTableDialog = false)}
        >cancel</ModalAction
      >
      <ModalAction
        primary
        onClick={() => onSubmitRenameForm(renameTableID, renameTableNewName)}
        >submit</ModalAction
      >
    </ModalActions>
  </Modal>
{/if}
