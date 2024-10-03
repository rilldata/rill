<script lang="ts">
  import { page } from "$app/stores";
  import AddDataModal from "@rilldata/web-common/features/sources/modal/AddDataModal.svelte";
  import FileDrop from "@rilldata/web-common/features/sources/modal/FileDrop.svelte";
  import SourceImportedModal from "@rilldata/web-common/features/sources/modal/SourceImportedModal.svelte";
  import { sourceImportedPath } from "@rilldata/web-common/features/sources/sources-store";
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";

  let showDropOverlay = false;

  $: ({
    url: { pathname },
  } = $page);

  function isEventWithFiles(event: DragEvent) {
    let types = event?.dataTransfer?.types;
    return types && types.indexOf("Files") != -1;
  }
</script>

<main
  role="application"
  class="index-body relative size-full flex overflow-hidden"
  on:drag|preventDefault|stopPropagation
  on:drop|preventDefault|stopPropagation
  on:dragenter|preventDefault|stopPropagation
  on:dragleave|preventDefault|stopPropagation
  on:dragover|preventDefault|stopPropagation={(e) => {
    if (isEventWithFiles(e)) showDropOverlay = true;
  }}
>
  {#if pathname !== "/welcome"}
    <Navigation />
  {/if}
  <section class="size-full overflow-hidden">
    <slot />
  </section>
</main>

{#if showDropOverlay}
  <FileDrop bind:showDropOverlay />
{/if}

<AddDataModal />
<SourceImportedModal sourcePath={$sourceImportedPath} />
