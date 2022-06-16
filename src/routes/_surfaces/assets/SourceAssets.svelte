<script lang="ts">
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";
  import { flip } from "svelte/animate";

  import ParquetIcon from "$lib/components/icons/Parquet.svelte";
  import AddIcon from "$lib/components/icons/Add.svelte";
  import CollapsibleTableSummary from "$lib/components/column-profile/CollapsibleTableSummary.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";

  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import type {
    DerivedSourceStore,
    PersistentSourceStore,
  } from "$lib/application-state-stores/source-stores";

  import { onSourceDrop, uploadFilesWithDialog } from "$lib/util/file-upload";

  const persistentSourceStore = getContext(
    "rill:app:persistent-source-store"
  ) as PersistentSourceStore;

  const derivedSourceStore = getContext(
    "rill:app:derived-source-store"
  ) as DerivedSourceStore;

  let showSources = true;
</script>

<div
  class="pl-4 pb-3 pr-4 pt-5 grid justify-between"
  style="grid-template-columns: auto max-content;"
  on:drop|preventDefault|stopPropagation={onSourceDrop}
  on:drag|preventDefault|stopPropagation
  on:dragenter|preventDefault|stopPropagation
  on:dragover|preventDefault|stopPropagation
  on:dragleave|preventDefault|stopPropagation
>
  <CollapsibleSectionTitle tooltipText={"sources"} bind:active={showSources}>
    <h4 class="flex flex-row items-center gap-x-2">
      <ParquetIcon size="16px" /> Sources
    </h4>
  </CollapsibleSectionTitle>

  <ContextButton
    id={"create-source-button"}
    tooltipText="import a csv or parquet file"
    on:click={uploadFilesWithDialog}
  >
    <AddIcon />
  </ContextButton>
</div>
{#if showSources}
  <div
    class="pb-6"
    transition:slide|local={{ duration: 200 }}
    on:drop|preventDefault|stopPropagation={onSourceDrop}
    on:drag|preventDefault|stopPropagation
    on:dragenter|preventDefault|stopPropagation
    on:dragover|preventDefault|stopPropagation
    on:dragleave|preventDefault|stopPropagation
  >
    {#if $persistentSourceStore?.entities && $derivedSourceStore?.entities}
      <!-- TODO: fix the object property access back to t.id from t["id"] once svelte fixes it -->
      {#each $persistentSourceStore.entities as { sourceName, id } (id)}
        {@const derivedSource = $derivedSourceStore.entities.find(
          (t) => t["id"] === id
        )}
        <div animate:flip={{ duration: 200 }} out:slide={{ duration: 200 }}>
          <CollapsibleTableSummary
            indentLevel={1}
            name={sourceName}
            cardinality={derivedSource?.cardinality ?? 0}
            profile={derivedSource?.profile ?? []}
            head={derivedSource?.preview ?? []}
            sizeInBytes={derivedSource?.sizeInBytes ?? 0}
            on:delete={() => {
              dataModelerService.dispatch("dropSource", [sourceName]);
            }}
          />
        </div>
      {/each}
    {/if}
  </div>
{/if}
