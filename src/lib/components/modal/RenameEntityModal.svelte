<script lang="ts">
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

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

  export let entityType:
    | EntityType.Table
    | EntityType.Model
    | EntityType.MetricsDefinition;
  export let openModal = false;
  export let closeModal: () => void;
  export let entityId = null;
  export let currentEntityName = null;

  let newAssetName = null;
  let error = null;
  let renameAction;
  let entityLabel: string;
  if (entityType === EntityType.Table) {
    renameAction = "updateTableName";
    entityLabel = "source";
  } else if (entityType === EntityType.Model) {
    renameAction = "updateModelName";
    entityLabel = "model";
  } else if (entityType === EntityType.MetricsDefinition) {
    renameAction = "updateMetricsDefinitionName";
    entityLabel = "dashboard";
  } else {
    throw new Error("assetType must be either 'Table' or 'Model'");
  }

  const resetVariablesAndCloseModal = () => {
    newAssetName = null;
    error = null;
    closeModal();
  };

  const submitHandler = (assetID: string, newAssetName: string) => {
    if (!newAssetName || newAssetName.length === 0) {
      error = `${entityType.toLowerCase()} name cannot be empty`;
      return;
    }
    if (newAssetName === currentEntityName) {
      error = "new name must be different from current name";
      return;
    }
    if (entityType === EntityType.Table || entityType === EntityType.Model) {
      dataModelerService
        .dispatch(renameAction, [assetID, newAssetName])
        .then((response) => {
          if (response.status === 0) {
            notifications.send({ message: response.messages[0].message });
            resetVariablesAndCloseModal();
          } else if (response.status === 1) {
            error = response.messages[0].message;
          }
        });
    }
    if (entityType === EntityType.MetricsDefinition) {
      // TODO: need to use the rillDeveloperService to access the metricsDefinition actions.
    }
  };
</script>

<Modal open={openModal} onBackdropClick={() => resetVariablesAndCloseModal()}>
  <ModalTitle>
    rename <span class="text-gray-500 italic">{currentEntityName}</span>
  </ModalTitle>
  <ModalContent>
    <form
      on:submit|preventDefault={() => submitHandler(entityId, newAssetName)}
    >
      <Input
        id="{entityLabel}-name"
        label="{entityLabel} name"
        bind:value={newAssetName}
        {error}
      />
    </form>
  </ModalContent>
  <ModalActions>
    <ModalAction on:click={() => resetVariablesAndCloseModal()}>
      cancel
    </ModalAction>
    <ModalAction primary on:click={() => submitHandler(entityId, newAssetName)}>
      submit
    </ModalAction>
  </ModalActions>
</Modal>
