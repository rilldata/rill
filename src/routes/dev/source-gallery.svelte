<script lang="ts">
  import { getContext } from "svelte";

  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "$lib/application-state-stores/table-stores";

  import ParquetIcon from "$lib/components/icons/Parquet.svelte";
  import CollapsibleTableSummary from "$lib/components/column-profile/CollapsibleTableSummary.svelte";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";

  let innerWidth = 0;
  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;
</script>

<svelte:window bind:innerWidth />

<div class="drawer-container bg-white p-2">
  <a
    href="/"
    class="flex flex-row gap-x-2 p-3 pt-1 pb-1 border border-black rounded w-max"
  >
    <span class="rotate-90 inline-block"><CaretDownIcon /></span> back
  </a>
  <!-- Drawer Handler -->
  <div class="assets">
    <div class="grid grid-cols-3">
      {#if $persistentTableStore && $derivedTableStore && $persistentTableStore?.entities?.length && $derivedTableStore?.entities?.length}
        {#each $persistentTableStore.entities as { tableName, path, id } (id)}
          {@const derivedTable = $derivedTableStore.entities.find(
            (t) => t["id"] === id
          )}
          <div
            class="source-list pl-3 pr-5 pt-3 pb-1"
            style:width="{innerWidth / 3 - 6}px"
          >
            <CollapsibleTableSummary
              icon={ParquetIcon}
              name={tableName}
              cardinality={derivedTable?.cardinality ?? 0}
              profile={derivedTable?.profile ?? []}
              head={derivedTable?.preview ?? []}
              {path}
              sizeInBytes={derivedTable?.sizeInBytes ?? 0}
            />
          </div>
        {/each}
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .drawer-container {
    height: calc(100vh - var(--header-height));
    overflow-y: auto;
  }

  /* .source-list {
    overflow-y: auto;
} */

  .assets {
    font-size: 12px;
    width: calc(100vw - 1rem);
  }
</style>
