<script lang="ts">
  import {
    Modal,
    ModalAction,
    ModalActions,
    ModalContent,
    ModalTitle,
  } from "$lib/components/modal";
  import Input from "$lib/components/Input.svelte";
  import { dataModelerService } from "$lib/application-state-stores/application-store";

  export let openDialog = false;
  export let closeDialog: () => void;
  export let tableID = null;
  export let currentTableName = null;

  let newTableName = null;
  let error = null;

  const clearVariables = () => {
    newTableName = null;
    error = null;
  };

  const submitHandler = (tableID: string, newTableName: string) => {
    if (!newTableName || newTableName.length === 0) {
      error = "source name cannot be empty";
      return;
    }
    if (newTableName === currentTableName) {
      error = "new name must be different from current name";
      return;
    }
    dataModelerService.dispatch("updateTableName", [tableID, newTableName]);

    clearVariables();
    closeDialog();
  };
</script>

<Modal open={openDialog} onBackdropClick={() => (openDialog = false)}>
  <ModalTitle>
    rename <span class="text-gray-500 italic">{currentTableName}</span>
  </ModalTitle>
  <ModalContent>
    <form on:submit|preventDefault={() => submitHandler(tableID, newTableName)}>
      <Input
        id="source-name"
        label="source name"
        bind:value={newTableName}
        {error}
      />
    </form>
  </ModalContent>
  <ModalActions>
    <ModalAction
      onClick={() => {
        clearVariables();
        closeDialog();
      }}
    >
      cancel
    </ModalAction>
    <ModalAction primary onClick={() => submitHandler(tableID, newTableName)}>
      submit
    </ModalAction>
  </ModalActions>
</Modal>
