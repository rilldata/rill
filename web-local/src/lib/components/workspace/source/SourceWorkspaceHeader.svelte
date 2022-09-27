<script lang="ts">
  import { dataModelerService } from "../../../application-state-stores/application-store";

  import type { PersistentTableStore } from "../../../application-state-stores/table-stores";

  import Source from "../../icons/Source.svelte";
  import { getContext } from "svelte";
  import WorkspaceHeader from "../WorkspaceHeader.svelte";

  export let id;

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;

  $: currentSource = $persistentTableStore?.entities?.find(
    (entity) => entity.id === id
  );

  const onChangeCallback = async (e) => {
    dataModelerService.dispatch("updateTableName", [id, e.target.value]);
  };

  $: titleInput = currentSource?.name;
</script>

<div
  class="grid gap-x-3 items-center pr-4"
  style:grid-template-columns="auto max-content"
>
  <WorkspaceHeader {...{ titleInput, onChangeCallback }} showStatus={false}>
    <svelte:fragment slot="icon">
      <Source />
    </svelte:fragment>
  </WorkspaceHeader>
</div>
