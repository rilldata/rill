<script lang="ts">
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import Viz from "@rilldata/web-common/components/icons/Viz.svelte";
  import Attachment from "@rilldata/web-common/components/icons/Attachment.svelte";
  import { extractFileName } from "@rilldata/web-common/features/entity-management/file-path-utils";
  import { slide } from "svelte/transition";

  export let value: string | string[] | undefined = undefined;
  export let error: string | Record<string | number, string[]> | undefined =
    undefined;
  export let multiple: boolean = false;
  export let accept: string | undefined = undefined;
  // Currently we upload either to runtime or cloud.
  // Implementation of it will be upto the caller of this component.
  export let uploadFile: (file: File) => Promise<string>;

  $: values = value ? (multiple ? (value as string[]) : [value as string]) : [];
  $: errors = error ? (multiple ? error : { 0: error }) : [];

  $: uploading = {};
  $: uploadErrors = {};
  let fileInput: HTMLInputElement;

  function uploadFiles(files: FileList) {
    uploading = {};
    uploadErrors = {};
    for (let i = 0; i < files.length; i++) {
      void uploadFileWrapper(files[i], i);
    }
  }

  async function uploadFileWrapper(file: File, i: number) {
    setFileUrl(file.name, i);
    uploading[i] = true;
    try {
      const url = await uploadFile(file);
      setFileUrl(url, i);
    } catch (err) {
      uploadErrors[i] = err.message;
    }
    uploading[i] = false;
  }

  function setFileUrl(fileUrl: string, i: number) {
    if (multiple) {
      if (value === undefined) {
        value = [];
      }
      (value as string[])[i] = fileUrl;
    } else {
      value = fileUrl;
    }
  }

  function handleInput() {
    if (!fileInput.files) return;
    uploadFiles(fileInput.files);
  }

  function handleFileDrop(event: DragEvent) {
    dragOver = false;

    if (!event.dataTransfer?.files?.length) return;
    uploadFiles(event.dataTransfer.files);
  }

  let dragOver = false;
</script>

<div class="container grid">
  <button
    class="upload-button"
    on:click={() => fileInput.click()}
    on:dragenter|preventDefault|stopPropagation={() => (dragOver = true)}
    on:dragleave|preventDefault|stopPropagation={() => (dragOver = false)}
    on:dragover|preventDefault|stopPropagation
    on:drop|preventDefault={handleFileDrop}
    class:bg-neutral-100={!dragOver}
    class:bg-primary-100={dragOver}
  >
    <Viz size="28px" class="text-gray-400 pointer-events-none" />
    <div class="container-flex-col pointer-events-none">
      <span class="upload-title"> Upload an image </span>
      {#if multiple}
        <span class="upload-description">
          Support for a single or bulk upload.
        </span>
      {/if}
    </div>
  </button>
  {#if values}
    {#each values as val, i (i)}
      {@const isUploading = !!uploading[i]}
      {@const hasError = (!!uploadErrors[i] || !!errors?.[i]) && !isUploading}
      <div class="container-flex-col">
        <div class="file-entry">
          {#if isUploading}
            <LoadingSpinner size="14px" />
          {:else}
            <Attachment size="14px" />
          {/if}
          <span
            class:text-primary-500={!hasError}
            class:text-red-600={hasError}
          >
            {extractFileName(val)}
          </span>
        </div>
        {#if hasError}
          <div in:slide={{ duration: 200 }} class="error">
            <div>{uploadErrors[i] ?? errors[i]}</div>
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
    @apply border border-neutral-400;
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
