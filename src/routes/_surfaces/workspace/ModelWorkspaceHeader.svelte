<script lang="ts">
  import {
    ApplicationStore,
    dataModelerService,
  } from "$lib/application-state-stores/application-store";
  import { getContext } from "svelte";

  import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";
  import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  import type { PersistentModelStore } from "$lib/application-state-stores/model-stores";
  import WorkspaceHeader from "./WorkspaceHeader.svelte";

  const store = getContext("rill:app:store") as ApplicationStore;
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;

  function formatModelName(str) {
    let output = str.trim().replaceAll(" ", "_").replace(/\.sql/, "");
    return output;
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
