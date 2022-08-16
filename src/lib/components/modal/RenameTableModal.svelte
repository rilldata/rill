<script lang="ts">
  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import Input from "$lib/components/Input.svelte";
  import {
    Modal,
    ModalAction,
    ModalActions,
    ModalContent,
    ModalTitle,
  } from "$lib/components/modal";
  import notifications from "$lib/components/notifications/";

  export let openModal = false;
  export let closeModal: () => void;
  export let tableID = null;
  export let currentTableName = null;

  let newTableName = null;
  let error = null;

  const resetVariablesAndCloseModal = () => {
    newTableName = null;
    error = null;
    closeModal();
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
    dataModelerService
      .dispatch("updateTableName", [tableID, newTableName])
      .then((response) => {
        if (response.status === 0) {
          notifications.send({ message: response.messages[0].message });
          resetVariablesAndCloseModal();
        } else if (response.status === 1) {
          error = response.messages[0].message;
        }
      });
  };
</script>

<Modal open={openModal} onBackdropClick={() => resetVariablesAndCloseModal()}>
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
    <ModalAction on:click={() => resetVariablesAndCloseModal()}>
      cancel
    </ModalAction>
    <ModalAction primary on:click={() => submitHandler(tableID, newTableName)}>
      submit
    </ModalAction>
  </ModalActions>
</Modal>
