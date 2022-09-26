<script lang="ts">
  import {
    ApplicationStore,
    dataModelerService,
  } from "../../application-state-stores/application-store";
  import { getContext } from "svelte";

  import { ActionStatus } from "$web-local/common/data-modeler-service/response/ActionResponse";
  import type { PersistentModelEntity } from "$web-local/common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  import type { PersistentModelStore } from "../../application-state-stores/model-stores";
  import WorkspaceHeader from "./WorkspaceHeader.svelte";

  const store = getContext("rill:app:store") as ApplicationStore;
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;

  function formatModelName(str) {
    return str?.trim().replaceAll(" ", "_").replace(/\.sql/, "");
  }

  let currentModel: PersistentModelEntity;
  $: if ($store?.activeEntity && $persistentModelStore?.entities)
    currentModel = $persistentModelStore.entities.find(
      (q) => q.id === $store.activeEntity.id
    );

  let titleInput = currentModel?.name;
  $: titleInput = currentModel?.name;

  // FIXME: this should eventually be a redux action dispatcher `onChangeAction`
  const onChangeCallback = async (e) => {
    if (currentModel?.id) {
      const resp = await dataModelerService.dispatch("updateModelName", [
        currentModel?.id,
        formatModelName(e.target.value),
      ]);
      if (resp.status === ActionStatus.Failure) {
        e.target.value = currentModel.name;
      }
    }
  };
</script>

<WorkspaceHeader
  {...{ titleInput: formatModelName(titleInput), onChangeCallback }}
/>
