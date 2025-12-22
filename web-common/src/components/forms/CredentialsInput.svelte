<script lang="ts">
  import InputLabel from "./InputLabel.svelte";
  import Trash from "../icons/Trash.svelte";
  import IconButton from "../button/IconButton.svelte";

  export let value: string | string[] | undefined = undefined;
  export let error: string | Record<string | number, string[]> | undefined =
    undefined;
  export let multiple: boolean = false;
  export let accept: string | undefined = ".json";
  export let uploadFile: (file: File) => Promise<string>;
  export let id: string | undefined = undefined;
  export let label: string | undefined = undefined;
  export let hint: string | undefined = undefined;
  export let optional: boolean = false;
  export let hideContent: boolean = false;
  export let hidden: boolean = false;
  export let filename: string | string[] | undefined = undefined;

  let fileInput: HTMLInputElement;

  $: errors = error ? (multiple ? error : { 0: error }) : [];
  $: uploading = {};
  $: uploadErrors = {};

  // File validation function
  function validateFile(file: File): string | null {
    if (!file.name.toLowerCase().endsWith(".json")) {
      return "File must be a JSON file";
    }
    return null;
  }

  // maintain a list of filenames to show in error messages.
  // since the final uploaded url set in `value` is usually not the same this is needed.
  let fileNames: string[] = [];
  $: selectedFileName = filename
    ? multiple
      ? (filename as string[])[0]
      : (filename as string)
    : fileNames.length > 0
      ? fileNames[0]
      : null;

  function uploadFiles(files: FileList) {
    uploading = {};
    uploadErrors = {};
    if (multiple) {
      value = new Array(files.length).fill("");
      fileNames = new Array<string>(files.length).fill("");
      filename = new Array(files.length).fill("");
    } else {
      value = "";
      fileNames = [""];
      filename = "";
    }
    for (let i = 0; i < files.length; i++) {
      void uploadFileWrapper(files[i], i);
    }
  }

  async function uploadFileWrapper(file: File, i: number) {
    uploading[i] = true;

    // Validate file before upload
    const validationError = validateFile(file);
    if (validationError) {
      uploadErrors[i] = validationError;
      uploading[i] = false;
      return;
    }

    try {
      fileNames[i] = file.name;
      const result = await uploadFile(file);

      // Store the JSON string content and filename
      if (multiple) {
        if (value === undefined) {
          value = [];
        }
        if (filename === undefined) {
          filename = [];
        }
        (value as string[])[i] = result;
        (filename as string[])[i] = file.name;
      } else {
        value = result;
        filename = file.name;
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

  function clearFile() {
    if (multiple) {
      value = [];
      fileNames = [];
      filename = [];
    } else {
      value = "";
      fileNames = [];
      filename = "";
    }
    // Clear the file input
    if (fileInput) {
      fileInput.value = "";
    }
  }

  let dragOver = false;
  $: errorMessages = Object.values({
    ...(errors as Record<string, any>),
    ...uploadErrors,
  })
    .map((e, i) => (fileNames[i] && e ? `${fileNames[i]}:${e}` : ""))
    .filter(Boolean);
</script>

<div class="container">
  {#if !hidden}
    {#if label && id}
      <InputLabel {id} {label} {hint} {optional} />
    {/if}

    <div class="file-input-wrapper">
      <button
        type="button"
        class="file-input-button"
        on:click={() => fileInput.click()}
        on:dragenter|preventDefault|stopPropagation={() => (dragOver = true)}
        on:dragleave|preventDefault|stopPropagation={() => (dragOver = false)}
        on:dragover|preventDefault|stopPropagation
        on:drop|preventDefault={handleFileDrop}
        class:drag-over={dragOver}
      >
        <span class="choose-file-text">Choose file</span>
        <span class="file-status-text">
          {#if Object.values(uploading).some((u) => u)}
            Uploading...
          {:else if selectedFileName && !hideContent}
            {selectedFileName}
          {:else}
            No file chosen
          {/if}
        </span>
      </button>
      {#if selectedFileName && !hideContent && !Object.values(uploading).some((u) => u)}
        <div class="trash-icon-button">
          <IconButton size={24} ariaLabel="Remove file" on:click={clearFile}>
            <Trash size="16px" />
            <div slot="tooltip-content">Remove file</div>
          </IconButton>
        </div>
      {/if}
    </div>
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
  {/if}
</div>

<style lang="postcss">
  .container {
    @apply flex flex-col gap-y-2;
  }

  .file-input-wrapper {
    @apply w-full relative;
  }

  .file-input-button {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    gap: 6px;
    padding: 0.5rem 0.75rem;
    border: 1px solid #d1d5db;
    border-radius: 0.125rem;
    background-color: white;
    text-align: left;
    cursor: pointer;
    transition: border-color 0.2s;
  }

  .file-input-button:focus {
    outline: none;
    ring: 2px;
    ring-color: #3b82f6;
    border-color: #3b82f6;
  }

  .file-input-button.drag-over {
    @apply border-blue-500 bg-blue-50;
  }

  .choose-file-text {
    @apply font-medium text-gray-900 text-sm;
  }

  .file-status-text {
    @apply text-gray-600 text-sm;
  }

  .trash-icon-button {
    position: absolute;
    right: 4px;
    top: 50%;
    transform: translateY(-50%);
  }

  .trash-icon-button :global(button) {
    color: #6b7280;
    transition: all 0.2s;
  }

  .trash-icon-button :global(button:hover) {
    color: #dc2626;
  }

  .error {
    @apply text-red-600 text-xs py-px mt-0.5;
  }
</style>
