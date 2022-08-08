<script lang="ts">
  import type { DerivedTableEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
  import type { PersistentTableEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
  import type { ApplicationStore } from "$lib/application-state-stores/application-store";

  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "$lib/application-state-stores/table-stores";
  import PreviewTable from "$lib/components/table/PreviewTable.svelte";

  import { getContext } from "svelte";
  import SourceWorkspaceHeader from "./SourceWorkspaceHeader.svelte";

  const store = getContext("rill:app:store") as ApplicationStore;

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;

  let currentSource: PersistentTableEntity;
  $: currentSource =
    $store?.activeEntity && $persistentTableStore?.entities
      ? $persistentTableStore.entities.find(
          (q) => q.id === $store.activeEntity.id
        )
      : undefined;
  let currentDerivedSource: DerivedTableEntity;
  $: currentDerivedSource =
    $store?.activeEntity && $derivedTableStore?.entities
      ? $derivedTableStore.entities.find(
          (q) => q.id === $store.activeEntity?.id
        )
      : undefined;
</script>

<div
  class="grid pb-6"
  style:grid-template-rows="max-content auto"
  style:height="100vh"
>
  <SourceWorkspaceHeader id={currentSource?.id} />
  <div
    style:overflow="auto"
    style:height="100%"
    class="m-3 border border-gray-300 rounded"
  >
    {#if currentDerivedSource}
      {#key currentDerivedSource.id}
        <!-- <PreviewTable
          rows={currentDerivedSource?.preview.slice(0, 100)}
          columnNames={currentDerivedSource?.profile}
        /> -->
        <PreviewTable
          rows={currentDerivedSource?.preview}
          columnNames={currentDerivedSource?.profile}
        />
      {/key}
    {/if}
  </div>
</div>
