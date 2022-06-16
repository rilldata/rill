<script lang="ts">
  import { getContext } from "svelte";

  import type {
    DerivedSourceStore,
    PersistentSourceStore,
  } from "$lib/application-state-stores/source-stores";

  import ParquetIcon from "$lib/components/icons/Parquet.svelte";
  import CollapsibleTableSummary from "$lib/components/column-profile/CollapsibleTableSummary.svelte";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";

  let innerWidth = 0;
  const persistentSourceStore = getContext(
    "rill:app:persistent-source-store"
  ) as PersistentSourceStore;
  const derivedSourceStore = getContext(
    "rill:app:derived-source-store"
  ) as DerivedSourceStore;
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
      {#if $persistentSourceStore && $derivedSourceStore && $persistentSourceStore?.entities?.length && $derivedSourceStore?.entities?.length}
        {#each $persistentSourceStore.entities as { sourceName, id } (id)}
          {@const derivedSource = $derivedSourceStore.entities.find(
            (t) => t["id"] === id
          )}
          <div
            class="source-list pl-3 pr-5 pt-3 pb-1"
            style:width="{innerWidth / 3 - 6}px"
          >
            <CollapsibleTableSummary
              name={sourceName}
              cardinality={derivedSource?.cardinality ?? 0}
              profile={derivedSource?.profile ?? []}
              head={derivedSource?.preview ?? []}
              sizeInBytes={derivedSource?.sizeInBytes ?? 0}
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
