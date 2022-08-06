<script lang="ts">
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "$lib/application-state-stores/table-stores";
  import { SelectMenu } from "$lib/components/menu";
  import PreviewTable2 from "$lib/components/table/PreviewTable2.svelte";

  import { getContext } from "svelte";

  let rowVirtualizer;
  let columnVirtualizer;
  let headerVirtualizer;
  let container;

  let ROWS = 1000;
  let COLUMNS = 1000;

  const data = Array.from({ length: ROWS }).map(() => {
    return Array.from({ length: COLUMNS }).map(() => {
      return { value: Math.random() };
    });
  });

  /** can I get any data set? */
  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;

  $: sources =
    $persistentTableStore?.entities?.map((entity) => {
      return {
        key: entity.id,
        main: entity.tableName,
      };
    }) || [];
  $: selection = sources[0];

  let currentPersistent;
  let currentDerived;
  let columnOrder;

  let columnSizes;
  $: if (selection?.key && $persistentTableStore && $derivedTableStore) {
    // get row count

    currentPersistent = $persistentTableStore.entities.find(
      (entity) => entity.id === selection.key
    );
    currentDerived = $derivedTableStore.entities.find(
      (entity) => entity.id === currentPersistent.id
    );

    columnOrder = currentDerived.profile.reduce((obj, profile, i) => {
      obj[i] = profile;
      return obj;
    }, {});
  }
</script>

<div class="p-3">
  <SelectMenu
    options={sources}
    {selection}
    on:select={(event) => {
      selection = event.detail;
    }}
  />
</div>
{#if currentDerived}
  {#key selection?.key}
    <PreviewTable2
      data={currentDerived.preview}
      columns={currentDerived.profile}
    />
  {/key}
{/if}
