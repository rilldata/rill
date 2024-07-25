<script lang="ts">
  import { page } from "$app/stores";
  import AddDataModal from "@rilldata/web-common/features/sources/modal/AddDataModal.svelte";
  import FileDrop from "@rilldata/web-common/features/sources/modal/FileDrop.svelte";
  import PreparingImport from "@rilldata/web-common/features/sources/modal/PreparingImport.svelte";
  import SourceImportedModal from "@rilldata/web-common/features/sources/modal/SourceImportedModal.svelte";
  import { sourceImportedPath } from "@rilldata/web-common/features/sources/sources-store";
  import BlockingOverlayContainer from "@rilldata/web-common/layout/BlockingOverlayContainer.svelte";
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import {
    importOverlayVisible,
    overlay,
  } from "@rilldata/web-common/layout/overlay-store";

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
  class="index-body absolute w-screen h-screen flex overflow-hidden"
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

{#if $importOverlayVisible}
  <PreparingImport />
{:else if showDropOverlay}
  <FileDrop bind:showDropOverlay />
{:else if $overlay !== null}
  <BlockingOverlayContainer
    bg="linear-gradient(to right, rgba(0,0,0,.6), rgba(0,0,0,.8))"
  >
    <div slot="title" class="font-bold">
      {$overlay?.title}
    </div>
    <svelte:fragment slot="detail">
      {#if $overlay?.detail}
        <svelte:component
          this={$overlay.detail.component}
          {...$overlay.detail.props}
        />
      {/if}
    </svelte:fragment>
  </BlockingOverlayContainer>
{/if}

<AddDataModal />
<SourceImportedModal sourcePath={$sourceImportedPath} />
