<script lang="ts">
  import Viz from "@rilldata/web-common/components/icons/Viz.svelte";
  import Attachment from "@rilldata/web-common/components/icons/Attachment.svelte";
  import { slide } from "svelte/transition";

  export let value: string | string[] | undefined = undefined;
  export let error: string | Record<string | number, string[]> | undefined =
    undefined;
  export let multiple: boolean = false;
  export let accept: string | undefined = undefined;
  export let onInput: (files: File[]) => void;

  $: values = value ? (multiple ? value : [value]) : [];
  $: errors = error ? (multiple ? error : { 0: error }) : [];
  let fileInput: HTMLInputElement;

  function handleInput() {
    onInput(Array.from(fileInput.files));
  }

  function handleFileDrop(event: DragEvent) {
    console.log(event);

    if (!event.dataTransfer?.files?.length) return;
    onInput(Array.from(event.dataTransfer.files));
  }
</script>

<div
  class="container grid"
  on:dragenter|preventDefault|stopPropagation
  on:dragleave|preventDefault|stopPropagation
  on:dragover|preventDefault|stopPropagation
  on:drop|preventDefault={handleFileDrop}
  role="presentation"
>
  <button class="upload-button" on:click={() => fileInput.click()}>
    <Viz size="28px" class="text-gray-400" />
    <div class="container-flex-col">
      <span class="upload-title">Upload an image</span>
      {#if multiple}
        <span class="upload-description">
          Support for a single or bulk upload.
        </span>
      {/if}
    </div>
  </button>
  {#if values}
    {#each values as val, i (i)}
      {@const hasError = !!errors?.[i]}
      <div class="container-flex-col">
        <div class="file-entry">
          <Attachment size="14px" />
          <span
            class:text-primary-500={!hasError}
            class:text-red-600={hasError}
          >
            {val}
          </span>
        </div>
        {#if hasError}
          <div in:slide={{ duration: 200 }} class="error">
            <div>{errors[i]}</div>
          </div>
        {/if}
      </div>
    {/each}
  {/if}
  <input
    type="file"
    {accept}
    hidden
    {multiple}
    bind:this={fileInput}
    on:input={handleInput}
  />
</div>

<style lang="postcss">
  .container {
    @apply flex flex-col gap-y-2;
  }

  .container-flex-col {
    @apply flex flex-col;
  }

  .upload-button {
    @apply flex flex-row gap-x-2.5 items-center justify-center py-5 min-h-10 w-80;
    @apply border border-neutral-400 bg-neutral-100;
  }

  .upload-title {
    @apply text-sm font-medium text-left text-gray-500;
  }

  .upload-description {
    @apply text-xs font-normal text-gray-400;
  }

  .file-entry {
    @apply flex flex-row items-center gap-x-1.5;
  }

  .error {
    @apply text-red-600 text-xs py-px mt-0.5;
  }
</style>
