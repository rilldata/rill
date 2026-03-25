<script lang="ts">
  import AddDataModal from "@rilldata/web-common/features/sources/modal/AddDataModal.svelte";
  import FileDrop from "@rilldata/web-common/features/sources/modal/FileDrop.svelte";
  import SourceImportedModal from "@rilldata/web-common/features/sources/modal/SourceImportedModal.svelte";
  import { sourceIngestionTracker } from "@rilldata/web-common/features/sources/sources-store";

  const ingestedPath = sourceIngestionTracker.ingestedPath;

  let showDropOverlay = false;

  function isEventWithFiles(event: DragEvent) {
    let types = event?.dataTransfer?.types;
    return types && types.indexOf("Files") != -1;
  }
</script>

<main
  role="application"
  class="index-body relative size-full flex flex-col overflow-hidden"
  ondrag={(e) => {
    e.preventDefault();
    e.stopPropagation();
  }}
  ondrop={(e) => {
    e.preventDefault();
    e.stopPropagation();
  }}
  ondragenter={(e) => {
    e.preventDefault();
    e.stopPropagation();
  }}
  ondragleave={(e) => {
    e.preventDefault();
    e.stopPropagation();
  }}
  ondragover={(e) => {
    e.preventDefault();
    e.stopPropagation();
    if (isEventWithFiles(e)) showDropOverlay = true;
  }}
>
  <slot />
</main>

{#if showDropOverlay}
  <FileDrop bind:showDropOverlay />
{/if}

<AddDataModal />
<SourceImportedModal sourcePath={$ingestedPath} />
