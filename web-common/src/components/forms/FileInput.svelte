<script lang="ts">
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import Viz from "@rilldata/web-common/components/icons/Viz.svelte";

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

  // maintain a list of filenames to show in error messages.
  // since the final uploaded url set in `value` is usually not the same this is needed.
  let fileNames: string[] = value
    ? multiple
      ? (value as string[])
      : [value as string]
    : [];
  $: hasValue = values.length > 0 || Object.values(uploading).some((u) => u);

  function uploadFiles(files: FileList) {
    uploading = {};
    uploadErrors = {};
    if (multiple) {
      value = new Array(files.length).fill("");
      fileNames = new Array<string>(files.length).fill("");
    } else {
      value = "";
      fileNames = [""];
    }
    for (let i = 0; i < files.length; i++) {
      void uploadFileWrapper(files[i], i);
    }
  }

  async function uploadFileWrapper(file: File, i: number) {
    uploading[i] = true;
    try {
      fileNames[i] = file.name;
      const url = await uploadFile(file);
      if (multiple) {
        if (value === undefined) {
          value = [];
        }
        (value as string[])[i] = url;
      } else {
        value = url;
      }
    } catch (err) {
      uploadErrors[i] = err.message;
    }
    uploading[i] = false;
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
  $: errorMessages = Object.values({
    ...(errors as Record<string, any>),
    ...uploadErrors,
  })
    .map((e, i) => (fileNames[i] && e ? `${fileNames[i]}:${e}` : ""))
    .filter(Boolean);
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
    {#if hasValue}
      <div class="upload-preview">
        {#each fileNames as _, i (i)}
          {@const isUploading = !!uploading[i]}
          {@const hasError =
            (!!uploadErrors[i] || !!errors?.[i]) && !isUploading}
          {@const val = values[i]}
          {#if (val || isUploading) && !hasError}
            <div class="border border-neutral-400 p-1">
              {#if isUploading}
                <LoadingSpinner size="36px" />
              {:else}
                <img src={val} alt="upload" class="h-10 w-fit" />
              {/if}
            </div>
          {/if}
        {/each}
      </div>
    {:else}
      <Viz size="28px" class="text-gray-400 pointer-events-none" />
      <div class="container-flex-col pointer-events-none">
        <span class="upload-title"> Upload an image </span>
        {#if multiple}
          <span class="upload-description">
            Support for a single or bulk upload.
          </span>
        {/if}
      </div>
    {/if}
  </button>
  {#if errorMessages.length > 0}
    <div class="error">
      {#each errorMessages as errorMessage, i (i)}
        <div>{errorMessage}</div>
      {/each}
    </div>
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
    @apply flex flex-row gap-x-2.5 items-center justify-center py-5 min-h-10 min-w-80;
    @apply border border-neutral-400;
  }

  .upload-title {
    @apply text-sm font-medium text-left text-gray-500;
  }

  .upload-description {
    @apply text-xs font-normal text-gray-400;
  }

  .upload-preview {
    @apply flex flex-wrap w-full gap-x-1 items-center justify-center pointer-events-none px-4;
  }

  .error {
    @apply text-red-600 text-xs py-px mt-0.5;
  }
</style>
