<script lang="ts">
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import {
    useRuntimeServiceGetCatalogObject,
    useRuntimeServiceMigrateSingle,
  } from "web-common/src/runtime-client";
  import {
    dataModelerService,
    runtimeStore,
  } from "../../application-state-stores/application-store";
  import { updateMetricsDefsWrapperApi } from "../../redux-store/metrics-definition/metrics-definition-apis";
  import { store } from "../../redux-store/store-root";
  import Input from "../Input.svelte";
  import { Dialog } from "../modal/index";
  import notifications from "../notifications";
  export let entityId = null;

  export let closeModal;
  export let entityType: EntityType;
  export let currentAssetName: string;

  let error: string;
  let newAssetName = currentAssetName;
  let renameAction;

  let entityLabel: string;
  if (entityType === EntityType.Table) {
    renameAction = "updateTableName";
    entityLabel = "source";
  } else if (entityType === EntityType.Model) {
    renameAction = "updateModelName";
    entityLabel = "model";
  } else if (entityType === EntityType.MetricsDefinition) {
    renameAction = ""; // not used in submitHandler for MetricsDefinitions
    entityLabel = "dashboard";
  } else {
    throw new Error("entityType must be either 'Table' or 'Model'");
  }

  const resetVariablesAndCloseModal = () => {
    newAssetName = null;
    error = null;
    closeModal();
  };

  $: runtimeInstanceId = $runtimeStore.instanceId;
  $: getCatalog = useRuntimeServiceGetCatalogObject(
    runtimeInstanceId,
    currentAssetName
  );
  const renameSource = useRuntimeServiceMigrateSingle();

  const submitHandler = (assetId: string, newAssetName: string) => {
    if (!newAssetName || newAssetName.length === 0) {
      error = `${entityType.toLowerCase()} name cannot be empty`;
      return;
    }
    // TODO: add validation for asset name
    if (newAssetName === currentAssetName) {
      error = "new name must be different from current name";
      return;
    }
    // TODO: remove this branching logic once we have a unified backend for all entities
    switch (entityType) {
      case EntityType.Table: {
        const currentSql = $getCatalog.data.object.source.sql;
        const newSql = currentSql.replace(
          `CREATE SOURCE ${currentAssetName}`,
          `CREATE SOURCE ${newAssetName}`
        );
        $renameSource.mutate(
          {
            instanceId: runtimeInstanceId,
            data: {
              sql: newSql,
              renameFrom: currentAssetName,
            },
          },
          {
            onSuccess: () => {
              resetVariablesAndCloseModal();
              notifications.send({
                message: `renamed ${entityLabel} ${currentAssetName} to ${newAssetName}`,
              });
            },
            onError: (err) => {
              error = err.message;
            },
          }
        );
        break;
      }
      case EntityType.Model: {
        dataModelerService
          .dispatch(renameAction, [assetId, newAssetName])
          .then((response) => {
            if (response.status === 0) {
              notifications.send({ message: response.messages[0].message });
              resetVariablesAndCloseModal();
            } else if (response.status === 1) {
              error = response.messages[0].message;
            }
          });
        break;
      }
      case EntityType.MetricsDefinition: {
        store.dispatch(
          updateMetricsDefsWrapperApi({
            id: assetId,
            changes: { metricDefLabel: newAssetName },
          })
        );
        resetVariablesAndCloseModal();
        notifications.send({ message: `dashboard renamed to ${newAssetName}` });
        break;
      }
      default:
        throw new Error(
          "entityType must be either 'Table', 'Model', or 'MetricsDefinition'"
        );
    }
  };
</script>

<Dialog
  compact
  showCancel
  size="sm"
  disabled={newAssetName === null || currentAssetName === newAssetName}
  on:cancel={resetVariablesAndCloseModal}
  on:click-outside={resetVariablesAndCloseModal}
  on:primary-action={() => submitHandler(entityId, newAssetName)}
>
  <svelte:fragment slot="title">Rename</svelte:fragment>
  <form
    slot="body"
    autocomplete="off"
    on:submit|preventDefault={() => submitHandler(entityId, newAssetName)}
  >
    <Input
      claimFocusOnMount
      id="{entityLabel}-name"
      label="{entityLabel} name"
      bind:value={newAssetName}
      {error}
    />
  </form>
  <svelte:fragment slot="primary-action-body">Change Name</svelte:fragment>
</Dialog>
