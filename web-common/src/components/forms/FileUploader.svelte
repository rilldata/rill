<script lang="ts">
  import {
    AlertCircleIcon,
    AlertTriangleIcon,
    FileText,
    Upload,
    X,
  } from "lucide-svelte";
  import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size.ts";

  export let files: FileList | undefined;
  export let error: string | Record<string | number, string[]> | undefined =
    undefined;
  export let multiple: boolean = false;
  export let accept: string | undefined = undefined;
  export let hint: string | undefined = undefined;
  export let fileSizeLimit: number | undefined = undefined;
  export let fileSizeSoftLimit: boolean = false;
  export let fileSizeSoftLimitMessage: string | undefined = undefined;

  let fileInput: HTMLInputElement;
  let dragOver = false;

  $: errors = error ? (multiple ? error : { 0: error }) : [];
  $: errorMessages = Object.values({
    ...(errors as Record<string, any>),
  })
    .map((e, i) => {
      if (!e) return "";
      // Prepend file name if multiple=true
      if (multiple && files?.[i]) return `${files[i].name}: ${e}`;
      return e as string;
    })
    .filter(Boolean);

  $: fileSizeLimitMessages = fileSizeLimit
    ? Array.from(files ?? [])
        .filter((f) => f.size > fileSizeLimit!)
        .map(
          (f) =>
            `${f.name}: exceeds the maximum size of ${formatMemorySize(fileSizeLimit!)}`,
        )
    : [];
  // Show a single warning message if multiple=false, else prepend the file name.
  $: normalizedErrorMessages =
    !multiple && fileSizeSoftLimitMessage
      ? [fileSizeSoftLimitMessage]
      : fileSizeLimitMessages;

  // Only add the file size limit message to warnings. Errors should be tracked through superforms.
  $: warningMessages = fileSizeSoftLimit ? normalizedErrorMessages : [];

  $: selectedFile = files?.[0];
  $: hasError = errorMessages.length > 0;
  $: hasWarning = warningMessages.length > 0;

  function clearFiles() {
    if (fileInput) fileInput.value = "";
    files = undefined;
  }

  function handleDrop(event: DragEvent) {
    dragOver = false;
    if (!event.dataTransfer?.files?.length) return;
    files = event.dataTransfer.files;
  }
</script>

{#if selectedFile}
  <div class="file-wrapper">
    <div
      class="file-row"
      class:has-error={hasError}
      class:has-warning={hasWarning && !hasError}
    >
      <div
        class="file-icon"
        class:has-error={hasError}
        class:has-warning={hasWarning && !hasError}
      >
        <FileText
          size={24}
          class={hasError ? "text-destructive" : "text-icon-default"}
          strokeWidth={1.5}
        />
      </div>
      <div class="file-info">
        <span class="file-name">{selectedFile.name}</span>
        <span
          class="file-size"
          class:has-error={hasError}
          class:has-warning={hasWarning && !hasError}
        >
          {formatMemorySize(selectedFile.size)}
        </span>
      </div>
      <button
        type="button"
        class="clear-button"
        aria-label="Remove file"
        onclick={clearFiles}
      >
        <X size={14} />
      </button>
    </div>
    {#if hasError}
      {#each errorMessages as message (message)}
        <div class="error-message">
          <AlertCircleIcon size={12} class="shrink-0" />
          <span>{message}</span>
        </div>
      {/each}
    {/if}
    {#if hasWarning}
      {#each warningMessages as message (message)}
        <div class="warning-message">
          <AlertTriangleIcon size={12} class="shrink-0" />
          <span>{message}</span>
        </div>
      {/each}
    {/if}
  </div>
{:else}
  <button
    type="button"
    class="file-uploader"
    class:drag-over={dragOver}
    onclick={() => fileInput.click()}
    ondragenter={(e) => {
      e.preventDefault();
      e.stopPropagation();
      dragOver = true;
    }}
    ondragleave={(e) => {
      e.preventDefault();
      e.stopPropagation();
      dragOver = false;
    }}
    ondragover={(e) => {
      e.preventDefault();
      e.stopPropagation();
    }}
    ondrop={(e) => {
      e.preventDefault();
      handleDrop(e);
    }}
    aria-label="Upload file"
  >
    <div class="inner">
      <div class="icon-wrapper">
        <Upload size={24} class="text-fg-secondary" />
      </div>
      <div class="text-section">
        <p class="upload-text">
          <span class="upload-cta">Click to upload</span>
          <span class="upload-drag"> or drag and drop</span>
        </p>
        {#if hint}<p class="upload-hint">{hint}</p>{/if}
      </div>
    </div>
  </button>
{/if}

<input
  type="file"
  {accept}
  hidden
  {multiple}
  bind:this={fileInput}
  bind:files
/>

<style lang="postcss">
  /* Upload prompt */
  .file-uploader {
    @apply w-full bg-surface-muted border border-dashed rounded-sm shadow-sm p-0.5;
    @apply cursor-pointer transition-colors;
  }

  .file-uploader:hover,
  .file-uploader.drag-over {
    @apply bg-primary-50;
  }

  .inner {
    @apply flex flex-col items-center justify-center gap-3 py-10 w-full;
  }

  .icon-wrapper {
    @apply bg-surface-card rounded-lg size-10 flex items-center justify-center;
  }

  .text-section {
    @apply flex flex-col items-center gap-1;
  }

  .upload-text {
    @apply text-sm font-medium leading-5;
  }

  .upload-cta {
    @apply text-accent-primary-action;
  }

  .upload-drag {
    @apply text-fg-primary;
  }

  .upload-hint {
    @apply text-xs text-fg-tertiary;
  }

  /* File selected state (shared) */
  .file-wrapper {
    @apply flex flex-col gap-1.5 w-full;
  }

  .file-row {
    @apply flex items-center gap-3 w-full px-4 py-4;
    @apply bg-surface-card border rounded-sm shadow-sm;
  }

  .file-row.has-error {
    @apply bg-destructive-foreground border-destructive;
  }

  .file-icon {
    @apply bg-surface-muted rounded-lg size-10 flex items-center justify-center shrink-0;
  }

  .file-icon.has-error {
    @apply bg-destructive/15;
  }

  .file-info {
    @apply flex flex-col gap-0.5 flex-1 min-w-0;
  }

  .file-name {
    @apply text-sm font-medium text-fg-primary leading-5 truncate;
  }

  .file-size {
    @apply text-xs text-fg-tertiary leading-[18px];
  }

  .file-size.has-error {
    @apply text-destructive;
  }

  .clear-button {
    @apply shrink-0 text-fg-secondary flex items-center justify-center cursor-pointer;
  }

  .error-message {
    @apply flex items-center gap-1.5 text-xs text-destructive;
  }

  .file-row.has-warning {
    @apply bg-yellow-50 border-yellow-400;
  }

  .file-icon.has-warning {
    @apply bg-yellow-200;
  }

  .file-size.has-warning {
    @apply text-amber-600;
  }

  .warning-message {
    @apply flex items-center gap-1.5 text-xs text-amber-600;
  }
</style>
